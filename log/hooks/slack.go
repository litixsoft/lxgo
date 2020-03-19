package lxLogHooks

import (
	"errors"
	"fmt"
	lxHelper "github.com/litixsoft/lxgo/helper"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

// Hook is a hook that writes logs of specified LogLevels to specified Writer
// Example:
// 	log.AddHook(&lxLogHooks.SlackHook{
//		LogLevels:       []logrus.Level{logrus.ErrorLevel},
//		ErrorReportsUrl: ErrorReportsUrl,
//		LogsUrl:         LogsUrl,
//		SlackBaseUrl:    WebHookBaseUrl,
//		SlackPathUrl:    WebHookPathUrl,
//		SlackUsername:   "echo-test",
//		SlackIcon:       ":ghost",
//	})
type SlackHook struct {
	LogLevels       []logrus.Level
	ErrorReportsUrl string
	LogsUrl         string
	SlackBaseUrl    string
	SlackPathUrl    string
	SlackUsername   string
	SlackIcon       string
}

// Fire will be called when some logging function is called with current hook
// It will format log entry to string and write it to appropriate writer
func (hook *SlackHook) Fire(entry *logrus.Entry) error {
	// Check, if nil than error as default
	if hook.LogLevels == nil {
		hook.LogLevels = []logrus.Level{logrus.ErrorLevel}
	}

	if hook.ErrorReportsUrl == "" {
		return errors.New("ErrorReportsUrl is empty")
	}

	if hook.LogsUrl == "" {
		return errors.New("LogsUrl is empty")
	}

	if hook.SlackBaseUrl == "" {
		return errors.New("SlackBaseUrl is empty")
	}

	if hook.SlackPathUrl == "" {
		return errors.New("SlackPathUrl is empty")
	}

	if hook.SlackUsername == "" {
		return errors.New("SlackUsername is empty")
	}

	if hook.SlackIcon == "" {
		return errors.New("SlackIcon is empty")
	}

	var title string
	var fields []lxHelper.M

	text := ""
	color := "good"

	// Check index for stack
	titleIdx := strings.Index(entry.Message, "\n")

	// Create title and text for message
	if titleIdx > 0 {
		title = entry.Message[0:titleIdx]
		text = entry.Message[(titleIdx + 1):]
	} else {
		title = entry.Message
	}

	// Add fields
	for k, v := range entry.Data {
		fields = append(fields, lxHelper.M{
			"title": k,
			"value": v,
			"short": true,
		})
	}

	// Check level for message color
	switch entry.Level.String() {
	case "warning":
		color = "warning"
	case "error":
		color = "danger"
	}

	header := http.Header{}
	header.Add("Content-Type", "application/json")

	// Create body for post request
	body := lxHelper.M{
		"username":   hook.SlackUsername,
		"icon_emoji": hook.SlackIcon,
		"attachments": []lxHelper.M{
			{
				"color":  color,
				"title":  title,
				"text":   text,
				"fields": fields,
			},
			{
				"fields": []lxHelper.M{
					{
						"title": "Logs",
						"value": fmt.Sprintf("<%s|error logs of %s>", hook.LogsUrl, hook.SlackUsername),
						"short": true,
					},
					{
						"title": "Reports",
						"value": fmt.Sprintf("<%s|global error reports>", hook.ErrorReportsUrl),
						"short": true,
					},
				},
			},
		},
	}

	// Http request to slack
	resp, err := lxHelper.Request(header, body, hook.SlackBaseUrl+hook.SlackPathUrl, http.MethodPost, 30*time.Second)
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
func (hook *SlackHook) Levels() []logrus.Level {
	return hook.LogLevels
}
