package lxaudit

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/litixsoft/lxgo/helper"
	"io/ioutil"
	"net/http"
	"sync"
)

const (
	Insert = "insert"
	Find   = "find"
	Update = "update"
	Upsert = "upsert"
	Remove = "remove"
)

// IAudit interface for lxaudit logger
type IAudit interface {
	Log(action string, user, data interface{}) error
}

// audit config struct
type auditConfig struct {
	client       *http.Client
	auditHost    string
	auditAuthKey string
	hostName     string
}

// audit struct
type audit struct {
	dbHost         string
	dbName         string
	collectionName string
}

var (
	// mux for instance lock
	auditMux = new(sync.Mutex)

	// audit instance
	auditConfigInstance *auditConfig
)

// InitAuditInstance, set instance for auditConfig
func InitAuditConfigInstance(client *http.Client, auditHost, auditAuthKey, hostName string) {
	auditMux.Lock()
	auditConfigInstance = &auditConfig{
		client:       client,
		auditHost:    auditHost,
		auditAuthKey: auditAuthKey,
		hostName:     hostName,
	}
	auditMux.Unlock()
}

// GetAuditInstance
func GetAuditInstance(dbHost, dbName, collectionName string) IAudit {
	// check config instance
	if auditConfigInstance == nil {
		panic(errors.New("auditConfigInstance was not initialized"))
	}

	// create audit instance
	return &audit{
		dbHost:         dbHost,
		dbName:         dbName,
		collectionName: collectionName,
	}
}

// Log, send post request to audit service
func (a *audit) Log(action string, user, data interface{}) error {
	// Set entry for request
	entry := lxHelper.M{
		"host":       auditConfigInstance.hostName,
		"db_host":    a.dbHost,
		"db":         a.dbName,
		"collection": a.collectionName,
		"action":     action,
		"user":       user,
		"data":       data,
	}

	// Convert entry to json
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	// Post to url
	req, err := http.NewRequest("POST", auditConfigInstance.auditHost+"/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// set header
	req.Header.Add("Authorization", "Bearer "+auditConfigInstance.auditAuthKey)
	req.Header.Add("Content-Type", "application/json")

	// send request
	resp, err := auditConfigInstance.client.Do(req)
	if err != nil {
		return err
	}

	// check validation or internal error
	if resp.StatusCode == http.StatusUnprocessableEntity || resp.StatusCode == http.StatusInternalServerError {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if err := resp.Body.Close(); err != nil {
			return err
		}

		// response error
		var respErr error

		// check response error
		switch resp.StatusCode {
		case http.StatusInternalServerError:
			respErr = errors.New(fmt.Sprintf("Internal-Server-Error: %v", string(body)))
		case http.StatusUnprocessableEntity:
			respErr = errors.New(fmt.Sprintf("Validation-Error: %v", string(body)))
		}

		return respErr
	}

	return nil
}
