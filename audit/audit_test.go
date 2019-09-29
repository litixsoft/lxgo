package lxAudit_test

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	testKey            = "c470e652-6d46-4f9d-960d-f32d84e682e7"
	testClientHost     = "test.host"
	testPath           = "/v1/log"
	testCollectionName = "test.collection"
	testUser           = bson.M{"name": "Timo Liebetrau", "age": float64(45)}
	testData           = bson.M{"customer": "Karl Lagerfeld"}
)

// getTestServer with return status
func getTestServer(t *testing.T, rtStatus int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testPath)
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)

		// convert body for check
		var jsonBody bson.M
		assert.NoError(t, json.Unmarshal(body, &jsonBody))

		// expected map
		expected := lxHelper.M{
			"host":       testClientHost,
			"collection": testCollectionName,
			"action":     lxAudit.Update,
			"user":       testUser,
			"data":       testData,
		}

		// check request body
		for k := range expected {
			if k != "user" && k != "data" {
				assert.Equal(t, expected[k], jsonBody[k])
			}
			if k == "user" {
				for k2, v2 := range jsonBody[k].(map[string]interface{}) {
					assert.Equal(t, expected[k].(bson.M)[k2], v2)
				}
			}
			if k == "data" {
				for k2, v2 := range jsonBody[k].(map[string]interface{}) {
					assert.Equal(t, expected[k].(bson.M)[k2], v2)
				}
			}
		}

		// Send http.StatusNoContent for successfully audit
		rw.WriteHeader(rtStatus)
	}))
}

func TestAudit_LogEntry(t *testing.T) {
	its := assert.New(t)

	t.Run("http.StatusOK", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusOK)
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)

		// log test entry and check error
		its.NoError(audit.LogEntry(lxAudit.Update, testUser, testData))
	})
	t.Run("http.StatusInternalServerError", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError)
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)

		// log test entry and check error
		err := audit.LogEntry(lxAudit.Update, testUser, testData)
		its.Error(err)
		its.IsType(&lxAudit.AuditLogEntryError{}, err)
		its.Equal(http.StatusInternalServerError, err.(*lxAudit.AuditLogEntryError).Code)
	})
	t.Run("http.StatusUnprocessableEntity", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusUnprocessableEntity)
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)

		// log test entry and check error
		err := audit.LogEntry(lxAudit.Update, testUser, testData)
		its.Error(err)
		its.IsType(&lxAudit.AuditLogEntryError{}, err)
		its.Equal(http.StatusUnprocessableEntity, err.(*lxAudit.AuditLogEntryError).Code)
	})
}
