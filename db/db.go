package lxDb

import (
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

// InsertOneResult
type InsertOneResult struct {
	InsertedID interface{}
}

// InsertManyResult
type InsertManyResult struct {
	InsertedIDs []interface{}
}

// UpdateResult
type UpdateResult struct {
	MatchedCount  int64
	ModifiedCount int64
	UpsertedCount int64
	UpsertedID    interface{}
}

// DeleteResult
type DeleteResult struct {
	DeletedCount int64 `bson:"n"`
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
