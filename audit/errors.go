package lxAudit

import (
	"fmt"
	"net/http"
)

// AuditLogError
type AuditLogEntryError struct {
	Code    int         `json:"-"`
	Message interface{} `json:"message"`
}

// AuditLogError creates a new AuditLogError instance.
func NewAuditLogEntryError(code int, message ...interface{}) *AuditLogEntryError {
	he := &AuditLogEntryError{Code: code, Message: http.StatusText(code)}
	if len(message) > 0 {
		he.Message = message[0]
	}
	return he
}

// Error
func (e *AuditLogEntryError) Error() string {
	return fmt.Sprintf("code:%d message:%s", e.Code, e.Message)
}
