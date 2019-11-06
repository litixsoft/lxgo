package lxLog

import (
	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
	"time"
)

type OutFormat int

const (
	FormatText OutFormat = iota + 1
	FormatJson
	FormatFluentd
)

var outFormat OutFormat
var logger *logrus.Logger
var once sync.Once
var hasInit bool

func GetLogger() *logrus.Logger {
	once.Do(func() {
		logger = logrus.New()
	})
	return logger
}

func GetOutFormat() OutFormat {
	return outFormat
}

func HasInit() bool {
	return hasInit
}

func InitLogger(output io.Writer, logLevel, logFormat string) (log *logrus.Logger) {
	log = GetLogger()
	log.SetOutput(output)

	// Setup log level
	switch logLevel {
	case "panic":
		log.SetLevel(logrus.PanicLevel)
	case "fatal":
		log.SetLevel(logrus.FatalLevel)
	case "error":
		log.SetLevel(logrus.ErrorLevel)
	case "warn":
		log.SetLevel(logrus.WarnLevel)
	case "info":
		log.SetLevel(logrus.InfoLevel)
	case "debug":
		log.SetLevel(logrus.DebugLevel)
	case "trace":
		log.SetLevel(logrus.TraceLevel)
	default:
		log.SetLevel(logrus.DebugLevel)
	}

	// Setup log format
	switch logFormat {
	case "text":
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
		outFormat = FormatText
	case "json":
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
		outFormat = FormatJson
	case "fluentd":
		log.SetFormatter(joonix.NewFormatter(joonix.StackdriverFormat, joonix.DisableTimestampFormat))
		// Set format for middleware config
		outFormat = FormatFluentd
	default:
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
		outFormat = FormatText
	}

	hasInit = true
	return log
}
