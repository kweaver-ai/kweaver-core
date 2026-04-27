package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
)

//go:generate mockgen -source=logics_business_domain.go -destination=../mocks/logics_business_domain.go -package=mocks

// IBusinessDomainService 业务域服务接口
type IBusinessDomainService interface {
	// 从Header中获取业务域并校验是否存在
	GetBusinessDomainFromHeader(c *gin.Context) (businessDomain string)
	// ValidateBusinessDomain 校验业务域是否存在
	ValidateBusinessDomain(ctx context.Context) (err error)
	// AssociateResource 关联资源到业务域
	AssociateResource(ctx context.Context, bdID, resourceID string, resourceType AuthResourceType) (err error)

	// DisassociateResource 取消资源与业务域的关联
	DisassociateResource(ctx context.Context, bdID, resourceID string, resourceType AuthResourceType) (err error)

	// BatchDisassociateResource 批量取消资源与业务域的关联
	BatchDisassociateResource(ctx context.Context, bdID string, resourceIds []string, resourceType AuthResourceType) (err error)

	// ResourceList 查询业务域下的资源列表
	ResourceList(ctx context.Context, bdID string, resourceType AuthResourceType) (resourceIDs []string, err error)

	// BatchResourceList 批量查询多业务域下的资源列表
	BatchResourceList(ctx context.Context, bdIds []string, resourceType AuthResourceType) (resourceToBdMap map[string]string, err error)
}
