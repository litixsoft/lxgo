package lxHelper

import (
	"github.com/litixsoft/lxgo/db"
)

type M map[string]interface{}

type ReqByQuery struct {
	Options    lxDb.Options           `json:"opts,omitempty"`
	IsFiltered bool                   `json:"isFiltered"`
	Query      map[string]interface{} `json:"query"`
}
