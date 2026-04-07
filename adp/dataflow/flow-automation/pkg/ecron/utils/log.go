package utils

import (
	"fmt"
	"os"
	"path"
	"sync"

	"github.com/sirupsen/logrus"
)

//go:generate mockgen -package mock -source ../utils/log.go -destination ../mock/mock_log.go

// Logger 定时服务日志服务，可适配其他日志组件
type Logger interface {
	Infof(format string, args ...interface{})
	Infoln(args ...interface{})
	Debugf(format string, args ...interface{})
	Debugln(args ...interface{})
	Errorf(format string, args ...interface{})
	Errorln(args ...interface{})
	Warnf(format string, args ...interface{})
	Warnln(args ...interface{})
	Tracef(format string, args ...interface{})
	Traceln(args ...interface{})
	Panicf(format string, args ...interface{})
	Panicln(args ...interface{})
	Fatalf(format string, args ...interface{})
	Fatalln(args ...interface{})
}

// NewLogger 获取日志句柄
func NewLogger() Logger {
	return newLogger()
}

type ecronLog struct {
	logger *logrus.Logger
}

var (
	logHandle *ecronLog   = &ecronLog{}
	logMutex  *sync.Mutex = new(sync.Mutex)
)

func newLogger() *ecronLog {
	logMutex.Lock()
	defer logMutex.Unlock()

	if nil != logHandle && nil != logHandle.logger {
		return logHandle
	}

	if nil == logHandle.logger {
		logHandle.logger = logrus.New()
		logHandle.logger.SetFormatter(&logrus.JSONFormatter{})
		logout := os.Getenv("LOGOUT")
		if len(logout) > 0 {
			logHandle.logger.SetOutput(os.Stdout)
		} else {
			logDir := "/var/log/ecron/"
			logName := "ecron.log"
			err := os.MkdirAll(logDir, 0755)
			if nil != err {
				fmt.Println("mkdir err", err)
				logHandle.logger.SetOutput(os.Stdout)
				return logHandle
			}
			logFileName := path.Join(logDir, logName)
			logFile, err := os.OpenFile(logFileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0)
			if nil != err {
				fmt.Println("open log file err", err)
				logHandle.logger.SetOutput(os.Stdout)
				return logHandle
			}
			logHandle.logger.SetOutput(logFile)
		}
	}

	return logHandle
}

// Infof 普通信息
func (l *ecronLog) Infof(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Infof(format, args...)
}

// Infoln 普通信息
func (l *ecronLog) Infoln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Infoln(args...)
}

// Warnf 警告信息
func (l *ecronLog) Warnf(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Warnf(format, args...)
}

// Warnln 警告信息
func (l *ecronLog) Warnln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Warnln(args...)
}

// Errorf 错误信息
func (l *ecronLog) Errorf(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Errorf(format, args...)
}

// Errorln 错误信息
func (l *ecronLog) Errorln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Errorln(args...)
}

// Debugf 调试信息
func (l *ecronLog) Debugf(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Debugf(format, args...)
}

// Debugln 调试信息
func (l *ecronLog) Debugln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Debugln(args...)
}

// Tracef 跟踪信息
func (l *ecronLog) Tracef(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Tracef(format, args...)
}

// Traceln 跟踪信息
func (l *ecronLog) Traceln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Traceln(args...)
}

// Fatalf 致命错误
func (l *ecronLog) Fatalf(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Fatalf(format, args...)
}

// Fatalln 致命错误
func (l *ecronLog) Fatalln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Fatalln(args...)
}

// Panicf 恐慌错误
func (l *ecronLog) Panicf(format string, args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Panicf(format, args...)
}

// Panicln 恐慌错误
func (l *ecronLog) Panicln(args ...interface{}) {
	if nil == l.logger {
		return
	}
	l.logger.Panicln(args...)
}
