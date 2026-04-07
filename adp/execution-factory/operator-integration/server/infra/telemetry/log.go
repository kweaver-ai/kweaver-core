// Package telemetry 可观测性相关包
package telemetry

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// LogExporterType 日志导出类型
type LogExporterType string

const (
	LogExporterTypeConsole LogExporterType = "console" // 控制台导出
	LogExporterTypeOTLP    LogExporterType = "http"    // http导出
)

// SamplerLogger 采样logger
type SamplerLogger struct {
	DefaultLogger interfaces.Logger
}

// NewSamplerLogger 创建logger
func NewSamplerLogger(defaultLogger interfaces.Logger) interfaces.Logger {
	s := &SamplerLogger{
		DefaultLogger: defaultLogger,
	}
	return s
}

func (l *SamplerLogger) Debug(v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Debug(v...)
}
func (l *SamplerLogger) Info(v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Info(v...)
}
func (l *SamplerLogger) Warn(v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Warn(v...)
}
func (l *SamplerLogger) Error(v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Error(v...)
}
func (l *SamplerLogger) Debugf(format string, v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Debugf(format, v...)
}
func (l *SamplerLogger) Infof(format string, v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Infof(format, v...)
}
func (l *SamplerLogger) Warnf(format string, v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Warnf(format, v...)
}
func (l *SamplerLogger) Errorf(format string, v ...interface{}) {
	s := &spanLogger{
		maker: l,
	}
	s.Errorf(format, v...)
}

// WithContext 传递context
func (l *SamplerLogger) WithContext(ctx context.Context) interfaces.Logger {
	return &spanLogger{
		ctx:   ctx,
		maker: l,
	}
}

type spanLogger struct {
	ctx   context.Context
	maker *SamplerLogger
}

func (s *spanLogger) Debug(v ...interface{}) {
	s.maker.DefaultLogger.Debug(v...)

	msg := fmt.Sprint(v...)
	if s.ctx != nil {
		o11y.Debug(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Debug(msg)
}
func (s *spanLogger) Info(v ...interface{}) {
	s.maker.DefaultLogger.Info(v...)
	msg := fmt.Sprint(v...)
	if s.ctx != nil {
		o11y.Info(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Info(msg)
}
func (s *spanLogger) Warn(v ...interface{}) {
	s.maker.DefaultLogger.Warn(v...)

	msg := fmt.Sprint(v...)
	if s.ctx != nil {
		o11y.Warn(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Warn(msg)
}
func (s *spanLogger) Error(v ...interface{}) {
	s.maker.DefaultLogger.Error(v...)

	msg := fmt.Sprint(v...)
	if s.ctx != nil {
		o11y.Error(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Error(msg)
}
func (s *spanLogger) Debugf(format string, v ...interface{}) {
	s.maker.DefaultLogger.Debugf(format, v...)

	msg := fmt.Sprintf(format, v...)
	if s.ctx != nil {
		o11y.Debug(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Debug(msg)
}
func (s *spanLogger) Infof(format string, v ...interface{}) {
	s.maker.DefaultLogger.Infof(format, v...)

	msg := fmt.Sprintf(format, v...)
	if s.ctx != nil {
		o11y.Info(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Info(msg)
}
func (s *spanLogger) Warnf(format string, v ...interface{}) {
	s.maker.DefaultLogger.Warnf(format, v...)

	msg := fmt.Sprintf(format, v...)
	if s.ctx != nil {
		o11y.Warn(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Warn(msg)
}
func (s *spanLogger) Errorf(format string, v ...interface{}) {
	s.maker.DefaultLogger.Errorf(format, v...)

	msg := fmt.Sprintf(format, v...)
	if s.ctx != nil {
		o11y.Error(s.ctx, msg)
		return
	}
	o11y.SystemLogger.Error(msg)
}

func (s *spanLogger) WithContext(ctx context.Context) interfaces.Logger {
	s.ctx = ctx
	return s
}
