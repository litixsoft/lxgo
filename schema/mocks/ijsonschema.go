// Code generated by MockGen. DO NOT EDIT.
// Source: schema/schema.go

// Package lxSchemaMocks is a generated GoMock package.
package lxSchemaMocks

import (
	gomock "github.com/golang/mock/gomock"
	lxSchema "github.com/litixsoft/lxgo/schema"
	gojsonschema "github.com/xeipuuv/gojsonschema"
	http "net/http"
	reflect "reflect"
)

// MockIJSONSchema is a mock of IJSONSchema interface
type MockIJSONSchema struct {
	ctrl     *gomock.Controller
	recorder *MockIJSONSchemaMockRecorder
}

// MockIJSONSchemaMockRecorder is the mock recorder for MockIJSONSchema
type MockIJSONSchemaMockRecorder struct {
	mock *MockIJSONSchema
}

// NewMockIJSONSchema creates a new mock instance
func NewMockIJSONSchema(ctrl *gomock.Controller) *MockIJSONSchema {
	mock := &MockIJSONSchema{ctrl: ctrl}
	mock.recorder = &MockIJSONSchemaMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIJSONSchema) EXPECT() *MockIJSONSchemaMockRecorder {
	return m.recorder
}

// SetSchemaRootDirectory mocks base method
func (m *MockIJSONSchema) SetSchemaRootDirectory(dirname string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SetSchemaRootDirectory", dirname)
	ret0, _ := ret[0].(error)
	return ret0
}

// SetSchemaRootDirectory indicates an expected call of SetSchemaRootDirectory
func (mr *MockIJSONSchemaMockRecorder) SetSchemaRootDirectory(dirname interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSchemaRootDirectory", reflect.TypeOf((*MockIJSONSchema)(nil).SetSchemaRootDirectory), dirname)
}

// HasSchema mocks base method
func (m *MockIJSONSchema) HasSchema(filename string) bool {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "HasSchema", filename)
	ret0, _ := ret[0].(bool)
	return ret0
}

// HasSchema indicates an expected call of HasSchema
func (mr *MockIJSONSchemaMockRecorder) HasSchema(filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "HasSchema", reflect.TypeOf((*MockIJSONSchema)(nil).HasSchema), filename)
}

// LoadSchema mocks base method
func (m *MockIJSONSchema) LoadSchema(filename string) (gojsonschema.JSONLoader, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LoadSchema", filename)
	ret0, _ := ret[0].(gojsonschema.JSONLoader)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LoadSchema indicates an expected call of LoadSchema
func (mr *MockIJSONSchemaMockRecorder) LoadSchema(filename interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LoadSchema", reflect.TypeOf((*MockIJSONSchema)(nil).LoadSchema), filename)
}

// ValidateBind mocks base method
func (m *MockIJSONSchema) ValidateBind(schema string, req *http.Request, s interface{}) (*lxSchema.JSONValidationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateBind", schema, req, s)
	ret0, _ := ret[0].(*lxSchema.JSONValidationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateBind indicates an expected call of ValidateBind
func (mr *MockIJSONSchemaMockRecorder) ValidateBind(schema, req, s interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateBind", reflect.TypeOf((*MockIJSONSchema)(nil).ValidateBind), schema, req, s)
}

// ValidateBindRaw mocks base method
func (m *MockIJSONSchema) ValidateBindRaw(schema string, data *[]byte, s interface{}) (*lxSchema.JSONValidationResult, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ValidateBindRaw", schema, data, s)
	ret0, _ := ret[0].(*lxSchema.JSONValidationResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ValidateBindRaw indicates an expected call of ValidateBindRaw
func (mr *MockIJSONSchemaMockRecorder) ValidateBindRaw(schema, data, s interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ValidateBindRaw", reflect.TypeOf((*MockIJSONSchema)(nil).ValidateBindRaw), schema, data, s)
}
