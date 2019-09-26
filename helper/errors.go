package lxHelper

// HandlePanicErr, handle error as panic
func HandlePanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

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
