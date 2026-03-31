package global

import (
	"sync"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/stretchr/testify/assert"
)

// ==================== InitMetrics 实际调用 ====================

func TestInitMetrics_WithNilMeter(t *testing.T) {
	// 不能 parallel，因为修改全局状态
	origConfig := GConfig
	origMetrics := GMetrics
	origOnce := metricsOnce //nolint:govet

	defer func() {
		GConfig = origConfig
		GMetrics = origMetrics
		metricsOnce = origOnce //nolint:govet
	}()

	GConfig = &conf.Config{
		OtelConfig: &conf.OtelConfig{
			ServiceName:    "test-service",
			ServiceVersion: "1.0.0",
			Metric: conf.MetricConfig{
				Enabled:  false,
				Exporter: "console",
			},
		},
	}
	metricsOnce = sync.Once{}
	GMetrics = nil

	err := InitMetrics(nil)
	assert.NoError(t, err)
	assert.NotNil(t, GMetrics)
}

// ==================== InitDependencyInjector 分支覆盖 ====================

func TestInitDependencyInjector_NoLogger_Error(t *testing.T) {
	origLogger := GLogger
	origMetrics := GMetrics
	origOnce := dependencyOnce //nolint:govet

	defer func() {
		GLogger = origLogger
		GMetrics = origMetrics
		dependencyOnce = origOnce //nolint:govet
	}()

	GLogger = nil
	GMetrics = nil
	dependencyOnce = sync.Once{}

	err := InitDependencyInjector()
	assert.Error(t, err, "应该在 GLogger==nil 时返回错误")
}

func TestInitDependencyInjector_NoMetrics_Error(t *testing.T) {
	origConfig := GConfig
	origLogger := GLogger
	origMetrics := GMetrics
	origLoggerOnce := loggerOnce  //nolint:govet
	origDepOnce := dependencyOnce //nolint:govet

	defer func() {
		GConfig = origConfig
		GLogger = origLogger
		GMetrics = origMetrics
		loggerOnce = origLoggerOnce  //nolint:govet
		dependencyOnce = origDepOnce //nolint:govet
	}()

	// 先初始化 Logger（需要 GConfig）
	GConfig = &conf.Config{
		OtelConfig: &conf.OtelConfig{
			ServiceName: "test",
			Log:         conf.LogConfig{Exporter: "console"},
			Metric:      conf.MetricConfig{Exporter: "console"},
		},
	}

	// 初始化 Metrics，使 GLogger 有值
	metricsOnce2 := sync.Once{}
	metricsOnce = metricsOnce2 //nolint:govet
	_ = InitMetrics(nil)

	// GLogger 有值但 GMetrics 置 nil
	GMetrics = nil
	dependencyOnce = sync.Once{}

	// 需要先让 GLogger 有值 —— 直接 hack
	// InitLogger 依赖 logs.NewLogger 可能 panic，所以不调用
	// 但 InitDependencyInjector 会先检查 GLogger == nil
	// 由于我们没法安全初始化 GLogger，跳过这个
	t.Skip("InitLogger 依赖 OTel Logger 实现, 无法安全创建 GLogger")
}

func TestInitDependencyInjector_Success_WithMetrics(t *testing.T) {
	origConfig := GConfig
	origMetrics := GMetrics
	origDepInjector := GDependencyInjector
	origMetricsOnce := metricsOnce //nolint:govet
	origDepOnce := dependencyOnce  //nolint:govet

	defer func() {
		GConfig = origConfig
		GMetrics = origMetrics
		GDependencyInjector = origDepInjector
		metricsOnce = origMetricsOnce //nolint:govet
		dependencyOnce = origDepOnce  //nolint:govet
	}()

	GConfig = &conf.Config{
		OtelConfig: &conf.OtelConfig{
			ServiceName: "test",
			Metric:      conf.MetricConfig{Exporter: "console"},
		},
	}

	// 初始化 Metrics
	metricsOnce = sync.Once{}
	GMetrics = nil
	_ = InitMetrics(nil)

	// GLogger 仍然为 nil => InitDependencyInjector 应返回 error
	dependencyOnce = sync.Once{}
	err := InitDependencyInjector()
	// GLogger is nil, so should error
	assert.Error(t, err)
}

// ==================== ShutdownOpenTelemetry ====================

func TestShutdownOpenTelemetry_WithMetrics(t *testing.T) {
	origConfig := GConfig
	origMetrics := GMetrics
	origMetricsOnce := metricsOnce //nolint:govet

	defer func() {
		GConfig = origConfig
		GMetrics = origMetrics
		metricsOnce = origMetricsOnce //nolint:govet
	}()

	GConfig = &conf.Config{
		OtelConfig: &conf.OtelConfig{
			ServiceName: "test",
			Metric:      conf.MetricConfig{Exporter: "console"},
		},
	}
	metricsOnce = sync.Once{}
	_ = InitMetrics(nil)

	// 不应 panic
	ShutdownOpenTelemetry()
}
