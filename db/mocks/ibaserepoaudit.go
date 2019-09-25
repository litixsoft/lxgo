// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/litixsoft/lxgo/db (interfaces: IBaseRepoAudit)

// Package lxDbMocks is a generated GoMock package.
package lxDbMocks

import (
	gomock "github.com/golang/mock/gomock"
	db "github.com/litixsoft/lxgo/db"
	reflect "reflect"
)

// MockIBaseRepoAudit is a mock of IBaseRepoAudit interface
type MockIBaseRepoAudit struct {
	ctrl     *gomock.Controller
	recorder *MockIBaseRepoAuditMockRecorder
}

// MockIBaseRepoAuditMockRecorder is the mock recorder for MockIBaseRepoAudit
type MockIBaseRepoAuditMockRecorder struct {
	mock *MockIBaseRepoAudit
}

// NewMockIBaseRepoAudit creates a new mock instance
func NewMockIBaseRepoAudit(ctrl *gomock.Controller) *MockIBaseRepoAudit {
	mock := &MockIBaseRepoAudit{ctrl: ctrl}
	mock.recorder = &MockIBaseRepoAuditMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIBaseRepoAudit) EXPECT() *MockIBaseRepoAuditMockRecorder {
	return m.recorder
}

// LogEntry mocks base method
func (m *MockIBaseRepoAudit) LogEntry(arg0 *db.AuditLogEntry) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LogEntry", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// LogEntry indicates an expected call of LogEntry
func (mr *MockIBaseRepoAuditMockRecorder) LogEntry(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogEntry", reflect.TypeOf((*MockIBaseRepoAudit)(nil).LogEntry), arg0)
}
