package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

var (
	fOnce sync.Once
	fa    interfaces.FlowAutomation
)

type flowAutomationClient struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
}

// NewFlowAutomationClient 创建流程自动化服务对象
func NewFlowAutomationClient() interfaces.FlowAutomation {
	fOnce.Do(func() {
		conf := config.NewConfigLoader()
		fa = &flowAutomationClient{
			baseURL: fmt.Sprintf("%s://%s:%d/api/automation", conf.FlowAutomation.PrivateProtocol,
				conf.FlowAutomation.PrivateHost, conf.FlowAutomation.PrivatePort),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return fa
}

// Export 导出流程
// http://{host}:{port}/api/automation/v1/operators/configs/export
func (f *flowAutomationClient) Export(ctx context.Context, dagIDs []string) (resp *interfaces.FlowAutomationExportResp, err error) {
	src := fmt.Sprintf("%s/v1/operators/configs/export?id=%s", f.baseURL, strings.Join(dagIDs, ","))
	headers := map[string]string{}
	_, respData, err := f.httpClient.Get(ctx, src, nil, headers)
	if err != nil {
		f.logger.WithContext(ctx).Warnf("export flow request failed, err: %v", err)
		return
	}
	resp = &interfaces.FlowAutomationExportResp{}
	err = utils.AnyToObject(respData, resp)
	if err != nil {
		f.logger.WithContext(ctx).Warnf("export flow response parse failed, err: %v", err)
		err = errors.DefaultHTTPError(ctx, http.StatusInternalServerError, err)
	}
	return
}

// Import 导入流程
// http://{host}:{port}/api/automation/v1/operators/configs/import
func (f *flowAutomationClient) Import(ctx context.Context, req *interfaces.FlowAutomationImportReq, userID string) (err error) {
	src := fmt.Sprintf("%s/v1/operators/configs/import", f.baseURL)
	headers := map[string]string{}
	headers["X-User"] = userID
	_, _, err = f.httpClient.Put(ctx, src, headers, req)
	if err != nil {
		f.logger.WithContext(ctx).Warnf("import flow request failed, err: %v", err)
	}
	return
}
