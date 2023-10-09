package logger

import (
	"io"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type (
	Logger struct {
		*zerolog.Logger
	}

	LoggerOptions struct {
		Debug bool
	}

	logLevel string

	// LoggingFn is a generic logging function with some context used by handlers when invoking services
	LoggingFn func(level logLevel) *zerolog.Event
)

const (
	DebugLevel logLevel = "DEBUG"
	InfoLevel  logLevel = "INFO"
	WarnLevel  logLevel = "WARN"
	ErrorLevel logLevel = "ERROR"
)

func NewLogger(options LoggerOptions) *Logger {
	var output io.Writer = os.Stdout
	var logLevel = zerolog.InfoLevel
	if options.Debug {
		logLevel = zerolog.DebugLevel
	}

	zerolog.TimestampFieldName = "T"
	zerolog.LevelFieldName = "L"
	zerolog.MessageFieldName = "M"
	zerolog.LevelDebugValue = string(DebugLevel)
	zerolog.LevelInfoValue = string(InfoLevel)
	zerolog.LevelWarnValue = string(WarnLevel)
	zerolog.LevelErrorValue = string(ErrorLevel)

	logger := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		Logger()

	return &Logger{&logger}
}

/*
ContextLoggingFn creates a LoggingFnz to be used in
places that do not necessarily need access to the gin context (deafult log level is info)
*/
func (logger *Logger) ContextLoggingFn(c *gin.Context) LoggingFn {
	return func(level logLevel) *zerolog.Event {
		switch level {
		case DebugLevel:
			return logger.Debugc(c)
		case InfoLevel:
			return logger.Infoc(c)
		case WarnLevel:
			return logger.Warnc(c)
		case ErrorLevel:
			return logger.Errorc(c)
		default:
			return logger.Infoc(c)
		}
	}
}

func (l *Logger) Infoc(c *gin.Context) *zerolog.Event {
	return injectContextVars(c, l.Info())
}

func (l *Logger) Warnc(c *gin.Context) *zerolog.Event {
	return injectContextVars(c, l.Warn())
}

func (l *Logger) Errorc(c *gin.Context) *zerolog.Event {
	return injectContextVars(c, l.Error())
}
func (l *Logger) Debugc(c *gin.Context) *zerolog.Event {
	return injectContextVars(c, l.Debug())
}

func injectContextVars(c *gin.Context, e *zerolog.Event) *zerolog.Event {
	if reqCount, exists := c.Get("requestcount"); exists {
		e.Interface("reqCount", reqCount)
	}
	if reqID, exists := c.Get("requestid"); exists {
		e.Interface("reqID", reqID)
	}
	return e
}
