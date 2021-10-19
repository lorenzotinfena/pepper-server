package main

import (
	"github.com/lorenzotinfena/chat-and-meet/proto" // Update
	"github.com/kellydunn/golang-geo"
)

type server struct {

}
func (*server) Match(matchRequest *proto.MatchRequest,serviceMatchServer proto.Service_MatchServer) error {
	//my_gender := matchRequest.MyInfo.Gender
	my_age := matchRequest.MyInfo.Age
	my_location := geo.NewPoint(matchRequest.MyInfo.Latitude, matchRequest.MyInfo.Longitude)

	//target_gender := matchRequest.Preferences.Gender
	min_age := matchRequest.Preferences.MinAge
	max_age := matchRequest.Preferences.MaxAge
	//max_distance_km := matchRequest.Preferences.KilometersRange

	// check parameters
	if my_age < 18 || my_age > 100 || my_location.Lat() < -90 || my_location.Lat() > 90 || my_location.Lng() < -180 || my_location.Lat() > 180 || min_age < 18 || max_age > 100 || min_age > max_age{
		return nil
	}
	return nil
	
}
func (*server) StartChat(proto.Service_StartChatServer) error {
	return nil
}