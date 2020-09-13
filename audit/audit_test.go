package lxAudit_test

import (
	"encoding/json"
	"errors"
	"fmt"
	lxAudit "github.com/litixsoft/lxgo/audit"
	lxHelper "github.com/litixsoft/lxgo/helper"
	lxLog "github.com/litixsoft/lxgo/log"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func getTestEntries() lxAudit.AuditEntries {
	// For testing log output by error,
	// user and data with map[string]interface{}
	var testEntries lxAudit.AuditEntries
	for i := 0; i < 3; i++ {
		testEntries = append(testEntries, lxAudit.AuditEntry{
			Host:       fmt.Sprintf("test_host_%d", i),
			Collection: "test_collection",
			Action:     lxAudit.Insert,
			User:       map[string]interface{}{"name": fmt.Sprintf("TestUser_%d", i)},
			Data:       map[string]interface{}{"location": fmt.Sprintf("Fab_%d", i)},
		})
	}

	return testEntries
}

func TestQueue_StartWorker(t *testing.T) {
	its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		ioutil.Discard,
		"debug",
		"text")

	testEntries := getTestEntries()

	t.Run("send_entries", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		queue := lxAudit.NewQueue(
			"test-host",
			server.URL,
			"",
			lxLog.GetLogger().WithFields(logrus.Fields{}),
			time.Millisecond*25)

		// StartWorker with ErrChan param for testing
		numWorker := 2
		for i := 0; i < numWorker; i++ {
			queue.StartWorker(queue.JobChan, queue.KillChan, queue.ErrChan)
		}

		// Send jobs over JobChan channel to worker
		numJobs := 3
		for i := 0; i < numJobs; i++ {
			go func() {
				queue.JobChan <- testEntries
			}()
		}

		// Wait of all ErrChan
		for i := 0; i < numJobs; i++ {
			its.NoError(<-queue.ErrChan)
		}

		// Stop the worker
		close(queue.KillChan)
		time.Sleep(time.Duration(time.Millisecond * 125))
	})
	t.Run("send_entries_error_!=_200", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		// test logger for test with json output by error
		testLogger, hook := test.NewNullLogger()

		queue := lxAudit.NewQueue(
			"test-host",
			server.URL,
			"",
			testLogger.WithFields(logrus.Fields{}))

		// StartWorker with ErrChan param for testing
		queue.StartWorker(queue.JobChan, queue.KillChan, queue.ErrChan)
		// Send sync job over JobChan channel to worker
		queue.JobChan <- testEntries

		// Wait of sig from errChan and check error
		err := <-queue.ErrChan
		its.Error(err)
		its.True(errors.Is(err, lxAudit.ErrStatus))

		// Analyse last entry data field in logger
		// Convert data field to []lxAudit.AuditEntry for comparison
		var resultEntries lxAudit.AuditEntries
		v, ok := hook.LastEntry().Data["audit"].(string)
		its.True(ok)
		err = json.Unmarshal([]byte(v), &resultEntries)
		its.NoError(err)

		// Compare result with transferred data
		its.Equal(testEntries, resultEntries)

		// Stop the worker
		close(queue.KillChan)
		time.Sleep(time.Duration(time.Millisecond * 125))
	})
}

func TestQueue_GetCountOfRunningWorkers(t *testing.T) {
	its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		ioutil.Discard,
		"debug",
		"text")

	t.Run("check running workers", func(t *testing.T) {
		queue := lxAudit.NewQueue(
			"test-host",
			"test",
			"",
			lxLog.GetLogger().WithFields(logrus.Fields{}))

		// Check running workers, should be 0
		its.Equal(0, queue.GetCountOfRunningWorkers())

		// Start worker min. 3
		numWorker := 3
		for i := 0; i < numWorker; i++ {
			queue.StartWorker(queue.JobChan, queue.KillChan)
		}

		// Check running workers after start, should be 3
		its.Equal(numWorker, queue.GetCountOfRunningWorkers())

		// Stop worker individually
		stop := 2
		for i := 0; i < stop; i++ {
			queue.KillChan <- true
		}

		time.Sleep(time.Duration(time.Millisecond * 125))

		// Check running workers after individual stop
		its.Equal(numWorker-stop, queue.GetCountOfRunningWorkers())

		// Close channel and stop all
		close(queue.KillChan)

		time.Sleep(time.Duration(time.Millisecond * 125))

		// Check running workers after close channel
		its.Equal(0, queue.GetCountOfRunningWorkers())

	})
}

