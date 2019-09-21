package lxErrors

// HandlePanicErr, handle error as panic
func HandlePanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
