package lxAudit_test

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/audit"
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
	testCollectionName = "test.collection"
	testUser           = bson.M{"name": "Timo Liebetrau", "age": float64(45)}
	testData           = bson.M{"customer": "Karl Lagerfeld"}
	expectedEntries    = bson.A{
		bson.M{
			"host":       testClientHost,
			"collection": testCollectionName,
			"action":     "insert",
			"user":       testUser,
			"data": bson.M{
				"firstname": "Timo_1",
				"lastname":  "Liebetrau_1",
			},
		},
		bson.M{
			"host":       testClientHost,
			"collection": testCollectionName,
			"action":     "update",
			"user":       testUser,
			"data": bson.M{
				"firstname": "Timo_2",
				"lastname":  "Liebetrau_2",
			},
		},
		bson.M{
			"host":       testClientHost,
			"collection": testCollectionName,
			"action":     "delete",
			"user":       testUser,
			"data": bson.M{
				"firstname": "Timo_3",
				"lastname":  "Liebetrau_3",
			},
		},
	}
	expectedEntry = bson.M{
		"host":       testClientHost,
		"collection": testCollectionName,
		"action":     lxAudit.Update,
		"user":       testUser,
		"data":       testData,
	}
)

// getTestServer with return status
func getTestServer(t *testing.T, rtStatus int, testPath string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		its := assert.New(t)
		// Test request parameters
		its.Equal(req.URL.String(), testPath)
		body, err := ioutil.ReadAll(req.Body)
		its.NoError(err)

		// convert body for check
		var jsonBody interface{}
		its.NoError(json.Unmarshal(body, &jsonBody))

		// Check actual request
		checkFunc := func(expected bson.M, actual map[string]interface{}) {
			for k := range expected {
				if k != "user" && k != "data" {
					its.Equal(expected[k], actual[k])
				}
				if k == "user" {
					for k2, v2 := range actual[k].(map[string]interface{}) {
						its.Equal(expected[k].(bson.M)[k2], v2)
					}
				}
				if k == "data" {
					for k2, v2 := range actual[k].(map[string]interface{}) {
						its.Equal(expected[k].(bson.M)[k2], v2)
					}
				}
			}
		}

		// Check jsonBody and cast request for check
		switch jBody := jsonBody.(type) {
		case map[string]interface{}:
			// Check single entry
			checkFunc(expectedEntry, jBody)
		case []interface{}:
			// Check collection of entries
			for i, exp := range expectedEntries {
				expected, ok := exp.(bson.M)
				its.True(ok)

				actual, ok := jBody[i].(map[string]interface{})
				its.True(ok)

				checkFunc(expected, actual)
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
		server := getTestServer(t, http.StatusOK, "/v1/log")
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)

		// log test entry and check error
		its.NoError(audit.LogEntry(lxAudit.Update, testUser, testData))
	})
	t.Run("http.StatusInternalServerError", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError, "/v1/log")
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
		server := getTestServer(t, http.StatusUnprocessableEntity, "/v1/log")
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

func TestAudit_LogEntries(t *testing.T) {
	its := assert.New(t)

	// testEntries for check
	testEntries := bson.A{
		bson.M{
			"action": "insert",
			"user": bson.M{
				"name": "Timo Liebetrau",
			},
			"data": bson.M{
				"firstname": "Timo_1",
				"lastname":  "Liebetrau_1",
			},
		},
		bson.M{
			"action": "update",
			"user": bson.M{
				"name": "Timo Liebetrau",
			},
			"data": bson.M{
				"firstname": "Timo_2",
				"lastname":  "Liebetrau_2",
			},
		},
		bson.M{
			"action": "delete",
			"user": bson.M{
				"name": "Timo Liebetrau",
			},
			"data": bson.M{
				"firstname": "Timo_3",
				"lastname":  "Liebetrau_3",
			},
		},
	}

	t.Run("http.StatusOK", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusOK, "/v1/bulk/log")
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
		its.NoError(audit.LogEntries(testEntries))
	})
	t.Run("http.StatusInternalServerError", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError, "/v1/bulk/log")
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)

		// log test entry and check error
		err := audit.LogEntries(testEntries)
		its.Error(err)
		its.IsType(&lxAudit.AuditLogEntryError{}, err)
		its.Equal(http.StatusInternalServerError, err.(*lxAudit.AuditLogEntryError).Code)
	})
	t.Run("http.StatusUnprocessableEntity", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusUnprocessableEntity, "/v1/bulk/log")
		defer server.Close()

		// instance
		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)

		// log test entry and check error
		err := audit.LogEntries(testEntries)
		its.Error(err)
		its.IsType(&lxAudit.AuditLogEntryError{}, err)
		its.Equal(http.StatusUnprocessableEntity, err.(*lxAudit.AuditLogEntryError).Code)
	})
}
