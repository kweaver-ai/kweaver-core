package opentelemetry

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
)

func minimalCfg() *conf.OtelConfig {
	return &conf.OtelConfig{
		ServiceName:    "test-svc",
		ServiceVersion: "v0.0.1",
		Trace:          conf.TraceConfig{Enabled: false},
		Log:            conf.LogConfig{Enabled: false},
		Metric:         conf.MetricConfig{Enabled: false},
	}
}

func TestNewProvider_AllDisabled(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(minimalCfg())
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.Nil(t, p.Tracer())
	assert.Nil(t, p.LoggerProvider())
	assert.Nil(t, p.MeterProvider())
}

func TestNewProvider_TraceConsole(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Trace.Enabled = true
	cfg.Trace.Exporter = "console"
	cfg.Trace.SamplingRate = 1.0

	p, err := NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.Tracer())

	_ = p.Shutdown(context.Background())
}

func TestNewProvider_LogConsole(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Log.Enabled = true
	cfg.Log.Exporter = "console"

	p, err := NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.LoggerProvider())

	_ = p.Shutdown(context.Background())
}

func TestNewProvider_MetricConsole(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Metric.Enabled = true
	cfg.Metric.Exporter = "console"
	cfg.Metric.ExportInterval = 10

	p, err := NewProvider(cfg)
	require.NoError(t, err)
	assert.NotNil(t, p)
	assert.NotNil(t, p.MeterProvider())

	_ = p.Shutdown(context.Background())
}

func TestNewProvider_UnsupportedTraceExporter(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Trace.Enabled = true
	cfg.Trace.Exporter = "unknown-exporter"
	cfg.Trace.SamplingRate = 1.0

	_, err := NewProvider(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported trace exporter")
}

func TestNewProvider_UnsupportedLogExporter(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Log.Enabled = true
	cfg.Log.Exporter = "unknown-exporter"

	_, err := NewProvider(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported log exporter")
}

func TestNewProvider_UnsupportedMetricExporter(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Metric.Enabled = true
	cfg.Metric.Exporter = "unknown-exporter"
	cfg.Metric.ExportInterval = 10

	_, err := NewProvider(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported metric exporter")
}

func TestProvider_Shutdown_NoProviders(t *testing.T) {
	t.Parallel()

	p, err := NewProvider(minimalCfg())
	require.NoError(t, err)

	err = p.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestProvider_Shutdown_WithTrace(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	cfg.Trace.Enabled = true
	cfg.Trace.Exporter = "console"
	cfg.Trace.SamplingRate = 1.0

	p, err := NewProvider(cfg)
	require.NoError(t, err)

	err = p.Shutdown(context.Background())
	require.NoError(t, err)
}

func TestCreateResource(t *testing.T) {
	t.Parallel()

	cfg := minimalCfg()
	res, err := createResource(cfg)
	require.NoError(t, err)
	assert.NotNil(t, res)
}
