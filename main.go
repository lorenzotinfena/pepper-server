package main

import (
	"log"
	"net"

	"github.com/lorenzotinfena/pepper-server/proto" // Update
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func run_grpc() error {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return err
	}
	creds, _ := credentials.NewServerTLSFromFile("certificate.pem", "key.pem")
	s := grpc.NewServer(grpc.Creds(creds))
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
