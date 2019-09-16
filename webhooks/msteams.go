package lxWebhooks

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/litixsoft/lxgo/helper"
)

type MsTeams struct {
	Client  *http.Client
	BaseUrl string
	Path    string
}

// SendSmall, small card for Microsoft Teams
func (api *MsTeams) SendSmall(title, msg, level string) ([]byte, error) {
	color := ""

	switch level {
	case Error:
		color = "#ED1B3E"
	case Warn:
		color = "#FBC940"
	case Info:
		color = "#88BC2B"
	}

	// Set entry for request
	entry := lxHelper.M{
		"@context":   "https://schema.org/extensions",
		"@type":      "MessageCard",
		"themeColor": color,
		"title":      title,
		"text":       msg,
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
