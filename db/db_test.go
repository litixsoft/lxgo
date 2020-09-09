package lxDb_test

import (
	lxDb "github.com/litixsoft/lxgo/db"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"testing"
)

func TestSetAuditAuth(t *testing.T) {
	its := assert.New(t)

	// Auth user
	au := &bson.M{"firstname": "test", "lastname": "test test"}

	// Run test
	res := lxDb.SetAuditAuth(au)

	// Should be type
	its.IsType(&lxDb.AuditAuth{}, res)

	// Test from args
	// Simulate wrap in the functions
	args := []interface{}{res}
	var authUser interface{}

	switch val := args[0].(type) {
	case *lxDb.AuditAuth:
		authUser = val.User
	}

	its.Equal(au, authUser)
}

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

func TestToBsonMap2(t *testing.T) {

}
