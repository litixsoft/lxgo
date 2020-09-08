package lxAudit_test

import "testing"

func TestNewAudit(t *testing.T) {
	t.Log("sadd")
}

//func TestStartWorker(t *testing.T) {
//	//its := assert.New(t)
//
//	// /dev/null for test
//	lxLog.InitLogger(
//		os.Stdout,
//		"debug",
//		"text")
//
//	// For testing log output by error,
//	// user and data with map[string]interface{}
//	var testEntries []lxAudit.AuditEntry
//	for i := 0; i < 3; i++ {
//		testEntries = append(testEntries, lxAudit.AuditEntry{
//			Host:       fmt.Sprintf("test_host_%d", i),
//			Collection: "",
//			Action:     lxAudit.Insert,
//			User:       map[string]interface{}{"name": fmt.Sprintf("TestUser_%d", i)},
//			Data:       map[string]interface{}{"location": fmt.Sprintf("Fab_%d", i)},
//		})
//	}
//
//	t.Run("send entries to worker", func(t *testing.T) {
//		// Handler for test
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(handler))
//		defer server.Close()
//
//		audit := lxAudit.NewAudit(
//			"test-host",
//			server.URL,
//			"",
//			lxLog.GetLogger().WithFields(logrus.Fields{}))
////
////		// InitJobConfig with test server for request
////		lxAudit.InitJobConfig(
////			"test-host",
////			server.URL,
////			"",
////			lxLog.GetLogger().WithFields(logrus.Fields{}),
////			10*time.Millisecond)
////
////		// Get ChanConfig singleton
////		cc := lxAudit.GetChanConfig()
////
//		// StartWorker with ErrChan param for testing
//		numWorker := 2
//		for i := 0; i < numWorker; i++ {
//			audit.StartWorker(audit.JobChan, audit.KillChan, audit.ErrChan)
//		}
////
////		// Send jobs over JobChan channel to worker
////		numJobs := 3
////		for i := 0; i < numJobs; i++ {
////			go func() {
////				cc.JobChan <- testEntries
////			}()
////		}
////
////		// Wait of all ErrChan
////		for i := 0; i < numJobs; i++ {
////			its.NoError(<-cc.ErrChan)
////		}
////
////		// Stop the worker
//		close(audit.KillChan)
////		important is a singleton channel not be closed
////		for i := 0; i < numWorker; i++ {
////			cc.KillChan <- true
////		}
//		time.Sleep(time.Duration(time.Millisecond * 10))
//	})
////	t.Run("send entries to worker with != 200 error", func(t *testing.T) {
////		// Handler for test
////		handler := func(w http.ResponseWriter, r *http.Request) {
////			w.WriteHeader(http.StatusUnprocessableEntity)
////		}
////
////		server := httptest.NewServer(http.HandlerFunc(handler))
////		defer server.Close()
////
////		// test logger for test with json output by error
////		testLogger, hook := test.NewNullLogger()
////
////		// InitJobConfig with test server for request
////		lxAudit.InitJobConfig(
////			"test-host",
////			server.URL,
////			"",
////			testLogger.WithFields(logrus.Fields{}))
////
////		// Get ChanConfig singleton
////		cc := lxAudit.GetChanConfig()
////
////		// StartWorker with ErrChan param for testing
////		lxAudit.StartWorker(cc.JobChan, cc.KillChan, cc.ErrChan)
////
////		// Send sync job over JobChan channel to worker
////		cc.JobChan <- testEntries
////
////		// Wait of sig from errChan and check error
////		err := <-cc.ErrChan
////		its.Error(err)
////		its.True(errors.Is(err, lxAudit.ErrStatus))
////
////		its.True(true)
////		t.Log(hook)
////
////		// Analyse last entry data field in logger
////		// Convert data field to []lxAudit.AuditEntry for comparison
////		var resultEntries []lxAudit.AuditEntry
////		v, ok := hook.LastEntry().Data["job"].([]byte)
////		its.True(ok)
////		err = json.Unmarshal(v, &resultEntries)
////		its.NoError(err)
////
////		// Compare result with transferred data
////		its.Equal(testEntries, resultEntries)
////
////		// Stop the worker
////		// important is a singleton channel not be closed
////		cc.KillChan <- true
////		time.Sleep(time.Duration(time.Millisecond * 10))
////	})
//}

