package lxDb

import (
	"context"
	"errors"
	"github.com/google/go-cmp/cmp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
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
	locale     *string
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
		locale:     nil,
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

// CreateIndexes, creates multiple indexes in the collection.
// The names of the created indexes are returned.
func (repo *mongoBaseRepo) CreateIndexes(indexes interface{}, args ...interface{}) ([]string, error) {
	timeout := DefaultTimeout
	opts := &options.CreateIndexesOptions{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.CreateIndexesOptions:
			opts = val
		}
	}

	// Convert indexModels
	indexModels, ok := indexes.([]mongo.IndexModel)
	if !ok {
		return []string{}, ErrIndexConvert
	}

	// create indexes
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return repo.collection.Indexes().CreateMany(ctx, indexModels, opts)
}

// InsertOne inserts a single document into the collection.
func (repo *mongoBaseRepo) InsertOne(doc interface{}, args ...interface{}) (interface{}, error) {
	timeout := DefaultTimeout
	opts := &options.InsertOneOptions{}
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.InsertOneOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := repo.collection.InsertOne(ctx, doc, opts)
	if err != nil {
		return nil, err
	}

	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {

		// Convert doc to bson.M
		bm, err := ToBsonMap(doc)
		if err != nil {
			return nil, err
		}

		// Add _id to data
		if _, ok := bm["_id"]; !ok {
			bm["_id"] = res.InsertedID
		}

		// Send to audit
		repo.audit.Send(bson.M{
			"collection": repo.collection.Name(),
			"action":     Insert,
			"user":       authUser,
			"data":       bm,
		})
	}

	return res.InsertedID, nil
}

// InsertMany inserts the provided documents.
func (repo *mongoBaseRepo) InsertMany(docs []interface{}, args ...interface{}) (*InsertManyResult, error) {
	timeout := DefaultTimeout
	opts := &options.InsertManyOptions{}
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.InsertManyOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	// Return UpdateManyResult
	insertManyResult := new(InsertManyResult)

	// Audit with insert, when audit is active than insert one and audit
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// InsertOne func for audit insert many
		insertOneFn := func(doc interface{}) (*mongo.InsertOneResult, error) {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return repo.collection.InsertOne(ctx, doc)
		}

		// Array for audits
		var auditEntries []bson.M

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
				if _, ok := bm["_id"]; !ok {
					bm["_id"] = res.InsertedID
				}

				// Audit only is inserted,
				auditEntries = append(auditEntries, bson.M{
					"collection": repo.collection.Name(),
					"action":     Insert,
					"user":       authUser,
					"data":       bm})
			}
		}

		// Check audit entries
		if auditEntries == nil || len(auditEntries) == 0 {
			return insertManyResult, nil
		}

		// Send to audit
		repo.audit.Send(auditEntries)

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

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
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

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Find and convert no documents error
	err := repo.collection.FindOne(ctx, filter, opts).Decode(result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// FindOneAndDelete find a single document and deletes it, returning the
// original in result.
func (repo *mongoBaseRepo) FindOneAndDelete(filter interface{}, result interface{}, args ...interface{}) error {
	timeout := DefaultTimeout
	opts := &options.FindOneAndDeleteOptions{}
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneAndDeleteOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := repo.collection.FindOneAndDelete(ctx, filter, opts).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}

	// Audit
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// Convert result to bson.M
		bm, err := ToBsonMap(result)
		if err != nil {
			return err
		}

		// Send to audit
		repo.audit.Send(bson.M{
			"collection": repo.collection.Name(),
			"action":     Delete,
			"user":       authUser,
			"data":       bson.M{"_id": bm["_id"]},
		})
	}

	return nil
}

