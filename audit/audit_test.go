package lxAudit_test

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

var (
	testKey            = "c470e652-6d46-4f9d-960d-f32d84e682e7"
	testClientHost     = "test.host"
	testPath           = "/log"
	testDbHost         = "test.dbhost"
	testDbName         = "test.dbName"
	testCollectionName = "test.collection"
	testUser           = lxHelper.M{"name": "Timo Liebetrau", "age": float64(45)}
	testData           = lxHelper.M{"customer": "Karl Lagerfeld"}
)

// getTestServer with return status
func getTestServer(t *testing.T, rtStatus int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), testPath)
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)

		// convert body for check
		var jsonBody lxHelper.M
		assert.NoError(t, json.Unmarshal(body, &jsonBody))

		// expected map
		expected := lxHelper.M{
			"db_host":    testDbHost,
			"host":       testClientHost,
			"db":         testDbName,
			"collection": testCollectionName,
			"action":     lxAudit.Remove,
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
					assert.Equal(t, expected[k].(lxHelper.M)[k2], v2)
				}
			}
			if k == "data" {
				for k2, v2 := range jsonBody[k].(map[string]interface{}) {
					assert.Equal(t, expected[k].(lxHelper.M)[k2], v2)
				}
			}
		}

		// Send http.StatusNoContent for successfully audit
		rw.WriteHeader(rtStatus)
	}))
}

func TestGetAuditInstance(t *testing.T) {
	t.Run("without init", func(t *testing.T) {
		// get audit instance, should be fail with panic
		assert.Panics(t, func() { lxAudit.GetAuditInstance(testDbHost, testDbName, testCollectionName) })
	})
	t.Run("with init", func(t *testing.T) {
		// init audit instance
		lxAudit.InitAuditConfigInstance(testClientHost, "http://test", testKey)

		// get audit instance
		audit := lxAudit.GetAuditInstance(testDbHost, testDbName, testCollectionName)

		// check instance, should not be nil
		assert.NotNil(t, audit)

		// check type
		assert.Equal(t, "*lxAudit.audit", reflect.TypeOf(audit).String())
	})
}

func TestAudit_Log(t *testing.T) {
	t.Run("http.StatusOK", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusOK)
		defer server.Close()

		// init logger
		lxAudit.InitAuditConfigInstance(testClientHost, server.URL, testKey)

		// log test entry and check error
		assert.NoError(t, lxAudit.GetAuditInstance(testDbHost, testDbName, testCollectionName).Log(lxAudit.Remove, testUser, testData))
	})
	t.Run("http.StatusInternalServerError", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError)
		defer server.Close()

		// init logger
		lxAudit.InitAuditConfigInstance(testClientHost, server.URL, testKey)

		// log test entry and check error
		assert.Error(t, lxAudit.GetAuditInstance(testDbHost, testDbName, testCollectionName).Log(lxAudit.Remove, testUser, testData))

	})
	t.Run("http.StatusUnprocessableEntity", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusUnprocessableEntity)
		defer server.Close()

		// init logger
		lxAudit.InitAuditConfigInstance(testClientHost, server.URL, testKey)

		// log test entry and check error
		assert.Error(t, lxAudit.GetAuditInstance(testDbHost, testDbName, testCollectionName).Log(lxAudit.Remove, testUser, testData))
	})
}
