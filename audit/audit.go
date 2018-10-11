package lxAudit

import (
	"time"
)

// AuditAction for analyses audit entries
type AuditAction int

const (
	Create = iota
	Read
	Update
	Delete
	Log
)

// IAudit, interface for audit repositories
type IAudit interface {
	SetupAudit() error
	Log(action int, user, message, data interface{}) (<-chan bool, <-chan error)
}

// AuditModel, model for audit entry
type AuditModel struct {
	TimeStamp   time.Time   `json:"timestamp"`
	ServiceName string      `json:"service_name"`
	ServiceHost string      `json:"service_host"`
	Action      int         `json:"action"`
	User        interface{} `json:"user"`
	Message     interface{} `json:"msg"`
	Data        interface{} `json:"data"`
}
