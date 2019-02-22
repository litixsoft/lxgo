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
	Log(dbHost, dbName, collectionName, action string, user, data interface{}) error
}

// audit struct
type audit struct {
	client       *http.Client
	auditHost    string
	auditAuthKey string
	hostName     string
}

var (
	// mux for instance lock
	auditMux = new(sync.Mutex)

	// audit instance
	auditInstance *audit
)

// InitAuditInstance, set instance for audit
func InitAuditInstance(client *http.Client, auditHost, auditAuthKey, hostName string) {
	auditMux.Lock()
	auditInstance = &audit{
		client:       client,
		auditHost:    auditHost,
		auditAuthKey: auditAuthKey,
		hostName:     hostName,
	}
	auditMux.Unlock()
}

// GetAuditInstance
func GetAuditInstance() IAudit {
	return auditInstance
}

// Log, send post request to audit service
func (a *audit) Log(dbHost, dbName, collectionName, action string, user, data interface{}) error {
	// Set entry for request
	entry := lxHelper.M{
		"host":       a.hostName,
		"db_host":    dbHost,
		"db":         dbName,
		"collection": collectionName,
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
	req, err := http.NewRequest("POST", auditInstance.auditHost+"/log", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// set header
	req.Header.Add("Authorization", "Bearer "+auditInstance.auditAuthKey)
	req.Header.Add("Content-Type", "application/json")

	// send request
	resp, err := auditInstance.client.Do(req)
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
