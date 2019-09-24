package lxDb

import (
	"context"
	lxErrors "github.com/litixsoft/lxgo/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
type AuthAudit struct {
	User interface{}
}

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
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.InsertOneOptions:
			opts = args[i].(*options.InsertOneOptions)
		case *AuthAudit:
			authUser = args[i].(*AuthAudit).User
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.InsertOne(ctx, doc, opts)
	if err != nil {
		return nil, err
	}

	if authUser != nil {
		go func(insertId interface{}, collection *mongo.Collection) {
			id, ok := insertId.(primitive.ObjectID)
			if !ok {
				log.Println("insert audit error, can't convert id")
				return
			}

			log.Println("authUser:", authUser)
			log.Println("db:", collection.Database().Name())
			log.Println("collection:", collection.Name())
			log.Println("action:", "insert")
			log.Println("id:", id)

			log.Println("Ich suche")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			var res bson.D
			err := collection.FindOne(ctx, bson.D{{"_id", id}}).Decode(&res)
			if err != nil {
				log.Printf("insert audit error:%v\n", err)
				return
			}
			log.Println("kein fehler")
			log.Println("data:", res)
		}(res.InsertedID, repo.collection)
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
