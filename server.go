package main

import (
	//"log"

	geo "github.com/kellydunn/golang-geo"
	"github.com/lorenzotinfena/chat-and-meet/proto" // Update
	"github.com/sethvargo/go-password/password"
)

type waiter struct {
	gender   proto.MatchRequest_Gender
	age      uint32
	location *geo.Point

	target_gender          proto.MatchRequest_Gender
	target_min_age         uint32
	target_max_age         uint32
	target_max_distance_km uint32

	callback func(key string)
}

func (me *waiter) can_match(dude waiter) bool {
	return (me.target_gender == proto.MatchRequest_Unknown || me.target_gender == dude.gender) && uint32(me.location.GreatCircleDistance(dude.location)) < me.target_max_distance_km && me.target_min_age <= dude.age && dude.age <= me.target_max_age
}

type server struct {
	queue []waiter
	chans map[string]string               // each client is identified by his key, his map is used to identifies the chain i have to read
	chats map[string](chan proto.Message) // used to map a key to a chain on which it has to write
}

func newServer() *server {
	return &server{chans: make(map[string]string),
		chats: make(map[string]chan proto.Message)}
}

func (server *server) Match(matchRequest *proto.MatchRequest, serviceMatchServer proto.Service_MatchServer) error {
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

	me := waiter{gender: gender,
		age:                    age,
		location:               location,
		target_gender:          target_gender,
		target_min_age:         target_min_age,
		target_max_age:         target_max_age,
		target_max_distance_km: target_max_distance_km}
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

			waiter.callback(key2)
			serviceMatchServer.Send(&proto.MatchResponse{ChatKey: &key1})
		}
	}
	// no match found, so I have to append me to the queue
	return nil
}
func (*server) StartChat(proto.Service_StartChatServer) error {
	return nil
}
