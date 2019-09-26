// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/litixsoft/lxgo/audit (interfaces: IAuditLogger)

// Package lxAuditMocks is a generated GoMock package.
package lxAuditMocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
	time "time"
)

// MockIAuditLogger is a mock of IAuditLogger interface
type MockIAuditLogger struct {
	ctrl     *gomock.Controller
	recorder *MockIAuditLoggerMockRecorder
}

// MockIAuditLoggerMockRecorder is the mock recorder for MockIAuditLogger
type MockIAuditLoggerMockRecorder struct {
	mock *MockIAuditLogger
}

// NewMockIAuditLogger creates a new mock instance
func NewMockIAuditLogger(ctrl *gomock.Controller) *MockIAuditLogger {
	mock := &MockIAuditLogger{ctrl: ctrl}
	mock.recorder = &MockIAuditLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIAuditLogger) EXPECT() *MockIAuditLoggerMockRecorder {
	return m.recorder
}

// LogEntry mocks base method
func (m *MockIAuditLogger) LogEntry(arg0 string, arg1, arg2 interface{}, arg3 ...time.Duration) error {
	m.ctrl.T.Helper()
	varargs := []interface{}{arg0, arg1, arg2}
	for _, a := range arg3 {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "LogEntry", varargs...)
	ret0, _ := ret[0].(error)
	return ret0
}

// LogEntry indicates an expected call of LogEntry
func (mr *MockIAuditLoggerMockRecorder) LogEntry(arg0, arg1, arg2 interface{}, arg3 ...interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]interface{}{arg0, arg1, arg2}, arg3...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LogEntry", reflect.TypeOf((*MockIAuditLogger)(nil).LogEntry), varargs...)
}
