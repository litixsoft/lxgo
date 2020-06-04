package lxSchema_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/litixsoft/lxgo/helper"
	"github.com/litixsoft/lxgo/schema"
	"github.com/litixsoft/lxgo/test-helper"
	"github.com/stretchr/testify/assert"
)

const (
	SchemaRootPath          = "fixtures"
	SchemaExistsFilename    = "schema_001.json"
	SchemaNotExistsFilename = "schema_002.json"
)

func buildHTTPContext(data lxHelper.M) (*http.Request, *httptest.ResponseRecorder) {
	// Convert login data to json
	//user := lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"}

	jsonData, err := json.Marshal(data)

	if err != nil {
		panic(err)
	}

	return lxTestHelper.GetTestReqAndRecJson(http.MethodPost, "/request_mock", bytes.NewBuffer(jsonData))
}

func TestJSONSchemaStruct_SetSchemaRootDirectory(t *testing.T) {
	t.Run("Set root schema directory", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)
	})
}

func TestJSONSchemaStruct_LoadSchema(t *testing.T) {
	t.Run("Load a json schema", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.LoadSchema(SchemaExistsFilename)
		assert.NotNil(t, res, "LoadSchema returns nil")
		assert.NoError(t, err)
	})

	t.Run("Return error if no json schema loader", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()
		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.LoadSchema(SchemaNotExistsFilename)
		assert.Nil(t, res, "LoadSchema return not nil")
		assert.Error(t, err)
	})
}

func TestJSONSchemaStruct_HasSchema(t *testing.T) {
	t.Run("Check if loaded schema exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()
		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.LoadSchema(SchemaExistsFilename)
		assert.NotNil(t, res, "LoadSchema returns nil")
		assert.NoError(t, err)

		bRes := lxSchema.Loader.HasSchema(SchemaExistsFilename)
		assert.Equal(t, bRes, true)
	})

	t.Run("Check if not loaded schema not exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()
		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res := lxSchema.Loader.HasSchema(SchemaNotExistsFilename)
		assert.Equal(t, res, false)
	})
}

func TestJSONSchemaStruct_ValidateBind(t *testing.T) {
	t.Run("Error if no schema given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBind("", nil, nil)

		assert.Nil(t, res, "Schema was loaded")
		assert.Error(t, err)
	})

	t.Run("Error if schema not exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBind(SchemaNotExistsFilename, nil, nil)

		assert.Nil(t, res, "schema was loaded")
		assert.Error(t, err)
	})

	t.Run("Error if no context given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBind(SchemaExistsFilename, nil, nil)

		assert.Nil(t, res, "schema was loaded")
		assert.Error(t, err)
		assert.Equal(t, err, fmt.Errorf("http.request is required"))
	})

	t.Run("validate with given context", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		req, _ := buildHTTPContext(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"})
		res, err := lxSchema.Loader.ValidateBind(SchemaExistsFilename, req, nil)

		assert.Nil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")
	})

	t.Run("validate successfully and map to structure", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		req, _ := buildHTTPContext(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"})
		res, err := lxSchema.Loader.ValidateBind(SchemaExistsFilename, req, &s)

		assert.Nil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")

		// validate schema
		assert.Equal(t, s.Name, "Otto", "s.name is invalid")
		assert.Equal(t, s.Loginname, "otto", "s.login_name is invalid")
		assert.Equal(t, s.Email, "otto@otto.com", "s.email is invalid")
	})

	t.Run("validate with error and dont map to structure", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		req, _ := buildHTTPContext(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "wrong"})
		res, err := lxSchema.Loader.ValidateBind(SchemaExistsFilename, req, &s)

		assert.NotNil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")

		var vErrs = res.Errors
		assert.Equal(t, len(vErrs), 1, "one validation error")

		// validate schema
		assert.Empty(t, s.Name, "s.name is given")
		assert.Empty(t, s.Loginname, "s.login_name is given")
		assert.Empty(t, s.Email, "s.email is given")
	})

	t.Run("return error while invalid json given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		// Invalid Json - { "login_name": Data", }
		jsonData := `{ "login_name": Data", }`

		req, _ := lxTestHelper.GetTestReqAndRecJson(http.MethodPost, "/request_mock", bytes.NewBufferString(jsonData))
		res, err := lxSchema.Loader.ValidateBind(SchemaExistsFilename, req, &s)

		assert.Nil(t, res, "no result expected")
		assert.Error(t, err, "error wantend while invalid json")
		assert.Equal(t, err.Error(), "invalid character 'D' looking for beginning of value", "error message")
	})
}

func TestJSONSchemaStruct_ValidateBindRaw(t *testing.T) {
	t.Run("Error if no schema given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBindRaw("", nil, nil)

		assert.Nil(t, res, "Schema was loaded")
		assert.Error(t, err)
	})

	t.Run("Error if schema not exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBindRaw(SchemaNotExistsFilename, nil, nil)

		assert.Nil(t, res, "schema was loaded")
		assert.Error(t, err)
	})

	t.Run("Error if no context given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBindRaw(SchemaExistsFilename, nil, nil)

		assert.Nil(t, res, "schema was loaded")
		assert.Error(t, err)
		assert.Equal(t, err, fmt.Errorf("data is required"))
	})

	t.Run("validate with given context", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		data, _ := json.Marshal(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"})
		res, err := lxSchema.Loader.ValidateBindRaw(SchemaExistsFilename, &data, nil)

		assert.Nil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")
	})

	t.Run("validate successfully and map to structure", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		data, _ := json.Marshal(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"})
		res, err := lxSchema.Loader.ValidateBindRaw(SchemaExistsFilename, &data, &s)

		assert.Nil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")

		// validate schema
		assert.Equal(t, s.Name, "Otto", "s.name is invalid")
		assert.Equal(t, s.Loginname, "otto", "s.login_name is invalid")
		assert.Equal(t, s.Email, "otto@otto.com", "s.email is invalid")
	})

	t.Run("validate with error and dont map to structure", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SchemaRootPath)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		data, _ := json.Marshal(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "wrong"})
		res, err := lxSchema.Loader.ValidateBindRaw(SchemaExistsFilename, &data, &s)

		assert.NotNil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")

		var vErrs = res.Errors
		assert.Equal(t, len(vErrs), 1, "one validation error")

		// validate schema
		assert.Empty(t, s.Name, "s.name is given")
		assert.Empty(t, s.Loginname, "s.login_name is given")
		assert.Empty(t, s.Email, "s.email is given")
	})
}
