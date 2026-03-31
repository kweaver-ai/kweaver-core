// Package drivenadapters 定义驱动适配器
// @file business_domain_management.go
// @description: 实现业务域管理服务
package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	jsoniter "github.com/json-iterator/go"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/utils"
)

var (
	bdOnce sync.Once
	bdm    interfaces.BusinessDomainManagement
)

type businessDomainManagementClient struct {
	baseURL    string
	logger     interfaces.Logger
	httpClient interfaces.HTTPClient
}

// NewBusinessDomainManagementClient 创建业务域管理服务对象
func NewBusinessDomainManagementClient() interfaces.BusinessDomainManagement {
	bdOnce.Do(func() {
		conf := config.NewConfigLoader()
		bdm = &businessDomainManagementClient{
			baseURL: fmt.Sprintf("%s://%s:%d/internal/api/business-system/v1", conf.BusinessDomainManagement.PrivateProtocol,
				conf.BusinessDomainManagement.PrivateHost, conf.BusinessDomainManagement.PrivatePort),
			logger:     conf.GetLogger(),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return bdm
}

// AssociateResource 关联资源到业务域
func (b *businessDomainManagementClient) AssociateResource(ctx context.Context, req *interfaces.BusinessDomainResourceAssociateRequest) error {
	src := fmt.Sprintf("%s/resource", b.baseURL)
	// header := common.GetHeaderFromCtx(ctx)
	header := map[string]string{}

	respCode, _, err := b.httpClient.Post(ctx, src, header, req)

	// 处理 403 权限不足
	if respCode == http.StatusForbidden {
		b.logger.Errorf("businessDomainManagementClient#AssociateResource failed:%v, url:%v", err, src)
		err = infraErr.NewHTTPError(ctx, http.StatusForbidden, infraErr.ErrExtBusinessDomainForbidden, err.Error())
		return err
	}

	// 处理 409 资源已关联冲突
	if respCode == http.StatusConflict {
		b.logger.Warnf("businessDomainManagementClient#AssociateResource conflict: resource already connected, resource_id:%s, domain_id:%s", req.ID, req.BDID)
		err = infraErr.NewHTTPError(ctx, http.StatusConflict, infraErr.ErrExtBusinessDomainResourceConflict, err.Error())
		return err
	}

	if err != nil {
		b.logger.Errorf("businessDomainManagementClient#AssociateResource failed:%v, url:%v", err, src)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return err
	}

	b.logger.Infof("businessDomainManagementClient#AssociateResource success, resource_id:%s, domain_id:%s", req.ID, req.BDID)
	return nil
}

// DisassociateResource 取消资源与业务域的关联
func (b *businessDomainManagementClient) DisassociateResource(ctx context.Context, req *interfaces.BusinessDomainResourceDisassociateRequest) error {
	// 构建查询参数
	queryParams := url.Values{}
	queryParams.Add("id", req.ID)
	queryParams.Add("type", req.Type)
	queryParams.Add("bd_id", req.BDID)
	src := fmt.Sprintf("%s/resource?%s", b.baseURL, queryParams.Encode())
	// header := common.GetHeaderFromCtx(ctx)
	header := map[string]string{}

	respCode, _, err := b.httpClient.Delete(ctx, src, header)
	if respCode == http.StatusForbidden {
		b.logger.Errorf("businessDomainManagementClient#DisassociateResource failed:%v, url:%v", err, src)
		err = infraErr.NewHTTPError(ctx, http.StatusForbidden, infraErr.ErrExtBusinessDomainForbidden, err.Error())
		return err
	}
	if err != nil {
		b.logger.Errorf("businessDomainManagementClient#DisassociateResource failed:%v, url:%v", err, src)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return err
	}

	b.logger.Infof("businessDomainManagementClient#DisassociateResource success, resource_id:%s, domain_id:%s", req.ID, req.BDID)
	return nil
}

// ResourceList 查询业务域下的资源列表
func (b *businessDomainManagementClient) ResourceList(ctx context.Context, req *interfaces.BusinessDomainResourceListRequest) (*interfaces.BusinessDomainResourceListResponse, error) {
	src := fmt.Sprintf("%s/resource", b.baseURL)
	// header := common.GetHeaderFromCtx(ctx)
	header := map[string]string{}

	// 构建查询参数
	queryParams := url.Values{}
	if req.BDID != "" {
		queryParams.Add("bd_id", req.BDID)
	}
	if req.ID != "" {
		queryParams.Add("id", req.ID)
	}
	if req.Type != "" {
		queryParams.Add("type", req.Type)
	}
	queryParams.Add("limit", fmt.Sprintf("%d", req.Limit))

	if req.Offset > 0 {
		queryParams.Add("offset", fmt.Sprintf("%d", req.Offset))
	}

	respCode, respParam, err := b.httpClient.Get(ctx, src, queryParams, header)
	if respCode == http.StatusForbidden {
		b.logger.Errorf("businessDomainManagementClient#ResourceList failed:%v, url:%v", err, src)
		err = infraErr.NewHTTPError(ctx, http.StatusForbidden, infraErr.ErrExtBusinessDomainForbidden, err.Error())
		return nil, err
	}
	if err != nil {
		b.logger.Errorf("businessDomainManagementClient#ResourceList failed:%v, url:%v", err, src)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}

	result := &interfaces.BusinessDomainResourceListResponse{}
	resultByt := utils.ObjectToByte(respParam)
	err = jsoniter.Unmarshal(resultByt, result)
	if err != nil {
		b.logger.Errorf("businessDomainManagementClient#ResourceList response unmarshal error:%s", err.Error())
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, err
	}

	b.logger.Infof("businessDomainManagementClient#ResourceList success, bd_id:%s, total:%d", req.BDID, result.Total)
	return result, nil
}
