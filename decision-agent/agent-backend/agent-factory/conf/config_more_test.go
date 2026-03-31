package conf

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ==================== MQConf.IsDebug ====================

func TestMQConf_IsDebug(t *testing.T) {
	t.Parallel()

	c := MQConf{}
	assert.True(t, c.IsDebug())
}

// ==================== setOtelDefaults ====================

func TestSetOtelDefaults_EmptyConfig(t *testing.T) {
	t.Parallel()

	config := &OtelConfig{}
	setOtelDefaults(config)

	assert.Equal(t, "agent-factory", config.ServiceName)
	assert.Equal(t, "1.0.7", config.ServiceVersion)

	assert.Equal(t, "console", config.Trace.Exporter)
	assert.Equal(t, "http://localhost:4318", config.Trace.HTTPEndpoint)
	assert.Equal(t, 1.0, config.Trace.SamplingRate)

	assert.Equal(t, "console", config.Log.Exporter)
	assert.Equal(t, "http://localhost:4318", config.Log.HTTPEndpoint)
	assert.Equal(t, "info", config.Log.Level)

	assert.Equal(t, "console", config.Metric.Exporter)
	assert.Equal(t, "http://localhost:4318", config.Metric.HTTPEndpoint)
	assert.Equal(t, 10, config.Metric.ExportInterval)
}

func TestSetOtelDefaults_PresetValues(t *testing.T) {
	t.Parallel()

	config := &OtelConfig{
		ServiceName:    "my-service",
		ServiceVersion: "2.0.0",
		Trace: TraceConfig{
			Exporter:     "http",
			HTTPEndpoint: "http://jaeger:4318",
			SamplingRate: 0.5,
		},
		Log: LogConfig{
			Exporter:     "http",
			HTTPEndpoint: "http://loki:4318",
			Level:        "debug",
		},
		Metric: MetricConfig{
			Exporter:       "http",
			HTTPEndpoint:   "http://prometheus:4318",
			ExportInterval: 30,
		},
	}
	setOtelDefaults(config)

	// Preset values should NOT be overwritten
	assert.Equal(t, "my-service", config.ServiceName)
	assert.Equal(t, "2.0.0", config.ServiceVersion)
	assert.Equal(t, "http", config.Trace.Exporter)
	assert.Equal(t, "http://jaeger:4318", config.Trace.HTTPEndpoint)
	assert.Equal(t, 0.5, config.Trace.SamplingRate)
	assert.Equal(t, "http", config.Log.Exporter)
	assert.Equal(t, "http://loki:4318", config.Log.HTTPEndpoint)
	assert.Equal(t, "debug", config.Log.Level)
	assert.Equal(t, "http", config.Metric.Exporter)
	assert.Equal(t, "http://prometheus:4318", config.Metric.HTTPEndpoint)
	assert.Equal(t, 30, config.Metric.ExportInterval)
}

func TestSetOtelDefaults_InvalidSamplingRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rate float64
	}{
		{"zero", 0},
		{"negative", -0.5},
		{"above one", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := &OtelConfig{
				Trace: TraceConfig{SamplingRate: tt.rate},
			}
			setOtelDefaults(config)
			assert.Equal(t, 1.0, config.Trace.SamplingRate)
		})
	}
}

func TestSetOtelDefaults_NegativeExportInterval(t *testing.T) {
	t.Parallel()

	config := &OtelConfig{
		Metric: MetricConfig{ExportInterval: -1},
	}
	setOtelDefaults(config)
	assert.Equal(t, 10, config.Metric.ExportInterval)
}
