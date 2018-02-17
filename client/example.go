package main

import (
	"log"

	rpcpb "github.com/shuoyang2016/mywish/rpc"
	context "golang.org/x/net/context"
	"google.golang.org/grpc"
)

func main() {
	// Set up a connection to the server.
	address := "192.168.29.108:8083"
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := rpcpb.NewMyWishServiceClient(conn)

	// Contact the server and print out its response.
	r, err := c.CheckOrCreateUser(context.Background(), &rpcpb.CheckOrCreateUserRequest{UserName: "123456", Password: "111111"})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r)
}
