package lxAudit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	lxHelper "github.com/litixsoft/lxgo/helper"
	"github.com/sirupsen/logrus"
	"net/http"

	//"net/http"
	"sync"
	"time"
)

type WorkerConfig struct {
	Queue        chan interface{}
	Done         chan bool
	Err          chan error
	Kill         chan bool
	clientHost   string
	auditHost    string
	auditAuthKey string
}

type AuditEntry struct {
	Collection string
	Action     string
	User       interface{}
	Data       interface{}
}

const (
	Insert         = "insert"
	Update         = "update"
	Delete         = "delete"
	DefaultTimeout = time.Second * 30
	PathLogEntry   = "/v1/log"
	PathLogEntries = "/v1/bulk/log"
)

var (
	once                   sync.Once
	hasInit                bool
	workerConfig           *WorkerConfig
	log                    *logrus.Entry
	ErrQueueAuditEntryType = errors.New("queue must be AuditEntry or []AuditEntry type")
	ErrByAuditService      = errors.New("audit service response error")
)

// GetAuditWorker, return singleton worker instance
// Usage:
//	log := lxLog.GetLogger()
//
// 	log.WithFields(logrus.Fields{
//		"omg":                   true,
//		"number":                100,
//		lxLog.StackdriverErrKey: lxLog.StackdriverErrorValue,
//	}).Error("Error without Stack and stackdriver error event!")
func GetWorkerConfig() *WorkerConfig {
	once.Do(func() {
		workerConfig = &WorkerConfig{
			Queue: make(chan interface{}),
			Done:  make(chan bool),
			Err:   make(chan error),
			Kill:  make(chan bool),
		}
	})
	return workerConfig
}

// InitAuditWorker
func InitAuditWorker(clientHost, auditHost, auditAuthKey string, logEntry *logrus.Entry, workers ...int) {
	awc := GetWorkerConfig()
	awc.clientHost = clientHost
	awc.auditHost = auditHost
	awc.auditAuthKey = auditAuthKey
	log = logEntry

	// When workers not set,
	// then default value is 1
	worker := 1
	if len(workers) > 0 {
		// workers is setting
		worker = workers[0]
	}

	// setup worker
	for i := 0; i < worker; i++ {
		go auditWorker()
	}

	hasInit = true
}

func HasAuditWorkerInit() bool {
	return hasInit
}

// auditWorker
func auditWorker() {
	for true {
		select {
		case q := <-workerConfig.Queue:
			// Convert type of queue
			switch val := q.(type) {
			default:
				log.Error(ErrQueueAuditEntryType)
				workerConfig.Err <- ErrQueueAuditEntryType
				workerConfig.Done <- true
			case AuditEntry:
				//if err := requestAuditEntry(val); err != nil {
				//
				//	workerConfig.Err <- nil
				//	workerConfig.Done <- true
				//
				//}
				time.Sleep(time.Millisecond * 100)
				workerConfig.Err <- nil
				workerConfig.Done <- true
			case []AuditEntry:
				log.Debug("doing work with AuditEntries!!", val)
				time.Sleep(time.Millisecond * 100)
				workerConfig.Err <- nil
				workerConfig.Done <- true
			}
		case <-workerConfig.Kill:
			log.Warn("shutdown worker with kill signal")
			return
		}
	}
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
func requestAuditEntry(entry AuditEntry, timeout ...time.Duration) error {
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	// Set client with timeout
	client := &http.Client{
		Timeout: to,
	}

	// Set entry for request
	jsonBody, err := json.Marshal(lxHelper.M{
		"host":       workerConfig.clientHost,
		"collection": entry.Collection,
		"action":     entry.Action,
		"user":       entry.User,
		"data":       entry.Data,
	})
	if err != nil {
		return err
	}

	// Set request
	req, err := http.NewRequest("POST", workerConfig.auditHost, bytes.NewBuffer(jsonBody))
	if err != nil {
		return err
	}

	// Set header for request
	req.Header.Set("Content-type", "application/json")
	req.Header.Set("Authorization", "Bearer "+workerConfig.auditAuthKey)

	// Start request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	// Check error
	var result lxHelper.M
	if resp.ContentLength > 0 {
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return err
		}

		return fmt.Errorf("status: %d error: %v", resp.StatusCode, result)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("status must be 200, actual: %d", resp.StatusCode)
	}

	//t.Log(result)

	// Header
	//header := http.Header{}
	//header.Add("Content-Type", "application/json")
	//header.Add("Authorization", "Bearer "+workerConfig.auditAuthKey)

	// Uri
	//uri := workerConfig.auditHost + PathLogEntry

	// Create a Resty Client
	//client := resty.New()
	//
	//ctx, cancel := context.WithTimeout(context.Background(), to)
	//defer cancel()
	//resp, err := client.R().
	//	SetContext(ctx).
	//	SetHeader("Content-Type", "application/json").
	//	SetAuthToken(workerConfig.auditAuthKey).
	//	SetBody(lxHelper.M{
	//		"host":       workerConfig.clientHost,
	//		"collection": entry.Collection,
	//		"action":     entry.Action,
	//		//"user":       entry.User,
	//		"data": entry.Data,
	//	}).
	//	Post(workerConfig.auditHost + PathLogEntry)
	//
	//if resp.StatusCode() != http.StatusOK {
	//
	//}
	//
	//log.Println("err:", err)
	//log.Println(resp.Status())
	//log.Println(resp.StatusCode())
	//log.Println(string(resp.Body()))
	//
	//log.Println("resp:", resp)
	//log.Printf("%T", resp)

	//// Request
	//resp, err := lxHelper.Request(header, body, uri, "POST", to)
	//if err != nil {
	//	return err
	//}
	//
	//// Check status
	//if resp.StatusCode != http.StatusOK {
	//	var ret bson.M
	//	if err := lxHelper.ReadResponseBody(resp, &ret); err != nil {
	//		return err
	//	}
	//
	//	// Check message exits
	//	if msg, ok := ret["message"]; ok {
	//		return NewAuditLogEntryError(resp.StatusCode, msg)
	//	}
	//
	//	return NewAuditLogEntryError(resp.StatusCode)
	//}

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
