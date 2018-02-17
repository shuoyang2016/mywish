package main

import (
	"flag"
	"net/http"

	"github.com/golang/glog"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/shuoyang2016/mywish/fe/server"
	rpc "github.com/shuoyang2016/mywish/rpc"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var _ server.Server
var _ glog.Level

var (
	echoEndpoint = flag.String("echo_endpoint", "localhost:8083", "endpoint of YourService")
)

func run() error {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err := rpc.RegisterMyWishServiceHandlerFromEndpoint(ctx, mux, *echoEndpoint, opts)
	if err != nil {
		return err
	}
	return http.ListenAndServe(":8082", mux)
}

func main() {
	flag.Parse()
	go server.StartServer(":8083")
	defer glog.Flush()
	if err := run(); err != nil {
		glog.Fatal(err)
	}
}
