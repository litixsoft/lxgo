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
)

func InitJobConfig(clientHost, auditHost, auditHostAuthKey string, logEntry *logrus.Entry) {
	jobConfig = &jobConfigType{
		clientHost:       clientHost,
		auditHost:        auditHost,
		auditHostAuthKey: auditHostAuthKey,
	}
	log = logEntry
}

// GetChanConfig return singleton channel config instance
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

// InitAuditWorker
//func InitAuditWorker(clientHost, auditHost, auditAuthKey string, logEntry *logrus.Entry, workers ...int) {
//	awc := GetWorkerConfig()
//	awc.clientHost = clientHost
//	awc.auditHost = auditHost
//	awc.auditAuthKey = auditAuthKey
//	log = logEntry
//
//	// When workers not set,
//	// then default value is 1
//	worker := 1
//	if len(workers) > 0 {
//		// workers is setting
//		worker = workers[0]
//	}
//
//	// setup worker
//	for i := 0; i < worker; i++ {
//		go auditWorker()
//	}
//
//	hasInit = true
//}

//func HasAuditWorkerInit() bool {
//	return hasInit
//}

// StartWorker starts a worker with singleton channel config
func StartWorker(testArgs ...interface{}) {
	go func(c *ChanConfig) {
		for true {
			select {
			case j := <-c.JobChan:
				log.Debug("start job", j)
				log.Debugf("job type:%T\n", j)
				time.Sleep(time.Millisecond * 500)
			case <-c.KillChan:
				log.Debug("shutdown worker with kill signal")
				return

				//	case q := <-workerConfig.Queue:
				//		// Convert type of queue
				//		switch val := q.(type) {
				//		default:
				//			log.Error(ErrQueueAuditEntryType)
				//			workerConfig.Err <- ErrQueueAuditEntryType
				//			workerConfig.Done <- true
				//		case AuditEntry:
				//			//if err := requestAuditEntry(val); err != nil {
				//			//
				//			//	workerConfig.Err <- nil
				//			//	workerConfig.Done <- true
				//			//
				//			//}
				//			time.Sleep(time.Millisecond * 100)
				//			workerConfig.Err <- nil
				//			workerConfig.Done <- true
				//		case []AuditEntry:
				//			log.Debug("doing work with AuditEntries!!", val)
				//			time.Sleep(time.Millisecond * 100)
				//			workerConfig.Err <- nil
				//			workerConfig.Done <- true
				//		}
				//	case <-workerConfig.Kill:
				//		log.Warn("shutdown worker with kill signal")
				//		return
			}
		}
	}(GetChanConfig())
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
