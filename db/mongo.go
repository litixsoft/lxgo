package lxDb

import (
	"context"
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
	Insert         = "insert"
	Update         = "update"
	Delete         = "delete"
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
	chanErr := make(chan error)

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.InsertOneOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		case chan bool:
			done = val
		case chan error:
			chanErr = val
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

			// Write to logger
			if err := repo.audit.LogEntry(Insert, authUser, doc); err != nil {
				log.Printf("insert audit error:%v\n", err)
				chanErr <- err
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
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.InsertManyOptions:
			opts = val
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
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.CountOptions:
			opts = val
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
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.EstimatedDocumentCountOptions:
			opts = val
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
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOptions:
			opts = val
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
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneOptions:
			opts = val
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Find and convert no documents error
	err := repo.collection.FindOne(ctx, filter, opts).Decode(result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return NewNotFoundError(err.Error())
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
	chanErr := make(chan error)

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneAndUpdateOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		case chan bool:
			done = val
		case chan error:
			chanErr = val
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Return UpdateResult
	var afterUpdate bson.D
	if err := repo.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&afterUpdate); err != nil {
		if err == mongo.ErrNoDocuments {
			return NewNotFoundError(err.Error())
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

			// Write to logger
			if err := repo.audit.LogEntry(Update, authUser, &afterUpdate); err != nil {
				log.Printf("update audit error:%v\n", err)
				chanErr <- err
				return
			}
		}()
	}

	return nil
}

// UpdateMany updates multiple documents in the collection.
func (repo *mongoBaseRepo) UpdateMany(filter interface{}, update interface{}, args ...interface{}) (*UpdateManyResult, error) {
	// Default values
	timeout := DefaultTimeout
	//opts := &options.UpdateOptions{}
	var authUser interface{}
	subIdName := "_id"

	//for i := 0; i < len(args); i++ {
	//	switch val := args[i].(type) {
	//	case time.Duration:
	//		//timeout = val
	//	case *options.UpdateOptions:
	//		//opts = val
	//	}
	//}

	// Return UpdateResult
	ret := new(UpdateManyResult)

	// UpdateOne func for audit update many
	updOneFn := func(filter bson.D) (*mongo.UpdateResult, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return repo.collection.UpdateOne(ctx, filter, update)
	}

	// Find all with filter for audit
	var allDocs []interface{}
	if err := repo.Find(filter, &allDocs); err != nil {
		return nil, err
	}

	var auditEntries bson.A
	for _, val := range allDocs {
		subFilter := bson.D{{subIdName, val.(bson.D).Map()[subIdName]}}
		res, err := updOneFn(subFilter)
		if err != nil {
			ret.FailedCount++
			ret.FailedIDs = append(ret.FailedIDs, val.(bson.D).Map()[subIdName])
		}
		if res != nil {
			ret.MatchedCount += res.MatchedCount
			ret.ModifiedCount += res.ModifiedCount
			ret.UpsertedCount += res.UpsertedCount
			if res.UpsertedID != nil {
				ret.UpsertedIDs = append(ret.UpsertedIDs, res.UpsertedID)
			}
		}

		//if err := repo.UpdateOne(subFilter, update); err != nil {
		//	ret.FailedCount++
		//	ret.FailedID = val.(bson.D).Map()["_id"]
		//}
		//ret.ModifiedCount++

		auditEntries = append(auditEntries, bson.M{"action": Update, "user": authUser, "data": val})
		ret.AuditCount++
	}

	//ctx, cancel := context.WithTimeout(context.Background(), timeout)
	//defer cancel()

	// Return UpdateResult
	//ret := new(UpdateResult)
	//res, err := repo.collection.UpdateMany(ctx, filter, update, opts)

	// Convert to UpdateResult
	//if res != nil {
	//	ret.MatchedCount = res.MatchedCount
	//	ret.ModifiedCount = res.ModifiedCount
	//	ret.UpsertedCount = res.UpsertedCount
	//	ret.UpsertedID = res.UpsertedID
	//}

	return ret, nil
}

// DeleteOne deletes a single document from the collection.
func (repo *mongoBaseRepo) DeleteOne(filter interface{}, args ...interface{}) error {
	// Default values
	timeout := DefaultTimeout
	opts := &options.FindOneAndDeleteOptions{}
	var authUser interface{}
	done := make(chan bool)
	chanErr := make(chan error)

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneAndDeleteOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		case chan bool:
			done = val
		case chan error:
			chanErr = val
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Find document before delete for audit
	var beforeDelete bson.D
	if err := repo.collection.FindOneAndDelete(ctx, filter, opts).Decode(&beforeDelete); err != nil {
		if err == mongo.ErrNoDocuments {
			return NewNotFoundError(err.Error())
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

			// Write to logger
			if err := repo.audit.LogEntry(Delete, authUser, &beforeDelete); err != nil {
				log.Printf("audit delete error: %v\n", err)
				chanErr <- err
				return
			}
		}()
	}

	return nil
}

// DeleteMany deletes multiple documents from the collection.
func (repo *mongoBaseRepo) DeleteMany(filter interface{}, args ...interface{}) (int64, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.DeleteOptions{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.DeleteOptions:
			opts = val
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

// GetCollection get instance of repo collection.
func (repo *mongoBaseRepo) GetCollection() interface{} {
	return repo.collection
}

// GetDb get instance of repo collection database.
func (repo *mongoBaseRepo) GetDb() interface{} {
	return repo.collection.Database()
}

// GetRepoName, get name of repo (database/collection)
func (repo *mongoBaseRepo) GetRepoName() string {
	return repo.collection.Database().Name() + "/" + repo.collection.Name()
}
