package lxAuditMocks_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/litixsoft/lxgo/audit/mocks"
	"github.com/litixsoft/lxgo/helper"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	testDefaultAction = "remove"
	testUser          = lxHelper.M{"name": "Timo Liebetrau", "age": 45}
	testData          = lxHelper.M{"customer": "Karl Lagerfeld"}
)

func TestMockIAudit_Log(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock audit for test
	mockIAudit := lxAuditMocks.NewMockIAudit(mockCtrl)

	// set expects for correct and error
	mockIAudit.EXPECT().Log(testDefaultAction, testUser, testData).Return(nil).Times(1)
	mockIAudit.EXPECT().Log(testDefaultAction, testUser, testData).Return(errors.New("test error")).Times(1)

	// check correct and error
	assert.NoError(t, mockIAudit.Log(testDefaultAction, testUser, testData))
	assert.Error(t, mockIAudit.Log(testDefaultAction, testUser, testData))
}
