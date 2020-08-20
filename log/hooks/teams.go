package lxLogHooks

import (
	"errors"
	"fmt"
	lxHelper "github.com/litixsoft/lxgo/helper"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
	//"time"
)

const (
	TeamsErrorColor = "#CC4A31"
	TeamsErrorImage = "https://test-medishop.lxcloud.de/error.png"
	TeamsWarnColor  = "#CCCC00"
	TeamsWarnImage  = "https://test-medishop.lxcloud.de/warning.png"
	TeamsInfoColor  = "#6264A7"
	TeamsInfoImage  = "https://test-medishop.lxcloud.de/info.png"
)

// Hook is a hook that writes logs of specified LogLevels to specified Writer
// Example:
// 	log.AddHook(&lxLogHooks.TeamsHook{
//		LogLevels:  []logrus.Level{logrus.ErrorLevel},
//		LogsUri:    LogsUri,
//		HookUri:    WebHookBaseUrl + WebHookPathUrl,
//		ReleaseName:   "echo-test-backend-prod",
//	})
type TeamsHook struct {
	LogLevels   []logrus.Level
	LogsUri     string
	HookUri     string
	ReleaseName string
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *TeamsHook) Fire(entry *logrus.Entry) error {
	// Check, if nil than error as default
	if hook.LogLevels == nil {
		hook.LogLevels = []logrus.Level{logrus.ErrorLevel}
	}

	if hook.LogsUri == "" {
		return errors.New("LogsUri is empty")
	}

	if hook.HookUri == "" {
		return errors.New("HookUri is empty")
	}

	if hook.ReleaseName == "" {
		return errors.New("ReleaseName is empty")
	}

	var title string
	var sections []lxHelper.M

	text := ""
	color := TeamsInfoColor
	image := TeamsInfoImage

	// Check index for stack
	titleIdx := strings.Index(entry.Message, "\n")

	// Create title and text for message
	if titleIdx > 0 {
		title = entry.Message[0:titleIdx]
		text = entry.Message[(titleIdx + 1):]
	} else {
		title = entry.Message
	}

	// Check http error
	// Find 3 keys that every http error has
	chk := 0
	if _, ok := entry.Data["user_agent"]; ok {
		chk++
	}
	if _, ok := entry.Data["remote_ip"]; ok {
		chk++
	}
	if _, ok := entry.Data["latency_human"]; ok {
		chk++
	}
	isHttpErr := chk == 3

	// Add all fields when not http error
	if !isHttpErr {
		for k, v := range entry.Data {

			sections = append(sections, lxHelper.M{
				"name":  k,
				"value": v,
			})
		}
	}

	// Check level for message color
	switch entry.Level.String() {
	case "warning":
		color = TeamsWarnColor
		image = TeamsWarnImage
	case "error":
		color = TeamsErrorColor
		image = TeamsErrorImage
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	// Create body for post request
	body := lxHelper.M{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": color,
		"summary":    "Error",
		"sections": []interface{}{
			lxHelper.M{
				"activityTitle":    fmt.Sprintf("%s<br/>%s", hook.ReleaseName, title),
				"activitySubtitle": text,
				"activityImage":    image,
				"facts":            sections,
				"markdown":         true,
			},
		},
		"potentialAction": []interface{}{
			lxHelper.M{
				"@type": "OpenUri",
				"name":  "View in logs",
				"targets": []lxHelper.M{
					{"os": "default", "uri": hook.LogsUri},
				},
			},
		},
	}

	// Http request to Teams
	resp, err := lxHelper.Request(header, body, hook.HookUri, http.MethodPost, 30*time.Second)
	if err != nil {
		return err
	}

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("error with code: %d", resp.StatusCode))
	}

	return nil
}

// Levels define on which log levels this hook would trigger
func (hook *TeamsHook) Levels() []logrus.Level {
	return hook.LogLevels
}
