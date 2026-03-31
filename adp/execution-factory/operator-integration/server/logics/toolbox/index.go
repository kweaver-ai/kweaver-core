// Package toolbox 工具箱、工具管理
// @file index.go
// @description: 实现工具箱、工具管理接口
package toolbox

import (
	"fmt"
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/category"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/intcomp"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/operator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/proxy"
)

var (
	tOnce       sync.Once
	toolService interfaces.IToolService

	validatorMethodPath = func(method, path string) string {
		return fmt.Sprintf("%s:%s", method, path)
	}
)

// ToolServiceImpl 工具箱
type ToolServiceImpl struct {
	DBTx                  model.DBTx
	ToolBoxDB             model.IToolboxDB
	ToolDB                model.IToolDB
	Proxy                 interfaces.ProxyHandler
	CategoryManager       interfaces.CategoryManager
	Logger                interfaces.Logger
	UserMgnt              interfaces.UserManagement
	Validator             interfaces.Validator
	OperatorMgnt          interfaces.OperatorManager
	IntCompConfigSvc      interfaces.IIntCompConfigService
	AuthService           interfaces.IAuthorizationService
	AuditLog              interfaces.LogModelOperator[*metric.AuditLogBuilderParams]
	BusinessDomainService interfaces.IBusinessDomainService
	MetadataService       interfaces.IMetadataService
}

// NewToolServiceImpl 创建工具箱服务
func NewToolServiceImpl() interfaces.IToolService {
	tOnce.Do(func() {
		conf := config.NewConfigLoader()
		toolService = &ToolServiceImpl{
			DBTx:                  dbaccess.NewBaseTx(),
			ToolBoxDB:             dbaccess.NewToolboxDB(),
			ToolDB:                dbaccess.NewToolDB(),
			Proxy:                 proxy.NewProxyServer(),
			Logger:                conf.GetLogger(),
			UserMgnt:              drivenadapters.NewUserManagementClient(),
			Validator:             validator.NewValidator(),
			CategoryManager:       category.NewCategoryManager(),
			OperatorMgnt:          operator.NewOperatorManager(),
			IntCompConfigSvc:      intcomp.NewIntCompConfigService(),
			AuthService:           auth.NewAuthServiceImpl(),
			AuditLog:              metric.NewAuditLogBuilder(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
			MetadataService:       metadata.NewMetadataService(),
		}
	})
	return toolService
}
