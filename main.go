package main

import (
	"log"
	"net"
	"github.com/lorenzotinfena/pepper-server/proto" // Update
	"google.golang.org/grpc"
)

func run_grpc() error {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	proto.RegisterServiceServer(s, newServer())
	return s.Serve(lis)
}
func main() {
	log.Println("Server has started!")
	defer log.Println("Server stopped!")
	if err := run_grpc(); err != nil {
		log.Fatal(err)
	}
}
