package log

import (
	"fmt"
	"log"
	"os"
)

var defLog Logger = &StdoutLogger{}

// SetLogger for testing
func SetLogger(log Logger) {
	defLog = log
}

// default logger
type StdoutLogger struct {
}

// Debug
func (s *StdoutLogger) Debug(msg string, fields ...interface{}) {
	log.Println("debug:", msg, fields)
}

// Debugf
func (s *StdoutLogger) Debugf(msg string, args ...interface{}) {
	log.Println("debug:", fmt.Sprintf(msg, args...))
}

// Info
func (s *StdoutLogger) Info(msg string, fields ...interface{}) {
	log.Println("info:", msg, fields)
}

// Infof
func (s *StdoutLogger) Infof(msg string, args ...interface{}) {
	log.Println("info:", fmt.Sprintf(msg, args...))
}

// Warn
func (s *StdoutLogger) Warn(msg string, fields ...interface{}) {
	log.Println("warn:", msg, fields)
}

// Warnf
func (s *StdoutLogger) Warnf(msg string, args ...interface{}) {
	log.Println("warn:", fmt.Sprintf(msg, args...))
}

// Error
func (s *StdoutLogger) Error(msg string, fields ...interface{}) {
	log.Println("error:", msg, fields)
}

// Errorf
func (s *StdoutLogger) Errorf(msg string, args ...interface{}) {
	log.Println("error:", fmt.Sprintf(msg, args...))
}

// Fatal
func (s *StdoutLogger) Fatal(msg string, fields ...interface{}) {
	log.Println("fatal:", msg, fields)
	os.Exit(1)
}

// Fatalf
func (s *StdoutLogger) Fatalf(msg string, args ...interface{}) {
	log.Println("fatal:", fmt.Sprintf(msg, args...))
	os.Exit(1)
}

// Logger
type Logger interface {
	Debug(msg string, fields ...interface{})
	Debugf(msg string, args ...interface{})
	Info(msg string, fields ...interface{})
	Infof(msg string, args ...interface{})
	Warn(msg string, fields ...interface{})
	Warnf(msg string, args ...interface{})
	Error(msg string, fields ...interface{})
	Errorf(msg string, args ...interface{})
	Fatal(msg string, fields ...interface{})
	Fatalf(msg string, args ...interface{})
}

// Debug
func Debug(msg string, fields ...interface{}) {
	defLog.Debug(msg, fields...)
}

// Debugf
func Debugf(msg string, args ...interface{}) {
	defLog.Debugf(msg, args...)
}

// Info
func Info(msg string, fields ...interface{}) {
	defLog.Info(msg, fields...)
}

// Infof
func Infof(msg string, args ...interface{}) {
	defLog.Infof(msg, args...)
}

// Warn
func Warn(msg string, fields ...interface{}) {
	defLog.Warn(msg, fields...)
}

// Warnf
func Warnf(msg string, args ...interface{}) {
	defLog.Warnf(msg, args...)
}

// Error
func Error(msg string, fields ...interface{}) {
	defLog.Error(msg, fields...)
}

// Errorf
func Errorf(msg string, args ...interface{}) {
	defLog.Errorf(msg, args...)
}

// Fatal
func Fatal(msg string, fields ...interface{}) {
	defLog.Fatal(msg, fields...)
}

// Fatalf
func Fatalf(msg string, args ...interface{}) {
	defLog.Fatalf(msg, args...)
}
