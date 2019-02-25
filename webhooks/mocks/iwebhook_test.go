package lxWebhooksMocks_test

import (
	"errors"
	"github.com/golang/mock/gomock"
	"github.com/litixsoft/lxgo/webhooks"
	"github.com/litixsoft/lxgo/webhooks/mocks"
	"github.com/stretchr/testify/assert"
	"testing"
)

var (
	title = "Test99"
	msg   = "test message"
	color = lxWebhooks.RedDark
)

func TestMockIWebhook_SendSmall(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock MsTeamsApi for test
	mockIWebhook := lxWebhooksMocks.NewMockIWebhook(mockCtrl)

	// set expects for correct and error
	mockIWebhook.EXPECT().SendSmall(title, msg, color).Return([]byte(""), nil).Times(1)
	mockIWebhook.EXPECT().SendSmall(title, msg, color).Return([]byte(""), errors.New("test error")).Times(1)

	// check correct and error
	_, err1 := mockIWebhook.SendSmall(title, msg, color)
	assert.NoError(t, err1)
	_, err2 := mockIWebhook.SendSmall(title, msg, color)
	assert.Error(t, err2)
}
