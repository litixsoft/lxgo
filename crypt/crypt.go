package lxCrypt

import (
	"golang.org/x/crypto/bcrypt"
)

// ICrypt,
// interface for mapping bcrypt
type ICrypt interface {
	GeneratePassword(plainPwd string) (string, error)
	ComparePassword(plainPwd, hashedPwd string) error
}

// Crypt,
// type for bcrypt mapper
type Crypt struct{}

// return crypt instance
func NewCrypt() *Crypt {
	return &Crypt{}
}

// GeneratePassword,
// mapper for create new encrypt password from plain password
func (c *Crypt) GeneratePassword(plainPwd string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPwd), nil
}

// ComparePassword,
// mapper for compare encrypt password with plain password
func (c *Crypt) ComparePassword(plainPwd, hashedPwd string) error {
	pPwd := []byte(plainPwd)
	hPwd := []byte(hashedPwd)
	return bcrypt.CompareHashAndPassword(hPwd, pPwd)
}
