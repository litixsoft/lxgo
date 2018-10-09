package lxDb

import (
	"github.com/globalsign/mgo"
	"log"
	"time"
)

// Db struct for mongodb
type MongoDb struct {
	Conn       *mgo.Session
	Name       string
	Collection string
}

// NewMongoDbConn, get a new db connection
func GetMongoDbConnection(dbHost, authDatabase, username, password string) *mgo.Session {
	// dial info
	dialInfo := &mgo.DialInfo{
		Addrs:   []string{dbHost},
		Timeout: 30 * time.Second,
	}

	// check params
	if authDatabase != "" && username != "" && password != "" {
		dialInfo.Database = authDatabase
		dialInfo.Username = username
		dialInfo.Password = password
	}

	// Create new connection
	conn, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMode(mgo.Monotonic, true)

	return conn
}

func NewMongoDb(connection *mgo.Session, dbName, collection string) *MongoDb {
	return &MongoDb{
		Conn:       connection,
		Name:       dbName,
		Collection: collection,
	}
}

// Setup create indexes for user collection.
func (db *MongoDb) Setup(indexes []mgo.Index) error {
	// Copy mongo session (thread safe) and close after function
	conn := db.Conn.Copy()
	defer conn.Close()

	// Ensure indexes
	col := conn.DB(db.Name).C(db.Collection)

	for _, i := range indexes {
		if err := col.EnsureIndex(i); err != nil {
			return err
		}
	}

	return nil
}
