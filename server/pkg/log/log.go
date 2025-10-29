package log

import (
	"log"
	"os"
)

// Logger provides structured logging
type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	Fatal(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Debugf(format string, args ...interface{})
}

type logger struct {
	debug bool
	*log.Logger
}

// New creates a new logger
func New(debug bool) Logger {
	return &logger{
		debug:  debug,
		Logger: log.New(os.Stderr, "[NerveCenter] ", log.LstdFlags),
	}
}

func (l *logger) Debug(format string, args ...interface{}) {
	if l.debug {
		l.Printf("[DEBUG] "+format, args...)
	}
}

func (l *logger) Info(format string, args ...interface{}) {
	l.Printf("[INFO] "+format, args...)
}

func (l *logger) Error(format string, args ...interface{}) {
	l.Printf("[ERROR] "+format, args...)
}

func (l *logger) Fatal(format string, args ...interface{}) {
	l.Error(format, args...)
	os.Exit(1)
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.Errorf(format, args...)
	os.Exit(1)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.Info(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.Error(format, args...)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.Debug(format, args...)
}

