package lxHelper_test

import (
	lxHelper "github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestNewRequestByQuery(t *testing.T) {
	its := assert.New(t)

	//var data lxHelper.RequestByQuery
	jsonStr := `{"opts":{"sort":{"name":1, "email": -1},"skip":5,"limit":10},"query":{"name":"Schulterglatze"}, "count":true}`
	data, err := lxHelper.NewRequestByQuery(jsonStr)
	its.NoError(err)

	// Check type
	expect := "*lxHelper.RequestByQuery"
	its.Equal(expect, reflect.TypeOf(data).String())

	// Check sub types
	expect = "lxDb.FindOptions"
	its.Equal(expect, reflect.TypeOf(data.FindOptions).String())

	expect = "map[string]interface {}"
	its.Equal(expect, reflect.TypeOf(data.Query).String())

	// Check values
	its.Equal(map[string]int{"name": 1, "email": -1}, data.FindOptions.Sort)
	its.Equal(int64(5), data.FindOptions.Skip)
	its.Equal(map[string]interface{}{"name": "Schulterglatze"}, data.Query)
	its.True(data.Count)
}
