package lxAudit_test

import (
	lxAudit "github.com/litixsoft/lxgo/audit"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestAuditLogEntryError_Error(t *testing.T) {
	its := assert.New(t)

	err := lxAudit.NewAuditLogEntryError(http.StatusBadRequest)
	err2 := lxAudit.NewAuditLogEntryError(http.StatusConflict, "test message")

	// Check types
	its.IsType(&lxAudit.AuditLogEntryError{}, err)
	its.IsType(&lxAudit.AuditLogEntryError{}, err2)

	// Check values
	its.Equal("code:400 message:Bad Request", err.Error())
	its.Equal("code:409 message:test message", err2.Error())
}
