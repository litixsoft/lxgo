package lxDb_test

import (
	lxDb "github.com/litixsoft/lxgo/db"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestFindOptions_ToMongoFindOptions(t *testing.T) {
	its := assert.New(t)

	expSort := map[string]int{"name": 1, "email": -1}
	expSkip := int64(2)
	expLimit := int64(5)

	// Create test options
	options := lxDb.FindOptions{
		Sort:  expSort,
		Skip:  expSkip,
		Limit: expLimit,
	}

	// Convert to mongo find options
	opts := options.ToMongoFindOptions()

	// Check types
	its.Equal("*options.FindOptions", reflect.TypeOf(opts).String())
	its.Equal("map[string]int", reflect.TypeOf(opts.Sort).String())
	its.Equal("*int64", reflect.TypeOf(opts.Skip).String())
	its.Equal("*int64", reflect.TypeOf(opts.Limit).String())

	// Check values
	its.Equal(expSort, opts.Sort)
	its.Equal(&expSkip, opts.Skip)
	its.Equal(&expLimit, opts.Limit)
}