// FindOneAndReplace finds a single document and replaces it, returning either
// the original or the replaced document.
func (repo *mongoBaseRepo) FindOneAndReplace(filter, replacement, result interface{}, args ...interface{}) error {
	timeout := DefaultTimeout
	opts := options.FindOneAndReplace()
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneAndReplaceOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	// Audit only with options.After
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// Check and set options
		if opts.ReturnDocument == nil || *opts.ReturnDocument != options.After && *opts.ReturnDocument != options.Before {
			// Set default to after
			opts.SetReturnDocument(options.After)
		}

		// Check option and replace with audit
		switch *opts.ReturnDocument {
		case options.After:
			// Save doc before replace for compare
			var beforeReplace bson.M
			// Set FindOne options
			findOneOpts := options.FindOne()
			// When FindOneAndUpdateOptions.Sort is set then set FindOneOptions
			if opts.Sort != nil {
				findOneOpts.SetSort(opts.Sort)
			}
			if err := repo.FindOne(filter, &beforeReplace, findOneOpts); err != nil {
				return err
			}

			// FindOne and update
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if err := repo.collection.FindOneAndReplace(ctx, filter, replacement, opts).Decode(result); err != nil {
				return err
			}

			// Audit only is replaced
			// Create after replace map for compare
			afterReplace, err := ToBsonMap(result)
			if err != nil {
				return err
			}
			// Compare and audit
			if !cmp.Equal(beforeReplace, afterReplace) {
				// Send to audit
				repo.audit.Send(bson.M{
					"collection": repo.collection.Name(),
					"action":     Update,
					"user":       authUser,
					"data":       afterReplace,
				})
			}
		case options.Before:
			// FindOne and replace
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if err := repo.collection.FindOneAndReplace(ctx, filter, replacement, opts).Decode(result); err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					return ErrNotFound
				}
				return err
			}

			// Audit only is replaced
			// Create before replace map for compare
			beforeReplace, err := ToBsonMap(result)
			if err != nil {
				return err
			}

			// Save doc after replace for compare
			var afterReplace bson.M
			if err := repo.FindOne(bson.D{{"_id", beforeReplace["_id"]}}, &afterReplace); err != nil {
				return err
			}

			// Compare and audit
			if !cmp.Equal(beforeReplace, afterReplace) {
				// Send to audit
				repo.audit.Send(bson.M{
					"collection": repo.collection.Name(),
					"action":     Update,
					"user":       authUser,
					"data":       afterReplace,
				})
			}
		}

		return nil
	}

	// Without audit simple FindOneAndUpdate with given opts
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := repo.collection.FindOneAndReplace(ctx, filter, replacement, opts).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// FindOneAndUpdate finds a single document and updates it, returning either
// the the updated.
func (repo *mongoBaseRepo) FindOneAndUpdate(filter, update, result interface{}, args ...interface{}) error {
	timeout := DefaultTimeout
	opts := &options.FindOneAndUpdateOptions{}
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneAndUpdateOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	// Audit only with options.After
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// Check and set options
		if opts.ReturnDocument == nil || *opts.ReturnDocument != options.After && *opts.ReturnDocument != options.Before {
			// Set default to after
			opts.SetReturnDocument(options.After)
		}

		// Check option and update with audit
		switch *opts.ReturnDocument {
		case options.After:
			// Save doc before update for compare
			var beforeUpdate bson.M
			// Set FindOne options
			findOneOpts := options.FindOne()
			// When FindOneAndUpdateOptions.Sort is set then set FindOneOptions
			if opts.Sort != nil {
				findOneOpts.SetSort(opts.Sort)
			}
			if err := repo.FindOne(filter, &beforeUpdate, findOneOpts); err != nil {
				return err
			}

			// FindOne and update
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if err := repo.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(result); err != nil {
				return err
			}

			// Audit only is updated
			// Create after update map for compare
			afterUpdate, err := ToBsonMap(result)
			if err != nil {
				return err
			}
			// Compare and audit
			if !cmp.Equal(beforeUpdate, afterUpdate) {
				// Send to audit
				repo.audit.Send(bson.M{
					"collection": repo.collection.Name(),
					"action":     Update,
					"user":       authUser,
					"data":       afterUpdate,
				})
			}
		case options.Before:
			// FindOne and update
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			if err := repo.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(result); err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					return ErrNotFound
				}
				return err
			}

			// Audit only is updated
			// Create before update map for compare
			beforeUpdate, err := ToBsonMap(result)
			if err != nil {
				return err
			}

			// Save doc after update for compare
			var afterUpdate bson.M
			if err := repo.FindOne(bson.D{{"_id", beforeUpdate["_id"]}}, &afterUpdate); err != nil {
				return err
			}

			// Compare and audit
			if !cmp.Equal(beforeUpdate, afterUpdate) {
				// Send to audit
				repo.audit.Send(bson.M{
					"collection": repo.collection.Name(),
					"action":     Update,
					"user":       authUser,
					"data":       afterUpdate,
				})
			}
		}

		return nil
	}

	// Without audit simple FindOneAndUpdate with given opts
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := repo.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(result); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}

	return nil
}

// UpdateOne updates a single document in the collection.
func (repo *mongoBaseRepo) UpdateOne(filter interface{}, update interface{}, args ...interface{}) error {
	timeout := DefaultTimeout
	opts := &options.UpdateOptions{}
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.UpdateOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// When audit then save doc before update for compare
		var beforeUpdate bson.M
		if err := repo.FindOne(filter, &beforeUpdate); err != nil {
			return err
		}

		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Find and update doc save doc after updated
		foaOpts := &options.FindOneAndUpdateOptions{}
		foaOpts.SetReturnDocument(options.After)
		var afterUpdate bson.M
		if err := repo.collection.FindOneAndUpdate(ctx, filter, update, foaOpts).Decode(&afterUpdate); err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return ErrNotFound
			}
			return err
		}

		// Audit only is updated
		if !cmp.Equal(beforeUpdate, afterUpdate) {
			// Send to audit
			repo.audit.Send(bson.M{
				"collection": repo.collection.Name(),
				"action":     Update,
				"user":       authUser,
				"data":       afterUpdate,
			})
		}
		return nil
	}

	// Without audit can simple update
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Simple update doc
	res, err := repo.collection.UpdateOne(ctx, filter, update, opts)

	if res != nil && res.MatchedCount == 0 {
		//return NewNotFoundError()
		return ErrNotFound
	}

	return err
}

