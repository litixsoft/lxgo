package lxAudit

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/sirupsen/logrus"

	//"errors"
	"fmt"
	//"github.com/sirupsen/logrus"
	"net/http"

	//"net/http"
	//"sync"
	"time"
)

// ChanConfig type for singleton channels
//type ChanConfig struct {
//	JobChan  chan interface{}
//	ErrChan  chan error
//	KillChan chan bool
//}

// jobConfigType type for for config audit service
//type auditConfig struct {
//	clientHost       string
//	auditHost        string
//	auditHostAuthKey string
//}

// AuditEntry for transport to service
type AuditEntry struct {
	Host       string      `json:"host"`
	Collection string      `json:"collection"`
	Action     string      `json:"action"`
	User       interface{} `json:"user"`
	Data       interface{} `json:"data"`
}

type IAudit interface {
	LogAudit(elem interface{})
	GetCountOfRunningWorkers() int
	StartWorker(jobChan chan interface{}, killSig chan bool, errChan ...chan error)
}

type audit struct {
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

//var (
//	once           sync.Once
//	chanConfig     *ChanConfig             // singleton channels
//	jobConfig      *jobConfigType          // configuration for the jobs
//	log            *logrus.Entry           // logger as entry for context from app
//	throttle       = time.Millisecond * 25 // default 25 milliseconds 40 per second
//	runningWorkers = 0                     // running workers
//)

// NewAudit
func NewAudit(clientHost, auditHost, auditHostAuthKey string, logEntry *logrus.Entry, workerThrottle ...time.Duration) *audit {
	// default 25 milliseconds 40 per second
	throttle := time.Millisecond * 25
	// When workerThrottle set overwrite default
	for _, v := range workerThrottle {
		throttle = v
	}

	return &audit{
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

	// Start the worker
	//for i:=0; i < workers; i++ {
	//	audit.worker(ac.jobChan, ac.killChan, ac.)
	//}

	//return &audit{
	//	clientHost:       clientHost,
	//	auditHost:        auditHost,
	//	auditHostAuthKey: auditHostAuthKey,
	//	log: logEntry,
	//}
	//
	//jobConfig = &jobConfigType{
	//	clientHost:       clientHost,
	//	auditHost:        auditHost,
	//	auditHostAuthKey: auditHostAuthKey,
	//}
	//log = logEntry
	//
	//// When workerThrottle set overwrite default
	//for _, v := range workerThrottle {
	//	throttle = v
	//}
	//
	//return &ChanConfig{
	//	JobChan:  make(chan interface{}),
	//	ErrChan:  make(chan error),
	//	KillChan: make(chan bool)}
}

// StartWorker starts a worker with singleton channel config.
// Before you can start that, an InitJobConfig must be made
// Example:
// lxAudit.StartWorker()
// or for testing send a chanErr
// lxAudit.StartWorker(chanErr)
func (ac *audit) StartWorker(jobChan chan interface{}, killSig chan bool, errChan ...chan error) {
	// increment worker
	ac.runningWorkers++

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
				err := RequestAudit(j, ac.clientHost, ac.auditHost, ac.auditHostAuthKey)
				if err != nil {
					// log error for manual insert
					jsonJob, jsonErr := json.Marshal(j)
					ctxLog := ac.log.WithField("func", "lxAudit.StartWorker")
					if jsonErr != nil {
						// error by convert job to json print in raw
						ctxLog.WithField("job", j).Error("error by RequestAudit can't convert job to json")
					} else {
						// log error as json for manual insert
						ac.log.WithField("job", jsonJob).Error("error by RequestAudit, can't send entries")
					}
				}
				// stop before end job
				time.Sleep(ac.throttle)
				// chk err and send to channel
				// important: caller has to wait for the signal
				if _errChan != nil {
					_errChan <- err
				}
			case <-killSig:
				ac.log.Infof("shutdown worker %d with kill signal", workerNum)
				ac.runningWorkers--
				return
			}
		}
	}(jobChan, killSig, _errChan, ac.runningWorkers)
}

// RunningWorkers shows how many workers are running
func (ac *audit) GetCountOfRunningWorkers() int {
	return ac.runningWorkers
}

// LogAudit
func (ac *audit) LogAudit(elem interface{}) {
	go func() {
		ac.JobChan <- elem
	}()
}

// InitJobConfig set the global vars for all jobs
// Example:
// lxAudit.InitJobConfig(
//		testClientHost,
//		"http://localhost:3000",
//		"c7a34742-0a91-5fb9-81c3-934c76f72436",
//		lxLog.GetLogger().WithFields(logrus.Fields{"client": "TestTheWest"}))
//func InitJobConfig(clientHost, auditHost, auditHostAuthKey string, logEntry *logrus.Entry, workerThrottle ...time.Duration) {
//	jobConfig = &jobConfigType{
//		clientHost:       clientHost,
//		auditHost:        auditHost,
//		auditHostAuthKey: auditHostAuthKey,
//	}
//	log = logEntry
//
//	// When workerThrottle set overwrite default
//	for _, v := range workerThrottle {
//		throttle = v
//	}
//}

// GetAuditQueueChan return singleton channel config instance
// Before you can start that, an InitJobConfig must be made
// Example for start workers:
// 		cc := lxAudit.GetChanConfig()
// 		lxAudit.StartWorker(cc.JobChan, cc.KillChan)
// 		lxAudit.StartWorker(cc.JobChan, cc.KillChan)
//
// Example for send signal to worker queue
//		cc := lxAudit.GetChanConfig()
//		go func() {
//			cc.JobChan <- lxAudit.AuditEntry{
//				Host: fmt.Sprintf("test_host_"),
//			}
//		}()
// Example for shutdown all worker
// 		cc := lxAudit.GetChanConfig()
//		close(cc.KillChan)
//func GetAuditQueueChan() *ChanConfig {
//	if jobConfig == nil || chanConfig == nil {
//		panic(errors.New("must  before NewAuditQueue"))
//	}
//	once.Do(func() {
//		chanConfig = &ChanConfig{
//
//			JobChan:  make(chan interface{}),
//			ErrChan:  make(chan error),
//			KillChan: make(chan bool),
//		}
//	})
//	return chanConfig
//}

// RequestAudit send entry or entries to audit service.
// This function can also be used independently of the worker.
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

		return fmt.Errorf("status: %v result: %v, %w", resp.Status, result, ErrRespContent)
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("%v, %w", resp.Status, ErrStatus)
	}

	return nil
}
