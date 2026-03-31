package global

import (
	"context"
	"sync"

	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/metric"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry/logs"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry/metrics"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/opentelemetry/middleware"

	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

var (
	GConfig             *conf.Config                   // 全局配置
	GDB                 *sqlx.DB                       // 全局 DB
	GLogger             *logs.Logger                   // 全局Logger
	GMetrics            *metrics.Metrics               // 全局Metrics
	GDependencyInjector *middleware.DependencyInjector // 全局DependencyInjector

	loggerOnce     sync.Once
	metricsOnce    sync.Once
	dependencyOnce sync.Once
)

// InitLogger 初始化全局Logger（线程安全）
func InitLogger(otelLogger otellog.Logger) error {
	var initErr error

	loggerOnce.Do(func() {
		var err error

		GLogger, err = logs.NewLogger(GConfig.OtelConfig, otelLogger)
		if err != nil {
			initErr = err
		}
	})

	return initErr
}

// GetLogger 获取全局Logger，如果未初始化则返回nil
func GetLogger() *logs.Logger {
	return GLogger
}

// InitMetrics 初始化全局Metrics（线程安全）
func InitMetrics(meter metric.Meter) error {
	var initErr error

	metricsOnce.Do(func() {
		var err error

		GMetrics, err = metrics.NewMetrics(GConfig.OtelConfig, meter)
		if err != nil {
			initErr = err
		}
	})

	return initErr
}

// GetMetrics 获取全局Metrics，如果未初始化则返回nil
func GetMetrics() *metrics.Metrics {
	return GMetrics
}

// InitDependencyInjector 初始化全局DependencyInjector（线程安全）
func InitDependencyInjector() error {
	var initErr error

	dependencyOnce.Do(func() {
		if GLogger == nil {
			initErr = context.DeadlineExceeded // 使用一个错误表示Logger未初始化
			return
		}

		if GMetrics == nil {
			initErr = context.DeadlineExceeded // 使用一个错误表示Metrics未初始化
			return
		}

		GDependencyInjector = middleware.NewDependencyInjector(GLogger, GMetrics)
	})

	return initErr
}

// GetDependencyInjector 获取全局DependencyInjector，如果未初始化则返回nil
func GetDependencyInjector() *middleware.DependencyInjector {
	return GDependencyInjector
}

// InitOpenTelemetry 初始化所有OpenTelemetry相关全局变量
func InitOpenTelemetry(otelProvider *opentelemetry.Provider) error {
	// 初始化Logger
	var otelLogger otellog.Logger
	if lp := otelProvider.LoggerProvider(); lp != nil {
		otelLogger = lp.Logger(
			GConfig.OtelConfig.ServiceName,
			otellog.WithInstrumentationVersion(GConfig.OtelConfig.ServiceVersion),
		)
	}
	// 如果otelLogger为nil，InitLogger会处理nil情况
	if err := InitLogger(otelLogger); err != nil {
		return err
	}

	// 初始化Metrics
	var meter metric.Meter
	if mp := otelProvider.MeterProvider(); mp != nil {
		meter = mp.Meter(
			GConfig.OtelConfig.ServiceName,
			metric.WithInstrumentationVersion(GConfig.OtelConfig.ServiceVersion),
		)
	}
	// 如果meter为nil，InitMetrics会处理nil情况
	if err := InitMetrics(meter); err != nil {
		return err
	}

	// 初始化DependencyInjector
	if err := InitDependencyInjector(); err != nil {
		return err
	}

	return nil
}

// ShutdownOpenTelemetry 关闭OpenTelemetry相关资源
func ShutdownOpenTelemetry() {
	// Logger和Metrics当前没有显式的关闭方法
	// 这里可以添加其他清理逻辑
}
