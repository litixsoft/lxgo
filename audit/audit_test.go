package lxaudit_test

import (
	"github.com/litixsoft/lxgo/audit"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"net/http"
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

func TestGetAuditInstance(t *testing.T) {
	t.Run("without init", func(t *testing.T) {
		// get audit instance, should be fail with panic
		assert.Panics(t, func() { lxaudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName) })
	})
	t.Run("with init", func(t *testing.T) {
		// init audit instance
		lxaudit.InitAuditConfigInstance(&http.Client{}, "http://test", TestKey, TestHost)

		// get audit instance
		audit := lxaudit.GetAuditInstance(TestDbHost, TestDbName, TestCollectionName)

		// check instance, should not be nil
		assert.NotNil(t, audit)

		// check type
		assert.Equal(t, "*lxaudit.audit", reflect.TypeOf(audit).String())
	})
}

//func TestAudit_Log(t *testing.T) {
//	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
//		// Test request parameters
//		assert.Equal(t, req.URL.String(), TestPath)
//		body, err := ioutil.ReadAll(req.Body)
//		assert.NoError(t, err)
//
//		// convert body for check
//		jsonBody := new(lxHelper.M)
//		assert.NoError(t, json.Unmarshal(body, jsonBody))
//
//		// expected map
//		expected := &lxHelper.M{
//			"db_host":    TestDbHost,
//			"host":       TestHost,
//			"db":         TestDbName,
//			"collection": TestCollectionName,
//			"action":     lxaudit.Remove,
//			"user":       testUser,
//			"data":       testData,
//		}
//
//		// check request body
//		assert.ObjectsAreEqual(expected, jsonBody)
//
//		// Send http.StatusNoContent for successfully audit
//		rw.WriteHeader(http.StatusOK)
//	}))
//	// Close the server when test finishes
//	defer server.Close()
//
//	// init logger
//	lxaudit.InitAuditInstance(server.Client(), server.URL, TestKey, TestHost)
//
//	// log test entry and check error
//	assert.NoError(t, lxaudit.GetAuditInstance().Log(TestDbHost, TestDbName, TestCollectionName, lxaudit.Remove, testUser, testData))
//}

//func TestAudit_Log2(t *testing.T) {
//	t.Run("InternalServerError", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
//			// Test request parameters
//			assert.Equal(t, req.URL.String(), TestPath)
//			body, err := ioutil.ReadAll(req.Body)
//			assert.NoError(t, err)
//
//			// convert body for check
//			jsonBody := new(lxHelper.M)
//			assert.NoError(t, json.Unmarshal(body, jsonBody))
//
//			// expected map
//			expected := &lxHelper.M{
//				"db_host":    TestDbHost,
//				"host":       TestHost,
//				"db":         TestDbName,
//				"collection": TestCollectionName,
//				"action":     lxaudit.Remove,
//				"user":       testUser,
//				"data":       testData,
//			}
//
//			// check request body
//			assert.ObjectsAreEqual(expected, jsonBody)
//
//			// Send http.StatusInternalServerError for error test
//			rw.WriteHeader(http.StatusInternalServerError)
//			_, err = rw.Write(body)
//			assert.NoError(t, err)
//		}))
//		// Close the server when test finishes
//		defer server.Close()
//
//		// init logger
//		lxaudit.InitAuditInstance(server.Client(), server.URL, TestKey, TestHost)
//
//		// log test entry and check error
//		err := lxaudit.GetAuditInstance().Log(TestDbHost, TestDbName, TestCollectionName, TestDefaultAction, testUser, testData)
//		assert.Error(t, err)
//	})
//	t.Run("StatusUnprocessableEntity", func(t *testing.T) {
//		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
//			// Test request parameters
//			assert.Equal(t, req.URL.String(), TestPath)
//			body, err := ioutil.ReadAll(req.Body)
//			assert.NoError(t, err)
//
//			// convert body for check
//			jsonBody := new(lxHelper.M)
//			assert.NoError(t, json.Unmarshal(body, jsonBody))
//
//			// expected map
//			expected := &lxHelper.M{
//				"db_host":    TestDbHost,
//				"host":       TestHost,
//				"db":         TestDbName,
//				"collection": TestCollectionName,
//				"action":     lxaudit.Remove,
//				"user":       testUser,
//				"data":       testData,
//			}
//
//			// check request body
//			assert.ObjectsAreEqual(expected, jsonBody)
//
//			// Send http.StatusInternalServerError for error test
//			rw.WriteHeader(http.StatusUnprocessableEntity)
//			_, err = rw.Write(body)
//			assert.NoError(t, err)
//		}))
//		// Close the server when test finishes
//		defer server.Close()
//
//		// init logger
//		lxaudit.InitAuditInstance(server.Client(), server.URL, TestKey, TestHost)
//
//		// log test entry and check error
//		err := lxaudit.GetAuditInstance().Log(TestDbHost, TestDbName, TestCollectionName, TestDefaultAction, testUser, testData)
//		assert.Error(t, err)
//	})
//}

//func TestAudit_Mock(t *testing.T) {
//	mockCtrl := gomock.NewController(t)
//	defer mockCtrl.Finish()
//
//	// Mock audit for test
//	mockIAudit := lxAuditMocks.NewMockIAudit(mockCtrl)
//
//	// return nil for correct log
//	mockIAudit.EXPECT().Log(gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any(),gomock.Any()).Return(nil).Times(1)
//
//	// Check
//	assert.NoError(t, mockIAudit.Log(TestDbHost, TestDbName, TestCollectionName, TestDefaultAction, testUser, testData))
//}
