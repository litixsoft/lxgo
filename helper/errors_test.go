package lxHelper_test

import (
	lxHelper "github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNotFoundError_Error(t *testing.T) {
	its := assert.New(t)

	err := lxHelper.NewNotFoundError()
	err2 := lxHelper.NewNotFoundError("test message")

	// Check types
	its.IsType(&lxHelper.NotFoundError{}, err)
	its.IsType(&lxHelper.NotFoundError{}, err2)

	// Check values
	its.Equal("not found", err.Error())
	its.Equal("test message", err2.Error())
}
