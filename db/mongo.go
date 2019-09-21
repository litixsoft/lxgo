package lxDb

import (
	"context"
	"github.com/globalsign/mgo"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

const (
	DefaultTimeout = time.Second * 30
)

type IMongoDBBaseRepo interface {
	Find(collection string, filter interface{}, result interface{}, args ...interface{}) error
	FindOne(collection string, filter interface{}, result interface{}, args ...interface{}) error
}

type mongoDbBaseRepo struct {
	db *mongo.Database
}

// GetMongoDbClient, return new mongo driver client
func GetMongoDbClient(uri string, timeout ...time.Duration) (*mongo.Client, error) {
	// Timeout
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return client, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return client, err
	}

	// Ping MongoDB
	ctx, _ = context.WithTimeout(context.Background(), to)
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		return client, err
	}

	return client, nil
}

// NewMongoBaseRepo, return base repo instance
func NewMongoBaseRepo(db *mongo.Database) IMongoDBBaseRepo {
	return &mongoDbBaseRepo{db: db}
}

// Find, find all matched by filter
func (repo *mongoDbBaseRepo) Find(collection string, filter interface{}, result interface{}, args ...interface{}) error {
	// Default values
	timeout := DefaultTimeout
	opts := &options.FindOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.FindOptions:
			opts = args[i].(*options.FindOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cur, err := repo.db.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		return err
	}

	return cur.All(ctx, result)
}

// Find, find all matched by filter
func (repo *mongoDbBaseRepo) FindOne(collection string, filter interface{}, result interface{}, args ...interface{}) error {
	// Default values
	timeout := DefaultTimeout
	opts := &options.FindOneOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.FindOneOptions:
			opts = args[i].(*options.FindOneOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return repo.db.Collection(collection).FindOne(ctx, filter, opts).Decode(result)
}

/////////////////////////////////////////////////
// deprecated, Will be removed in a later version
/////////////////////////////////////////////////
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
