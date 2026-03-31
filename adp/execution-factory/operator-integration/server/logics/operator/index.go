// Package operator 实现算子操作接口
// @file index.go 初始化
// @description: 实现算子操作管理
package operator

import (
	"sync"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/dbaccess"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/mq"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces/model"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/auth"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/category"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/intcomp"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metadata"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/metric"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/proxy"
)

type operatorManager struct {
	Logger                interfaces.Logger
	DBOperatorManager     model.IOperatorRegisterDB
	DBTx                  model.DBTx
	CategoryManager       interfaces.CategoryManager
	UserMgnt              interfaces.UserManagement
	Validator             interfaces.Validator
	Proxy                 interfaces.ProxyHandler
	OpReleaseDB           model.IOperatorReleaseDB
	OpReleaseHistoryDB    model.IOperatorReleaseHistoryDB
	IntCompConfigSvc      interfaces.IIntCompConfigService
	AuthService           interfaces.IAuthorizationService
	AuditLog              interfaces.LogModelOperator[*metric.AuditLogBuilderParams]
	FlowAutomation        interfaces.FlowAutomation
	MQClient              mq.MQClient
	BusinessDomainService interfaces.IBusinessDomainService
	MetadataService       interfaces.IMetadataService
}

var (
	once sync.Once
	om   interfaces.OperatorManager
)

// NewOperatorManager 算子操作接口
func NewOperatorManager() interfaces.OperatorManager {
	once.Do(func() {
		conf := config.NewConfigLoader()
		om = &operatorManager{
			Logger:                conf.GetLogger(),
			DBOperatorManager:     dbaccess.NewOperatorManagerDB(),
			DBTx:                  dbaccess.NewBaseTx(),
			CategoryManager:       category.NewCategoryManager(),
			UserMgnt:              drivenadapters.NewUserManagementClient(),
			Validator:             validator.NewValidator(),
			Proxy:                 proxy.NewProxyServer(),
			OpReleaseDB:           dbaccess.NewOperatorReleaseDB(),
			OpReleaseHistoryDB:    dbaccess.NewOperatorReleaseHistoryDB(),
			IntCompConfigSvc:      intcomp.NewIntCompConfigService(),
			AuthService:           auth.NewAuthServiceImpl(),
			AuditLog:              metric.NewAuditLogBuilder(),
			FlowAutomation:        drivenadapters.NewFlowAutomationClient(),
			MQClient:              mq.NewMQClient(),
			BusinessDomainService: business_domain.NewBusinessDomainService(),
			MetadataService:       metadata.NewMetadataService(),
		}
	})
	return om
}
