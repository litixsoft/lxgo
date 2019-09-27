package lxDb_test

import (
	lxDb "github.com/litixsoft/lxgo/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestToBsonDoc(t *testing.T) {
	its := assert.New(t)

	type TestType struct {
		FirstName string
		LastName  string
		Age       int
	}

	testType := TestType{
		FirstName: "Karl",
		LastName:  "Ransauer",
		Age:       44,
	}

	doc, err := lxDb.ToBsonDoc(&testType)
	its.NoError(err)
	its.IsType(&bson.D{}, doc)
	its.Equal(testType.FirstName, doc.Map()["firstname"])
	its.Equal(testType.LastName, doc.Map()["lastname"])
}

func TestToBsonMap(t *testing.T) {
	its := assert.New(t)

	type TestType struct {
		FirstName string
		LastName  string
		Age       int
	}

	testType := TestType{
		FirstName: "Karl",
		LastName:  "Ransauer",
		Age:       44,
	}

	doc, err := lxDb.ToBsonMap(&testType)
	its.NoError(err)
	its.IsType(bson.M{}, doc)
	its.Equal(testType.FirstName, doc["firstname"])
	its.Equal(testType.LastName, doc["lastname"])
}
