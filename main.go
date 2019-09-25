package main

import "fmt"

type LogEntryConfig struct {
	AuthUser   interface{} `json:"auth_user,omitempty" bson:"auth_user,omitempty"`
	Db         string      `json:"db,omitempty" bson:"db"`
	Collection string      `json:"collection,omitempty" bson:"collection"`
	Ident      string      `json:"ident,omitempty" bson:"ident"`
	Action     string      `json:"action,omitempty" bson:"action"`
	Data       interface{} `json:"data,omitempty" bson:"data,omitempty"`
}

func main() {

	le := LogEntryConfig{}
	fmt.Println(le)

}