// UpdateMany updates multiple documents in the collection.
func (repo *mongoBaseRepo) UpdateMany(filter interface{}, update interface{}, args ...interface{}) (*UpdateManyResult, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.UpdateOptions{}
	var authUser interface{}

	// Check args
	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.UpdateOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	// Return UpdateManyResult
	updateManyResult := new(UpdateManyResult)

	// Audit
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// UpdateOne func for audit update many
		updOneFn := func(subFilter bson.D, afterUpdate *bson.M) error {
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			return repo.collection.FindOneAndUpdate(ctx, subFilter, update,
				options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(afterUpdate)
		}

		// Find all with filter for audit
		var allDocs []bson.M
		if err := repo.Find(filter, &allDocs); err != nil {
			return nil, err
		}

		// Array for audits
		var auditEntries []bson.M

		// Update docs and log entries
		for _, val := range allDocs {
			subFilter := bson.D{{"_id", val["_id"]}}
			// UpdateOne
			var afterUpdate bson.M
			if err := updOneFn(subFilter, &afterUpdate); err != nil {
				// Error, add subId to failedCount and Ids
				updateManyResult.FailedCount++
				updateManyResult.FailedIDs = append(updateManyResult.FailedIDs, subFilter.Map()["_id"])
			} else {
				updateManyResult.MatchedCount++
				// Is Modified use DeepEqual
				if !cmp.Equal(val, afterUpdate) {
					updateManyResult.ModifiedCount++
					// Audit only is modified
					auditEntries = append(auditEntries, bson.M{
						"collection": repo.collection.Name(),
						"action":     Update,
						"user":       authUser,
						"data":       afterUpdate})
				}
			}
		}

		// Check audit entries
		if auditEntries == nil || len(auditEntries) == 0 {
			return updateManyResult, nil
		}

		// Send to audit
		repo.audit.Send(auditEntries)

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

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.FindOneAndDeleteOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Find document before delete for audit
	var beforeDelete struct {
		ID primitive.ObjectID `bson:"_id"`
	}
	if err := repo.collection.FindOneAndDelete(ctx, filter, opts).Decode(&beforeDelete); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return ErrNotFound
		}
		return err
	}

	// Audit
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// Send to audit
		repo.audit.Send(bson.M{
			"collection": repo.collection.Name(),
			"action":     Delete,
			"user":       authUser,
			"data":       bson.M{"_id": beforeDelete.ID},
		})
	}

	return nil
}

// DeleteMany deletes multiple documents from the collection.
func (repo *mongoBaseRepo) DeleteMany(filter interface{}, args ...interface{}) (*DeleteManyResult, error) {
	// Default values
	timeout := DefaultTimeout
	opts := &options.DeleteOptions{}
	var authUser interface{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.DeleteOptions:
			opts = val
		case *AuditAuth:
			authUser = val.User
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	// Return UpdateManyResult
	deleteManyResult := new(DeleteManyResult)

	// Audit
	if authUser != nil && repo.audit != nil && repo.audit.IsActive() {
		// Find all (only id field) with filter for audit
		var allDocs []bson.M
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
		if allDocs != nil && len(allDocs) > 0 {
			// create audit entries
			var auditEntries []bson.M
			for _, doc := range allDocs {
				// data save only sub id by deleted
				data := bson.M{"_id": doc["_id"]}
				auditEntries = append(auditEntries, bson.M{
					"collection": repo.collection.Name(),
					"action":     Delete,
					"user":       authUser,
					"data":       data})
			}

			// Send to audit
			repo.audit.Send(auditEntries)
		}

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

// SetLocale, sets locale string for collation repository wide, empty string will be remove settings
func (repo *mongoBaseRepo) SetLocale(code string) {
	if code == "" {
		repo.locale = nil
	} else {
		repo.locale = &code
	}
}

// Aggregate, performs a aggregation with binding to result
func (repo *mongoBaseRepo) Aggregate(pipeline interface{}, result interface{}, args ...interface{}) error {
	// Default values
	timeout := DefaultTimeout
	opts := &options.AggregateOptions{}

	for i := 0; i < len(args); i++ {
		switch val := args[i].(type) {
		case time.Duration:
			timeout = val
		case *options.AggregateOptions:
			opts = val
		}
	}

	if repo.locale != nil && opts.Collation == nil {
		opts.SetCollation(&options.Collation{
			Locale: *repo.locale,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cur, err := repo.collection.Aggregate(ctx, pipeline, opts)
	if err != nil {
		return err
	}

	if result != nil {
		ctxCursor, cancelCursor := context.WithTimeout(context.Background(), timeout)
		defer cancelCursor()

		if err := cur.All(ctxCursor, result); err != nil {
			return err
		}
	}

	return nil
}
