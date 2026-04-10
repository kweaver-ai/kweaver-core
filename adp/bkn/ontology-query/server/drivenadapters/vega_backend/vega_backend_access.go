// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package vega_backend

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/common"
	"ontology-query/interfaces"
)

var (
	vbOnce sync.Once
	vb     interfaces.VegaBackendAccess
)

type vegaBackendAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
	baseURL    string
}

// NewVegaBackendAccess creates vega-backend client (aligned with bkn-backend).
func NewVegaBackendAccess(appSetting *common.AppSetting) interfaces.VegaBackendAccess {
	vbOnce.Do(func() {
		vb = &vegaBackendAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
			baseURL:    appSetting.VegaBackendUrl,
		}
	})
	return vb
}

func (v *vegaBackendAccess) buildHeaders(ctx context.Context) map[string]string {
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	// 用当前token的用户去访问vega
	return map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
}

func (v *vegaBackendAccess) QueryResourceData(ctx context.Context, resourceID string, params *interfaces.ResourceDataQueryParams) (*interfaces.DatasetQueryResponse, error) {
	httpURL := fmt.Sprintf("%s/resources/%s/data", v.baseURL, url.PathEscape(resourceID))
	headers := v.buildHeaders(ctx)
	headers[o11y.HTTP_HEADER_METHOD_OVERRIDE] = http.MethodGet

	respCode, respData, err := v.httpClient.PostNoUnmarshal(ctx, httpURL, headers, params)
	paramsJSON, _ := sonic.Marshal(params)
	logger.Debugf("QueryResourceData [%s] request [%s] code [%d] err [%v]", httpURL, string(paramsJSON), respCode, err)
	if err != nil {
		return nil, fmt.Errorf("QueryResourceData http request failed: %w", err)
	}
	if respCode != http.StatusOK {
		return nil, fmt.Errorf("QueryResourceData failed: %s", respData)
	}
	var response interfaces.DatasetQueryResponse
	if err := json.Unmarshal([]byte(respData), &response); err != nil {
		return nil, fmt.Errorf("unmarshal QueryResourceData response: %w", err)
	}
	return &response, nil
}
