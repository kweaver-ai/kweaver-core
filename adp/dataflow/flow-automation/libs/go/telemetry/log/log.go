// Package log 日志文件
package log

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_log"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/public"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/resource"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/encoder"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/field"
	spanLog "github.com/kweaver-ai/TelemetrySDK-Go/span/v2/log"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/open_standard"
	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/runtime"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/common"
)

// Logger 日志服务，可适配其他日志组件
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

var (
	sl     spanLog.Logger
	slOnce sync.Once
)

// NewLogger 获取日志句柄
func WithContext(ctx context.Context) Logger {
	return &spanServerLog{ctx: ctx}
}

type spanServerLog struct {
	ctx context.Context
}

// Infof 普通信息
func (l *spanServerLog) Infof(format string, args ...interface{}) {
	sl.Info(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Infof(format, args...)
}

// Infoln 普通信息
func (l *spanServerLog) Infoln(args ...interface{}) {
	sl.Info(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Infoln(args...)
}

// Warnf 警告信息
func (l *spanServerLog) Warnf(format string, args ...interface{}) {
	sl.Warn(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Warnf(format, args...)
}

// Warnln 警告信息
func (l *spanServerLog) Warnln(args ...interface{}) {
	sl.Warn(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Warnln(args...)
}

// Errorf 错误信息
func (l *spanServerLog) Errorf(format string, args ...interface{}) {
	sl.Error(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Errorf(format, args...)
}

// Errorln 错误信息
func (l *spanServerLog) Errorln(args ...interface{}) {
	sl.Error(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Errorln(args...)
}

// Debugf 调试信息
func (l *spanServerLog) Debugf(format string, args ...interface{}) {
	sl.Debug(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Debugf(format, args...)
}

// Debugln 调试信息
func (l *spanServerLog) Debugln(args ...interface{}) {
	sl.Debug(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Debugln(args...)
}

// Tracef 跟踪信息
func (l *spanServerLog) Tracef(format string, args ...interface{}) {
	sl.Trace(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Tracef(format, args...)
}

// Traceln 跟踪信息
func (l *spanServerLog) Traceln(args ...interface{}) {
	sl.Trace(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Traceln(args...)
}

// Fatalf 致命错误
func (l *spanServerLog) Fatalf(format string, args ...interface{}) {
	sl.Fatal(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Fatalf(format, args...)
}

// Fatalln 致命错误
func (l *spanServerLog) Fatalln(args ...interface{}) {
	sl.Fatal(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Fatalln(args...)
}

// Panicf 恐慌错误
func (l *spanServerLog) Panicf(format string, args ...interface{}) {
	sl.Error(fmt.Sprintf(format, args...), field.WithContext(l.ctx))
	log.NewLogger().Panicf(format, args...)
}

// Panicln 恐慌错误
func (l *spanServerLog) Panicln(args ...interface{}) {
	sl.Error(contact(args...), field.WithContext(l.ctx))
	log.NewLogger().Panicln(args...)
}

// InitARLog 用于ut初始化使用
func InitARLog(config *common.TelemetryConf) {
	slOnce.Do(func() {
		// ARLogger 程序日志记录器，使用异步发送模式，无返回值。
		//spanLog.AllLevel
		logLevel := config.GetLogLevel()
		var ARLogger = spanLog.NewSamplerLogger(spanLog.WithSample(1.0), spanLog.WithLevel(logLevel))
		// 初始化ar_log
		hostName, _ := os.Hostname()
		public.SetServiceInfo(config.ServerName, config.ServerVersion, hostName)
		var systemLogClient public.Client
		if config.TraceURL == "" { // 没上报地址就 输出到标准输出, 一般用于无环境调试
			systemLogClient = public.NewConsoleClient()
		} else {
			systemLogClient = public.NewHTTPClient(public.WithAnyRobotURL(config.LogURL),
				public.WithCompression(1), public.WithTimeout(10*time.Second),
				public.WithRetry(true, 5*time.Second, 30*time.Second, 1*time.Minute))
		}

		systemLogExporter := ar_log.NewExporter(systemLogClient)
		systemLogWriter := open_standard.OpenTelemetryWriter(
			encoder.NewJsonEncoderWithExporters(systemLogExporter),
			resource.LogResource())
		systemLogRunner := runtime.NewRuntime(systemLogWriter, field.NewSpanFromPool)
		systemLogRunner.SetUploadInternalAndMaxLog(3*time.Second, 10)
		// 运行SystemLogger日志器。
		go systemLogRunner.Run()

		ARLogger.SetLevel(logLevel)
		ARLogger.SetRuntime(systemLogRunner)
		sl = ARLogger
	})
}

// InitLogger 初始化logger
func InitLogger(serverName string) {
	ar_log.InitLogger("cm", "anyshare-telemetry-sdk", serverName)
	sl = ar_log.Logger
}

// Close 释放ar_log实例
func Close() {
	sl.Close()
}

func contact(args ...interface{}) string {
	var res []string
	for _, val := range args {
		res = append(res, fmt.Sprintf("%v", val))
	}
	contactRes := strings.Join(res, " ")
	return contactRes
}
