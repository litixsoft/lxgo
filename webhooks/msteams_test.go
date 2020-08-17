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

//func TestMsTeams_Send(t *testing.T) {
//
//
//	uri := "https://outlook.office.com/webhook/60583ba8-c5ce-4430-bbd3-2fa334fae87d@6c6c46b4-fb1d-475e-8011-684739c7ca7e/IncomingWebhook/fc551013af6e4aec959df574b58be45d/249c6ecb-24f4-45ee-a889-436388713e0a"
//
//	// Set entry for request
//	//entry := lxHelper.M{
//	//	"@context":   "https://schema.org/extensions",
//	//	"@type":      "MessageCard",
//	//	"themeColor": "0076D7",
//	//	"summary": "Larry Bryant created a new task",
//	//	"sections": []lxHelper.M{
//	//		{
//	//			"activityTitle": "Larry Bryant created a new task",
//	//			"activitySubtitle": "On Project Tango",
//	//			"activityImage": "https://teamsnodesample.azurewebsites.net/static/img/image5.png",
//	//			"markdown": true,
//	//			"facts": []lxHelper.M{
//	//				{
//	//					"name": "Assigned to",
//	//					"value": "Unassigned",
//	//				},
//	//				{
//	//					"name": "Due date",
//	//					"value": "Mon May 01 2017 17:07:18 GMT-0700 (Pacific Daylight Time)",
//	//				},
//	//				{
//	//					"name": "Status",
//	//					"value": "Not started",
//	//				},
//	//			},
//	//		},
//	//	},
//	//	"potentialAction": []lxHelper.M{
//	//		{
//	//			"@type": "ActionCard",
//	//			"name": "Add a comment",
//	//			"inputs": []lxHelper.M{
//	//				{
//	//					"@type": "TextInput",
//	//					"id": "comment",
//	//					"isMultiline": false,
//	//					"title": "Add a comment here for this task",
//	//				},
//	//			},
//	//			"actions": []lxHelper.M{
//	//				{
//	//					"@type": "HttpPOST",
//	//					"name": "Add comment",
//	//					"target": "http://...",
//	//				},
//	//			},
//	//		},
//	//		{
//	//			"@type": "ActionCard",
//	//			"name": "Set due date",
//	//			"inputs": []lxHelper.M{
//	//				{
//	//					"@type": "DataInput",
//	//					"id": "dueDate",
//	//					"title": "Enter a due date for this task",
//	//				},
//	//			},
//	//			"actions": []lxHelper.M{
//	//				{
//	//					"@type": "HttpPOST",
//	//					"name": "Save",
//	//					"target": "http://...",
//	//				},
//	//			},
//	//		},
//	//		{
//	//			"@type": "ActionCard",
//	//			"name": "Change status",
//	//			"inputs": []lxHelper.M{
//	//				{
//	//					"@type": "MultichoiceInput",
//	//					"id": "list",
//	//					"title": "Select a status",
//	//					"isMultiSelect": "false",
//	//					"choices": []lxHelper.M{
//	//						{
//	//							"display": "In Progress",
//	//							"value": "1",
//	//						},
//	//						{
//	//							"display": "Active",
//	//							"value": "2",
//	//						},
//	//						{
//	//							"display": "Closed",
//	//							"value": "3",
//	//						},
//	//					},
//	//				},
//	//			},
//	//			"actions": []lxHelper.M{
//	//				{
//	//					"@type": "HttpPOST",
//	//					"name": "Save",
//	//					"target": "http://...",
//	//				},
//	//			},
//	//		},
//	//	},
//	//}
//	const logUri = "https://console.cloud.google.com/logs/viewer?project=litixsoft&customFacets=undefined&limitCustomFacetWidth=true&minLogLevel=0&expandAll=false&timestamp=2020-08-17T15:56:09.716000000Z&advancedFilter=resource.type%3D%22k8s_container%22%0Aresource.labels.cluster_name%3D%22cluster-1%22%0Aresource.labels.namespace_name%3D%22prod%22%0Aresource.labels.container_name%3D%22urogister-backend-prod%22%0Aseverity%3D%22ERROR%22&dateRangeStart=2020-08-17T14:56:10.080Z&dateRangeEnd=2020-08-17T15:56:10.080Z&interval=PT1H"
//	//entry := lxHelper.M{
//	//	"@type": "MessageCard",
//	//	"@context": "http://schema.org/extensions",
//	//	"themeColor": "#CC4A31",
//	//	"summary": "Error",
//	//	"sections": []interface{}{
//	//		lxHelper.M{
//	//			"activityTitle": "<span style=\"color:#CC4A31\">500</span> | GET | /import/patientDetails/5f28ee112685e1e264b27253",
//	//			"activitySubtitle": "Error by https://www.urogister.de/import",
//	//			//"activityImage": "data:image/jpeg;base64,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",
//	//			"facts": []interface{}{
//	//				lxHelper.M{
//	//					"name": "Assigned to",
//	//					"value": "Unassigned",
//	//				},
//	//				lxHelper.M{
//	//					"name": "Due date",
//	//					"value": "Mon May 01 2017 17:07:18 GMT-0700 (Pacific Daylight Time)",
//	//				},
//	//				lxHelper.M{
//	//					"name": "Status",
//	//					"value": "Not started",
//	//				},
//	//			},
//	//			"markdown": true,
//	//		},
//	//	},
//	//	"potentialAction": []interface{}{
//	//		//lxHelper.M{
//	//		//	"@type": "ActionCard",
//	//		//	"name": "Add a comment",
//	//		//	"inputs": []interface{}{
//	//		//		lxHelper.M{
//	//		//			"@type": "TextInput",
//	//		//			"id": "comment",
//	//		//			"isMultiline": false,
//	//		//			"title": "Add a comment here for this task",
//	//		//		},
//	//		//	},
//	//		//	"actions": []interface{}{
//	//		//		lxHelper.M{
//	//		//			"@type": "HttpPOST",
//	//		//			"name": "Add comment",
//	//		//			"target": "http://...",
//	//		//		},
//	//		//	},
//	//		//},
//	//		//lxHelper.M{
//	//		//	"@type": "ActionCard",
//	//		//	"name": "Set due date",
//	//		//	"inputs": []interface{}{
//	//		//		lxHelper.M{
//	//		//			"@type": "DateInput",
//	//		//			"id": "dueDate",
//	//		//			"title": "Enter a due date for this task",
//	//		//		},
//	//		//	},
//	//		//	"actions": []interface{}{
//	//		//		lxHelper.M{
//	//		//			"@type": "HttpPOST",
//	//		//			"name": "Save",
//	//		//			"target": "http://...",
//	//		//		},
//	//		//	},
//	//		//},
//	//		//lxHelper.M{
//	//		//	"@type": "ActionCard",
//	//		//	"name": "Change status",
//	//		//	"inputs": []interface{}{
//	//		//		lxHelper.M{
//	//		//			"@type": "MultichoiceInput",
//	//		//			"id": "list",
//	//		//			"title": "Select a status",
//	//		//			"isMultiSelect": "false",
//	//		//			"choices": []interface{}{
//	//		//				lxHelper.M{
//	//		//					"display": "In Progress",
//	//		//					"value": "1",
//	//		//				},
//	//		//				lxHelper.M{
//	//		//					"display": "Active",
//	//		//					"value": "2",
//	//		//				},
//	//		//				lxHelper.M{
//	//		//					"display": "Closed",
//	//		//					"value": "3",
//	//		//				},
//	//		//			},
//	//		//		},
//	//		//	},
//	//		//	"actions": []interface{}{
//	//		//		lxHelper.M{
//	//		//			"@type": "HttpPOST",
//	//		//			"name": "Save",
//	//		//			"target": "http://...",
//	//		//		},
//	//		//	},
//	//		//},
//	//		lxHelper.M {
//	//			"@type": "OpenUri",
//	//			"name": "View error in logs",
//	//			"targets": []lxHelper.M{
//	//				{ "os": "default", "uri": logUri },
//	//			},
//	//		},
//	//	},
//	//}
//
//
//	// Convert entry to json
//	jsonData, err := json.Marshal(entry)
//	assert.NoError(t, err)
//
//	t.Log(uri)
//	t.Log(string(jsonData))
//
//	// Post to url
//	cl := http.DefaultClient
//
//	response, err := cl.Post(uri, "application/json", bytes.NewBuffer(jsonData))
//	assert.NoError(t, err)
//
//	defer func() {
//		if err := response.Body.Close(); err != nil {
//			t.Fatal(err)
//		}
//	}()
//
//	body, err := ioutil.ReadAll(response.Body)
//	assert.NoError(t, err)
//
//	t.Log(string(body))
//
//}
