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
	audit      IBaseRepoAudit
}

// NewMongoBaseRepo, return base repo instance
func NewMongoBaseRepo(collection *mongo.Collection, baseRepoAudit ...IBaseRepoAudit) IBaseRepo {
	// Default audit is nil
	var audit IBaseRepoAudit

	// Optional audit in args
	if len(baseRepoAudit) > 0 {
		audit = baseRepoAudit[0]
	}

	return &mongoBaseRepo{
		collection: collection,
		audit:      audit,
	}
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

// InsertOne inserts a single document into the collection.
func (repo *mongoBaseRepo) InsertOne(doc interface{}, args ...interface{}) (interface{}, error) {
	timeout := DefaultTimeout
	opts := &options.InsertOneOptions{}
	var authUser interface{}
	done := make(chan bool)

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.InsertOneOptions:
			opts = args[i].(*options.InsertOneOptions)
		case *AuditAuth:
			authUser = args[i].(*AuditAuth).User
		case chan bool:
			done = args[i].(chan bool)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.InsertOne(ctx, doc, opts)
	if err != nil {
		return nil, err
	}

	if authUser != nil && repo.audit != nil {
		// Start audit async
		go func(id primitive.ObjectID) {
			defer func() {
				done <- true
			}()

			// Log entry
			logEntry := &AuditLogEntry{
				AuthUser:       authUser,
				DbName:         repo.collection.Database().Name(),
				CollectionName: repo.collection.Name(),
				Ident:          id.Hex(),
				Action:         "insert",
				Data:           doc,
			}

			// Write to logger
			if err := repo.audit.LogEntry(logEntry); err != nil {
				log.Printf("insert audit error:%v\n", err)
				return
			}
		}(res.InsertedID.(primitive.ObjectID))
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
func (repo *mongoBaseRepo) UpdateOne(filter interface{}, update interface{}, args ...interface{}) error {
	timeout := DefaultTimeout
	opts := &options.FindOneAndUpdateOptions{}
	opts.SetReturnDocument(options.After)
	var authUser interface{}
	done := make(chan bool)

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[i].(time.Duration)
		case *options.FindOneAndUpdateOptions:
			opts = args[i].(*options.FindOneAndUpdateOptions)
		case *AuditAuth:
			authUser = args[i].(*AuditAuth).User
		case chan bool:
			done = args[i].(chan bool)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Return UpdateResult
	var afterUpdate bson.D
	if err := repo.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&afterUpdate); err != nil {
		if err == mongo.ErrNoDocuments {
			return lxErrors.NewNotFoundError(err.Error())
		}
		return err
	}

	// Audit
	if authUser != nil && repo.audit != nil {
		// Start audit async
		go func() {
			defer func() {
				done <- true
			}()

			// Log entry
			logEntry := &AuditLogEntry{
				AuthUser:       authUser,
				DbName:         repo.collection.Database().Name(),
				CollectionName: repo.collection.Name(),
				Ident:          afterUpdate.Map()["_id"].(primitive.ObjectID).Hex(),
				Action:         "update",
				Data:           afterUpdate,
			}

			// Write to logger
			if err := repo.audit.LogEntry(logEntry); err != nil {
				log.Printf("update audit error:%v\n", err)
				return
			}
		}()
	}

	return nil
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
