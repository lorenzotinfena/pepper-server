package main

import (
	"context"
	"log"
	"time"

	"github.com/lorenzotinfena/pepper-server/proto" // Update
	"google.golang.org/grpc"
)

func _main() {
	time.Sleep(15 * time.Second)
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
	stream, err := c.StartChat(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	stream.Send(&proto.Message{Text: res.GetChatKey()})
	go func() {
		for {
			time.Sleep(2 * time.Second)
			stream.Send(&proto.Message{Text: "Hei ciao sono lùlù"})
		}
	}()
	for {
		for {
			mes, err := stream.Recv()
			if err != nil {
				stream.Send(&proto.Message{Text: "Ho ricevuto: " + mes.Text})
			}
		}

	}
}
