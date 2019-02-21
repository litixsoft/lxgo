package lxSchema

import (
	"encoding/json"
	"fmt"
	"github.com/xeipuuv/gojsonschema"
	"io/ioutil"
	"net/http"
	"path/filepath"
)

type JSONValidationErrors struct {
	Field   string                 `json:"field"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details"`
}

type JSONValidationResult struct {
	Errors []JSONValidationErrors `json:"errors"`
}

type IJSONSchema interface {
	SetSchemaRootDirectory(dirname string) error
	HasSchema(filename string) bool
	LoadSchema(filename string) (gojsonschema.JSONLoader, error)
	ValidateBind(schema string, req *http.Request, s interface{}) (*JSONValidationResult, error)
}

type JSONSchema struct {
	root    string
	schemas map[string]gojsonschema.JSONLoader
}

var Loader IJSONSchema

// InitJsonSchemaLoader - initiate JSONSchemaLoader globally as Loader
func InitJsonSchemaLoader() {
	Loader = &JSONSchema{
		root:    "",
		schemas: make(map[string]gojsonschema.JSONLoader),
	}
}

// SetSchemaRootDirectory - sets default root schema directory
func (js *JSONSchema) SetSchemaRootDirectory(dirname string) error {
	var err error = nil

	if !filepath.IsAbs(dirname) {
		dirname, err = filepath.Abs(dirname)
	}

	if err == nil {
		js.root = "file:///" + filepath.ToSlash(dirname) + "/"
	}

	return err
}

// HasSchema - checks, if schema already loaded
func (js *JSONSchema) HasSchema(filename string) bool {
	return js.schemas[filename] != nil
}

// LoadSchema - loads a schema and holds it, return schema or error
func (js *JSONSchema) LoadSchema(filename string) (gojsonschema.JSONLoader, error) {
	if !js.HasSchema(filename) {
		var jsonURILoader = gojsonschema.NewReferenceLoader(js.root + filename)

		if _, err := jsonURILoader.LoadJSON(); err != nil {
			return nil, err
		}

		js.schemas[filename] = jsonURILoader
	}

	return js.schemas[filename], nil
}

// ValidateBind - validate given c:echo.Context with schema and binds if success to s:interface{} - if schema not loaded func will perform it
func (js *JSONSchema) ValidateBind(schema string, req *http.Request, s interface{}) (*JSONValidationResult, error) {
	if req == nil {
		return nil, fmt.Errorf("http.request is required")
	}

	schemaLoader, err := js.LoadSchema(schema)

	if err != nil {
		return nil, err
	}

	// No schema
	if schemaLoader == nil {
		return nil, fmt.Errorf("schema could not be loaded %s", schema)
	}

	var (
		jsonRAWDoc []byte
	)

	// Read []byte stream from Request Body
	if jsonRAWDoc, err = ioutil.ReadAll(req.Body); err != nil {
		return nil, err
	}

	// Transfer []byte stream to gojsonschema.documentLoader and Validate
	documentLoader := gojsonschema.NewBytesLoader(jsonRAWDoc)
	res, err := gojsonschema.Validate(schemaLoader, documentLoader)

	if err != nil {
		return nil, err
	}

	// Is schema valid; then
	if !res.Valid() {
		valErrors := res.Errors()
		valResults := &JSONValidationResult{}

		for i := range valErrors {
			valResults.Errors = append(valResults.Errors, JSONValidationErrors{
				Field:   valErrors[i].Field(),
				Message: valErrors[i].Type(),
				Details: valErrors[i].Details(),
			})
		}

		return valResults, nil
	}

	if s != nil {
		// if schema valid; and S given; translate
		if err = json.Unmarshal(documentLoader.JsonSource().([]byte), s); err != nil {
			return nil, err
		}
	}

	// Return result
	return nil, nil
}
