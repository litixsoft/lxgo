package lxHelper

import "encoding/json"

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
