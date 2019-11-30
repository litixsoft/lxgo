package lxPdf_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	lxPdf "github.com/litixsoft/lxgo/pdf"
	"github.com/stretchr/testify/assert"
)

var (
	testServiceKey   = "c470e652-6d46-4f9d-960d-f32d84e682e7"
	testServiceUrl   = "test.host"
	testTemplateDir  = "fixtures"
	testTemplatePath = "testTemplate.tmpl.html"
	testFooterPath   = "testFooter.html"
	testCssPath      = "testStyle.css"

	testData = map[string]interface{}{
		"test": "testFoo",
	}
)

func getTestServer(t *testing.T, resStatus int, testPath string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		its := assert.New(t)

		// Test request parameters
		its.Equal(req.URL.String(), testPath)

		rw.WriteHeader(resStatus)
	}))
}

func TestPdf_CreatePdf(t *testing.T) {
	its := assert.New(t)

	t.Run("return error when template is missing", func(t *testing.T) {
		// instance
		pdf := lxPdf.NewPdfService(testServiceUrl, testServiceKey, testTemplateDir)

		response, err := pdf.CreatePdf("", nil)

		its.Nil(response)
		its.EqualError(err, "invalid params")
	})

	t.Run("return error when data is missing", func(t *testing.T) {
		// instance
		pdf := lxPdf.NewPdfService(testServiceUrl, testServiceKey, testTemplateDir)

		response, err := pdf.CreatePdf("template", nil)

		its.Nil(response)
		its.EqualError(err, "invalid params")
	})

	t.Run("return error when data type is not supported", func(t *testing.T) {
		// instance
		pdf := lxPdf.NewPdfService(testServiceUrl, testServiceKey, testTemplateDir)

		response, err := pdf.CreatePdf("template", map[string]interface{}{"test": func() {}})

		its.Nil(response)
		its.EqualError(err, "malformed data: json: unsupported type: func()")
	})

	t.Run("return error when template not found", func(t *testing.T) {
		// instance
		pdf := lxPdf.NewPdfService(testServiceUrl, testServiceKey, testTemplateDir)

		response, err := pdf.CreatePdf("template", testData)

		its.Nil(response)
		its.EqualError(err, "add template: open file: open fixtures/template: no such file or directory")
	})

	t.Run("return error with invalid url", func(t *testing.T) {
		// instance
		pdf := lxPdf.NewPdfService(testServiceUrl, testServiceKey, testTemplateDir)

		opts := lxPdf.PdfOptions().AddFooterFile("footer").AddImageFile("image")

		response, err := pdf.CreatePdf(testTemplatePath, testData, opts)

		its.Nil(response)
		its.EqualError(err, "do request: Post test.host/create: unsupported protocol scheme \"\"")
	})

	t.Run("return error when remote service return with status 500", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusInternalServerError, "/create")
		defer server.Close()

		// instance
		pdf := lxPdf.NewPdfService(server.URL, testServiceKey, testTemplateDir)

		opts := lxPdf.PdfOptions().AddFooterFile(testFooterPath).AddCssFile(testCssPath)

		response, err := pdf.CreatePdf(testTemplatePath, testData, opts)

		its.Nil(response)
		its.EqualError(err, "service return with code 500: ")
	})

	t.Run("return object on success", func(t *testing.T) {
		// get server and close the server when test finishes
		server := getTestServer(t, http.StatusOK, "/create")
		defer server.Close()

		// instance
		pdf := lxPdf.NewPdfService(server.URL, testServiceKey, testTemplateDir)

		opts := lxPdf.PdfOptions().AddFooterFile(testFooterPath).AddCssFile(testCssPath).SetLandscape(true)

		response, err := pdf.CreatePdf(testTemplatePath, testData, opts)

		its.NotNil(response)
		its.NoError(err)
	})
}
