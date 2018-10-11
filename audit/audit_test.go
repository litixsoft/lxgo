package lxAudit_test

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestAuditModel_ToJson(t *testing.T) {
	// test time
	testTime := time.Now()

	// entry for test
	testEntry := lxAudit.AuditModel{
		TimeStamp:   testTime,
		ServiceName: "TestService",
		ServiceHost: "TestHost",
		Action:      lxAudit.Log,
		User:        lxHelper.M{"name": "Timo Liebetrau"},
		Message:     "TestMessage",
		Data:        lxHelper.M{"data": "TestData"},
	}

	// convert entry for expected
	conEntry := lxHelper.M{
		"timestamp":    lxHelper.M{"$date": testEntry.TimeStamp.UTC().Format(time.RFC3339)},
		"service_name": testEntry.ServiceName,
		"service_host": testEntry.ServiceHost,
		"action":       testEntry.Action,
		"user":         testEntry.User,
		"msg":          testEntry.Message,
		"data":         testEntry.Data,
	}

	// generate json for expected
	jentry, _ := json.Marshal(conEntry)
	expected := string(jentry)

	// test
	result := testEntry.ToJson()

	// check
	assert.Equal(t, expected, result)
}
