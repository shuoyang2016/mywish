package db

import (
	"errors"

	"github.com/globalsign/mgo"
)

type MongoConnection struct {
	BaseSession        *mgo.Session
	DB                 string
	PlayerSCollection  string
	ProductsCollection string
}

type Option struct {
	URL                string
	DB                 string
	PlayerSCollection  string
	ProductsCollection string
}

func (m *MongoConnection) Close() {
	if m.BaseSession != nil {
		m.BaseSession.Close()
	}
}

func StartMongoConnection(option *Option) (*MongoConnection, error) {
	if option.URL == "" {
		return nil, errors.New("The mongo db url cannot be empty.")
	}
	session, err := mgo.Dial(option.URL)
	if err != nil {
		return nil, err
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	mongoConn := MongoConnection{BaseSession: session, DB: option.DB, PlayerSCollection: option.PlayerSCollection,
		ProductsCollection: option.ProductsCollection,
	}
	return &mongoConn, nil
}
