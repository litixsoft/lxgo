package lxTestHelper

import (
	"io"
	"net/http"
	"net/http/httptest"
)

const (
	HeaderContentType   = "Content-Type"
	MIMEApplicationJSON = "application/json"
)

// SetRequest, setup json request returns recorder and echo context
//func SetEchoRequest(method, target string, body io.Reader) (*httptest.ResponseRecorder, echo.Context) {
//	e := echo.New()
//	e.Logger.SetOutput(ioutil.Discard)
//	rec := httptest.NewRecorder()
//	req := httptest.NewRequest(method, target, body)
//	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
//	c := e.NewContext(req, rec)
//	return rec, c
//}

// GetTestReqAndRec, returns a test request an recorder
func GetTestReqAndRecJson(method, target string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)

	return req, rec
}
