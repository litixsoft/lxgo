package lxAudit_test

import (
	"encoding/json"
	"errors"
	"fmt"
	lxAudit "github.com/litixsoft/lxgo/audit"
	lxHelper "github.com/litixsoft/lxgo/helper"
	lxLog "github.com/litixsoft/lxgo/log"
	"github.com/sirupsen/logrus"
	//"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestRunningWorkers(t *testing.T) {
	its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		ioutil.Discard,
		"debug",
		"text")

	t.Run("without started worker", func(t *testing.T) {
		its.Equal(0, lxAudit.RunningWorkers())
	})
	t.Run("with started workers", func(t *testing.T) {
		// InitJobConfig with test server for request
		//lxAudit.InitJobConfig(
		//	"test-host",
		//	"http://foo.de",
		//	"",
		//	lxLog.GetLogger().WithFields(logrus.Fields{}),
		//	time.Millisecond * 1)

		// StartWorker creates the ChanConfig singleton
		// Channel error for waiting by testing
		// Worker1
		//chanErr1 := make(chan error)
		//chanErr2 := make(chan error)
		//
		//lxAudit.StartWorker(chanErr1)
		//lxAudit.StartWorker(chanErr2)

		// Worker2
		//chanErr2 := make(chan error)
		//lxAudit.StartWorker(chanErr2)

		// Check
		//its.Equal(2, lxAudit.RunningWorkers())
		//its.NoError(<-chanErr1)

		// Quit all workers
		//cc := lxAudit.GetChanConfig()
		//cc.KillChan <- true

		// Wait of kill all workers
		//its.NoError(<-chanErr1)
		//its.NoError(<-chanErr2)

		//cc := lxAudit.GetChanConfig()
		//cc.KillChan<- true
		//
		//time.Sleep(time.Second * 2)
		//
		//t.Log(lxAudit.RunningWorkers())
		//its.NoError(<-chanErr)
		//its.NoError(<-chanErr)
		//its.NoError(<-chanErr)
		//
		//// Check running workers
		//its.Equal(0, lxAudit.RunningWorkers())
	})
}

func TestRequestAudit(t *testing.T) {
	its := assert.New(t)

	var testEntries []lxAudit.AuditEntry
	for i := 0; i < 3; i++ {
		testEntries = append(testEntries, lxAudit.AuditEntry{
			Host:       fmt.Sprintf("test_host_%d", i),
			Collection: "",
			Action:     lxAudit.Insert,
			User:       lxHelper.M{"name": fmt.Sprintf("TestUser_%d", i)},
			Data:       lxHelper.M{"location": fmt.Sprintf("Fab_%d", i)},
		})
	}

	t.Run("test entries", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(testEntries, "", server.URL, "", time.Second*5)
		its.NoError(err)
	})
	t.Run("test entry", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(testEntries[1], "", server.URL, "")
		its.NoError(err)
	})
	t.Run("test error != 200", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(testEntries[1], "", server.URL, "")
		its.Error(err)
		its.True(errors.Is(err, lxAudit.ErrStatus))
	})
	t.Run("test error with content", func(t *testing.T) {
		b, err := json.Marshal([]lxHelper.M{{"email": "invalid"}})
		its.NoError(err)

		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Header().Set("content-type", "application/json")
			w.Write(b)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err = lxAudit.RequestAudit(testEntries[1], "", server.URL, "")
		its.Error(err)
		its.True(errors.Is(err, lxAudit.ErrRespContent))
	})
	t.Run("test error entry with wrong data type", func(t *testing.T) {
		wrongEntry := lxHelper.M{"name": "foo"}

		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(wrongEntry, "", server.URL, "")
		its.Error(err)
		its.True(errors.Is(lxAudit.ErrAuditEntryType, err))
	})
}

func TestSendJobToWorker(t *testing.T) {
	//its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		os.Stdout,
		"debug",
		"text")

	// For testing log output by error,
	// user and data with map[string]interface{}
	var testEntries []lxAudit.AuditEntry
	for i := 0; i < 3; i++ {
		testEntries = append(testEntries, lxAudit.AuditEntry{
			Host:       fmt.Sprintf("test_host_%d", i),
			Collection: "",
			Action:     lxAudit.Insert,
			User:       map[string]interface{}{"name": fmt.Sprintf("TestUser_%d", i)},
			Data:       map[string]interface{}{"location": fmt.Sprintf("Fab_%d", i)},
		})
	}

	t.Run("send entries to worker", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		// InitJobConfig with test server for request
		lxAudit.InitJobConfig(
			"test-host",
			server.URL,
			"",
			lxLog.GetLogger().WithFields(logrus.Fields{}),
			time.Millisecond*1)

		// Channel error for waiting by testing
		//chanErr := make(chan error)

		// StartWorker creates the ChanConfig singleton
		go lxAudit.StartWorker("alpha")
		go lxAudit.StartWorker("beta")

		// Get ChanConfig singleton
		// Send job over JobChan channel to worker
		cc := lxAudit.GetChanConfig()
		// ErrChan for testing
		cc.ErrChan = make(chan error)
		go func() {
			cc.JobChan <- testEntries
		}()

		// Wait err
		t.Log(<-cc.ErrChan)

		//errChan := <- cc.ErrChan
		//t.Log(errChan)
		// Wait of sig
		//its.NoError(<-chanErr)

		// Kill the worker
		//cc.KillChan <- true
		//time.Sleep(time.Duration(time.Second * 5))

		// Wait of sig
		//its.NoError(<-chanErr)
	})
	//t.Run("send entries to worker with != 200 error", func(t *testing.T) {
	//	// Handler for test
	//	handler := func(w http.ResponseWriter, r *http.Request){
	//		w.WriteHeader(http.StatusUnprocessableEntity)
	//	}
	//
	//	server := httptest.NewServer(http.HandlerFunc(handler))
	//	defer server.Close()
	//
	//	// test logger for test with json output by error
	//	testLogger, hook := test.NewNullLogger()
	//
	//	// InitJobConfig with test server for request
	//	lxAudit.InitJobConfig(
	//		"test-host",
	//		server.URL,
	//		"",
	//		testLogger.WithFields(logrus.Fields{}),
	//		time.Millisecond * 1)
	//
	//	// Channel error for waiting by testing
	//	chanErr := make(chan error)
	//
	//	// StartWorker creates the ChanConfig singleton
	//	lxAudit.StartWorker(chanErr)
	//
	//	// Get ChanConfig singleton
	//	// Send job over JobChan channel to worker
	//	cc := lxAudit.GetChanConfig()
	//	go func() {
	//		cc.JobChan <- testEntries
	//	}()
	//
	//	// Wait of sig and check error
	//	err := <-chanErr
	//	its.Error(err)
	//	its.True(errors.Is(err, lxAudit.ErrStatus))
	//
	//	// Analyse last entry data field in logger
	//	// Convert data field to []lxAudit.AuditEntry for comparison
	//	var resultEntries []lxAudit.AuditEntry
	//	v, ok := hook.LastEntry().Data["job"].([]byte)
	//	its.True(ok)
	//	err = json.Unmarshal(v, &resultEntries)
	//	its.NoError(err)
	//
	//	// Compare result with transferred data
	//	its.Equal(testEntries, resultEntries)
	//
	//	// Kill the worker
	//	cc.KillChan <- true
	//
	//	// Wait of sig
	//	its.NoError(<-chanErr)
	//})
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
