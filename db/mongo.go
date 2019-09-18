package lxDb

import (
	"context"
	"github.com/globalsign/mgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

// GetMongoDbClient, return new mongo driver client
func GetMongoDbClient(uri string) (*mongo.Client, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return client, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return client, err
	}

	return client, nil
}

// deprecated, Will be removed in a later version
// Db struct for mongodb
type MongoDb struct {
	Conn       *mgo.Session
	Name       string
	Collection string
}

// NewMongoDbConn, get a new db connection
func GetMongoDbConnection(url string) *mgo.Session {
	// dial info
	dialInfo, err := mgo.ParseURL(url)
	if err != nil {
		log.Fatal(err)
	}

	// Connection timeout
	dialInfo.Timeout = 30 * time.Second

	// Create new connection
	conn, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}
	conn.SetMode(mgo.Primary, true)

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
