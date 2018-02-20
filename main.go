package main

import (
	"flag"

	"github.com/golang/glog"
	"github.com/shuoyang2016/mywish/server"
	"github.com/shuoyang2016/mywish/server/config"
)

var _ glog.Level

func main() {
	cfg := config.NewConfig()
	flag.Parse()
	stop := server.StartServer(cfg)
	<-stop
}
