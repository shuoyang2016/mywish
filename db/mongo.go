package db

import "github.com/globalsign/mgo"

const (
	kLocalMongoURL = "localhost:27017"
)

type MongoConnection struct {
	BaseSession *mgo.Session
	DB          string
}

func (m *MongoConnection) Close() {
	if m.BaseSession != nil {
		m.BaseSession.Close()
	}
}

func StartMongoConnection(db string, url string) *MongoConnection {
	var dbUrl string
	if url == "" {
		dbUrl = kLocalMongoURL
	}
	session, err := mgo.Dial(dbUrl)
	if err != nil {
		panic(err)
	}
	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	mongoConn := MongoConnection{BaseSession: session, DB: db}
	return &mongoConn
}
