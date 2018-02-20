package config

import (
	"flag"

	"github.com/golang/glog"
)

var _ = glog.V

const (
	// Mongo DB consts
	kPlayersCollection  = "players"
	kProductsCollection = "products"
)

type Config struct {
	RestPort int
	GrpcPort int

	// Mongo DB configs
	MongoAddress       string
	PlayersCollection  string
	ProductsCollection string
	DBName             string

	// Auth DB configs
	SqlAddress string
}

func NewConfig() *Config {
	cfg := &Config{}
	flag.IntVar(&cfg.RestPort, "rest_port", 8082, "The port of REST API.")
	flag.IntVar(&cfg.GrpcPort, "grpc_port", 8083, "The port of GRPC.")
	flag.StringVar(&cfg.MongoAddress, "mongo_addr", "localhost:27017", "The address of mongo db.")
	flag.StringVar(&cfg.SqlAddress, "sql_address", "mywishtest:mywishtest@/mywishtest?parseTime=true",
		"The MySQL address for authentication module.")
	flag.StringVar(&cfg.DBName, "mongo_db_name", "mywish_mongo", "The database name of mongo DB.")
	cfg.PlayersCollection = kPlayersCollection
	cfg.ProductsCollection = kProductsCollection
	return cfg
}
