package lxDb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	DefaultTimeout = time.Second * 30
)

type IMongoDBBaseRepo interface {
	Find(collection string, filter interface{}, result interface{}, args ...interface{}) error
}

type mongoDbBaseRepo struct {
	db *mongo.Database
}

func NewMongoBaseRepo(db *mongo.Database) IMongoDBBaseRepo {
	return &mongoDbBaseRepo{db: db}
}

// Find
func (repo *mongoDbBaseRepo) Find(collection string, filter interface{}, result interface{}, args ...interface{}) error {
	// Default values
	timeout := DefaultTimeout
	opts := &options.FindOptions{}

	for i := 0; i < len(args); i++ {
		switch args[i].(type) {
		case time.Duration:
			timeout = args[0].(time.Duration)
		case *options.FindOptions:
			opts = args[0].(*options.FindOptions)
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
