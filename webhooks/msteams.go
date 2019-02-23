package lxWebhooks

import (
	"bytes"
	"encoding/json"
	"github.com/litixsoft/lxgo/helper"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	RedDark     = "#ED1B3E"
	MagentaDark = "#F236B8"
	YellowDark  = "#FBC940"
	GreenDark   = "#88BC2B"
)

type IMsTeamsApi interface {
	SendSmall(title, msg, color string) ([]byte, error)
}

type MsTeamsApi struct {
	Client  *http.Client
	BaseUrl string
	Path    string
}

// SendSmall, small card for Microsoft Teams
func (api *MsTeamsApi) SendSmall(title, msg, color string) ([]byte, error) {
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
