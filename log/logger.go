package lxLog

import (
	"github.com/sirupsen/logrus"
	"sync"
)

var logger *logrus.Logger
var once sync.Once

func GetLogger() *logrus.Logger {
	once.Do(func() {
		logger = logrus.New()
	})
	return logger
}
