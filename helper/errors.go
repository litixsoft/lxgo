package lxHelper

// HandlePanicErr, handle error as panic
func HandlePanicErr(err error) {
	if err != nil {
		panic(err)
	}
}

// DuplicateKeyError
type DuplicateKeyError struct {
	Err       string
	FieldName string
	Value     interface{}
}

func (e *DuplicateKeyError) Error() string {
	return "DUPLICATE_KEY_ERROR"
}
