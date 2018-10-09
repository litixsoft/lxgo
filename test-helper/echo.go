package lxTestHelper

import (
	"github.com/labstack/echo"
	"io"
	"io/ioutil"
	"net/http/httptest"
)

// SetRequest, setup json request returns recorder and echo context
func SetEchoRequest(method, target string, body io.Reader) (*httptest.ResponseRecorder, echo.Context) {
	e := echo.New()
	e.Logger.SetOutput(ioutil.Discard)
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, body)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c := e.NewContext(req, rec)
	return rec, c
}
