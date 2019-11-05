package lxLogMiddleware

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type Out int

const (
	FormatText Out = iota + 1
	FormatFluentd
)

type (
	// EchoLogConfig defines the config for EchoLogConfig middleware.
	EchoLogConfig struct {
		// Skipper defines a function to skip middleware.
		middleware.Skipper

		// Logger instance
		Logger *logrus.Logger

		// OutFormat format for output.
		OutFormat Out
	}
)

// DefaultEchoLogConfig
var (
	DefaultEchoLogConfig = EchoLogConfig{
		Skipper:   middleware.DefaultSkipper,
		OutFormat: FormatText,
	}
)

// EchoLog returns an EchoLog middleware with default config.
// Example for use:
// e.Use(lxLogMiddleware.EchoLog(log))
func EchoLog(logger *logrus.Logger) echo.MiddlewareFunc {
	c := DefaultEchoLogConfig
	c.Logger = logger
	return EchoLogWithConfig(c)
}

// EchoLogWithConfig returns an EchoLog middleware with custom config.
// Example for use:
// e.Use(lxLogMiddleware.EchoLogWithConfig(lxLogMiddleware.EchoLogConfig {
//     Logger: log,
//     OutFormat: lxLogMiddleware.FormatText,
// }))
func EchoLogWithConfig(config EchoLogConfig) echo.MiddlewareFunc {
	// Check logger, is required
	if config.Logger == nil {
		panic(errors.New("logger not available, an instance must be present in the configuration"))
	}

	// Defaults
	if config.Skipper == nil {
		config.Skipper = DefaultEchoLogConfig.Skipper
	}
	if config.OutFormat == 0 {
		config.OutFormat = FormatText
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
			switch config.OutFormat {
			case FormatText:
				// Create text message
				msg = fmt.Sprintf("| %3d | %13v | %15s | %-7s %s",
					res.Status,
					stop.Sub(start).String(),
					c.RealIP(),
					req.Method,
					req.RequestURI,
				)
			case FormatFluentd:
				// Create fluentd message
				msg = fmt.Sprintf("| %3d | %13v | %15s | %-7s %s",
					res.Status,
					stop.Sub(start).String(),
					c.RealIP(),
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
