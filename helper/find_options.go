package lxHelper

import "go.mongodb.org/mongo-driver/mongo/options"

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

// ToMongoFindOneAndDeleteOptions
func (fo *FindOptions) ToMongoFindOneAndDeleteOptions() *options.FindOneAndDeleteOptions {
	opts := options.FindOneAndDelete()
	opts.SetSort(fo.Sort)
	return opts
}

// ToMongoFindOneAndReplaceOptions
func (fo *FindOptions) ToMongoFindOneAndReplaceOptions() *options.FindOneAndReplaceOptions {
	opts := options.FindOneAndReplace()
	opts.SetSort(fo.Sort)
	return opts
}

// ToMongoFindOneAndDeleteOptions
func (fo *FindOptions) ToMongoFindOneAndUpdateOptions() *options.FindOneAndUpdateOptions {
	opts := options.FindOneAndUpdate()
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
