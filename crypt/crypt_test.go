package lxCrypt_test

import (
	"github.com/litixsoft/lxgo/crypt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewCrypt(t *testing.T) {
	c := lxCrypt.NewCrypt()
	assert.IsType(t, &lxCrypt.Crypt{}, c)
}

func TestCrypt_GeneratePassword(t *testing.T) {
	c := lxCrypt.Crypt{}

	cryptPwd, err := c.GeneratePassword("plain-pwd")
	assert.NoError(t, err)

	err = c.ComparePassword(cryptPwd, "plain-pwd")
	assert.NoError(t, err)
}
