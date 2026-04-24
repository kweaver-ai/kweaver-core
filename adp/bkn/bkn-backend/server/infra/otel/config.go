package otel

import "bkn-backend/version"

// OtelConfig 新版 OTel Collector 配置
type OtelConfig struct {
	ServiceName    string    `yaml:"service_name" mapstructure:"service_name"`
	ServiceVersion string    `yaml:"service_version" mapstructure:"service_version"`
	Environment    string    `yaml:"environment" mapstructure:"environment"`
	OTLPEndpoint   string    `yaml:"otlp_endpoint" mapstructure:"otlp_endpoint"` // e.g. "otel-collector:4318"
	Trace          TraceConf `yaml:"trace" mapstructure:"trace"`
	Log            LogConf   `yaml:"log" mapstructure:"log"`
}

// TraceConf trace 子配置
type TraceConf struct {
	Enabled      bool    `yaml:"enabled" mapstructure:"enabled"`
	SamplingRate float64 `yaml:"sampling_rate" mapstructure:"sampling_rate"`
}

// LogConf log 子配置
type LogConf struct {
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Level   string `yaml:"level" mapstructure:"level"`
}

// SetDefaults 设置 OtelConfig 默认值
func (c *OtelConfig) SetDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = version.ServerName
	}

	if c.ServiceVersion == "" {
		c.ServiceVersion = version.ServerVersion
	}

	if c.Environment == "" {
		c.Environment = "production"
	}

	if c.OTLPEndpoint == "" {
		c.OTLPEndpoint = "localhost:4318"
	}

	if c.Trace.SamplingRate <= 0 || c.Trace.SamplingRate > 1 {
		c.Trace.SamplingRate = 1.0
	}

	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
}
