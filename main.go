package main

import (

	"golang.org/x/net/context"
	"net/http"
	"flag"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/shuoyang2016/mywish/fe/server"
	"github.com/golang/glog"
	"google.golang.org/grpc"
	rpc "github.com/shuoyang2016/mywish/rpc"
)

var _ server.Server
var _ glog.Level

var (
	echoEndpoint = flag.String("echo_endpoint", "localhost:9090", "endpoint of YourService")
)

/*
func main() {
	session, e := mgo.Dial(default_url);
	if e != nil {
		fmt.Println("Got error %v", e)
	}
	database := "iwish"
	collection := "products"
	c := session.DB(database).C(collection)
	p := product.Product{ID: 123, Name: "foo"}
	c.Insert(p)

	//p1 := db.Product{}
	num, err := c.Find(nil).Count()
	if err != nil {
		fmt.Println("The number of count is %v.", num)
	}
}
*/

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