package lxDb

// ChangeInfo holds details about the outcome of an update operation.
type ChangeInfo struct {
	Updated int // Number of documents updated
	Removed int // Number of documents removed
	Matched int // Number of documents matched but not necessarily changed
}

type Options struct {
	Sort  string `json:"sort,omitempty"`
	Skip  int    `json:"skip"`
	Limit int    `json:"limit"`
	Count bool   `json:"count"`
}
