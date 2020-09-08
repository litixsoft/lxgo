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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func getTestEntries() []lxAudit.AuditEntry {
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

	t.Run("send entries to worker", func(t *testing.T) {
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
	t.Run("send entries to worker with != 200 error", func(t *testing.T) {
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
		var resultEntries []lxAudit.AuditEntry
		v, ok := hook.LastEntry().Data["job"].([]byte)
		its.True(ok)
		err = json.Unmarshal(v, &resultEntries)
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

func TestQueue_LogAudit(t *testing.T) {
	its := assert.New(t)

	// /dev/null for test
	lxLog.InitLogger(
		ioutil.Discard,
		"debug",
		"text")

	testEntries := getTestEntries()

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

	// Testing with error for check transport
	queue.Send(testEntries)

	// Wait for error over the error channel
	err := <-queue.ErrChan
	its.Error(err)
	its.True(errors.Is(err, lxAudit.ErrStatus))
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
