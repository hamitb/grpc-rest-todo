package grpc

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"

	v1 "github.com/hamitb/go-grpc-http-rest-microservice/pkg/api/v1"
	"google.golang.org/grpc"
)

// RunServer runs gRPC service to publish Todo service
func RunServer(ctx context.Context, v1API v1.TodoServiceServer, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	// register service
	server := grpc.NewServer()
	v1.RegisterTodoServiceServer(server, v1API)

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handling
			log.Println("shutting down gRPC server...")

			server.GracefulStop()

			<-ctx.Done()
		}
	}()

	// start gRPC server
	log.Println("starting gRPC server...")
	return server.Serve(lis)
}
