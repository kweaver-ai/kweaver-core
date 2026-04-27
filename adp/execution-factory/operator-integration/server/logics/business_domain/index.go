package business_domain

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	infracommon "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

type businessDomainServiceImpl struct {
	logger       interfaces.Logger
	bdManagement interfaces.BusinessDomainManagement
}

// NewBusinessDomainService 创建业务域服务实例
func NewBusinessDomainService() interfaces.IBusinessDomainService {
	if !config.GetBusinessDomainEnabled() {
		return &noopBusinessDomainService{}
	}
	return &businessDomainServiceImpl{
		logger:       config.NewConfigLoader().GetLogger(),
		bdManagement: drivenadapters.NewBusinessDomainManagementClient(),
	}
}

// GetBusinessDomainFromHeader 从Header中获取业务域并校验是否存在
func (s *businessDomainServiceImpl) GetBusinessDomainFromHeader(c *gin.Context) (businessDomain string) {
	businessDomain = c.GetHeader(string(interfaces.HeaderXBusinessDomain))
	return businessDomain
}

// ValidateBusinessDomain 校验业务域是否存在
func (s *businessDomainServiceImpl) ValidateBusinessDomain(ctx context.Context) (err error) {
	_, ok := infracommon.GetBusinessDomainFromCtx(ctx)
	if !ok {
		err = errors.NewHTTPError(ctx, http.StatusBadRequest, errors.ErrExtBusinessDomainIDRequired, "x-business-domain-id is required")
		return err
	}
	return nil
}

// AssociateResource 关联资源到业务域
func (s *businessDomainServiceImpl) AssociateResource(ctx context.Context, bdId, resourceId string, resourceType interfaces.AuthResourceType) error {
	req := &interfaces.BusinessDomainResourceAssociateRequest{
		BDID: bdId,
		ID:   resourceId,
		Type: string(resourceType),
	}

	err := s.bdManagement.AssociateResource(ctx, req)
	if err != nil {
		s.logger.Errorf("AssociateResource failed: %v, bdId: %s, resourceID: %s, resourceType: %s",
			err, bdId, resourceId, resourceType)
		return err
	}

	s.logger.Infof("AssociateResource success, bdId: %s, resourceID: %s, resourceType: %s",
		bdId, resourceId, resourceType)
	return nil
}

// BatchDisassociateResource 批量取消资源与业务域的关联
func (s *businessDomainServiceImpl) BatchDisassociateResource(ctx context.Context, bdID string, resourceIds []string, resourceType interfaces.AuthResourceType) (err error) {
	if len(resourceIds) == 0 {
		return
	}

	for _, resourceId := range resourceIds {
		err = s.DisassociateResource(ctx, bdID, resourceId, resourceType)
		if err != nil {
			return err
		}
	}
	return
}

// DisassociateResource 取消资源与业务域的关联
func (s *businessDomainServiceImpl) DisassociateResource(ctx context.Context, bdId, resourceId string, resourceType interfaces.AuthResourceType) error {
	req := &interfaces.BusinessDomainResourceDisassociateRequest{
		BDID: bdId,
		ID:   resourceId,
		Type: string(resourceType),
	}

	err := s.bdManagement.DisassociateResource(ctx, req)
	if err != nil {
		s.logger.Errorf("DisassociateResource failed: %v, bdId: %s, resourceID: %s, resourceType: %s",
			err, bdId, resourceId, resourceType)
		return err
	}

	s.logger.Infof("DisassociateResource success, bdId: %s, resourceID: %s, resourceType: %s",
		bdId, resourceId, resourceType)
	return nil
}

// ResourceList 查询业务域下的资源列表
func (s *businessDomainServiceImpl) ResourceList(ctx context.Context, bdId string, resourceType interfaces.AuthResourceType) ([]string, error) {
	req := &interfaces.BusinessDomainResourceListRequest{
		BDID:   bdId,
		Type:   string(resourceType),
		Limit:  -1, // 设置为-1表示不分页，获取所有数据
		Offset: 0,
	}

	resp, err := s.bdManagement.ResourceList(ctx, req)
	if err != nil {
		s.logger.Errorf("ResourceList failed: %v, bdId: %s, resourceType: %s",
			err, bdId, resourceType)
		return nil, err
	}

	// 提取资源ID列表
	resourceIDs := make([]string, 0, len(resp.Items))
	for _, item := range resp.Items {
		resourceIDs = append(resourceIDs, item.ID)
	}

	s.logger.Infof("ResourceList success, bdId: %s, resourceType: %s, count: %d",
		bdId, resourceType, len(resourceIDs))
	return resourceIDs, nil
}

// BatchResourceList 批量查询多业务域下的资源列表
func (s *businessDomainServiceImpl) BatchResourceList(ctx context.Context, bdIds []string, resourceType interfaces.AuthResourceType) (resourceToBdMap map[string]string, err error) {
	// 初始化返回结果
	resourceToBdMap = make(map[string]string)

	// 遍历所有业务域ID
	for _, bdId := range bdIds {
		// 调用单个业务域的资源列表方法
		resourceIds, err := s.ResourceList(ctx, bdId, resourceType)
		if err != nil {
			s.logger.Errorf("BatchResourceList failed for bdId %s: %v", bdId, err)
			// 返回错误，不继续处理其他业务域
			return nil, err
		}

		// 将资源ID和业务域ID的映射关系添加到结果中
		for _, resourceId := range resourceIds {
			resourceToBdMap[resourceId] = bdId
		}
	}
	return resourceToBdMap, nil
}
