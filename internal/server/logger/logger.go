package logger

import (
	"io"
	"os"

	"github.com/rs/zerolog"
)

type (
	Logger struct {
		zerolog.Logger
	}

	LoggerOptions struct {
		Debug bool
	}

	logLevel string
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

	return &Logger{logger}
}

func (l *Logger) SpawnForService(serviceName string) *Logger {
	return &Logger{l.With().Str("service", serviceName).Logger()}
}

func (l *Logger) SpawnForContext(reqCount string, reqID string) *Logger {
	return &Logger{l.With().Str("reqCount", reqCount).Str("reqID", reqID).Logger()}
}
