package drivenadapters

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	commonLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/log"
	traceCommon "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/common"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
)

func init() {
	if commonLog.NewLogger() == nil {
		logout := "1"
		logDir := "/var/log/contentAutoMation/ut"
		logName := "contentAutoMation.log"
		commonLog.InitLogger(logout, logDir, logName)
	}
	traceLog.InitARLog(&traceCommon.TelemetryConf{LogLevel: "all"})
}

func TestVegaBackend_WriteDatasetDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v1/resources/dataset/test-dataset-id/docs", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// 使用 mock server URL
	client := &vegaBackend{baseURL: server.URL, httpClient: NewOtelHTTPClient()}
	err := client.WriteDatasetDocuments(context.Background(), "test-dataset-id", []map[string]any{
		{"id": "doc1", "name": "test"},
	})
	assert.Equal(t, nil, err)
}

func TestVegaBackend_WriteDatasetDocuments_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := &vegaBackend{baseURL: server.URL, httpClient: NewOtelHTTPClient()}
	err := client.WriteDatasetDocuments(context.Background(), "test-dataset-id", []map[string]any{})
	assert.NotEqual(t, nil, err)
}

func TestVegaBackend_WriteDatasetDocuments_Created(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := &vegaBackend{baseURL: server.URL, httpClient: NewOtelHTTPClient()}
	err := client.WriteDatasetDocuments(context.Background(), "test-dataset-id", []map[string]any{
		{"id": "doc1", "name": "test"},
	})
	assert.Equal(t, nil, err)
}
