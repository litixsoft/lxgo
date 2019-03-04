package lxDb

import (
	"time"
)

const (
	ActionInsert = "insert"
	ActionFind   = "find"
	ActionUpdate = "update"
	ActionUpsert = "upsert"
	ActionRemove = "remove"
)

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

type AuditModel struct {
	TimeStamp  time.Time   `json:"timestamp" bson:"timestamp"`
	Collection string      `json:"collection" bson:"collection"`
	Action     string      `json:"action" bson:"action"`
	User       interface{} `json:"user" bson:"user"`
	Data       interface{} `json:"data" bson:"data"`
}
