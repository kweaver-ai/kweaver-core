package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/kweaver-core/adp/dataflow/flow-automation/libs/go/telemetry/log"
)

// VegaBackend 接口定义
type VegaBackend interface {
	WriteDatasetDocuments(ctx context.Context, datasetID string, documents []map[string]any, userID string, userType string) error
}

var (
	vegaBackendOnce     sync.Once
	vegaBackendInstance VegaBackend
)

// NewVegaBackend 创建 VegaBackend 单例
func NewVegaBackend() VegaBackend {
	vegaBackendOnce.Do(func() {
		config := common.NewConfig()
		// 使用内部 API 路径 /api/vega-backend/in
		baseURL := fmt.Sprintf("http://%s:%d/api/vega-backend/in",
			config.VegaBackendConfig.Host, config.VegaBackendConfig.Port)
		vegaBackendInstance = &vegaBackend{
			baseURL:    baseURL,
			httpClient: NewOtelHTTPClient(),
		}
	})
	return vegaBackendInstance
}

type vegaBackend struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

// WriteDatasetDocuments 向指定 dataset 写入文档
func (v *vegaBackend) WriteDatasetDocuments(ctx context.Context, datasetID string, documents []map[string]any, userID string, userType string) error {
	log := traceLog.WithContext(ctx)

	// 使用内部 API 路径: /api/vega-backend/in/v1/resources/:id/data
	src := fmt.Sprintf("%s/v1/resources/%s/data", v.baseURL, url.PathEscape(datasetID))
	headers := map[string]string{
		"Content-Type":            "application/json",
		"X-Account-ID":            userID,
		"X-Account-Type":          userType,
		"X-HTTP-Method-Override": http.MethodPost,
	}

	log.Infof("WriteDatasetDocuments: dataset_id=%s, documents=%d, url=%s, user_id=%s", datasetID, len(documents), src, userID)

	reqBytes, err := json.Marshal(documents)
	if err != nil {
		log.Warnf("WriteDatasetDocuments marshal failed: %v", err)
		return fmt.Errorf("WriteDatasetDocuments marshal failed: %w", err)
	}

	respCode, respData, err := v.httpClient.Request(ctx, src, http.MethodPost, headers, &reqBytes)
	if err != nil {
		log.Warnf("WriteDatasetDocuments request failed: %v", err)
		return fmt.Errorf("WriteDatasetDocuments request failed: %w", err)
	}

	if respCode != http.StatusCreated && respCode != http.StatusOK {
		log.Warnf("WriteDatasetDocuments failed: code=%d, body=%s", respCode, string(respData))
		return fmt.Errorf("WriteDatasetDocuments failed: %s", string(respData))
	}

	log.Infof("WriteDatasetDocuments success: dataset_id=%s, documents=%d", datasetID, len(documents))
	return nil
}
