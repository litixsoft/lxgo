package lxHelper

import (
	"github.com/litixsoft/lxgo/db"
)

/////////////////////////////////////////////////
// deprecated, Will be removed in a later version
/////////////////////////////////////////////////
type M map[string]interface{}

type ReqByQuery struct {
	Options    lxDb.Options           `json:"opts,omitempty"`
	IsFiltered bool                   `json:"isFiltered"`
	Query      map[string]interface{} `json:"query"`
}
