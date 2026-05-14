package drivenadapters

import (
	"context"
	"encoding/json"
	"io"
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
		assert.Equal(t, "/v1/resources/test-dataset-id/data", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-user-id", r.Header.Get("X-Account-ID"))
		assert.Equal(t, "user", r.Header.Get("X-Account-Type"))
		assert.Equal(t, http.MethodPost, r.Header.Get("X-HTTP-Method-Override"))

		body, err := io.ReadAll(r.Body)
		assert.Equal(t, nil, err)
		defer r.Body.Close()

		var payload map[string]any
		err = json.Unmarshal(body, &payload)
		if err == nil {
			t.Fatalf("expected array body, got object: %#v", payload)
		}

		var documents []map[string]any
		err = json.Unmarshal(body, &documents)
		assert.Equal(t, nil, err)
		assert.Equal(t, 1, len(documents))
		assert.Equal(t, "doc1", documents[0]["id"])
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	// 使用 mock server URL
	client := &vegaBackend{baseURL: server.URL, httpClient: NewOtelHTTPClient()}
	err := client.WriteDatasetDocuments(context.Background(), "test-dataset-id", []map[string]any{
		{"id": "doc1", "name": "test"},
	}, "test-user-id", "user")
	assert.Equal(t, nil, err)
}

func TestVegaBackend_WriteDatasetDocuments_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := &vegaBackend{baseURL: server.URL, httpClient: NewOtelHTTPClient()}
	err := client.WriteDatasetDocuments(context.Background(), "test-dataset-id", []map[string]any{}, "test-user-id", "user")
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
	}, "test-user-id", "user")
	assert.Equal(t, nil, err)
}
