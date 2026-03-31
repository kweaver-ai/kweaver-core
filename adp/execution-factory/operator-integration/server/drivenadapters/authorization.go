package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

type authorization struct {
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
	baseURL    string
}

var (
	authOnce sync.Once
	auth     *authorization
)

const (
	authOperationCheckURI = "/v1/operation-check"
	authResourceFilterURI = "/v1/resource-filter"
	authResourceListURI   = "/v1/resource-list"
	authCreatePolicyURI   = "/v1/policy"
	authDeletePolicyURI   = "/v1/policy-delete"
	authReaultFilterURI   = "/v1/resource-filter"
)

// NewAuthorization 创建鉴权服务对象
func NewAuthorization() interfaces.Authorization {
	authOnce.Do(func() {
		config := config.NewConfigLoader()
		auth = &authorization{
			logger:     config.GetLogger(),
			httpClient: rest.NewHTTPClient(),
			baseURL: fmt.Sprintf("%s://%s:%d/api/authorization",
				config.Authorization.PrivateProtocol,
				config.Authorization.PrivateHost,
				config.Authorization.PrivatePort),
		}
	})
	return auth
}

// OperationCheck 操作鉴权
func (a *authorization) OperationCheck(ctx context.Context, req *interfaces.AuthOperationCheckRequest) (resp *interfaces.AuthOperationCheckResponse, err error) {
	url := fmt.Sprintf("%s%s", a.baseURL, authOperationCheckURI)
	header := map[string]string{"Content-Type": "application/json"}
	_, respBody, err := a.httpClient.Post(ctx, url, header, req)
	if err != nil {
		a.logger.WithContext(ctx).Warnf("[OperationCheck] operation check failed, err: %v", err)
		return
	}
	resp = &interfaces.AuthOperationCheckResponse{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[OperationCheck] Unmarshal %s err:%v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// ResourceList 资源列举
func (a *authorization) ResourceList(ctx context.Context, req *interfaces.ResourceListRequest) (resp []*interfaces.AuthResourceResult, err error) {
	url := fmt.Sprintf("%s%s", a.baseURL, authResourceListURI)
	header := map[string]string{"Content-Type": "application/json"}
	_, respBody, err := a.httpClient.Post(ctx, url, header, req)
	if err != nil {
		a.logger.WithContext(ctx).Warnf("[ResourceList] resource list failed, err: %v", err)
		return nil, err
	}
	resp = []*interfaces.AuthResourceResult{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, &resp)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[ResourceList] Unmarshal %s err:%v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// ResourceFilter 资源过滤
func (a *authorization) ResourceFilter(ctx context.Context, req *interfaces.AuthResourceFilterRequest) (resp []*interfaces.AuthResourceResult, err error) {
	url := fmt.Sprintf("%s%s", a.baseURL, authResourceFilterURI)
	header := map[string]string{"Content-Type": "application/json"}
	_, respBody, err := a.httpClient.Post(ctx, url, header, req)
	if err != nil {
		a.logger.WithContext(ctx).Warnf("[ResourceFilter] resource filter failed, err: %v", err)
		return
	}
	resp = []*interfaces.AuthResourceResult{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, &resp)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[ResourceFilter] Unmarshal %s err:%v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// CreatePolicy 创建策略
func (a *authorization) CreatePolicy(ctx context.Context, req []*interfaces.AuthCreatePolicyRequest) (err error) {
	url := fmt.Sprintf("%s%s", a.baseURL, authCreatePolicyURI)
	header := map[string]string{"Content-Type": "application/json"}
	respCode, respData, err := a.httpClient.Post(ctx, url, header, req)
	if err != nil {
		a.logger.WithContext(ctx).Warnf("[CreatePolicy] create policy failed, err: %v", err)
		return
	}
	if respCode != http.StatusNoContent {
		err = infraErr.DefaultHTTPError(ctx, respCode, respData)
	}
	return
}

// DeletePolicy 删除策略
func (a *authorization) DeletePolicy(ctx context.Context, req *interfaces.AuthDeletePolicyRequest) (err error) {
	url := fmt.Sprintf("%s%s", a.baseURL, authDeletePolicyURI)
	header := map[string]string{"Content-Type": "application/json"}
	respCode, respBody, err := a.httpClient.Post(ctx, url, header, req)
	if err != nil {
		a.logger.WithContext(ctx).Warnf("[DeletePolicy] delete policy failed, err: %v", err)
		return
	}
	if respCode != http.StatusNoContent {
		a.logger.WithContext(ctx).Warnf("[DeletePolicy] delete policy failed, respCode: %d, respBody: %s", respCode, respBody)
		err = infraErr.DefaultHTTPError(ctx, respCode, respBody)
	}
	return
}
