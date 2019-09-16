package lxWebhooks_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	lxHelper "github.com/litixsoft/lxgo/helper"
	lxWebhooks "github.com/litixsoft/lxgo/webhooks"
	"github.com/stretchr/testify/assert"
)

var (
	slackPath     = "/hook/0815"
	slackTitle    = "SlackTest99"
	slackUsername = "test"
	slackIcon     = ":exclamation:"
	slackMsg      = "test message"
	slackLevel    = lxWebhooks.Warn
)

func TestSlack_SendSmall(t *testing.T) {
	// test server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), slackPath)
		body, err := ioutil.ReadAll(req.Body)
		assert.NoError(t, err)

		// convert body for check
		jsonBody := new(lxHelper.M)
		assert.NoError(t, json.Unmarshal(body, jsonBody))

		// expected map
		expected := &lxHelper.M{
			"username":   slackUsername,
			"icon_emoji": slackIcon,
			"attachments": []interface{}{map[string]interface{}{
				"fallback": slackMsg,
				"color":    "warning",
				"fields": []interface{}{map[string]interface{}{
					"title": slackTitle,
					"value": slackMsg,
					"short": false,
				}},
			}},
		}

		// check request body
		assert.Equal(t, expected, jsonBody)

		// Send http.StatusNoContent for successfully audit
		rw.WriteHeader(http.StatusOK)
	}))
	// Close the server when test finishes
	defer server.Close()

	// test ms teams
	api := &lxWebhooks.Slack{
		Client:   server.Client(),
		BaseUrl:  server.URL,
		Path:     slackPath,
		Username: slackUsername,
		Icon:     slackIcon,
	}

	// send request with client
	_, err := api.SendSmall(slackTitle, slackMsg, slackLevel)
	assert.NoError(t, err)
}
