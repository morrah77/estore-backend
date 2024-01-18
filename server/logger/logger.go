package logger

import (
	"log"
)

/*
*
A simplest logger generally compatible with https://github.com/sirupsen/logrus which is now deprecated
Packages like zerolog and zap have weird interface

TODO Consider to find and use some descent logging library instead of the meanness below!
*/
const (
	LOG_ERROR = iota
	LOG_WARN
	LOG_INFO
	LOG_DEBUG
)

var nameToLevel map[string]int = map[string]int{
	"error": LOG_ERROR,
	"ERROR": LOG_ERROR,
	"warn":  LOG_WARN,
	"WARN":  LOG_WARN,
	"info":  LOG_INFO,
	"INFO":  LOG_INFO,
	"debug": LOG_DEBUG,
	"DEBUG": LOG_DEBUG,
}

var logLevel int

type SimpleLogger func(s string, i ...interface{})

func (l SimpleLogger) Error(s string, i ...interface{}) {
	if logLevel < LOG_ERROR {
		return
	}
	l(s, i...)
}

func (l SimpleLogger) Warn(s string, i ...interface{}) {
	if logLevel < LOG_WARN {
		return
	}
	l(s, i...)
}

func (l SimpleLogger) Info(s string, i ...interface{}) {
	if logLevel < LOG_INFO {
		return
	}
	l(s, i...)
}

func (l SimpleLogger) Debug(s string, i ...interface{}) {
	if logLevel < LOG_DEBUG {
		return
	}
	l(s, i...)
}

func (l SimpleLogger) Log(logLevel int, s string, i ...interface{}) {
	l(s, i...)
}

var Log SimpleLogger = func(s string, i ...interface{}) {
	log.Printf(s, i...)
}

func (l SimpleLogger) WithLogLevel(level int) {
	logLevel = level
}

func (l SimpleLogger) SetLevel(level int) {
	logLevel = level
}

func New() SimpleLogger {
	return Log
}

func ParseLogLevel(level string) int {
	l, ok := nameToLevel[level]
	if !ok {
		panic("Unknown log level!")
	}
	return l
}
