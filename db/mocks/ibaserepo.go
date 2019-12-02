// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/litixsoft/lxgo/db (interfaces: IBaseRepo)

// Package lxDbMocks is a generated GoMock package.
package lxDbMocks

import (
	gomock "github.com/golang/mock/gomock"
	db "github.com/litixsoft/lxgo/db"
	reflect "reflect"
)

// MockIBaseRepo is a mock of IBaseRepo interface
type MockIBaseRepo struct {
	ctrl     *gomock.Controller
	recorder *MockIBaseRepoMockRecorder
}

// MockIBaseRepoMockRecorder is the mock recorder for MockIBaseRepo
type MockIBaseRepoMockRecorder struct {
	mock *MockIBaseRepo
}

// NewMockIBaseRepo creates a new mock instance
func NewMockIBaseRepo(ctrl *gomock.Controller) *MockIBaseRepo {
	mock := &MockIBaseRepo{ctrl: ctrl}
	mock.recorder = &MockIBaseRepoMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIBaseRepo) EXPECT() *MockIBaseRepoMockRecorder {
	return m.recorder
}

// CountDocuments mocks base method
func (m *MockIBaseRepo) CountDocuments(arg0 interface{}, arg1 ...interface{}) (int64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CountDocuments", varargs...)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CountDocuments indicates an expected call of CountDocuments
func (mr *MockIBaseRepoMockRecorder) CountDocuments(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CountDocuments", reflect.TypeOf((*MockIBaseRepo)(nil).CountDocuments), varargs...)
}

// CreateIndexes mocks base method
func (m *MockIBaseRepo) CreateIndexes(arg0 interface{}, arg1 ...interface{}) ([]string, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "CreateIndexes", varargs...)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateIndexes indicates an expected call of CreateIndexes
func (mr *MockIBaseRepoMockRecorder) CreateIndexes(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateIndexes", reflect.TypeOf((*MockIBaseRepo)(nil).CreateIndexes), varargs...)
}

// DeleteMany mocks base method
func (m *MockIBaseRepo) DeleteMany(arg0 interface{}, arg1 ...interface{}) (*db.DeleteManyResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteMany", varargs...)
	ret0, _ := ret[0].(*db.DeleteManyResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// DeleteMany indicates an expected call of DeleteMany
func (mr *MockIBaseRepoMockRecorder) DeleteMany(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteMany", reflect.TypeOf((*MockIBaseRepo)(nil).DeleteMany), varargs...)
}

// DeleteOne mocks base method
func (m *MockIBaseRepo) DeleteOne(arg0 interface{}, arg1 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "DeleteOne", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteOne indicates an expected call of DeleteOne
func (mr *MockIBaseRepoMockRecorder) DeleteOne(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteOne", reflect.TypeOf((*MockIBaseRepo)(nil).DeleteOne), varargs...)
}

// EstimatedDocumentCount mocks base method
func (m *MockIBaseRepo) EstimatedDocumentCount(arg0 ...interface{}) (int64, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{}
	for _, a := range arg0 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "EstimatedDocumentCount", varargs...)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// EstimatedDocumentCount indicates an expected call of EstimatedDocumentCount
func (mr *MockIBaseRepoMockRecorder) EstimatedDocumentCount(arg0 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EstimatedDocumentCount", reflect.TypeOf((*MockIBaseRepo)(nil).EstimatedDocumentCount), arg0...)
}

// Find mocks base method
func (m *MockIBaseRepo) Find(arg0, arg1 interface{}, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Find", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// Find indicates an expected call of Find
func (mr *MockIBaseRepoMockRecorder) Find(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Find", reflect.TypeOf((*MockIBaseRepo)(nil).Find), varargs...)
}

// FindOne mocks base method
func (m *MockIBaseRepo) FindOne(arg0, arg1 interface{}, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindOne", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// FindOne indicates an expected call of FindOne
func (mr *MockIBaseRepoMockRecorder) FindOne(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOne", reflect.TypeOf((*MockIBaseRepo)(nil).FindOne), varargs...)
}

// FindOneAndDelete mocks base method
func (m *MockIBaseRepo) FindOneAndDelete(arg0, arg1 interface{}, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindOneAndDelete", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// FindOneAndDelete indicates an expected call of FindOneAndDelete
func (mr *MockIBaseRepoMockRecorder) FindOneAndDelete(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneAndDelete", reflect.TypeOf((*MockIBaseRepo)(nil).FindOneAndDelete), varargs...)
}

// FindOneAndReplace mocks base method
func (m *MockIBaseRepo) FindOneAndReplace(arg0, arg1, arg2 interface{}, arg3 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindOneAndReplace", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// FindOneAndReplace indicates an expected call of FindOneAndReplace
func (mr *MockIBaseRepoMockRecorder) FindOneAndReplace(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneAndReplace", reflect.TypeOf((*MockIBaseRepo)(nil).FindOneAndReplace), varargs...)
}

// FindOneAndUpdate mocks base method
func (m *MockIBaseRepo) FindOneAndUpdate(arg0, arg1, arg2 interface{}, arg3 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "FindOneAndUpdate", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// FindOneAndUpdate indicates an expected call of FindOneAndUpdate
func (mr *MockIBaseRepoMockRecorder) FindOneAndUpdate(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindOneAndUpdate", reflect.TypeOf((*MockIBaseRepo)(nil).FindOneAndUpdate), varargs...)
}

// GetCollection mocks base method
func (m *MockIBaseRepo) GetCollection() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetCollection")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// GetCollection indicates an expected call of GetCollection
func (mr *MockIBaseRepoMockRecorder) GetCollection() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetCollection", reflect.TypeOf((*MockIBaseRepo)(nil).GetCollection))
}

// GetDb mocks base method
func (m *MockIBaseRepo) GetDb() interface{} {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetDb")
	ret0, _ := ret[0].(interface{})
	return ret0
}

// GetDb indicates an expected call of GetDb
func (mr *MockIBaseRepoMockRecorder) GetDb() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetDb", reflect.TypeOf((*MockIBaseRepo)(nil).GetDb))
}

// GetRepoName mocks base method
func (m *MockIBaseRepo) GetRepoName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetRepoName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetRepoName indicates an expected call of GetRepoName
func (mr *MockIBaseRepoMockRecorder) GetRepoName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetRepoName", reflect.TypeOf((*MockIBaseRepo)(nil).GetRepoName))
}

// InsertMany mocks base method
func (m *MockIBaseRepo) InsertMany(arg0 []interface{}, arg1 ...interface{}) (*db.InsertManyResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "InsertMany", varargs...)
	ret0, _ := ret[0].(*db.InsertManyResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertMany indicates an expected call of InsertMany
func (mr *MockIBaseRepoMockRecorder) InsertMany(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertMany", reflect.TypeOf((*MockIBaseRepo)(nil).InsertMany), varargs...)
}

// InsertOne mocks base method
func (m *MockIBaseRepo) InsertOne(arg0 interface{}, arg1 ...interface{}) (interface{}, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0}
	for _, a := range arg1 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "InsertOne", varargs...)
	ret0, _ := ret[0].(interface{})
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// InsertOne indicates an expected call of InsertOne
func (mr *MockIBaseRepoMockRecorder) InsertOne(arg0 interface{}, arg1 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0}, arg1...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InsertOne", reflect.TypeOf((*MockIBaseRepo)(nil).InsertOne), varargs...)
}

// UpdateMany mocks base method
func (m *MockIBaseRepo) UpdateMany(arg0, arg1 interface{}, arg2 ...interface{}) (*db.UpdateManyResult, error) {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateMany", varargs...)
	ret0, _ := ret[0].(*db.UpdateManyResult)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UpdateMany indicates an expected call of UpdateMany
func (mr *MockIBaseRepoMockRecorder) UpdateMany(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateMany", reflect.TypeOf((*MockIBaseRepo)(nil).UpdateMany), varargs...)
}

// UpdateOne mocks base method
func (m *MockIBaseRepo) UpdateOne(arg0, arg1 interface{}, arg2 ...interface{}) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1}
	for _, a := range arg2 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "UpdateOne", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOne indicates an expected call of UpdateOne
func (mr *MockIBaseRepoMockRecorder) UpdateOne(arg0, arg1 interface{}, arg2 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1}, arg2...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOne", reflect.TypeOf((*MockIBaseRepo)(nil).UpdateOne), varargs...)
}
