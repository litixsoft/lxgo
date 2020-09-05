package lxAudit

import (
	"bytes"
	"encoding/json"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"

	//"errors"
	"fmt"
	//"github.com/sirupsen/logrus"
	"net/http"

	//"net/http"
	//"sync"
	"time"
)

type ChanConfig struct {
	JobChan  chan interface{}
	KillChan chan bool
}

type jobConfigType struct {
	clientHost       string
	auditHost        string
	auditHostAuthKey string
}

type AuditEntry struct {
	Host       string      `json:"host"`
	Collection string      `json:"collection"`
	Action     string      `json:"action"`
	User       interface{} `json:"user"`
	Data       interface{} `json:"data"`
}

const (
	Insert         = "insert"
	Update         = "update"
	Delete         = "delete"
	DefaultTimeout = time.Second * 15
	PathLogEntry   = "/v1/log"
	PathLogEntries = "/v1/bulk/log"
)

var (
	once              sync.Once
	chanConfig        *ChanConfig
	jobConfig         *jobConfigType
	log               *logrus.Entry
	ErrAuditEntryType = errors.New("must be AuditEntry or []AuditEntry type")
	throttle          = time.Millisecond * 50 // Default 50 milliseconds 20 per second
)

// InitJobConfig set the global vars for all jobs
// Example:
// lxAudit.InitJobConfig(
//		testClientHost,
//		"http://localhost:3000",
//		"c7a34742-0a91-5fb9-81c3-934c76f72436",
//		lxLog.GetLogger().WithFields(logrus.Fields{"client": "TestTheWest"}))
func InitJobConfig(clientHost, auditHost, auditHostAuthKey string, logEntry *logrus.Entry, workerThrottle ...time.Duration) {
	jobConfig = &jobConfigType{
		clientHost:       clientHost,
		auditHost:        auditHost,
		auditHostAuthKey: auditHostAuthKey,
	}
	log = logEntry

	// When workerThrottle set overwrite default
	if len(workerThrottle) > 0 {
		throttle = workerThrottle[0]
	}
}

// GetChanConfig return singleton channel config instance
// Before you can start that, an InitJobConfig must be made
// Usage:
//	c := lxAudit.GetChanConfig()
// 	c.JobChan <- interface{}
func GetChanConfig() *ChanConfig {
	if jobConfig == nil {
		panic(errors.New("jobConfig is nil, InitJobConfig before GetChanConfig"))
	}
	once.Do(func() {
		chanConfig = &ChanConfig{
			JobChan:  make(chan interface{}),
			KillChan: make(chan bool),
		}
	})
	return chanConfig
}

// StartWorker starts a worker with singleton channel config.
// Before you can start that, an InitJobConfig must be made
// Example:
// lxAudit.StartWorker()
// or for testing send a chanErr
// lxAudit.StartWorker(chanErr)
func StartWorker(chanErr ...chan error) {
	// Optional param chan error for testing
	// By many params use latest param
	var _chanErr chan error
	for _, v := range chanErr {
		_chanErr = v
	}

	// Get singleton chan config
	cc := GetChanConfig()
	log.Info("start lxAudit worker")

	// go func for worker
	go func(cc *ChanConfig, _chanErr chan error) {
		for true {
			select {
			case j := <-cc.JobChan:
				// send the entries or entry to audit host
				err := RequestAudit(j, jobConfig.clientHost, jobConfig.auditHost, jobConfig.auditHostAuthKey)
				if err != nil {
					// log error for manual insert
					jsonJob, jsonErr := json.Marshal(j)
					ctxLog := log.WithField("func", "lxAudit.StartWorker")
					if jsonErr != nil {
						// error by convert job to json print in raw
						ctxLog.WithField("job", j).Error("error by RequestAudit can't convert job to json")
					} else {
						// log error as json for manual insert
						log.WithField("job", jsonJob).Error("error by RequestAudit, can't send entries")
					}
				}
				// stop before end job
				time.Sleep(throttle)
				// chk err and send to channel
				// important: caller has to wait for the signal
				if _chanErr != nil {
					_chanErr <- err
				}
			case <-cc.KillChan:
				log.Info("shutdown worker with kill signal")
				return
			}
		}
	}(cc, _chanErr)
}

