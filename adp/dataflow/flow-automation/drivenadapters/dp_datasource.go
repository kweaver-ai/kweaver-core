package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/common"
	otelHttp "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/http"
	traceLog "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/telemetry/log"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
)

//go:generate mockgen -package mock_drivenadapters -source ../drivenadapters/dp_datasource.go -destination ../tests/mock_drivenadapters/dp_datasource_mock.go

// DataSourceBinData 数据源连接信息
type DataSourceBinData struct {
	CatalogName     string `json:"catalog_name"`
	DataViewSource  string `json:"data_view_source"`
	DatabaseName    string `json:"database_name"`
	Schema          string `json:"schema"`
	ConnectProtocol string `json:"connect_protocol"`
	Host            string `json:"host"`
	Port            int    `json:"port"`
	Account         string `json:"account"`
	Password        string `json:"password"`
}

// DataSourceCatalog 数据源目录信息
type DataSourceCatalog struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Type              string             `json:"type"`
	BinData           *DataSourceBinData `json:"bin_data"`
	Comment           string             `json:"comment"`
	Operations        []string           `json:"operations"`
	CreatedByUID      string             `json:"created_by_uid"`
	CreatedByUsername string             `json:"created_by_username"`
	CreatedAt         int64              `json:"created_at"`
	UpdatedByUID      string             `json:"updated_by_uid"`
	UpdatedByUsername string             `json:"updated_by_username"`
	UpdatedAt         int64              `json:"updated_at"`
}

// IDataSource 接口定义
type IDataSource interface {
	GetDataSourceCatalog(ctx context.Context, dataSourceID, token, ipStr string) (*DataSourceCatalog, error)
	GetBaseURL() string
}

// DataSourceImpl 数据源实现结构体
type DataSourceImpl struct {
	baseURL    string
	httpClient otelHttp.HTTPClient
}

var (
	dataSourceOnce sync.Once
	ds             IDataSource
)

// NewDataSource 创建数据源实例
func NewDataSource() IDataSource {
	dataSourceOnce.Do(func() {
		config := common.NewConfig()
		baseURL := ""

		if config.DPDataSource.Host != "" {
			baseURL = fmt.Sprintf("%s://%s:%d", config.DPDataSource.Protocol, config.DPDataSource.Host, config.DPDataSource.Port)
		}

		ds = &DataSourceImpl{
			baseURL:    baseURL,
			httpClient: NewOtelHTTPClient(),
		}
	})
	return ds
}

// GetDataSourceCatalog 获取数据源目录信息
func (d *DataSourceImpl) GetDataSourceCatalog(ctx context.Context, dataSourceID, token, ipStr string) (*DataSourceCatalog, error) {
	// 构建请求URL
	target := fmt.Sprintf("%s/api/data-connection/v1/datasource/%s", d.baseURL, dataSourceID)

	// 构建请求头
	headers := map[string]string{
		"content-type":    "application/json;charset=UTF-8",
		"accept-language": "zh-CN",
		"Authorization":   utils.TokenParser(token), // 直接添加token到headers
		"X-Forwarded-For": ipStr,
	}

	// 发送GET请求
	status, resp, err := d.httpClient.Get(ctx, target, headers)

	if err != nil {
		traceLog.WithContext(ctx).Errorf("GetDataSourceCatalog request failed: %v, url: %s", err, target)
		return nil, err
	}

	if status != http.StatusOK {
		traceLog.WithContext(ctx).Errorf("GetDataSourceCatalog unexpected status code: %d, url: %s", status, target)
		return nil, fmt.Errorf("unexpected status code: %d", status)
	}

	// 解析响应
	var response DataSourceCatalog
	bytes, _ := json.Marshal(resp)
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		traceLog.WithContext(ctx).Errorf("GetDataSourceCatalog decode response error: %v", err)
		return nil, err
	}

	return &response, nil
}

// GetBaseURL 获取基础URL
func (d *DataSourceImpl) GetBaseURL() string {
	return d.baseURL
}
