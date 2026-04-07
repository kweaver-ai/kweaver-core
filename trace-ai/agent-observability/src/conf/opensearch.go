package conf

import (
	"os"
	"time"
)

type OpenSearchConfig struct {
	Endpoint   string
	TraceIndex string
	Timeout    time.Duration
	Auth       OpenSearchAuthConfig
}

type OpenSearchAuthConfig struct {
	Enabled  bool
	Username string
	Password string
}

func NewOpenSearchConfig() OpenSearchConfig {
	endpoint := os.Getenv("OPENSEARCH_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://127.0.0.1:9200"
	}

	traceIndex := os.Getenv("OPENSEARCH_TRACE_INDEX")
	if traceIndex == "" {
		traceIndex = "ss4o_traces-default-namespace"
	}

	return OpenSearchConfig{
		Endpoint:   endpoint,
		TraceIndex: traceIndex,
		Timeout:    3 * time.Second,
		Auth: OpenSearchAuthConfig{
			Enabled:  os.Getenv("OPENSEARCH_AUTH_ENABLED") == "true",
			Username: os.Getenv("OPENSEARCH_AUTH_USERNAME"),
			Password: os.Getenv("OPENSEARCH_AUTH_PASSWORD"),
		},
	}
}
