package lxAudit

import (
	"github.com/globalsign/mgo/bson"
	lxHelper "github.com/litixsoft/lxgo/helper"
	"net/http"
	"time"
)

const (
	Insert         = "insert"
	Update         = "update"
	Delete         = "delete"
	DefaultTimeout = time.Second * 30
	PathLogEntry   = "/v1/logEntry"
)

// IAudit interface for audit logger
type IAudit interface {
	LogEntry(action string, user, data interface{}, timeout ...time.Duration) error
}

type audit struct {
	clientHost     string
	collectionName string
	auditHost      string
	auditAuthKey   string
}

func NewAudit(clientHost, collectionName, auditHost, auditAuthKey string) IAudit {
	return &audit{
		clientHost:     clientHost,
		collectionName: collectionName,
		auditHost:      auditHost,
		auditAuthKey:   auditAuthKey,
	}
}

// Log, send post request to audit service
func (al *audit) LogEntry(action string, user, data interface{}, timeout ...time.Duration) error {
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	// Set entry for request
	body := bson.M{
		"host":       al.clientHost,
		"collection": al.collectionName,
		"action":     action,
		"user":       user,
		"data":       data,
	}

	// Header
	header := http.Header{}
	header.Add("Content-Type", "application/json")
	header.Add("Authorization", "Bearer "+al.auditAuthKey)

	// Uri
	uri := al.auditHost + PathLogEntry

	// Request
	resp, err := lxHelper.Request(header, body, uri, "POST", to)
	if err != nil {
		return err
	}

	// Check status
	if resp.StatusCode != http.StatusOK {
		var ret bson.M
		if err := lxHelper.ReadResponseBody(resp, &ret); err != nil {
			return err
		}

		// Check message exits
		if msg, ok := ret["message"]; ok {
			return NewAuditLogEntryError(resp.StatusCode, msg)
		}

		return NewAuditLogEntryError(resp.StatusCode)
	}

	return nil
}
