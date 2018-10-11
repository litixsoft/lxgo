package lxAudit

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/helper"
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
	Log(action int, user, message, data interface{}) chan bool
}

// AuditModel, model for audit entry
type AuditModel struct {
	TimeStamp   time.Time   `json:"timestamp" bson:"timestamp"`
	ServiceName string      `json:"service_name" bson:"service_name"`
	ServiceHost string      `json:"service_host" bson:"service_host"`
	Action      int         `json:"action" bson:"action"`
	User        interface{} `json:"user" bson:"user"`
	Message     interface{} `json:"msg" bson:"msg"`
	Data        interface{} `json:"data" bson:"data"`
}

// ToJson, convert model to json string for log
func (am *AuditModel) ToJson() string {

	conEntry := lxHelper.M{
		"timestamp":    lxHelper.M{"$date": am.TimeStamp.UTC().Format(time.RFC3339)},
		"service_name": am.ServiceName,
		"service_host": am.ServiceHost,
		"action":       am.Action,
		"user":         am.User,
		"msg":          am.Message,
		"data":         am.Data,
	}

	jentry, _ := json.Marshal(conEntry)

	return string(jentry)
}
