package lxWebhooks_test

import (
	"encoding/json"
	"github.com/litixsoft/lxgo/helper"
	"github.com/litixsoft/lxgo/webhooks"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMsTeamsApi_SendSmall(t *testing.T) {
	path := "/webhook/64112b37-462c-4c47"

	// test server
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// Test request parameters
		assert.Equal(t, req.URL.String(), path)

		// convert request body into map
		cReqBody := new(lxHelper.M)
		if err := json.NewDecoder(req.Body).Decode(cReqBody); err != nil {
			log.Fatal(err)
		}

		// convert request body to json byte array
		jbody, _ := json.Marshal(cReqBody)

		// Send json request body back to client
		rw.Write(jbody)
	}))
	// Close the server when test finishes
	defer server.Close()

	// test ms teams
	api := &lxWebhooks.MsTeamsApi{
		Client:  server.Client(),
		BaseUrl: server.URL,
		Path:    path,
	}

	// vars for test
	title := "Test99"
	msg := "test message"
	color := lxWebhooks.RedDark

	// send request with client
	body, err := api.SendSmall(title, msg, color)
	assert.NoError(t, err)

	// check response
	actual := new(lxHelper.M)
	assert.NoError(t, json.Unmarshal(body, actual))

	expected := &lxHelper.M{
		"@context":   "https://schema.org/extensions",
		"@type":      "MessageCard",
		"themeColor": color,
		"title":      title,
		"text":       msg,
	}

	assert.Equal(t, expected, actual)
}
