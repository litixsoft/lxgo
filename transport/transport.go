package transport

import (
	"encoding/json"
	"fmt"
	"github.com/globalsign/mgo/bson"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
	"io/ioutil"
	"net/http"
)

const (
	TRANSPORT_INTERNAL_CONTENTTYPE_KEYNAME = "Content-Type"

	TRANSPORT_CONTENTTYPE_JSON    = "application/json"
	TRANSPORT_CONTENTTYPE_BSON    = "application/x-bson"
	TRANSPORT_CONTENTTYPE_MSGPACK = "application/x-msgpack"
)

// public

func Bind(req *http.Request, out interface{}) error {
	// Get content-type from header
	sContentType := req.Header.Get(TRANSPORT_INTERNAL_CONTENTTYPE_KEYNAME)

	// get body as []byte
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		return err
	}

	if len(body) == 0 {
		return nil
	}

	// Bind by Content Type
	switch sContentType {
	case TRANSPORT_CONTENTTYPE_BSON:
		return bson.Unmarshal(body, out)
	case TRANSPORT_CONTENTTYPE_JSON:
		return json.Unmarshal(body, out)
	case TRANSPORT_CONTENTTYPE_MSGPACK:
		return msgpack.Unmarshal(body, out)
	}

	return errors.New(fmt.Sprintf("Unknow content type \"%s\"", sContentType))
}

func Send(res http.ResponseWriter, code int, i interface{}, contenttype string) error {
	var out []byte
	var err error

	if i != nil {
		switch contenttype {
		case TRANSPORT_CONTENTTYPE_BSON:
			out, err = bson.Marshal(i)
		case TRANSPORT_CONTENTTYPE_JSON:
			out, err = bson.Marshal(i)
		case TRANSPORT_CONTENTTYPE_MSGPACK:
			out, err = msgpack.Marshal(i)
		default:
			return errors.New(fmt.Sprintf("Unknow content-type \"%s\"to marshal", contenttype))
		}

		if err != nil {
			return err
		}
	}

	header := res.Header()
	header.Set(TRANSPORT_INTERNAL_CONTENTTYPE_KEYNAME, contenttype)

	res.WriteHeader(code)
	_, err = res.Write(out)

	return err
}

func SendBSON(res http.ResponseWriter, code int, i interface{}) error {
	return Send(res, code, i, TRANSPORT_CONTENTTYPE_BSON)
}

func SendJSON(res http.ResponseWriter, code int, i interface{}) error {
	return Send(res, code, i, TRANSPORT_CONTENTTYPE_JSON)
}

func SendMSGPACK(res http.ResponseWriter, code int, i interface{}) error {
	return Send(res, code, i, TRANSPORT_CONTENTTYPE_MSGPACK)
}
