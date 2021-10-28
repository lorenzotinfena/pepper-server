package main

import (
	"log"

	"context"
	"sync"
	"time"

	geo "github.com/kellydunn/golang-geo"
	"github.com/lorenzotinfena/chat-and-meet/proto" // Update
	"github.com/sethvargo/go-password/password"
)

const MAX_WAITING_TIME_SECONDS = 1000000

type waiter struct {
	Gender   proto.MatchRequest_Gender
	Age      uint32
	Location *geo.Point

	Target_gender          proto.MatchRequest_Gender
	Target_min_age         uint32
	Target_max_age         uint32
	Target_max_distance_km uint32

	Callback      func(key string)
	InsertionTime time.Time
}

func (me *waiter) can_match(dude waiter) bool {
	return (me.Target_gender == proto.MatchRequest_Unknown || me.Target_gender == dude.Gender) && uint32(me.Location.GreatCircleDistance(dude.Location)) < me.Target_max_distance_km && me.Target_min_age <= dude.Age && dude.Age <= me.Target_max_age
}

type server struct {
	queue []waiter
	chans map[string]string        // each client is identified by his key, his map is used to identifies the chain i have to read
	chats map[string](chan string) // used to map a key to a chain on which it has to write
}

func newServer() *server {
	server := server{
		chans: make(map[string]string),
		chats: make(map[string]chan string)}
	go func() {
		for {
			time.Sleep(2 * time.Second)
			now := time.Now()
			whatToRemove := make([]int, 0)
			for i, waiter := range server.queue {
				if now.Sub(waiter.InsertionTime).Seconds() > MAX_WAITING_TIME_SECONDS {
					whatToRemove = append(whatToRemove, i)
				}
			}
			for _, i := range whatToRemove {
				server.queue = append(server.queue[:i], server.queue[i+1:]...)
			}
		}
	}()
	return &server
}

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
	mu := sync.Mutex{}
	for i, waiter := range server.queue {
		mu.Lock()
		if me.can_match(waiter) && waiter.can_match(me) {
			// it's a match!
			log.Println("new match!")
			var key1, key2 string
			var err error
			for {
				key1, err = password.Generate(20, 10, 10, false, false)
				if err != nil {
					return nil, err
				}
				if _, exist := server.chans[key1]; !exist {
					break
				}
			}
			for {
				key2, err = password.Generate(20, 10, 10, false, false)
				if err != nil {
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

			server.queue = append(server.queue[:i], server.queue[i+1:]...) //remove this element, will be better using a mutax
			mu.Unlock()
			waiter.Callback(key2)
			return &proto.MatchResponse{ChatKey: &key1}, nil
		}
		mu.Unlock()
	}
	// no match found, so I have to append me to the queue
	c := make(chan string)
	me.Callback = func(key string) {
		c <- key
	}
	me.InsertionTime = time.Now()
	server.queue = append(server.queue, me)
	key := <-c
	return &proto.MatchResponse{ChatKey: &key}, nil
}
func (server *server) cleanChat(key string) {
	delete(server.chans, key)
	delete(server.chats, key)
}

//TODO: ogni tanto controllare i chans e le chat non ancora partire, ed eliminarle, e magari passare la chiave nei metadati del context
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
	to_write_chan := server.chats[to_write_key]
	to_read_chan := server.chats[to_read_key]
	// start chat
	go func() {
		for {
			mes, err := stream.Recv()
			if err != nil {
				to_write_chan <- "EOF"
				server.cleanChat(to_write_key)
				break
			}
			to_write_chan <- mes.GetText()
		}
	}()
	for {
		for text := range to_read_chan {
			if text == "EOF" {
				to_write_chan <- "EOF"
				server.cleanChat(to_write_key)
				return nil
			}
			stream.Send(&proto.Message{Text: text})
		}
	}
}
