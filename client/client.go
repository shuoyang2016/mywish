package client

import (
	"log"

	"github.com/shuoyang2016/mywish/rpc"
	"google.golang.org/grpc"
)

func NewClient(addr string) *rpc.MyWishServiceClient {
	if addr == "" {
		addr = "192.168.29.108:8083"
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := rpc.NewMyWishServiceClient(conn)
	return &c
}
