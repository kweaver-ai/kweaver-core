// Package operator 算子操作适配器
// @file operator.go
// @description: 算子操作适配器
package operator

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/drivenadapters"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	lcategory "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/category"
	loperator "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/operator"
)

// OperatorHandler 算子注册接口
type OperatorHandler interface {
	/*算子管理接口*/
	OperatorRegister(c *gin.Context)
	OperatorQueryByOperatorID(c *gin.Context)
	OperatorQueryPage(c *gin.Context)
	OperatorUpdateByOpenAPI(c *gin.Context)
	OperatorEdit(c *gin.Context)
	OperatorDelete(c *gin.Context)
	OperatorStatusUpdate(c *gin.Context)
	DebugOperator(c *gin.Context)
	ExecuteOperator(c *gin.Context)

	/*历史记录查询操作*/
	QueryOperatorHistoryDetail(c *gin.Context) // 已发布版本详情（从历史记录中查询）
	QueryOperatorHistoryList(c *gin.Context)   // 历史版本列表

	/*算子市场查询操作*/
	QueryOperatorMarketList(c *gin.Context)
	QueryOperatorMarketDetail(c *gin.Context)

	/*内部算子注册*/
	RegisterInternalOperator(c *gin.Context)
}

var (
	once sync.Once
	h    OperatorHandler
)

type operatorHandle struct {
	Logger          interfaces.Logger
	OperatorManager interfaces.OperatorManager
	CategoryManager interfaces.CategoryManager
	Hydra           interfaces.Hydra
	Validator       interfaces.Validator
}

// NewOperatorHandler 算子操作接口
func NewOperatorHandler() OperatorHandler {
	once.Do(func() {
		confLoader := config.NewConfigLoader()
		h = &operatorHandle{
			Hydra:           drivenadapters.NewHydra(),
			Logger:          confLoader.GetLogger(),
			OperatorManager: loperator.NewOperatorManager(),
			CategoryManager: lcategory.NewCategoryManager(),
			Validator:       validator.NewValidator(),
		}
	})
	return h
}
