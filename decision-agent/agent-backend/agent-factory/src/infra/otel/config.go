package otel

// OtelV2Config 新版 OTel Collector 配置
type OtelV2Config struct {
	ServiceName    string      `yaml:"service_name"`
	ServiceVersion string      `yaml:"service_version"`
	Environment    string      `yaml:"environment"`
	OTLPEndpoint   string      `yaml:"otlp_endpoint"` // e.g. "otel-collector:4318"
	Trace          TraceV2Conf `yaml:"trace"`
	Log            LogV2Conf   `yaml:"log"`
}

// TraceV2Conf trace 子配置
type TraceV2Conf struct {
	Enabled      bool    `yaml:"enabled"`
	SamplingRate float64 `yaml:"sampling_rate"`
}

// LogV2Conf log 子配置
type LogV2Conf struct {
	Enabled bool   `yaml:"enabled"`
	Level   string `yaml:"level"`
}

// SetDefaults 设置 OtelV2Config 默认值
func (c *OtelV2Config) SetDefaults() {
	if c.ServiceName == "" {
		c.ServiceName = "agent-factory"
	}

	if c.ServiceVersion == "" {
		c.ServiceVersion = "1.0.0"
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
