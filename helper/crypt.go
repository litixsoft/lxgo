package lxHelper

import "golang.org/x/crypto/bcrypt"

/////////////////////////////////////////////////
// deprecated, Will be removed in a later version
/////////////////////////////////////////////////
// GenerateFromPassword,
// create a new encrypted password from plain password
func GenerateFromPassword(plainPwd string) (string, error) {
	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(plainPwd), bcrypt.MinCost)
	if err != nil {
		return "", err
	}
	return string(hashedPwd), nil
}

// CompareHashAndPassword,
// compare encrypt password with plain password
func CompareHashAndPassword(hashedPwd, plainPwd string) error {
	hPwd := []byte(hashedPwd)
	pPwd := []byte(plainPwd)
	return bcrypt.CompareHashAndPassword(hPwd, pPwd)
}
