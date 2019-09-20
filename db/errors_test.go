package lxDb_test

import (
	lxDb "github.com/litixsoft/lxgo/db"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotFoundError_Error(t *testing.T) {
	its := assert.New(t)

	err := lxDb.NewNotFoundError()
	err2 := lxDb.NewNotFoundError("test message")

	// Check types
	its.IsType(&lxDb.NotFoundError{}, err)
	its.IsType(&lxDb.NotFoundError{}, err2)

	// Check values
	its.Equal("not found", err.Error())
	its.Equal("test message", err2.Error())
}
