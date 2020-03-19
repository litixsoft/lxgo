package lxLogMiddleware

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	lxLog "github.com/litixsoft/lxgo/log"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type (
	// EchoLoggerConfig defines the config for EchoLoggerConfig middleware.
	EchoLoggerConfig struct {
		// Skipper defines a function to skip middleware.
		middleware.Skipper

		// Logger instance
		Logger *logrus.Logger

		// LoggerOutFormat format for output.
		LoggerOutFormat lxLog.OutFormat
	}
)

// DefaultEchoLoggerConfig
var (
	DefaultEchoLoggerConfig = EchoLoggerConfig{
		Skipper:         middleware.DefaultSkipper,
		Logger:          lxLog.GetLogger(),
		LoggerOutFormat: lxLog.GetOutFormat(),
	}
)

// EchoLogger returns an EchoLogger middleware with default config.
// Default config based on logger init
// Usage:
// e.Use(lxLogMiddleware.EchoLogger())
func EchoLogger() echo.MiddlewareFunc {
	if !lxLog.HasInit() {
		panic(errors.New("logger is not initialized, an instance must be must be initialized"))
	}
	c := DefaultEchoLoggerConfig
	return EchoLoggerWithConfig(c)
}

// EchoLoggerWithConfig returns an EchoLogger middleware with custom config.
// Usage:
// e.Use(lxLogMiddleware.EchoLoggerWithConfig(lxLogMiddleware.EchoLoggerConfig {
//     Logger: log,
//     OutFormat: lxLog.FormatText,
// }))
func EchoLoggerWithConfig(config EchoLoggerConfig) echo.MiddlewareFunc {
	// Check logger, is required
	if config.Logger == nil {
		if !lxLog.HasInit() {
			panic(errors.New("logger is not initialized, an instance must be must be initialized"))
		}
		config.Logger = lxLog.GetLogger()
	}

	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultEchoLoggerConfig.Skipper
	}
	if config.LoggerOutFormat == 0 {
		config.LoggerOutFormat = lxLog.GetOutFormat()
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()
			start := time.Now()

			var err error
			if err = next(c); err != nil {
				c.Error(err)
			}
			stop := time.Now()

			id := req.Header.Get(echo.HeaderXRequestID)
			if id == "" {
				id = res.Header().Get(echo.HeaderXRequestID)
			}
			reqSize := req.Header.Get(echo.HeaderContentLength)
			if reqSize == "" {
				reqSize = "0"
			}

			// Create fields and messages
			var msg string
			fields := logrus.Fields{}

			// Create output format
			switch config.LoggerOutFormat {
			case lxLog.FormatText:
				// Create text message
				msg = fmt.Sprintf("| %3d | %13v | %15s | %-7s %s",
					res.Status,
					stop.Sub(start).String(),
					c.RealIP(),
					req.Method,
					req.RequestURI,
				)
			case lxLog.FormatJson, lxLog.FormatFluentd:
				// Create fluentd message
				msg = fmt.Sprintf("%3d | %s %s",
					res.Status,
					req.Method,
					req.RequestURI,
				)

				fields = logrus.Fields{
					"id":            id,
					"remote_ip":     c.RealIP(),
					"host":          req.Host,
					"method":        req.Method,
					"uri":           req.RequestURI,
					"status":        res.Status,
					"bytes_in":      reqSize,
					"bytes_out":     strconv.FormatInt(res.Size, 10),
					"referer":       req.Referer(),
					"protocol":      req.Proto,
					"latency":       strconv.FormatInt(int64(stop.Sub(start)), 10),
					"latency_human": stop.Sub(start).String(),
					"user_agent":    req.UserAgent(),
				}
			}

			// Log with status
			switch {
			case res.Status >= 500:
				config.Logger.WithFields(fields).Error(msg)
			case res.Status >= 300:
				config.Logger.WithFields(fields).Warn(msg)
			default:
				config.Logger.WithFields(fields).Info(msg)
			}

			return err
		}
	}
}
