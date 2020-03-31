package lxDb

import "errors"

// Error by convert index model
var ErrIndexConvert = errors.New("can't convert index models")

// Not Found error
var ErrNotFound = errors.New("not found")

// NotFoundError
type NotFoundError struct {
	Message string
}

// NewApiPostError creates a new HTTPError instance.
func NewNotFoundError(message ...interface{}) *NotFoundError {
	he := &NotFoundError{Message: "not found"}
	if len(message) > 0 {
		he.Message = message[0].(string)
	}
	return he
}

func (e *NotFoundError) Error() string {
	return e.Message
}
