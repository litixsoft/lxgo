package lxHelper

import (
	"encoding/json"
	lxDb "github.com/litixsoft/lxgo/db"
)

// RequestByQuery
type RequestByQuery struct {
	FindOptions lxDb.FindOptions       `json:"opts,omitempty"`
	Query       map[string]interface{} `json:"query"`
	Count       bool                   `json:"count"`
}

// NewRequestByQuery
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

/////////////////////////////////////////////////
// deprecated, Will be removed in a later version
/////////////////////////////////////////////////
func NewReqByQuery(opts string) (*ReqByQuery, error) {
	var data ReqByQuery

	if len(opts) > 0 {
		err := json.Unmarshal([]byte(opts), &data)
		if err != nil {
			return nil, err
		}
	}

	return &data, nil
}
