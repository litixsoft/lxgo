package lxDb

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IBaseRepo interface {
	InsertOne(collection string, doc interface{}, args ...interface{}) (interface{}, error)
	InsertMany(collection string, docs []interface{}, args ...interface{}) ([]interface{}, error)
	CountDocuments(collection string, filter interface{}, args ...interface{}) (int64, error)
	EstimatedDocumentCount(collection string, args ...interface{}) (int64, error)
	Find(collection string, filter interface{}, result interface{}, args ...interface{}) error
	FindOne(collection string, filter interface{}, result interface{}, args ...interface{}) error
	UpdateOne(collection string, filter interface{}, update interface{}, args ...interface{}) (*UpdateResult, error)
	UpdateMany(collection string, filter interface{}, update interface{}, args ...interface{}) (*UpdateResult, error)
	DeleteOne(collection string, filter interface{}, args ...interface{}) (int64, error)
	DeleteMany(collection string, filter interface{}, args ...interface{}) (int64, error)
}

// FindOptions
type FindOptions struct {
	Sort  map[string]int `json:"sort,omitempty"`
	Skip  int64          `json:"skip"`
	Limit int64          `json:"limit"`
}

// ToMongoFindOptions
func (fo *FindOptions) ToMongoFindOptions() *options.FindOptions {
	opts := options.Find()
	opts.SetSkip(fo.Skip)
	opts.SetLimit(fo.Limit)
	opts.SetSort(fo.Sort)
	return opts
}

// ToMongoFindOneOptions
func (fo *FindOptions) ToMongoFindOneOptions() *options.FindOneOptions {
	opts := options.FindOne()
	opts.SetSkip(fo.Skip)
	opts.SetSort(fo.Sort)
	return opts
}

// ToMongoCountOptions
func (fo *FindOptions) ToMongoCountOptions() *options.CountOptions {
	opts := options.Count()
	opts.SetSkip(fo.Skip)
	opts.SetLimit(fo.Limit)
	return opts
}

// UpdateResult
type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    interface{}
}

// ToBsonDoc, convert interface to bson.D
func ToBsonDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

/////////////////////////////////////////////////
// deprecated, Will be removed in a later version
/////////////////////////////////////////////////
// ChangeInfo holds details about the outcome of an update operation.
type ChangeInfo struct {
	Updated int // Number of documents updated
	Removed int // Number of documents removed
	Matched int // Number of documents matched but not necessarily changed
}

type Options struct {
	Sort  string `json:"sort,omitempty"`
	Skip  int    `json:"skip"`
	Limit int    `json:"limit"`
	Count bool   `json:"count"`
}