func TestQueue_IsActive(t *testing.T) {
	its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		ioutil.Discard,
		"debug",
		"text")

	t.Run("check running workers", func(t *testing.T) {
		queue := lxAudit.NewQueue(
			"test-host",
			"test",
			"",
			lxLog.GetLogger().WithFields(logrus.Fields{}))

		// Check active, should be false
		its.False(queue.IsActive())

		// Start worker
		for i := 0; i < 2; i++ {
			queue.StartWorker(queue.JobChan, queue.KillChan)
		}

		// Check active after start, should be true
		its.True(queue.IsActive())

		// Close channel and stop all
		close(queue.KillChan)

		time.Sleep(time.Duration(time.Millisecond * 125))

		// Check active after close channel
		its.False(queue.IsActive())
	})
}

func TestQueue_Send(t *testing.T) {
	its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		ioutil.Discard,
		"debug",
		"text")

	var testEntries []bson.M
	for i := 0; i < 3; i++ {
		testEntries = append(testEntries, bson.M{
			"host":       fmt.Sprintf("test_host_%d", i),
			"collection": "",
			"action":     "insert",
			"user":       map[string]interface{}{"name": fmt.Sprintf("TestUser_%d", i)},
			"data":       map[string]interface{}{"location": fmt.Sprintf("Fab_%d", i)},
		})
	}

	// Handler for test
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	queue := lxAudit.NewQueue(
		"test-host",
		server.URL,
		"",
		lxLog.GetLogger().WithFields(logrus.Fields{}))

	// StartWorker with ErrChan param for testing
	queue.StartWorker(queue.JobChan, queue.KillChan, queue.ErrChan)

	t.Run("entries", func(t *testing.T) {
		// Testing with error for check transport
		queue.Send(testEntries)

		// Wait for error over the error channel
		err := <-queue.ErrChan
		its.Error(err)
		its.True(errors.Is(err, lxAudit.ErrStatus))
	})
	t.Run("entry", func(t *testing.T) {
		// Testing with error for check transport
		queue.Send(testEntries[1])

		// Wait for error over the error channel
		err := <-queue.ErrChan
		its.Error(err)
		its.True(errors.Is(err, lxAudit.ErrStatus))
	})

	// Stop the worker
	close(queue.KillChan)
	time.Sleep(time.Duration(time.Millisecond * 125))
}

func TestRequestAudit(t *testing.T) {
	its := assert.New(t)

	testEntries := getTestEntries()

	t.Run("test entries", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(testEntries, server.URL, "", time.Second*5)
		its.NoError(err)
	})
	t.Run("test entry", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(testEntries[1], server.URL, "")
		its.NoError(err)
	})
	t.Run("test error != 200", func(t *testing.T) {
		// Handler for test
		handler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnprocessableEntity)
		}

		server := httptest.NewServer(http.HandlerFunc(handler))
		defer server.Close()

		err := lxAudit.RequestAudit(testEntries[1], server.URL, "")
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

		err = lxAudit.RequestAudit(testEntries[1], server.URL, "")
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

		err := lxAudit.RequestAudit(wrongEntry, server.URL, "")
		its.Error(err)
		its.True(errors.Is(lxAudit.ErrAuditEntryType, err))
	})
}

/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////

