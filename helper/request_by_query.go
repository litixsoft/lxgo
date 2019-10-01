package lxHelper

import (
	"encoding/json"
)

// RequestByQuery
type RequestByQuery struct {
	FindOptions FindOptions            `json:"opts,omitempty"`
	Query       map[string]interface{} `json:"query"`
	Count       bool                   `json:"count"`
}

// NewRequestByQuery, convert query string
// ?config={"opts":{"sort":{"name":1, "email": -1},"skip":5,"limit":10},"query":{"name":"Schulterglatze"}, "count":true}
func NewRequestByQuery(queryStr string) (data *RequestByQuery, err error) {
	data = new(RequestByQuery)

	// Exit when empty string
	if len(queryStr) == 0 {
		return data, err
	}

	// Convert
	err = json.Unmarshal([]byte(queryStr), data)

	return data, err
}
