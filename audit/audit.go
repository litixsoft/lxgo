package lxAudit

import "time"

// IAudit, interface for audit repositories
type IAudit interface {
	SetupAudit() error
	Log(user, message, data interface{}) (<-chan bool, <-chan error)
}

// AuditModel, model for audit entry
type AuditModel struct {
	TimeStamp   time.Time   `json:"timestamp"`
	ServiceName string      `json:"service_name"`
	ServiceHost string      `json:"service_host"`
	User        interface{} `json:"user"`
	Message     interface{} `json:"msg"`
	Data        interface{} `json:"data"`
}
