package lxWebhooks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	lxHelper "github.com/litixsoft/lxgo/helper"
)

type Slack struct {
	Client   *http.Client
	BaseUrl  string
	Path     string
	Username string
	Icon     string
}

// SendSmall, send message to slack
func (api *Slack) SendSmall(title, msg, level string) ([]byte, error) {
	color := ""

	switch level {
	case Error:
		color = "danger"
	case Warn:
		color = "warning"
	case Info:
		color = "good"
	}

	// Set entry for request
	entry := lxHelper.M{
		"username":   api.Username,
		"icon_emoji": api.Icon,
		"attachments": []lxHelper.M{{
			"fallback": msg,
			"color":    color,
			"fields": []lxHelper.M{{
				"title": title,
				"value": msg,
				"short": false,
			}},
		}},
	}

	// Convert entry to json
	jsonData, err := json.Marshal(entry)
	if err != nil {
		return nil, err
	}

	// Post to url
	response, err := api.Client.Post(api.BaseUrl+api.Path, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()
	return ioutil.ReadAll(response.Body)
}
