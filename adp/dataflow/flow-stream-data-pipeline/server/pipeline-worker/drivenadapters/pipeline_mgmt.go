package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-worker/interfaces"
)

var (
	clientOnce sync.Once
	client     interfaces.PipelineMgmtAccess
)

type pipelineMgmtAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewPipelineMgmt(appSetting *common.AppSetting) interfaces.PipelineMgmtAccess {
	clientOnce.Do(func() {
		client = &pipelineMgmtAccess{
			appSetting: appSetting,
			// 设置超时时间 40s
			httpClient: common.NewHTTPClientWithOptions(rest.HttpClientOptions{TimeOut: 40}),
		}
	})
	return client
}

// 请求 flow-stream-data-pipeline, 获取pipelineInfo
func (pmAccess *pipelineMgmtAccess) GetConfigs(ctx context.Context, pipelineID string, isListen bool) (*interfaces.Pipeline, bool, error) {
	urlStr := fmt.Sprintf("%s/%s", pmAccess.appSetting.PipelineMgmtUrl, pipelineID)
	var queryValues url.Values = make(url.Values)
	queryValues.Add("is_listen", fmt.Sprintf("%t", isListen))

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	resCode, respData, err := pmAccess.httpClient.GetNoUnmarshal(ctx, urlStr, queryValues, headers)
	logger.Debugf("post [%s] finished, request headers is [%v], response code is [%d], result is [%s], error is [%v]",
		urlStr, headers, resCode, respData, err)
	if err != nil {
		logger.Errorf("failed to get pipeline %s info, error: %s", pipelineID, err.Error())
		return nil, false, err
	}

	if resCode == http.StatusNotFound {
		return nil, false, nil
	}

	if resCode != http.StatusOK {
		var baseError rest.BaseError
		if err = sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err.Error())
			return nil, false, err
		}

		return nil, false, fmt.Errorf("get pipeline %s info failed, error: %s, %v",
			pipelineID, baseError.Description, baseError.ErrorDetails)
	}

	var pipeline interfaces.Pipeline
	err = sonic.Unmarshal(respData, &pipeline)
	if err != nil {
		logger.Errorf("umarshal get pipeline %s info failed, error: %s", pipelineID, err.Error())
		return nil, false, err
	}

	return &pipeline, true, nil
}

// 请求 flow-stream-data-pipeline，更新任务状态
func (pmAccess *pipelineMgmtAccess) UpdatePipelineStatus(ctx context.Context, pipelineID string, status *interfaces.PipelineStatusInfo) error {
	url := fmt.Sprintf("%s/%s/attrs/status,status_details?is_inner_request=true", pmAccess.appSetting.PipelineMgmtUrl, pipelineID)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, resp, err := pmAccess.httpClient.PutNoUnmarshal(ctx, url, headers, status)
	if err != nil {
		return fmt.Errorf("http request for updating pipeline status '%s' failed, %s", pipelineID, err.Error())
	}

	if respCode != http.StatusNoContent {
		var baseError *rest.BaseError
		err = sonic.Unmarshal(resp, &baseError)
		if err != nil {
			return fmt.Errorf("unmarshal baseError failed, %s", err.Error())
		}

		return fmt.Errorf("update pipeline status '%s' failed, errDetails: %v", pipelineID, baseError.ErrorDetails)
	}

	return nil
}
