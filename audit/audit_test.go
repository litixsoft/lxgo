package lxAudit_test

import (
	"encoding/json"
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

const (
	TestKey            = "c470e652-6d46-4f9d-960d-f32d84e682e7"
	TestHost           = "test.host"
	TestPath           = "/log"
	TestDbHost         = "test.dbhost"
	TestDbName         = "test.dbName"
	TestCollectionName = "test.collection"
	TestDefaultAction  = "remove"
)

var (
	testUser = lxHelper.M{"name": "Timo Liebetrau", "age": 45}
	testData = lxHelper.M{"customer": "Karl Lagerfeld"}
)

// getTestServer with return status
func getTestServer(t *testing.T, rtStatus int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), TestPath)
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)

		// convert body for check
		jsonBody := new(lxHelper.M)
		assert.NoError(t, json.Unmarshal(body, jsonBody))

		// expected map
		expected := &lxHelper.M{
			"db_host":    TestDbHost,
			"host":       TestHost,
			"db":         TestDbName,
			"collection": TestCollectionName,
			"action":     lxAudit.Remove,
			"user":       testUser,
			"data":       testData,
		}

		// check request body
		assert.ObjectsAreEqual(expected, jsonBody)

		// Send http.StatusNoContent for successfully audit
		rw.WriteHeader(rtStatus)
	}))
}

func TestGetAuditInstance(t *testing.T) {
	t.Run("without init", func(t *testing.T) {
		// get audit instance, should be fail with panic
		assert.Panics(t, func() { lxAudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName) })
	})
	t.Run("with init", func(t *testing.T) {
		// init audit instance
		lxAudit.InitAuditConfigInstance(&http.Client{}, "http://test", TestKey, TestHost)

		// get audit instance
		audit := lxAudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName)

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
		lxAudit.InitAuditConfigInstance(server.Client(), server.URL, TestKey, TestHost)

		// log test entry and check error
		assert.NoError(t, lxAudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName).Log(lxAudit.Remove, testUser, testData))
	})
	t.Run("http.StatusInternalServerError", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError)
		defer server.Close()

		// init logger
		lxAudit.InitAuditConfigInstance(server.Client(), server.URL, TestKey, TestHost)

		// log test entry and check error
		assert.Error(t, lxAudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName).Log(lxAudit.Remove, testUser, testData))

	})
	t.Run("http.StatusUnprocessableEntity", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusUnprocessableEntity)
		defer server.Close()

		// init logger
		lxAudit.InitAuditConfigInstance(server.Client(), server.URL, TestKey, TestHost)

		// log test entry and check error
		assert.Error(t, lxAudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName).Log(lxAudit.Remove, testUser, testData))
	})
}

func TestMockIAudit_EXPECT(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock audit for test
	mockIAudit := lxAudit.NewMockIAudit(mockCtrl)

	// set expects for correct and error
	mockIAudit.EXPECT().Log(TestDefaultAction, testUser, testData).Return(nil).Times(1)
	mockIAudit.EXPECT().Log(TestDefaultAction, testUser, testData).Return(errors.New("test error")).Times(1)

	// check correct and error
	assert.NoError(t, mockIAudit.Log(TestDefaultAction, testUser, testData))
	assert.Error(t, mockIAudit.Log(TestDefaultAction, testUser, testData))
}
