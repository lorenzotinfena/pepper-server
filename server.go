package main

import (
	"errors"
	"log"

	"context"
	"sync"
	"time"

	geo "github.com/kellydunn/golang-geo"
	"github.com/lorenzotinfena/pepper-server/proto" // Update
	"github.com/sethvargo/go-password/password"
)

const MAX_WAITING_TIME_SECONDS = 15

type waiter struct {
	Gender   proto.MatchRequest_Gender
	Age      uint32
	Location *geo.Point

	Target_gender          proto.MatchRequest_Gender
	Target_min_age         uint32
	Target_max_age         uint32
	Target_max_distance_km uint32

	Callback func(key string)
	id       uint64
}

func (me *waiter) can_match(dude waiter) bool {
	return (me.Target_gender == proto.MatchRequest_Unknown || me.Target_gender == dude.Gender) && uint32(me.Location.GreatCircleDistance(dude.Location)) <= me.Target_max_distance_km && me.Target_min_age <= dude.Age && dude.Age <= me.Target_max_age
}

type server struct {
	queue         []waiter
	chans         map[string]string        // each client is identified by his key, his map is used to identifies the chain i have to read
	chats         map[string](chan string) // used to map a key to a chain on which it has to write
	waitingToChat map[string](chan interface{})
	mutex         sync.Mutex
}

func newServer() *server {
	server := server{
		chans:         make(map[string]string),
		chats:         make(map[string]chan string),
		waitingToChat: make(map[string]chan interface{}),
		mutex:         sync.Mutex{}}
	return &server
}

var id uint64 = 0

func (server *server) Match(ctx context.Context, matchRequest *proto.MatchRequest) (*proto.MatchResponse, error) {
	log.Println("Match call")

	gender := matchRequest.GetMyInfo().GetGender()
	age := matchRequest.GetMyInfo().GetAge()
	location := geo.NewPoint(matchRequest.GetMyInfo().GetLatitude(), matchRequest.GetMyInfo().GetLongitude())

	target_gender := matchRequest.GetPreferences().GetGender()
	target_min_age := matchRequest.GetPreferences().GetMinAge()
	target_max_age := matchRequest.GetPreferences().GetMaxAge()
	target_max_distance_km := matchRequest.GetPreferences().GetKilometersRange()

	// check parameters
	if gender == proto.MatchRequest_Unknown || age < 18 || age > 100 || location.Lat() < -90 || location.Lat() > 90 || location.Lng() < -180 || location.Lat() > 180 || target_min_age < 18 || target_max_age > 100 || target_min_age > target_max_age {
		return nil, nil
	}

	me := waiter{Gender: gender,
		Age:                    age,
		Location:               location,
		Target_gender:          target_gender,
		Target_min_age:         target_min_age,
		Target_max_age:         target_max_age,
		Target_max_distance_km: target_max_distance_km}

	// search for a match
	server.mutex.Lock()
	for i, waiter := range server.queue {
		if me.can_match(waiter) && waiter.can_match(me) {
			// it's a match!
			log.Println("new match!")
			var key1, key2 string
			var err error
			for {
				key1, err = password.Generate(20, 10, 10, false, false)
				if err != nil {
					server.mutex.Unlock()
					return nil, err
				}
				if _, exist := server.chans[key1]; !exist {
					break
				}
			}
			for {
				key2, err = password.Generate(20, 10, 10, false, false)
				if err != nil {
					server.mutex.Unlock()
					return nil, err
				}
				if _, exist := server.chans[key2]; !exist && key2 != key1 {
					break
				}
			}
			server.chans[key1] = key2
			server.chans[key2] = key1
			server.chats[key1] = make(chan string)
			server.chats[key2] = make(chan string)

			server.queue = append(server.queue[:i], server.queue[i+1:]...)
			server.mutex.Unlock()
			checkClientwaitToStartChat := func(key string) {
				defer delete(server.waitingToChat, key)
				select {
				case <-server.waitingToChat[key]:
					return
				case <-time.After(3 * time.Second):
					server.chats[server.chans[key]] <- "EOFEOF"
					server.cleanChat(key)
				}
			}
			server.waitingToChat[key1] = make(chan interface{})
			server.waitingToChat[key2] = make(chan interface{})
			go checkClientwaitToStartChat(key1)
			go checkClientwaitToStartChat(key2)
			waiter.Callback(key2)
			return &proto.MatchResponse{ChatKey: &key1}, nil
		}
	}
	// no match found, so I have to append me to the queue
	c := make(chan string)
	me.Callback = func(key string) {
		c <- key
	}
	me.id = id
	id++
	server.queue = append(server.queue, me)
	server.mutex.Unlock()

	remove_me_from_queue := func(){
		server.mutex.Lock()
		var index int
		for i, el := range server.queue {
			if el.id == me.id {
				index = i
				break
			}
		}
		server.queue = append(server.queue[:index], server.queue[index+1:]...)
		server.mutex.Unlock()
	}
	select {
	case <-ctx.Done(): // handle client disconnection
		remove_me_from_queue()
		return nil, errors.New("client disconnected")
	case <-time.After(MAX_WAITING_TIME_SECONDS * time.Second):
		remove_me_from_queue()
		return &proto.MatchResponse{}, nil
	case key := <-c:
		return &proto.MatchResponse{ChatKey: &key}, nil
	}
}
func (server *server) cleanChat(key string) {
	server.mutex.Lock()
	delete(server.chans, key)
	delete(server.chats, key)
	server.mutex.Unlock()
}

func (server *server) StartChat(stream proto.Service_StartChatServer) error {
	log.Println("StartChat call")
	// check validity
	first, err := stream.Recv()
	if err != nil {
		return nil
	}
	to_write_key := first.GetText()
	to_read_key, exist := server.chans[to_write_key]
	if !exist {
		return nil
	}
	server.waitingToChat[to_write_key] <- true
	defer server.cleanChat(to_write_key)

	to_write_chan := server.chats[to_write_key]
	to_read_chan := server.chats[to_read_key]
	// start chat
	go func() {
		for {
			mes, err := stream.Recv()
			if err != nil || mes.GetText() == "EOFEOF" {
				to_write_chan <- "EOFEOF"
				return
			}
			to_write_chan <- mes.GetText()
		}
	}()
	for {
		text := <- to_read_chan
		if text == "EOFEOF" {
			log.Println("Chat ended")
			stream.Context().Err()
			return nil
		}
		stream.Send(&proto.Message{Text: text})
	}
}
