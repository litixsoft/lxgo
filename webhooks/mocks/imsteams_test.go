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

func TestMockIMsTeams_SendSmall(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	// Mock MsTeamsApi for test
	mockIMsTeams := lxWebhooksMocks.NewMockIMsTeams(mockCtrl)

	// set expects for correct and error
	mockIMsTeams.EXPECT().SendSmall(title, msg, color).Return([]byte(""), nil).Times(1)
	mockIMsTeams.EXPECT().SendSmall(title, msg, color).Return([]byte(""), errors.New("test error")).Times(1)

	// check correct and error
	_, err1 := mockIMsTeams.SendSmall(title, msg, color)
	assert.NoError(t, err1)
	_, err2 := mockIMsTeams.SendSmall(title, msg, color)
	assert.Error(t, err2)
}
