package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
)

type Logger interface {
	Debug(msg string, args ...interface{})
	Info(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
	Error(msg string, args ...interface{})
}

type SlogWrapper struct {
	logger *slog.Logger
}

func (s *SlogWrapper) Debug(msg string, args ...interface{}) { s.logger.Debug(msg, args...) }
func (s *SlogWrapper) Info(msg string, args ...interface{})  { s.logger.Info(msg, args...) }
func (s *SlogWrapper) Warn(msg string, args ...interface{})  { s.logger.Warn(msg, args...) }
func (s *SlogWrapper) Error(msg string, args ...interface{}) { s.logger.Error(msg, args...) }

func SetupLogger(logFileName, level string, prod bool) (Logger, func()) {

	var logLevel slog.Level

	switch strings.ToUpper(level) {
	case "DEBUG":
		logLevel = slog.LevelDebug
	case "INFO":
		logLevel = slog.LevelInfo
	case "WARN":
		logLevel = slog.LevelWarn
	case "ERROR":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelError
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	logOutput, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("failed to open log file: ", err)
		panic(err)
	}

	mw := io.MultiWriter(logOutput, os.Stdout)
	handler := slog.NewTextHandler(mw, opts)

	if prod {
		handler = slog.NewTextHandler(logOutput, opts)
	}

	logger := slog.New(handler)

	slogWrapper := &SlogWrapper{logger}

	// constructor - cleanup
	return slogWrapper, func() {
		fmt.Println("closing log file")
		logOutput.Close()
	}
}

