package lxDb

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
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
	subIdName := "_id"

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
		case *SubIdName:
			subIdName = val.Name
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
		go func() {
			defer func() {
				done <- true
			}()

			// Convert for audit
			bm, err := ToBsonMap(doc)
			if err != nil {
				log.Printf("insert audit error:%v\n", err)
				chanErr <- err
				return
			}

			// Check id exists and not empty
			if _, ok := bm[subIdName]; !ok {
				bm[subIdName] = res.InsertedID
			}

			// Write to logger
			if err := repo.audit.LogEntry(Insert, authUser, bm); err != nil {
				log.Printf("insert audit error:%v\n", err)
				chanErr <- err
				return
			}
		}()
	}

	return res.InsertedID, nil
}

// InsertMany inserts the provided documents.
func (repo *mongoBaseRepo) InsertMany(docs []interface{}, args ...interface{}) (*InsertManyResult, error) {
	timeout := DefaultTimeout
	opts := &options.InsertManyOptions{}
	var authUser interface{}
	done := make(chan bool)
	chanErr := make(chan error)
	subIdName := "_id"

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.InsertManyOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		case chan bool:
			done = val
		case chan error:
			chanErr = val
		case *SubIdName:
			subIdName = val.Name
		}
	}

	// Return UpdateManyResult
	insertManyResult := new(InsertManyResult)

	// Audit
	if authUser != nil && repo.audit != nil {
		// InsertOne func for audit insert many
		insertOneFn := func(doc interface{}) (*mongo.InsertOneResult, error) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return repo.collection.InsertOne(ctx, doc)
		}

		// Array for audits
		var auditEntries bson.A

		// Insert docs and create log entries
		for _, doc := range docs {
			// InsertOne
			res, err := insertOneFn(&doc)
			if err != nil || res.InsertedID == nil {
				// Error, increment FailedCount
				insertManyResult.FailedCount++
			} else {
				// Add inserted id
				insertManyResult.InsertedIDs = append(insertManyResult.InsertedIDs, res.InsertedID)

				// Convert for audit
				bm, err := ToBsonMap(doc)
				if err != nil {
					return insertManyResult, err
				}

				// Check id exists and not empty
				if _, ok := bm[subIdName]; !ok {
					bm[subIdName] = res.InsertedID
				}

				// Audit only is inserted,
				auditEntries = append(auditEntries, bson.M{"action": Insert, "user": authUser, "data": bm})
			}
		}
		// Start audit async
		go func() {
			defer func() {
				done <- true
			}()

			// Write to logger
			if err := repo.audit.LogEntries(auditEntries); err != nil {
				log.Printf("insert many audit error:%v\n", err)
				chanErr <- err
				return
			}
		}()

		return insertManyResult, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.InsertMany(ctx, docs, opts)
	if err != nil {
		return insertManyResult, err
	}

	// Convert
	if res != nil {
		insertManyResult.InsertedIDs = res.InsertedIDs
	}

	return insertManyResult, nil
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
	opts := &options.UpdateOptions{}
	var authUser interface{}
	done := make(chan bool)
	chanErr := make(chan error)
	subIdName := "_id"

	// Check args
	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.UpdateOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		case chan bool:
			done = val
		case chan error:
			chanErr = val
		case *SubIdName:
			subIdName = val.Name
		}
	}

	// Return UpdateManyResult
	updateManyResult := new(UpdateManyResult)

	//options.Find().SetProjection(bson.D{{"_id", 1}})

	// Audit
	if authUser != nil && repo.audit != nil {
		// UpdateOne func for audit update many
		updOneFn := func(subFilter bson.D, afterUpdate *bson.D) error {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return repo.collection.FindOneAndUpdate(ctx, subFilter, update,
				options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(afterUpdate)
		}

		// Find all with filter for audit
		var allDocs []interface{}
		if err := repo.Find(filter, &allDocs); err != nil {
			return nil, err
		}

		// Array for audits
		var auditEntries bson.A

		// Update docs and log entries
		for _, val := range allDocs {
			subFilter := bson.D{{subIdName, val.(bson.D).Map()[subIdName]}}
			// UpdateOne
			var afterUpdate bson.D
			if err := updOneFn(subFilter, &afterUpdate); err != nil {
				// Error, add subId to failedCount and Ids
				updateManyResult.FailedCount++
				updateManyResult.FailedIDs = append(updateManyResult.FailedIDs, subFilter.Map()[subIdName])
			} else {
				updateManyResult.MatchedCount++
				// Is Modified use DeepEqual
				if !cmp.Equal(val, afterUpdate) {
					updateManyResult.ModifiedCount++
					// Audit only is modified
					auditEntries = append(auditEntries, bson.M{"action": Update, "user": authUser, "data": afterUpdate})
				}
			}
		}

		// Start audit async
		go func() {
			defer func() {
				done <- true
			}()

			// Write to logger
			if err := repo.audit.LogEntries(auditEntries); err != nil {
				log.Printf("update audit error:%v\n", err)
				chanErr <- err
				return
			}
		}()

		return updateManyResult, nil
	}

	// Context for update
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Without audit UpdateMany will be performed
	res, err := repo.collection.UpdateMany(ctx, filter, update, opts)

	// Convert to UpdateManyResult
	if res != nil {
		updateManyResult.MatchedCount = res.MatchedCount
		updateManyResult.ModifiedCount = res.ModifiedCount
		updateManyResult.UpsertedCount = res.UpsertedCount
		updateManyResult.UpsertedID = res.UpsertedID
	}

	return updateManyResult, err
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
func (repo *mongoBaseRepo) DeleteMany(filter interface{}, args ...interface{}) (*DeleteManyResult, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.DeleteOptions{}
	var authUser interface{}
	done := make(chan bool)
	chanErr := make(chan error)
	subIdName := "_id"

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.DeleteOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		case chan bool:
			done = val
		case chan error:
			chanErr = val
		case *SubIdName:
			subIdName = val.Name
		}
	}

	// Return UpdateManyResult
	deleteManyResult := new(DeleteManyResult)

	// Audit
	if authUser != nil && repo.audit != nil {
		// Find all (only id field) with filter for audit
		var allDocs []interface{}
		if err := repo.Find(filter, &allDocs, options.Find().SetProjection(bson.D{{"_id", 1}})); err != nil {
			return deleteManyResult, err
		}

		// DeleteMany
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		res, err := repo.collection.DeleteMany(ctx, filter, opts)
		if err != nil {
			return deleteManyResult, err
		}
		if res != nil {
			deleteManyResult.DeletedCount = res.DeletedCount
		}

		// Start audit async
		go func(allDocs []interface{}) {
			defer func() {
				done <- true
			}()

			// create audit entries
			var auditEntries bson.A
			for _, doc := range allDocs {
				// data save only sub id by deleted
				data := bson.M{subIdName: doc.(bson.D).Map()[subIdName]}
				auditEntries = append(auditEntries, bson.M{"action": Delete, "user": authUser, "data": data})
			}

			// Write to logger
			if err := repo.audit.LogEntries(auditEntries); err != nil {
				log.Printf("delete audit error:%v\n", err)
				chanErr <- err
				return
			}
		}(allDocs)

		return deleteManyResult, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.DeleteMany(ctx, filter, opts)
	if res != nil {
		deleteManyResult.DeletedCount = res.DeletedCount
	}

	return deleteManyResult, err
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
