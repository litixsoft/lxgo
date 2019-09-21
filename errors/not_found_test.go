package lxErrors_test

import (
	lxErrors "github.com/litixsoft/lxgo/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotFoundError_Error(t *testing.T) {
	its := assert.New(t)

	err := lxErrors.NewNotFoundError()
	err2 := lxErrors.NewNotFoundError("test message")

	// Check types
	its.IsType(&lxErrors.NotFoundError{}, err)
	its.IsType(&lxErrors.NotFoundError{}, err2)

	// Check values
	its.Equal("not found", err.Error())
	its.Equal("test message", err2.Error())
}
