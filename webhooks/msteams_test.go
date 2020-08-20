package lxWebhooks_test

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/helper"
	"github.com/litixsoft/lxgo/webhooks"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

type M map[string]interface{}

var (
	path  = "/webhook/64112b37-462c-4c47"
	title = "Test99"
	msg   = "test message"
	color = lxWebhooks.Error
)

func TestMsTeams_SendSmall(t *testing.T) {
	// test server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), path)
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)

		// convert body for check
		jsonBody := new(lxHelper.M)
		assert.NoError(t, json.Unmarshal(body, jsonBody))

		// expected map
		expected := &lxHelper.M{
			"@context":   "https://schema.org/extensions",
			"@type":      "MessageCard",
			"themeColor": "#ED1B3E",
			"title":      title,
			"text":       msg,
		}

		// check request body
		assert.Equal(t, expected, jsonBody)

		// Send http.StatusNoContent for successfully audit
		rw.WriteHeader(http.StatusOK)
	}))
	// Close the server when test finishes
	defer server.Close()

	// test ms teams
	api := &lxWebhooks.MsTeams{
		Client:  server.Client(),
		BaseUrl: server.URL,
		Path:    path,
	}

	// send request with client
	_, err := api.SendSmall(title, msg, color)
	assert.NoError(t, err)
}
