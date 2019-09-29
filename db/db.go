package lxDb

import (
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// IBaseRepo
type IBaseRepo interface {
	InsertOne(doc interface{}, args ...interface{}) (interface{}, error)
	InsertMany(docs []interface{}, args ...interface{}) ([]interface{}, error)
	CountDocuments(filter interface{}, args ...interface{}) (int64, error)
	EstimatedDocumentCount(args ...interface{}) (int64, error)
	Find(filter interface{}, result interface{}, args ...interface{}) error
	FindOne(filter interface{}, result interface{}, args ...interface{}) error
	UpdateOne(filter interface{}, update interface{}, args ...interface{}) error
	UpdateMany(filter interface{}, update interface{}, args ...interface{}) (*UpdateResult, error)
	DeleteOne(filter interface{}, args ...interface{}) (int64, error)
	DeleteMany(filter interface{}, args ...interface{}) (int64, error)
	GetCollection() interface{}
	GetDb() interface{}
	GetRepoName() string
}

type IBaseRepoAudit interface {
	LogEntry(action string, user, data interface{}, timeout ...time.Duration) error
}

// AuthAudit, auth user for audit
type AuditAuth struct {
	User interface{}
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

// ToBsonMap, convert interface to bson.M
func ToBsonMap(v interface{}) (doc bson.M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}
