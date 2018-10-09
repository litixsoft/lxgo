package lxHelper

// DuplicateKeyError
type DuplicateKeyError struct {
	Err       string
	FieldName string
	Value     interface{}
}

func (e *DuplicateKeyError) Error() string {
	return "DUPLICATE_KEY_ERROR"
}
