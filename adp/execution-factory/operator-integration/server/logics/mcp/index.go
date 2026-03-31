// Package mcp 实现MCP Server操作接口
// @file index.go 初始化
// @description: 实现MCP Server操作管理
package mcp

import (
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
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/mcpinstance"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/toolbox"
)

var (
	mOnce      sync.Once
	mcpService interfaces.IMCPService
)

type mcpServiceImpl struct {
	logger                    interfaces.Logger
	DBTx                      model.DBTx
	DBMCPServerConfig         model.DBMCPServerConfig
	DBMCPServerRelease        model.DBMCPServerRelease
	DBMCPServerReleaseHistory model.DBMCPServerReleaseHistory
	DBMCPTool                 model.DBMCPTool
	IntCompConfigService      interfaces.IIntCompConfigService
	UserMgnt                  interfaces.UserManagement
	Validator                 interfaces.Validator
	CategoryManager           interfaces.CategoryManager
	AuthService               interfaces.IAuthorizationService
	ToolService               interfaces.IToolService
	AuditLog                  interfaces.LogModelOperator[*metric.AuditLogBuilderParams]
	MCPInstanceService        interfaces.InstanceService
	BusinessDomainService     interfaces.IBusinessDomainService
}

// NewMCPServiceImpl 初始化MCP服务
func NewMCPServiceImpl() interfaces.IMCPService {
	mOnce.Do(func() {
		s := &mcpServiceImpl{
			logger:                    config.NewConfigLoader().GetLogger(),
			DBTx:                      dbaccess.NewBaseTx(),
			DBMCPServerConfig:         dbaccess.NewMCPServerConfigDBSingleton(),
			DBMCPServerRelease:        dbaccess.NewMCPServerReleaseDBSingleton(),
			DBMCPServerReleaseHistory: dbaccess.NewMCPServerReleaseHistoryDBSingleton(),
			DBMCPTool:                 dbaccess.NewMCPToolDBSingleton(),
			IntCompConfigService:      intcomp.NewIntCompConfigService(),
			UserMgnt:                  drivenadapters.NewUserManagementClient(),
			Validator:                 validator.NewValidator(),
			CategoryManager:           category.NewCategoryManager(),
			AuthService:               auth.NewAuthServiceImpl(),
			ToolService:               toolbox.NewToolServiceImpl(),
			AuditLog:                  metric.NewAuditLogBuilder(),
			BusinessDomainService:     business_domain.NewBusinessDomainService(),
		}
		s.MCPInstanceService = mcpinstance.NewMCPInstanceService(s)
		mcpService = s
	})
	return mcpService
}