//func TestPanicsByGetChanConfig(t *testing.T) {
//	t.Run("Test GetChanConfig without InitJobConfig", func(t *testing.T) {
//		//  should be panic
//		assert.Panics(t, func() {
//			lxAudit.GetChanConfig()
//		})
//	})
//	t.Run("Test start worker without GetChanConfig", func(t *testing.T) {
//		//  should be panic
//		assert.Panics(t, func() {
//			lxAudit.StartWorker(nil, nil)
//		})
//	})
//}

//func TestRunningWorkers(t *testing.T) {
//	its := assert.New(t)
//
//	// /dev/null for test
//	lxLog.InitLogger(
//		ioutil.Discard,
//		"debug",
//		"text")
//
//	t.Run("without started worker", func(t *testing.T) {
//		its.Equal(0, lxAudit.RunningWorkers())
//	})
//	t.Run("with started workers", func(t *testing.T) {
//		// InitJobConfig with test server for request
//		lxAudit.InitJobConfig(
//			"test-host",
//			"http://foo.de",
//			"",
//			lxLog.GetLogger().WithFields(logrus.Fields{}),
//			time.Millisecond*1)
//
//		// Check running workers, should be 0
//		its.Equal(0, lxAudit.RunningWorkers())
//
//		cc := lxAudit.GetChanConfig()
//		numWorker := 3
//		for i := 0; i < numWorker; i++ {
//			lxAudit.StartWorker(cc.JobChan, cc.KillChan)
//		}
//
//		// Check running workers after start, should be 3
//		its.Equal(numWorker, lxAudit.RunningWorkers())
//
//		// Stop the worker
//		// Important: a singleton channel don't be closed
//		for i := 0; i < numWorker; i++ {
//			cc.KillChan <- true
//		}
//		time.Sleep(time.Duration(time.Millisecond * 1))
//
//		// Check running workers after start, should be 3
//		its.Equal(0, lxAudit.RunningWorkers())
//	})
//}

//func TestRequestAudit(t *testing.T) {
//	its := assert.New(t)
//
//	var testEntries []lxAudit.AuditEntry
//	for i := 0; i < 3; i++ {
//		testEntries = append(testEntries, lxAudit.AuditEntry{
//			Host:       fmt.Sprintf("test_host_%d", i),
//			Collection: "",
//			Action:     lxAudit.Insert,
//			User:       lxHelper.M{"name": fmt.Sprintf("TestUser_%d", i)},
//			Data:       lxHelper.M{"location": fmt.Sprintf("Fab_%d", i)},
//		})
//	}
//
//	t.Run("test entries", func(t *testing.T) {
//		// Handler for test
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(handler))
//		defer server.Close()
//
//		err := lxAudit.RequestAudit(testEntries, "", server.URL, "", time.Second*5)
//		its.NoError(err)
//	})
//	t.Run("test entry", func(t *testing.T) {
//		// Handler for test
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(handler))
//		defer server.Close()
//
//		err := lxAudit.RequestAudit(testEntries[1], "", server.URL, "")
//		its.NoError(err)
//	})
//	t.Run("test error != 200", func(t *testing.T) {
//		// Handler for test
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusUnprocessableEntity)
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(handler))
//		defer server.Close()
//
//		err := lxAudit.RequestAudit(testEntries[1], "", server.URL, "")
//		its.Error(err)
//		its.True(errors.Is(err, lxAudit.ErrStatus))
//	})
//	t.Run("test error with content", func(t *testing.T) {
//		b, err := json.Marshal([]lxHelper.M{{"email": "invalid"}})
//		its.NoError(err)
//
//		// Handler for test
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusUnprocessableEntity)
//			w.Header().Set("content-type", "application/json")
//			w.Write(b)
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(handler))
//		defer server.Close()
//
//		err = lxAudit.RequestAudit(testEntries[1], "", server.URL, "")
//		its.Error(err)
//		its.True(errors.Is(err, lxAudit.ErrRespContent))
//	})
//	t.Run("test error entry with wrong data type", func(t *testing.T) {
//		wrongEntry := lxHelper.M{"name": "foo"}
//
//		// Handler for test
//		handler := func(w http.ResponseWriter, r *http.Request) {
//			w.WriteHeader(http.StatusOK)
//		}
//
//		server := httptest.NewServer(http.HandlerFunc(handler))
//		defer server.Close()
//
//		err := lxAudit.RequestAudit(wrongEntry, "", server.URL, "")
//		its.Error(err)
//		its.True(errors.Is(lxAudit.ErrAuditEntryType, err))
//	})
//}
