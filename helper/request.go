package lxHelper

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"
)

// Request
func Request(header http.Header, body interface{}, uri, method string, timeout time.Duration) (*http.Response, error) {
	// Request
	var err error
	req := new(http.Request)

	// Check body and set request
	if body != nil {
		// Convert reqBody to json
		jsonData, err := json.Marshal(body)
		if err != nil {
			return &http.Response{}, err
		}

		// Post to url
		req, err = http.NewRequest(method, uri, bytes.NewBuffer(jsonData))
		if err != nil {
			return &http.Response{}, err
		}
	} else {
		req, err = http.NewRequest(method, uri, nil)
		if err != nil {
			return &http.Response{}, err
		}
	}

	// Set header
	req.Header = header

	// Send request
	client := &http.Client{Timeout: timeout}

	return client.Do(req)
}

// ReadResponseBody, read and convert body
func ReadResponseBody(resp *http.Response, result interface{}) error {
	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Check respBody
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, result); err != nil {
			return err
		}
	}

	return resp.Body.Close()
}
