package conf

import (
	"testing"

	otelcfg "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel"
	"github.com/stretchr/testify/assert"
)

// ==================== MQConf.IsDebug ====================

func TestMQConf_IsDebug(t *testing.T) {
	t.Parallel()

	c := MQConf{}
	assert.True(t, c.IsDebug())
}

// ==================== OtelV2Config.SetDefaults ====================

func TestOtelV2ConfigSetDefaults_EmptyConfig(t *testing.T) {
	t.Parallel()

	config := &otelcfg.OtelV2Config{}
	config.SetDefaults()

	assert.Equal(t, "agent-factory", config.ServiceName)
	assert.Equal(t, "1.0.0", config.ServiceVersion)
	assert.Equal(t, "production", config.Environment)
	assert.Equal(t, "localhost:4318", config.OTLPEndpoint)
	assert.Equal(t, 1.0, config.Trace.SamplingRate)
	assert.Equal(t, "info", config.Log.Level)
}

func TestOtelV2ConfigSetDefaults_PresetValues(t *testing.T) {
	t.Parallel()

	config := &otelcfg.OtelV2Config{
		ServiceName:    "my-service",
		ServiceVersion: "2.0.0",
		Environment:    "staging",
		OTLPEndpoint:   "otel:4318",
		Trace: otelcfg.TraceV2Conf{
			Enabled:      true,
			SamplingRate: 0.5,
		},
		Log: otelcfg.LogV2Conf{
			Enabled: true,
			Level:   "debug",
		},
	}
	config.SetDefaults()

	assert.Equal(t, "my-service", config.ServiceName)
	assert.Equal(t, "2.0.0", config.ServiceVersion)
	assert.Equal(t, "staging", config.Environment)
	assert.Equal(t, "otel:4318", config.OTLPEndpoint)
	assert.Equal(t, 0.5, config.Trace.SamplingRate)
	assert.Equal(t, "debug", config.Log.Level)
	assert.True(t, config.Trace.Enabled)
	assert.True(t, config.Log.Enabled)
}

func TestOtelV2ConfigSetDefaults_InvalidSamplingRate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rate float64
	}{
		{"zero", 0},
		{"negative", -0.5},
		{"above_one", 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			config := &otelcfg.OtelV2Config{
				Trace: otelcfg.TraceV2Conf{
					SamplingRate: tt.rate,
				},
			}
			config.SetDefaults()

			assert.Equal(t, 1.0, config.Trace.SamplingRate)
		})
	}
}
