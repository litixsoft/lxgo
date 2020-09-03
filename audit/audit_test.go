package lxAudit_test

import (
	"encoding/json"
	lxAudit "github.com/litixsoft/lxgo/audit"
	lxHelper "github.com/litixsoft/lxgo/helper"
	lxLog "github.com/litixsoft/lxgo/log"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
	//"time"
)

var (
	testKey            = "d6a34742-0a91-4fa9-81c3-934c76f72634"
	testClientHost     = "test.host"
	testCollectionName = "test.collection"
	testUser           = lxHelper.M{"name": "Timo Liebetrau", "age": float64(45)}
	testData           = lxHelper.M{"customer": "Karl Lagerfeld"}
	expectedEntries    = []lxHelper.M{
		{
			"host":       testClientHost,
			"collection": testCollectionName,
			"action":     lxAudit.Insert,
			"user":       testUser,
			"data":       testData,
		},
		{
			"host": testClientHost,
			//"collection": testCollectionName,
			"action": lxAudit.Update,
			"user":   testUser,
			"data":   testData,
		},
		{
			"host":       testClientHost,
			"collection": testCollectionName,
			"action":     lxAudit.Delete,
			"user":       testUser,
			"data":       testData,
		},
	}
	expectedEntry = lxHelper.M{
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
		checkFunc := func(expected lxHelper.M, actual map[string]interface{}) {
			for k := range expected {
				if k != "user" && k != "data" {
					its.Equal(expected[k], actual[k])
				}
				if k == "user" {
					for k2, v2 := range actual[k].(map[string]interface{}) {
						its.Equal(expected[k].(lxHelper.M)[k2], v2)
					}
				}
				if k == "data" {
					for k2, v2 := range actual[k].(map[string]interface{}) {
						its.Equal(expected[k].(lxHelper.M)[k2], v2)
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
				actual, ok := jBody[i].(map[string]interface{})
				its.True(ok)
				checkFunc(exp, actual)
			}
		}

		switch rtStatus {
		default:
			// Send http.StatusNoContent with 200 for successfully audit
			rw.WriteHeader(http.StatusOK)
		case http.StatusInternalServerError:
			// http.StatusInternalServerError with content
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusInternalServerError)

			data := lxHelper.M{"message": "ganz schlimm"}
			jsonData, err := json.Marshal(data)
			its.NoError(err)

			_, err = io.WriteString(rw, string(jsonData))
			its.NoError(err)
		case http.StatusUnprocessableEntity:
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusUnprocessableEntity)

			data := []lxHelper.M{
				{"message1": "ganz schlimm1"},
				{"message2": "ganz schlimm2"},
			}
			jsonData, err := json.Marshal(data)
			its.NoError(err)

			_, err = io.WriteString(rw, string(jsonData))
			its.NoError(err)
		}
	}))
}

func TestRequestAudit(t *testing.T) {
	its := assert.New(t)

	auditEntries := []lxAudit.AuditEntry{
		{
			Collection: testCollectionName,
			Action:     lxAudit.Insert,
			User:       testUser,
			Data:       testData,
		},
		{
			Collection: testCollectionName,
			Action:     lxAudit.Update,
			User:       testUser,
			Data:       testData,
		},
		{
			Collection: testCollectionName,
			Action:     lxAudit.Delete,
			User:       testUser,
			Data:       testData,
		},
	}
	//auditEntry := lxAudit.AuditEntry{
	//	Collection: testCollectionName,
	//	Action:     lxAudit.Update,
	//	User:       testUser,
	//	Data:       testData,
	//}

	t.Run("auditEntries http.StatusOK", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusOK, "/v1/bulk/log")
		defer server.Close()

		// Test request entry and check error
		its.NoError(lxAudit.RequestAudit(auditEntries, testClientHost, server.URL, testKey))
	})
	t.Run("auditEntries http.StatusInternalServerError", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError, "/v1/bulk/log")
		defer server.Close()

		// Test request entries and check error
		err := lxAudit.RequestAudit(auditEntries, testClientHost, server.URL, testKey)
		its.Error(err)
		//its.IsType(&lxAudit.AuditLogEntryError{}, err)
		//its.Equal(http.StatusInternalServerError, err.(*lxAudit.AuditLogEntryError).Code)
	})
	t.Run("auditEntries http.StatusUnprocessableEntity", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusUnprocessableEntity, "/v1/bulk/log")
		defer server.Close()

		// Test request entries and check error
		err := lxAudit.RequestAudit(auditEntries, testClientHost, server.URL, testKey)
		its.Error(err)
		//its.IsType(&lxAudit.AuditLogEntryError{}, err)
		//its.Equal(http.StatusInternalServerError, err.(*lxAudit.AuditLogEntryError).Code)
	})
}

