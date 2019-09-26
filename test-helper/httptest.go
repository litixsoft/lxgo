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

// GetTestReqAndRec, returns a test request an recorder
func GetTestReqAndRecJson(method, target string, body io.Reader) (*http.Request, *httptest.ResponseRecorder) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set(HeaderContentType, MIMEApplicationJSON)

	return req, rec
}
