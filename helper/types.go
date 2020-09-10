package lxHelper

import "go.mongodb.org/mongo-driver/bson"

type M map[string]interface{}
type A []M

// ToMap, convert interface to lxHelper.M
func ToMap(v interface{}) (ret M, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &ret)
	return
}