func TestTheWest(t *testing.T) {
	// logger init
	lxLog.InitLogger(
		os.Stdout,
		"debug",
		"text")

	lxAudit.InitJobConfig(
		testClientHost,
		"http://localhost:3030",
		testKey,
		lxLog.GetLogger().WithFields(logrus.Fields{"fun": "TestTheWest"}))

	lxAudit.StartWorker()

	//job.Queue <- lxHelper.M{"message":"HelloWorld"}

	numOfJobs := 20
	c := lxAudit.GetChanConfig()
	for i := 0; i < numOfJobs; i++ {
		go func(num int) {
			c.JobChan <- lxAudit.AuditEntry{
				Collection: testCollectionName,
				Action:     lxAudit.Delete,
				User:       testUser,
				Data:       testData,
			}
		}(i)
	}

	time.Sleep(time.Second * 10)
	c.KillChan <- true
	time.Sleep(time.Second * 2)

	//auditEntry := lxAudit.AuditEntry{
	//	Collection: testCollectionName,
	//	Action:     lxAudit.Update,
	//	User:       testUser,
	//	Data:       testData,
	//}
	//
	//err := lxAudit.RequestAudit(auditEntry,testClientHost, "http://localhost:3030", "")
	//t.Log(err)
}

//func TestAudit_LogEntry(t *testing.T) {
//	its := assert.New(t)
//
//	t.Run("http.StatusOK", func(t *testing.T) {
//		// get server and close the server when test finishes
//		server := getTestServer(t, http.StatusOK, "/v1/log")
//		defer server.Close()
//
//		// instance
//		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
//
//		// log test entry and check error
//		its.NoError(audit.LogEntry(lxAudit.Update, testUser, testData))
//	})
//	t.Run("http.StatusInternalServerError", func(t *testing.T) {
//		// get server and close the server when test finishes
//		server := getTestServer(t, http.StatusInternalServerError, "/v1/log")
//		defer server.Close()
//
//		// instance
//		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
//
//		// log test entry and check error
//		err := audit.LogEntry(lxAudit.Update, testUser, testData)
//		its.Error(err)
//		its.IsType(&lxAudit.AuditLogEntryError{}, err)
//		its.Equal(http.StatusInternalServerError, err.(*lxAudit.AuditLogEntryError).Code)
//	})
//	t.Run("http.StatusUnprocessableEntity", func(t *testing.T) {
//		// get server and close the server when test finishes
//		server := getTestServer(t, http.StatusUnprocessableEntity, "/v1/log")
//		defer server.Close()
//
//		// instance
//		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
//
//		// log test entry and check error
//		err := audit.LogEntry(lxAudit.Update, testUser, testData)
//		its.Error(err)
//		its.IsType(&lxAudit.AuditLogEntryError{}, err)
//		its.Equal(http.StatusUnprocessableEntity, err.(*lxAudit.AuditLogEntryError).Code)
//	})
//}

