package main

import (
	//"log"

	"time"

	geo "github.com/kellydunn/golang-geo"
	"github.com/lorenzotinfena/chat-and-meet/proto" // Update
	"github.com/sethvargo/go-password/password"
)

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
	chans map[string]string               // each client is identified by his key, his map is used to identifies the chain i have to read
	chats map[string](chan proto.Message) // used to map a key to a chain on which it has to write
}

func newServer() *server {
	server := server{chans: make(map[string]string),
		chats: make(map[string]chan proto.Message)}
	go func() {
		for {
			time.Sleep(2 * time.Second)
			now := time.Now()
			whatToRemove := make([]int, 0)
			for i, waiter := range server.queue {
				if now.Sub(waiter.InsertionTime).Seconds() > 10 {
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

func (server *server) Match(matchRequest *proto.MatchRequest, stream proto.Service_MatchServer) error {
	gender := matchRequest.GetMyInfo().GetGender()
	age := matchRequest.GetMyInfo().GetAge()
	location := geo.NewPoint(matchRequest.GetMyInfo().GetLatitude(), matchRequest.GetMyInfo().GetLongitude())

	target_gender := matchRequest.GetPreferences().GetGender()
	target_min_age := matchRequest.GetPreferences().GetMinAge()
	target_max_age := matchRequest.GetPreferences().GetMaxAge()
	target_max_distance_km := matchRequest.GetPreferences().GetKilometersRange()

	// check parameters
	if gender == proto.MatchRequest_Unknown || age < 18 || age > 100 || location.Lat() < -90 || location.Lat() > 90 || location.Lng() < -180 || location.Lat() > 180 || target_min_age < 18 || target_max_age > 100 || target_min_age > target_max_age {
		return nil
	}

	me := waiter{Gender: gender,
		Age:                    age,
		Location:               location,
		Target_gender:          target_gender,
		Target_min_age:         target_min_age,
		Target_max_age:         target_max_age,
		Target_max_distance_km: target_max_distance_km}
	// search for a match
	for _, waiter := range (*server).queue {
		if me.can_match(waiter) && waiter.can_match(me) {
			// it's a match!
			var key1, key2 string
			for {
				key1, err := password.Generate(20, 10, 10, false, false)
				if err == nil {
					return err
				}
				if _, exist := (*server).chans[key1]; !exist {
					break
				}
			}
			for {
				key2, err := password.Generate(20, 10, 10, false, false)
				if err == nil {
					return err
				}
				if _, exist := (*server).chans[key2]; !exist {
					break
				}
			}
			(*server).chans[key1] = key2
			(*server).chans[key2] = key1
			(*server).chats[key1] = make(chan proto.Message)
			(*server).chats[key2] = make(chan proto.Message)

			waiter.Callback(key2)
			stream.Send(&proto.MatchResponse{ChatKey: &key1})
		}
	}
	// no match found, so I have to append me to the queue
	me.Callback = func(key string) {
		stream.Send(&proto.MatchResponse{ChatKey: &key})
	}
	(*server).queue = append((*server).queue, me)
	return nil
}
func (server *server) StartChat(stream proto.Service_StartChatServer) error {
	return nil
}
