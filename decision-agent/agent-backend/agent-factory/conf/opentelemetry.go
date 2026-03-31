package conf

type OtelConfig struct {
	ServiceName    string       `yaml:"service_name"`
	ServiceVersion string       `yaml:"service_version"`
	Trace          TraceConfig  `yaml:"trace"`
	Log            LogConfig    `yaml:"log"`
	Metric         MetricConfig `yaml:"metric"`
}

// TraceConfig Trace 配置
type TraceConfig struct {
	Enabled      bool    `yaml:"enabled"`
	Exporter     string  `yaml:"exporter"` // console http
	HTTPEndpoint string  `yaml:"http_endpoint"`
	SamplingRate float64 `yaml:"sampling_rate"`
}

// LogConfig Log 配置
type LogConfig struct {
	Enabled      bool   `yaml:"enabled"`
	Exporter     string `yaml:"exporter"` // console http
	HTTPEndpoint string `yaml:"http_endpoint"`
	Level        string `yaml:"level"`
}

// MetricConfig Metric 配置
type MetricConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Exporter       string `yaml:"exporter"` // console http
	HTTPEndpoint   string `yaml:"http_endpoint"`
	ExportInterval int    `yaml:"export_interval"`
}

// setDefaults 设置配置默认值
func setOtelDefaults(config *OtelConfig) {
	if config.ServiceName == "" {
		config.ServiceName = "agent-factory"
	}

	if config.ServiceVersion == "" {
		config.ServiceVersion = "1.0.7"
	}
	// Trace 默认值
	if config.Trace.Exporter == "" {
		config.Trace.Exporter = "console"
	}

	if config.Trace.HTTPEndpoint == "" {
		config.Trace.HTTPEndpoint = "http://localhost:4318"
	}

	if config.Trace.SamplingRate <= 0 || config.Trace.SamplingRate > 1 {
		config.Trace.SamplingRate = 1.0
	}

	// Log 默认值
	if config.Log.Exporter == "" {
		config.Log.Exporter = "console"
	}

	if config.Log.HTTPEndpoint == "" {
		config.Log.HTTPEndpoint = "http://localhost:4318"
	}

	if config.Log.Level == "" {
		config.Log.Level = "info"
	}
	// Metric 默认值
	if config.Metric.Exporter == "" {
		config.Metric.Exporter = "console"
	}

	if config.Metric.HTTPEndpoint == "" {
		config.Metric.HTTPEndpoint = "http://localhost:4318"
	}

	if config.Metric.ExportInterval <= 0 {
		config.Metric.ExportInterval = 10
	}
}
