package lxSchema_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo"
	"github.com/litixsoft/lxgo/helper"
	"github.com/litixsoft/lxgo/schema"
	"github.com/litixsoft/lxgo/test-helper"
	"github.com/stretchr/testify/assert"
	"testing"
)

const (
	SCHEMAROOTPATH            = "fixtures"
	SCHEMA_EXISTS_FILENAME    = "schema_001.json"
	SCHEMA_NOTEXISTS_FILENAME = "schema_002.json"
)

func buildEchoContext(data lxHelper.M) echo.Context {
	// Convert login data to json
	//user := lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"}

	jsonData, err := json.Marshal(data)

	if err != nil {
		panic(err)
	}

	_, c := lxTestHelper.SetEchoRequest(echo.POST, "/request_mock", bytes.NewBuffer(jsonData))
	return c
}

func TestJSONSchemaStruct_SetSchemaRootDirectory(t *testing.T) {
	t.Run("Set root schema directory", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)
	})
}

func TestJSONSchemaStruct_LoadSchema(t *testing.T) {
	t.Run("Load a json schema", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.LoadSchema(SCHEMA_EXISTS_FILENAME)
		assert.NotNil(t, res, "LoadSchema returns nil")
		assert.NoError(t, err)
	})

	t.Run("Return error if no json schema loader", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()
		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.LoadSchema(SCHEMA_NOTEXISTS_FILENAME)
		assert.Nil(t, res, "LoadSchema return not nil")
		assert.Error(t, err)
	})
}

func TestJSONSchemaStruct_HasSchema(t *testing.T) {
	t.Run("Check if loaded schema exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()
		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.LoadSchema(SCHEMA_EXISTS_FILENAME)
		assert.NotNil(t, res, "LoadSchema returns nil")
		assert.NoError(t, err)

		bRes := lxSchema.Loader.HasSchema(SCHEMA_EXISTS_FILENAME)
		assert.Equal(t, bRes, true)
	})

	t.Run("Check if not loaded schema not exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()
		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res := lxSchema.Loader.HasSchema(SCHEMA_NOTEXISTS_FILENAME)
		assert.Equal(t, res, false)
	})
}

func TestJSONSchemaStruct_ValidateBind(t *testing.T) {
	t.Run("Error if no schema given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBind("", nil, nil)

		assert.Nil(t, res, "Schema was loaded")
		assert.Error(t, err)
	})

	t.Run("Error if schema not exists", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBind(SCHEMA_NOTEXISTS_FILENAME, nil, nil)

		assert.Nil(t, res, "schema was loaded")
		assert.Error(t, err)
	})

	t.Run("Error if no context given", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		res, err := lxSchema.Loader.ValidateBind(SCHEMA_EXISTS_FILENAME, nil, nil)

		assert.Nil(t, res, "schema was loaded")
		assert.Error(t, err)
		assert.Equal(t, err, fmt.Errorf("context is required"))
	})

	t.Run("validate with given context", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		c := buildEchoContext(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"})
		res, err := lxSchema.Loader.ValidateBind(SCHEMA_EXISTS_FILENAME, c, nil)

		assert.Nil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")
	})

	t.Run("validate successfully and map to structure", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		c := buildEchoContext(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "otto@otto.com"})
		res, err := lxSchema.Loader.ValidateBind(SCHEMA_EXISTS_FILENAME, c, &s)

		assert.Nil(t, res, "no valid result")
		assert.NoError(t, err, "returns error")

		// validate schema
		assert.Equal(t, s.Name, "Otto", "s.name is invalid")
		assert.Equal(t, s.Loginname, "otto", "s.login_name is invalid")
		assert.Equal(t, s.Email, "otto@otto.com", "s.email is invalid")
	})

	t.Run("validate with error and dont map to structure", func(t *testing.T) {
		lxSchema.InitJsonSchemaLoader()

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		c := buildEchoContext(lxHelper.M{"name": "Otto", "login_name": "otto", "email": "wrong"})
		res, err := lxSchema.Loader.ValidateBind(SCHEMA_EXISTS_FILENAME, c, &s)

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

		err := lxSchema.Loader.SetSchemaRootDirectory(SCHEMAROOTPATH)
		assert.NoError(t, err)

		var s struct {
			Name      string `json:"name"`
			Loginname string `json:"login_name"`
			Email     string `json:"email"`
		}

		// Invalid Json - { "login_name": Data", }
		jsonData := `{ "login_name": Data", }`

		_, c := lxTestHelper.SetEchoRequest(echo.POST, "/request_mock", bytes.NewBufferString(jsonData))
		res, err := lxSchema.Loader.ValidateBind(SCHEMA_EXISTS_FILENAME, c, &s)

		assert.Nil(t, res, "no result expected")
		assert.Error(t, err, "error wantend while invalid json")
		assert.Equal(t, err.Error(), "invalid character 'D' looking for beginning of value", "error message")
	})
}
