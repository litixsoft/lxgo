package lxHelper

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// FindOptions
type FindOptions struct {
	Fields map[string]int `json:"fields"`
	Sort   map[string]int `json:"sort,omitempty"`
	Skip   int64          `json:"skip"`
	Limit  int64          `json:"limit"`
	Locale *string        `json:"locale,omitempty"`
}

// ToMongoFindOptions
func (fo *FindOptions) ToMongoFindOptions() *options.FindOptions {
	opts := options.Find()
	opts.SetSkip(fo.Skip)
	opts.SetSort(fo.Sort)

	if fo.Locale != nil {
		opts.SetCollation(&options.Collation{Locale: *fo.Locale})
	}

	if fo.Limit > 0 {
		opts.SetLimit(fo.Limit)
	}

	fields := make(bson.D, 0, len(fo.Fields))
	for key, value := range fo.Fields {
		fields = append(fields, bson.E{Key: key, Value: value})
	}
	opts.SetProjection(fields)
	return opts
}

// ToMongoFindOneOptions
func (fo *FindOptions) ToMongoFindOneOptions() *options.FindOneOptions {
	opts := options.FindOne()
	opts.SetSkip(fo.Skip)
	opts.SetSort(fo.Sort)

	if fo.Locale != nil {
		opts.SetCollation(&options.Collation{Locale: *fo.Locale})
	}

	fields := make(bson.D, 0, len(fo.Fields))
	for key, value := range fo.Fields {
		fields = append(fields, bson.E{Key: key, Value: value})
	}
	opts.SetProjection(fields)
	return opts
}

// ToMongoFindOneAndDeleteOptions
func (fo *FindOptions) ToMongoFindOneAndDeleteOptions() *options.FindOneAndDeleteOptions {
	opts := options.FindOneAndDelete()
	opts.SetSort(fo.Sort)

	if fo.Locale != nil {
		opts.SetCollation(&options.Collation{Locale: *fo.Locale})
	}

	fields := make(bson.D, 0, len(fo.Fields))
	for key, value := range fo.Fields {
		fields = append(fields, bson.E{Key: key, Value: value})
	}
	opts.SetProjection(fields)
	return opts
}

// ToMongoFindOneAndReplaceOptions
func (fo *FindOptions) ToMongoFindOneAndReplaceOptions() *options.FindOneAndReplaceOptions {
	opts := options.FindOneAndReplace()
	opts.SetSort(fo.Sort)

	if fo.Locale != nil {
		opts.SetCollation(&options.Collation{Locale: *fo.Locale})
	}

	fields := make(bson.D, 0, len(fo.Fields))
	for key, value := range fo.Fields {
		fields = append(fields, bson.E{Key: key, Value: value})
	}
	opts.SetProjection(fields)
	return opts
}

// ToMongoFindOneAndDeleteOptions
func (fo *FindOptions) ToMongoFindOneAndUpdateOptions() *options.FindOneAndUpdateOptions {
	opts := options.FindOneAndUpdate()
	opts.SetSort(fo.Sort)

	if fo.Locale != nil {
		opts.SetCollation(&options.Collation{Locale: *fo.Locale})
	}

	fields := make(bson.D, 0, len(fo.Fields))
	for key, value := range fo.Fields {
		fields = append(fields, bson.E{Key: key, Value: value})
	}
	opts.SetProjection(fields)
	return opts
}

// ToMongoCountOptions
func (fo *FindOptions) ToMongoCountOptions() *options.CountOptions {
	opts := options.Count()
	opts.SetSkip(fo.Skip)

	if fo.Locale != nil {
		opts.SetCollation(&options.Collation{Locale: *fo.Locale})
	}

	if fo.Limit > 0 {
		opts.SetLimit(fo.Limit)
	}
	return opts
}
