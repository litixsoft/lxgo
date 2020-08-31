package lxAudit

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	lxHelper "github.com/litixsoft/lxgo/helper"
	"log"
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

var once sync.Once
var hasInit bool
var workerConfig *WorkerConfig

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
func InitAuditWorker(clientHost, auditHost, auditAuthKey string, workers ...int) *WorkerConfig {
	awc := GetWorkerConfig()
	awc.clientHost = clientHost
	awc.auditHost = auditHost
	awc.auditAuthKey = auditAuthKey

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
	return awc
}

func HasInit() bool {
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
				fmt.Println("queue must be AuditEntry or []AuditEntry")
				workerConfig.Err <- errors.New("queue type must be AuditEntry or []AuditEntry")
				workerConfig.Done <- true
			case AuditEntry:
				fmt.Println("doing work with AuditEntry!!", val)
				workerConfig.Err <- nil
				workerConfig.Done <- true
			case []AuditEntry:
				fmt.Println("doing work with AuditEntries!!", val)
				workerConfig.Err <- nil
				workerConfig.Done <- true
			}

			time.Sleep(time.Millisecond * 100)

		case <-workerConfig.Kill:
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
func LogEntry(entry AuditEntry, timeout ...time.Duration) error {
	to := DefaultTimeout
	if len(timeout) > 0 {
		to = timeout[0]
	}

	// Set entry for request
	//body := bson.M{
	//	"host":       workerConfig.clientHost,
	//	"collection": entry.Collection,
	//	"action":     entry.Action,
	//	"user":       entry.User,
	//	"data":       entry.Data,
	//}

	// Header
	//header := http.Header{}
	//header.Add("Content-Type", "application/json")
	//header.Add("Authorization", "Bearer "+workerConfig.auditAuthKey)

	// Uri
	//uri := workerConfig.auditHost + PathLogEntry

	// Create a Resty Client
	client := resty.New()

	ctx, cancel := context.WithTimeout(context.Background(), to)
	defer cancel()
	resp, err := client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetAuthToken(workerConfig.auditAuthKey).
		SetBody(lxHelper.M{
			"host":       workerConfig.clientHost,
			"collection": entry.Collection,
			"action":     entry.Action,
			//"user":       entry.User,
			"data": entry.Data,
		}).
		Post(workerConfig.auditHost + PathLogEntry)

	if resp.StatusCode() != http.StatusOK {

	}

	log.Println("err:", err)
	log.Println(resp.Status())
	log.Println(resp.StatusCode())
	log.Println(string(resp.Body()))

	log.Println("resp:", resp)
	log.Printf("%T", resp)

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
