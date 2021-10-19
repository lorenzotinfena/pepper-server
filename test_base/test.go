package main

import (
	"log"
	
		"context"

		"github.com/lorenzotinfena/chat-and-meet/proto" // Update
		"google.golang.org/grpc")

func main() {
	conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure(), grpc.WithBlock())
	log.Println("connected")
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := proto.NewServiceClient(conn)

	// Contact the server and print out its response.

	//ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	//defer cancel()
	res, err := c.Match(context.Background(), &proto.MatchRequest{MyInfo: &proto.MatchRequest_MyInfo{Age: 20, Gender: proto.MatchRequest_Male, Longitude: 0, Latitude: 0}, Preferences: &proto.MatchRequest_Preferences{KilometersRange: 120, MinAge: 18, MaxAge: 22}})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Println("chat key: " + res.GetChatKey())
}