package lxLog

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"
)

type OutFormat int

const (
	StackdriverErrKey     = "@type"
	StackdriverErrorValue = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"
	LogLevelPanic         = "panic"
	LogLevelFatal         = "fatal"
	LogLevelError         = "error"
	LogLevelWarn          = "warn"
	LogLevelInfo          = "info"
	LogLevelDebug         = "debug"
	LogLevelTrace         = "trace"

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

// GetMessageWithStack, return message string with stack trace
func GetMessageWithStack(message string) string {
	stackSlice := make([]byte, 512)
	s := runtime.Stack(stackSlice, false)

	return fmt.Sprintf("%s\n%s", message, stackSlice[0:s])
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
		log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "severity",
				logrus.FieldKeyMsg:   "message",
			},
		})
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
