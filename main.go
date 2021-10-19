/*package main

import (
	"fmt"
	//
	//"google.golang.org/grpc"
)

func main() {
	fmt.Println("dvxcvcxd")

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Println("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	RegisterServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Println("failed to serve: %v", err)
	}
}*/
package main

import (
	"context"
	"flag"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/lorenzotinfena/chat-and-meet/proto" // Update
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
)

var (
	// command-line options:
	// gRPC server endpoint
	grpcServerEndpoint = flag.String("grpc-server-endpoint", "localhost:9090", "gRPC server endpoint")
)

func run_grpc_gateway() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := proto.RegisterServiceHandlerFromEndpoint(ctx, mux, *grpcServerEndpoint, opts)
	if err != nil {
		return err
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	return http.ListenAndServe(":"+ port, mux)
}
func run_grpc() error {
	lis, err := net.Listen("tcp", "localhost:9090")
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	proto.RegisterServiceServer(s, &server{})
	return s.Serve(lis)
}
func main() {
	log.Println("Server has started!")
	defer log.Println("Server crashed!")
	go func() {
		if err := run_grpc(); err != nil {
			log.Fatal(err)
		}
	}()

	if err := run_grpc_gateway(); err != nil {
		log.Fatal(err)
	}
}
