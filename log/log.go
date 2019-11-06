package lxLog

import (
	joonix "github.com/joonix/log"
	"github.com/sirupsen/logrus"
	"io"
	"strings"
	"sync"
	"time"
)

type OutFormat int

const (
	LogLevelPanic = "panic"
	LogLevelFatal = "fatal"
	LogLevelError = "error"
	LogLevelWarn  = "warn"
	LogLevelInfo  = "info"
	LogLevelDebug = "debug"
	LogLevelTrace = "trace"

	LogFormatText    = "text"
	LogFormatJson    = "json"
	LogFormatFluentd = "fluentd"

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
	switch strings.ToLower(logLevel) {
	case LogLevelPanic:
		log.SetLevel(logrus.PanicLevel)
	case LogLevelFatal:
		log.SetLevel(logrus.FatalLevel)
	case LogLevelError:
		log.SetLevel(logrus.ErrorLevel)
	case LogLevelWarn:
		log.SetLevel(logrus.WarnLevel)
	case LogLevelInfo:
		log.SetLevel(logrus.InfoLevel)
	case LogLevelDebug:
		log.SetLevel(logrus.DebugLevel)
	case LogLevelTrace:
		log.SetLevel(logrus.TraceLevel)
	default:
		log.SetLevel(logrus.DebugLevel)
	}

	// Setup log format
	switch strings.ToLower(logFormat) {
	case LogFormatText:
		log.SetFormatter(&logrus.TextFormatter{
			ForceColors:     true,
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
		outFormat = FormatText
	case LogFormatJson:
		log.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339})
		outFormat = FormatJson
	case LogFormatFluentd:
		log.SetFormatter(joonix.NewFormatter(joonix.StackdriverFormat, joonix.DisableTimestampFormat))
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
