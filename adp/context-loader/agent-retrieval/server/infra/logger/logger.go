// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package logger
// @description: 定义日志接口
// @file logger.go
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// Level 日志level
type Level int

const (
	// LevelDebug debug
	LevelDebug Level = iota
	// LevelInfo info
	LevelInfo
	// LevelWarn warn
	LevelWarn
	// LevelError error
	LevelError

	levelDebugPrefix = "DEG"
	levelWarnPrefix  = "WAR"
	levelInfoPrefix  = "INF"
	levelErrorPrefix = "ERR"
)

// SimpleLogger 默认日志
type SimpleLogger struct {
	level     Level
	calldepth int
	writer    io.Writer
	log       *log.Logger
}

// DefaultLogger 默认log
func DefaultLogger() (l *SimpleLogger) {
	return NewLogger(LevelInfo, DefaultCalldepth)
}

const (
	DefaultCalldepth = 2
	MaxCalldepth     = 3
)

// NewLogger 新建logger
func NewLogger(level Level, calldepth int) (l *SimpleLogger) {
	l = &SimpleLogger{level: level, writer: os.Stdout, calldepth: calldepth}
	f := log.Ldate | log.Ltime | log.Lshortfile
	l.log = log.New(l.writer, "", f)
	return
}

func addLvl(lvl, str string) string {
	return lvl + " " + str
}

// WithContext 携带上下文
func (l *SimpleLogger) WithContext(ctx context.Context) interfaces.Logger { //nolint:revive
	// not support
	return l
}

// Output 输出
func (l *SimpleLogger) Output(calldepth int, s string) error {
	return log.Output(calldepth, s)
}

// Debug debug log
func (l *SimpleLogger) Debug(v ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	_ = l.log.Output(l.calldepth, addLvl(levelDebugPrefix, fmt.Sprint(v...)))
}

// Debugf debugf log
func (l *SimpleLogger) Debugf(format string, v ...interface{}) {
	if l.level > LevelDebug {
		return
	}
	_ = l.log.Output(l.calldepth, addLvl(levelDebugPrefix, fmt.Sprintf(format, v...)))
}

// Info info log
func (l *SimpleLogger) Info(v ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	_ = l.log.Output(l.calldepth, addLvl(levelInfoPrefix, fmt.Sprint(v...)))
}

// Infof infof log
func (l *SimpleLogger) Infof(format string, v ...interface{}) {
	if l.level > LevelInfo {
		return
	}
	_ = l.log.Output(l.calldepth, addLvl(levelInfoPrefix, fmt.Sprintf(format, v...)))
}

// Warn warn log
func (l *SimpleLogger) Warn(v ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	_ = l.log.Output(l.calldepth, addLvl(levelWarnPrefix, fmt.Sprint(v...)))
}

// Warnf warnf log
func (l *SimpleLogger) Warnf(format string, v ...interface{}) {
	if l.level > LevelWarn {
		return
	}
	_ = l.log.Output(l.calldepth, addLvl(levelWarnPrefix, fmt.Sprintf(format, v...)))
}

// Error error log
func (l *SimpleLogger) Error(v ...interface{}) {
	// error is the highest level
	_ = l.log.Output(l.calldepth, addLvl(levelErrorPrefix, fmt.Sprint(v...)))
}

// Errorf errorf log
func (l *SimpleLogger) Errorf(format string, v ...interface{}) {
	// error is the highest level
	_ = l.log.Output(l.calldepth, addLvl(levelErrorPrefix, fmt.Sprintf(format, v...)))
}