// TODO TestRequestAudit2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestRequestAudit2(t *testing.T) {
//	its := assert.New(t)
//
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	var testEntries lxAudit.AuditEntries
//	for i := 0; i < 3; i++ {
//		testEntries = append(testEntries, lxAudit.AuditEntry{
//			Host:       fmt.Sprintf("test_host_%d", i),
//			Collection: "TestCollection",
//			Action:     lxAudit.Insert,
//			User:       lxHelper.M{"name": fmt.Sprintf("TestUser_%d", i)},
//			Data:       lxHelper.M{"location": fmt.Sprintf("Fab_%d", i)},
//		})
//	}
//	t.Run("entries", func(t *testing.T) {
//
//		err := lxAudit.RequestAudit(testEntries, auditHost, authKey)
//		its.NoError(err)
//	})
//	t.Run("entry", func(t *testing.T) {
//
//		err := lxAudit.RequestAudit(testEntries[1], auditHost, authKey)
//		its.NoError(err)
//	})
//}

// TODO TestQueue_StartWorker2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestQueue_StartWorker2(t *testing.T) {
//	its := assert.New(t)
//
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	// /dev/null for test
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	// TestData
//	testEntries := getTestEntries()
//
//	queue := lxAudit.NewQueue(
//		"test-host",
//		auditHost,
//		authKey,
//		lxLog.GetLogger().WithFields(logrus.Fields{"client":"tester"}))
//
//	// StartWorker with ErrChan param for testing
//	numWorker := 1
//	for i := 0; i < numWorker; i++ {
//		queue.StartWorker(queue.JobChan, queue.KillChan, queue.ErrChan)
//	}
//
//	t.Run("send_entries", func(t *testing.T) {
//		go func() {
//			queue.JobChan <- testEntries
//		}()
//
//		// Wait of all ErrChan
//		for i := 0; i < 1; i++ {
//			its.NoError(<-queue.ErrChan)
//		}
//	})
//	t.Run("send_entry", func(t *testing.T) {
//		go func() {
//			queue.JobChan <- testEntries[1]
//		}()
//
//		// Wait of all ErrChan
//		for i := 0; i < 1; i++ {
//			its.NoError(<-queue.ErrChan)
//		}
//	})
//
//	// Stop the worker
//	close(queue.KillChan)
//	time.Sleep(time.Duration(time.Millisecond * 125))
//
//}

// TODO TestQueue_Send2
// TODO manual real test without mocks for check audit entry
// TODO Important, always comment out and start manually
// TODO comment in again after the test

//func TestQueue_Send2(t *testing.T) {
//	its := assert.New(t)
//
//	auditHost := "http://localhost:3000"
//	// Todo important, set key only local for test, delete in audit after test
//	authKey := "69b85f72-4568-483a-8f24-fc7a140a46ad"
//
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	var testEntries []bson.M
//	for i := 0; i < 3; i++ {
//		testEntries = append(testEntries, bson.M{
//			"host":       fmt.Sprintf("test_host_%d", i),
//			"collection": "test_collection",
//			"action":     "insert",
//			"user":       map[string]interface{}{"name": fmt.Sprintf("TestUser_%d", i)},
//			"data":       map[string]interface{}{"location": fmt.Sprintf("Fab_%d", i)},
//		})
//	}
//
//	queue := lxAudit.NewQueue(
//		"test-host",
//		auditHost,
//		authKey,
//		lxLog.GetLogger().WithFields(logrus.Fields{"client": "tester"}))
//
//	// StartWorker with ErrChan param for testing
//	numWorker := 1
//	for i := 0; i < numWorker; i++ {
//		queue.StartWorker(queue.JobChan, queue.KillChan, queue.ErrChan)
//	}
//
//	t.Run("entries", func(t *testing.T) {
//		queue.Send(testEntries)
//
//		// Wait of all ErrChan
//		for i := 0; i < 1; i++ {
//			its.NoError(<-queue.ErrChan)
//		}
//	})
//	t.Run("entry", func(t *testing.T) {
//		queue.Send(testEntries[1])
//
//		// Wait of all ErrChan
//		for i := 0; i < 1; i++ {
//			its.NoError(<-queue.ErrChan)
//		}
//	})
//}