//
//import (
//	"errors"
//	"fmt"
//	lxHelper "github.com/litixsoft/lxgo/helper"
//	"go.mongodb.org/mongo-driver/bson"
//	"net/http"
//	"time"
//)
//
//const (
//	Insert         = "insert"
//	Update         = "update"
//	Delete         = "delete"
//	DefaultTimeout = time.Second * 30
//	PathLogEntry   = "/v1/log"
//	PathLogEntries = "/v1/bulk/log"
//)
//
//// IAudit interface for audit logger
//type IAudit interface {
//	LogEntry(action string, user, data interface{}, timeout ...time.Duration) error
//	LogEntries(entries []interface{}, timeout ...time.Duration) error
//	GetWorkerChannels() auditWorkerChannels
//}
//
//type audit struct {
//	clientHost     string
//	collectionName string
//	auditHost      string
//	auditAuthKey   string
//	auditWorkerChannels auditWorkerChannels
//}
//
//func NewAudit(clientHost, collectionName, auditHost, auditAuthKey string, channels auditWorkerChannels) IAudit {
//	return &audit{
//		clientHost:     clientHost,
//		collectionName: collectionName,
//		auditHost:      auditHost,
//		auditAuthKey:   auditAuthKey,
//		auditWorkerChannels: channels,
//	}
//}
//
////////////////////////////////////
//type AuditEntry struct {
//	Action string
//	User   interface{}
//	Data   interface{}
//}
//
//type auditWorkerChannels struct {
//	Queue chan interface{}
//	Done  chan bool
//	Err   chan error
//	Kill  chan bool
//}
//
//func (al *audit)GetWorkerChannels() auditWorkerChannels {
//	return al.auditWorkerChannels
//}
//
//func NewAuditWorkerChannels() auditWorkerChannels {
//	return auditWorkerChannels{
//		Queue: make(chan interface{}),
//		Done:  make(chan bool),
//		Err:   make(chan error),
//		Kill:  make(chan bool),
//	}
//}
//
//func AuditWorker(channels auditWorkerChannels) {
//	for true {
//		select {
//		case q := <-channels.Queue:
//			// Convert type of queue
//			switch val := q.(type) {
//			default:
//				fmt.Println("queue must be AuditEntry or []AuditEntry")
//				channels.Err <- errors.New("queue type must be AuditEntry or []AuditEntry")
//				channels.Done <- true
//			case AuditEntry:
//				fmt.Println("doing work with AuditEntry!!", val)
//				channels.Err <- nil
//				channels.Done <- true
//			case []AuditEntry:
//				fmt.Println("doing work with AuditEntries!!", val)
//				channels.Err <- nil
//				channels.Done <- true
//			}
//
//			time.Sleep(time.Millisecond * 100)
//
//		case <-channels.Kill:
//			return
//		}
//	}
//}
//
//
///////////////////////////////////

// Log, send post request to audit service
func RequestAudit(auditEntries interface{}, clientHost, auditHost, auditAuthKey string, timeout ...time.Duration) error {
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	// Request
	var req *http.Request

	switch val := auditEntries.(type) {
	default:
		// Wrong type return with error
		return ErrAuditEntryType

	case []AuditEntry:
		// Add host to entries
		for i := range val {
			val[i].Host = clientHost
		}

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
		// Add host to entry
		val.Host = clientHost

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

		return fmt.Errorf("status: %d result: %v", resp.StatusCode, result)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("status must be 200, actual: %d", resp.StatusCode)
	}

	return nil
}

//// LogEntries, send post request with entries to audit service
//func (al *audit) LogEntries(entries []interface{}, timeout ...time.Duration) error {
//	to := DefaultTimeout
//	if len(timeout) > 0 {
//		to = timeout[0]
//	}
//
//	for _, entry := range entries {
//		if val, ok := entry.(bson.M); ok {
//			val["host"] = al.clientHost
//			val["collection"] = al.collectionName
//		}
//	}
//
//	// Header
//	header := http.Header{}
//	header.Add("Content-Type", "application/json")
//	header.Add("Authorization", "Bearer "+al.auditAuthKey)
//
//	// Uri
//	uri := al.auditHost + PathLogEntries
//
//	// Request
//	resp, err := lxHelper.Request(header, entries, uri, "POST", to)
//	if err != nil {
//		return err
//	}
//
//	// Check status
//	if resp.StatusCode != http.StatusOK {
//		var ret bson.M
//		if err := lxHelper.ReadResponseBody(resp, &ret); err != nil {
//			return err
//		}
//
//		// Check message exits
//		if msg, ok := ret["message"]; ok {
//			return NewAuditLogEntryError(resp.StatusCode, msg)
//		}
//
//		return NewAuditLogEntryError(resp.StatusCode)
//	}
//
//	return nil
//}
