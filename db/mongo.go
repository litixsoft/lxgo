package lxDb

import (
	"context"
	"github.com/globalsign/mgo"
	lxErrors "github.com/litixsoft/lxgo/errors"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

const (
	DefaultTimeout = time.Second * 30
)

type mongoBaseRepo struct {
	collection *mongo.Collection
}

type AuditFn func(user interface{})

// GetMongoDbClient, return new mongo driver client
func GetMongoDbClient(uri string) (client *mongo.Client, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return
	}

	// Check connection
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return client, err
	}

	return client, nil
}

// NewMongoBaseRepo, return base repo instance
func NewMongoBaseRepo(collection *mongo.Collection) IBaseRepo {
	return &mongoBaseRepo{collection: collection}
}

// InsertOne inserts a single document into the collection.
func (repo *mongoBaseRepo) InsertOne(doc interface{}, args ...interface{}) (interface{}, error) {
	timeout := DefaultTimeout
	opts := &options.InsertOneOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.InsertOneOptions:
			opts = args[i].(*options.InsertOneOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.InsertOne(ctx, doc, opts)
	if err != nil {
		return nil, err
	}

	return res.InsertedID, nil
}

// InsertMany inserts the provided documents.
func (repo *mongoBaseRepo) InsertMany(docs []interface{}, args ...interface{}) ([]interface{}, error) {
	timeout := DefaultTimeout
	opts := &options.InsertManyOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.InsertManyOptions:
			opts = args[i].(*options.InsertManyOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.InsertMany(ctx, docs, opts)
	if err != nil {
		return nil, err
	}

	return res.InsertedIDs, nil
}

// CountDocuments gets the number of documents matching the filter.
// For a fast count of the total documents in a collection see EstimatedDocumentCount.
func (repo *mongoBaseRepo) CountDocuments(filter interface{}, args ...interface{}) (int64, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.CountOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.CountOptions:
			opts = args[i].(*options.CountOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return repo.collection.CountDocuments(ctx, filter, opts)
}

// EstimatedDocumentCount gets an estimate of the count of documents in a collection using collection metadata.
func (repo *mongoBaseRepo) EstimatedDocumentCount(args ...interface{}) (int64, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.EstimatedDocumentCountOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.EstimatedDocumentCountOptions:
			opts = args[i].(*options.EstimatedDocumentCountOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return repo.collection.EstimatedDocumentCount(ctx, opts)
}

// Find, find all matched by filter
func (repo *mongoBaseRepo) Find(filter interface{}, result interface{}, args ...interface{}) error {
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

	cur, err := repo.collection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}

	return cur.All(ctx, result)
}

// Find, find all matched by filter
func (repo *mongoBaseRepo) FindOne(filter interface{}, result interface{}, args ...interface{}) error {
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

	// Find and convert no documents error
	err := repo.collection.FindOne(ctx, filter, opts).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return lxErrors.NewNotFoundError(err.Error())
		}

		return err
	}

	return nil
}

// UpdateOne updates a single document in the collection.
func (repo *mongoBaseRepo) UpdateOne(filter interface{}, update interface{}, args ...interface{}) (*UpdateResult, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.UpdateOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.UpdateOptions:
			opts = args[i].(*options.UpdateOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Return UpdateResult
	ret := new(UpdateResult)
	res, err := repo.collection.UpdateOne(ctx, filter, update, opts)

	// Convert to UpdateResult
	if res != nil {
		ret.MatchedCount = res.MatchedCount
		ret.ModifiedCount = res.ModifiedCount
		ret.UpsertedCount = res.UpsertedCount
		ret.UpsertedID = res.UpsertedID
	}

	return ret, err
}

// UpdateMany updates multiple documents in the collection.
func (repo *mongoBaseRepo) UpdateMany(filter interface{}, update interface{}, args ...interface{}) (*UpdateResult, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.UpdateOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.UpdateOptions:
			opts = args[i].(*options.UpdateOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Return UpdateResult
	ret := new(UpdateResult)
	res, err := repo.collection.UpdateMany(ctx, filter, update, opts)

	// Convert to UpdateResult
	if res != nil {
		ret.MatchedCount = res.MatchedCount
		ret.ModifiedCount = res.ModifiedCount
		ret.UpsertedCount = res.UpsertedCount
		ret.UpsertedID = res.UpsertedID
	}

	return ret, err
}

// DeleteOne deletes a single document from the collection.
func (repo *mongoBaseRepo) DeleteOne(filter interface{}, args ...interface{}) (int64, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.DeleteOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.DeleteOptions:
			opts = args[i].(*options.DeleteOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Return DeletedCount
	count := int64(0)
	res, err := repo.collection.DeleteOne(ctx, filter, opts)
	if res != nil {
		count = res.DeletedCount
	}

	return count, err
}

// DeleteMany deletes multiple documents from the collection.
func (repo *mongoBaseRepo) DeleteMany(filter interface{}, args ...interface{}) (int64, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.DeleteOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.DeleteOptions:
			opts = args[i].(*options.DeleteOptions)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Return DeletedCount
	count := int64(0)
	res, err := repo.collection.DeleteMany(ctx, filter, opts)
	if res != nil {
		count = res.DeletedCount
	}

	return count, err
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