//func TestAudit_LogEntries(t *testing.T) {
//	its := assert.New(t)
//
//	// testEntries for check
//	testEntries := bson.A{
//		bson.M{
//			"action": "insert",
//			"user": bson.M{
//				"name": "Timo Liebetrau",
//			},
//			"data": bson.M{
//				"firstname": "Timo_1",
//				"lastname":  "Liebetrau_1",
//			},
//		},
//		bson.M{
//			"action": "update",
//			"user": bson.M{
//				"name": "Timo Liebetrau",
//			},
//			"data": bson.M{
//				"firstname": "Timo_2",
//				"lastname":  "Liebetrau_2",
//			},
//		},
//		bson.M{
//			"action": "delete",
//			"user": bson.M{
//				"name": "Timo Liebetrau",
//			},
//			"data": bson.M{
//				"firstname": "Timo_3",
//				"lastname":  "Liebetrau_3",
//			},
//		},
//	}
//
//	t.Run("http.StatusOK", func(t *testing.T) {
//		// get server and close the server when test finishes
//		server := getTestServer(t, http.StatusOK, "/v1/bulk/log")
//		defer server.Close()
//
//		// instance
//		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
//		its.NoError(audit.LogEntries(testEntries))
//	})
//	t.Run("http.StatusInternalServerError", func(t *testing.T) {
//		// get server and close the server when test finishes
//		server := getTestServer(t, http.StatusInternalServerError, "/v1/bulk/log")
//		defer server.Close()
//
//		// instance
//		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
//
//		// log test entry and check error
//		err := audit.LogEntries(testEntries)
//		its.Error(err)
//		its.IsType(&lxAudit.AuditLogEntryError{}, err)
//		its.Equal(http.StatusInternalServerError, err.(*lxAudit.AuditLogEntryError).Code)
//	})
//	t.Run("http.StatusUnprocessableEntity", func(t *testing.T) {
//		// get server and close the server when test finishes
//		server := getTestServer(t, http.StatusUnprocessableEntity, "/v1/bulk/log")
//		defer server.Close()
//
//		// instance
//		audit := lxAudit.NewAudit(testClientHost, testCollectionName, server.URL, testKey)
//
//		// log test entry and check error
//		err := audit.LogEntries(testEntries)
//		its.Error(err)
//		its.IsType(&lxAudit.AuditLogEntryError{}, err)
//		its.Equal(http.StatusUnprocessableEntity, err.(*lxAudit.AuditLogEntryError).Code)
//	})
//}

func TestAudit_Worker(t *testing.T) {
	//message := lxHelper.M{
	//	"hello": "world",
	//	"life":  42,
	//	"embedded": lxHelper.M{
	//		"yes": "of course!",
	//	},
	//}
	//
	//jsonData, err := json.Marshal(message)
	//if err != nil {
	//	t.Log(err)
	//	t.FailNow()
	//}
	//
	//client := &http.Client{
	//	Timeout: time.Second * 10,
	//}
	//
	//req, err := http.NewRequest("POST", "http://localhost:3030", bytes.NewBuffer(jsonData))
	//if err != nil {
	//	t.Log(err)
	//	t.FailNow()
	//}
	//req.Header.Set("Content-type", "application/json")
	//resp, err := client.Do(req)
	//if err != nil {
	//	t.Log(err)
	//	t.FailNow()
	//}
	//
	//// Check error
	//var result lxHelper.M
	//if resp.ContentLength > 0 {
	//	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
	//		t.Log(err)
	//		t.Fail()
	//	}
	//}
	//
	//t.Log(result)

	// get server and close the server when test finishes
	//server := getTestServer(t, http.StatusOK, "/v1/log")
	//defer server.Close()
	//
	//// Global
	//log := lxLog.InitLogger(os.Stderr, "debug", "text")
	//logEntry := log.WithFields(logrus.Fields{
	//	"app":  "lxgo audit_test",
	//	"func": "worker queue",
	//})
	//lxAudit.InitAuditWorker(testClientHost, "http://localhost:3000", testKey, logEntry, 2)
	//
	//// instance
	//if !lxAudit.HasAuditWorkerInit() {
	//	t.Log("not init worker queue")
	//	t.FailNow()
	//}
	//worker := lxAudit.GetWorkerConfig()
	//numberOfJobs := 20
	//for j := 0; j < numberOfJobs; j++ {
	//	go func(j int) {
	//		t.Log("Start Job:", j)
	//		worker.Queue <- lxAudit.AuditEntry{
	//			Collection: testCollectionName,
	//			Action:     fmt.Sprintf("%s_%d", lxAudit.Update, j),
	//			User:       lxHelper.M{"name": fmt.Sprintf("tester_%d", j)},
	//			Data:       lxHelper.M{"foo": fmt.Sprintf("bar_%d", j)},
	//		}
	//	}(j)
	//}
	//
	//t.Log("Wait for jobs:", numberOfJobs)
	//for c := 0; c < numberOfJobs; c++ {
	//	<-worker.Err
	//	<-worker.Done
	//}
	//
	//// cleaning workers
	//close(worker.Kill)

	//t.Log("Post")
	//lxAudit.LogEntry(lxAudit.AuditEntry{
	//	Collection: testCollectionName,
	//	Action:     lxAudit.Update,
	//	User:       testUser,
	//	Data:       testData,
	//})

}
