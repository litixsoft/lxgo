package lxAudit

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"

	//"errors"
	"fmt"
	//"github.com/sirupsen/logrus"
	"net/http"

	//"net/http"
	//"sync"
	"time"
)

// AuditEntry transport type for service
type AuditEntry struct {
	Host       string      `json:"host"`
	Collection string      `json:"collection"`
	Action     string      `json:"action"`
	User       interface{} `json:"user"`
	Data       interface{} `json:"data"`
}

// AuditEntries transport type for service
type AuditEntries []AuditEntry

type queue struct {
	clientHost       string
	auditHost        string
	auditHostAuthKey string
	log              *logrus.Entry
	throttle         time.Duration
	runningWorkers   int
	JobChan          chan interface{}
	ErrChan          chan error
	KillChan         chan bool
}

const (
	// Actions
	Insert = "insert"
	Update = "update"
	Delete = "delete"

	// Timeout for cancel request
	DefaultTimeout = time.Second * 15

	// Route paths
	PathLogEntry   = "/v1/log"
	PathLogEntries = "/v1/bulk/log"
)

// Errors
var (
	ErrAuditEntryType = errors.New("must be AuditEntry or []AuditEntry type")
	ErrRespContent    = errors.New("response shouldn't have any content")
	ErrStatus         = errors.New("status must have 200")
)

// NewQueue create instance of queue
// Example:
// queue := lxAudit.NewQueue(
// 		"test-host",
//		server.URL,
//		"<auth key for host>",
//		lxLog.GetLogger().WithFields(logrus.Fields{}),
//		10 * time.Millisecond)
// optional workerThrottle is time.Sleep after job
func NewQueue(clientHost, auditHost, auditHostAuthKey string, logEntry *logrus.Entry, workerThrottle ...time.Duration) *queue {
	// default 25 milliseconds 40 per second
	throttle := time.Millisecond * 25
	// When workerThrottle set overwrite default
	for _, v := range workerThrottle {
		throttle = v
	}

	return &queue{
		clientHost:       clientHost,
		auditHost:        auditHost,
		auditHostAuthKey: auditHostAuthKey,
		log:              logEntry,
		throttle:         throttle,
		runningWorkers:   0,
		JobChan:          make(chan interface{}),
		ErrChan:          make(chan error),
		KillChan:         make(chan bool),
	}
}

// StartWorker starts a worker
// Example:
// queue := NewQueue(....)
// queue.StartWorker(queue.JobChan, queue.KillChan)
// or for testing send a chanErr
// queue.StartWorker(queue.JobChan, queue.KillChan)
func (qu *queue) StartWorker(jobChan chan interface{}, killSig chan bool, errChan ...chan error) {
	// increment worker
	qu.runningWorkers++

	// Set _errChan, when not given than is nil
	var _errChan chan error
	for _, v := range errChan {
		_errChan = v
	}

	// go func for worker
	go func(jobChan chan interface{}, killSig chan bool, errChan chan error, workerNum int) {
		for true {
			select {
			case j := <-jobChan:
				// send the entries or entry to audit service
				err := RequestAudit(j, qu.auditHost, qu.auditHostAuthKey)
				if err != nil {
					// log error for manual insert
					jsonJob, jsonErr := json.Marshal(j)
					ctxLog := qu.log.WithField("func", "lxAudit.StartWorker")
					if jsonErr != nil {
						// error by convert job to json print in raw
						ctxLog.WithField("job", j).Error("error by RequestAudit can't convert job to json")
					} else {
						// log error as json for manual insert
						qu.log.WithField("job", jsonJob).Error("error by RequestAudit, can't send entries")
					}
				}
				// stop before end job
				time.Sleep(qu.throttle)
				// chk err and send to channel
				// important: caller has to wait for the signal
				if _errChan != nil {
					_errChan <- err
				}
			case <-killSig:
				qu.log.Infof("shutdown worker %d with kill signal", workerNum)
				qu.runningWorkers--
				return
			}
		}
	}(jobChan, killSig, _errChan, qu.runningWorkers)
}

// GetCountOfRunningWorkers shows how many workers are running
// Example:
// queue := NewQueue(....)
// queue.GetCountOfRunningWorkers()
func (qu *queue) GetCountOfRunningWorkers() int {
	return qu.runningWorkers
}

// IsActive return true when worker running,
// IBaseRepoAudit
func (qu *queue) IsActive() bool {
	return qu.GetCountOfRunningWorkers() > 0
}

// Send async convert and send elem to worker,
// elm must be bson.M or []bson.M type
// Example:
// queue := NewQueue(....)
// queue.Send(bson.M{...})
// IBaseRepoAudit
func (qu *queue) Send(elem interface{}) {
	go func() {
		switch val := elem.(type) {
		case bson.M:
			// Send entry to worker
			qu.JobChan <- AuditEntry{
				Host:       qu.clientHost,
				Collection: val["collection"].(string),
				Action:     val["action"].(string),
				User:       val["user"],
				Data:       val["data"],
			}
		case []bson.M:
			entries := make(AuditEntries, len(val))
			for i, e := range val {
				entries[i] = AuditEntry{
					Host:       qu.clientHost,
					Collection: e["collection"].(string),
					Action:     e["action"].(string),
					User:       e["user"],
					Data:       e["data"],
				}
			}
			// Send entries to worker
			qu.JobChan <- entries
		}
	}()
}

// RequestAudit send entry or entries to audit service.
// This function can also be used independently of the worker.
// Example:
// err := lxAudit.RequestAudit(...)
func RequestAudit(elem interface{}, auditHost, auditAuthKey string, timeout ...time.Duration) error {
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	// Request
	var req *http.Request

	switch val := elem.(type) {
	default:
		// Wrong type return with error
		return ErrAuditEntryType

	case AuditEntries:
		// Convert entries to json
		jsonBody, err := json.Marshal(val)
		if err != nil {
			return err
		}

		// Make request
		req, err = http.NewRequest("POST", auditHost+PathLogEntries, bytes.NewBuffer(jsonBody))
		if err != nil {
			return err
		}

	case AuditEntry:
		// Convert entry to json
		jsonBody, err := json.Marshal(val)
		if err != nil {
			return err
		}

		// Make request
		req, err = http.NewRequest("POST", auditHost+PathLogEntry, bytes.NewBuffer(jsonBody))
		if err != nil {
			return err
		}
	}

	// Set client with timeout
	client := &http.Client{
		Timeout: to,
	}

	// Set header for request
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+auditAuthKey)

	// Start request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check error
	var result interface{}
	if resp.StatusCode != http.StatusOK && resp.ContentLength > 0 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		return fmt.Errorf("status: %v result: %v, %w", resp.Status, result, ErrRespContent)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%v, %w", resp.Status, ErrStatus)
	}

	return nil
}
