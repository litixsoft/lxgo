// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/litixsoft/lxgo/webhooks (interfaces: IMsTeams)

// Package lxWebhooksMocks is a generated GoMock package.
package lxWebhooksMocks

import (
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockIMsTeams is a mock of IMsTeams interface
type MockIMsTeams struct {
	ctrl     *gomock.Controller
	recorder *MockIMsTeamsMockRecorder
}

// MockIMsTeamsMockRecorder is the mock recorder for MockIMsTeams
type MockIMsTeamsMockRecorder struct {
	mock *MockIMsTeams
}

// NewMockIMsTeams creates a new mock instance
func NewMockIMsTeams(ctrl *gomock.Controller) *MockIMsTeams {
	mock := &MockIMsTeams{ctrl: ctrl}
	mock.recorder = &MockIMsTeamsMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockIMsTeams) EXPECT() *MockIMsTeamsMockRecorder {
	return m.recorder
}

// SendSmall mocks base method
func (m *MockIMsTeams) SendSmall(arg0, arg1, arg2 string) ([]byte, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendSmall", arg0, arg1, arg2)
	ret0, _ := ret[0].([]byte)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// SendSmall indicates an expected call of SendSmall
func (mr *MockIMsTeamsMockRecorder) SendSmall(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendSmall", reflect.TypeOf((*MockIMsTeams)(nil).SendSmall), arg0, arg1, arg2)
}