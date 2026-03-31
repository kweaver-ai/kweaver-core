package driveradapters

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/category"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/driveradapters/operator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/business_domain"
)

// OperatorRestHandler operator RESTfual API Handler 接口
type OperatorRestHandler interface {
	// RegisterPrivate 注册内部API
	RegisterPrivate(engine *gin.RouterGroup)

	// RegisterPublic 注册外部API
	RegisterPublic(engine *gin.RouterGroup)
}

type operatorRestHandler struct {
	OperatorHandler       operator.OperatorHandler
	CategoryHandler       category.CategoryHandler
	businessDomainService interfaces.IBusinessDomainService
}

var (
	oOnce    sync.Once
	oHandler OperatorRestHandler
)

func NewOperatorRestHandler() OperatorRestHandler {
	oOnce.Do(func() {
		oHandler = &operatorRestHandler{
			OperatorHandler:       operator.NewOperatorHandler(),
			CategoryHandler:       category.NewCategoryHandler(),
			businessDomainService: business_domain.NewBusinessDomainService(),
		}
	})
	return oHandler
}

// RegisterPrivate 注册内部API
func (o *operatorRestHandler) RegisterPrivate(engine *gin.RouterGroup) {
	/*算子管理相关接口*/
	// POST /api/agent-operator-integration/internal-v1/operator/register 注册算子
	engine.POST("/operator/register", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.OperatorRegister)
	// GET /api/agent-operator-integration/internal-v1/operator/info/{operator_id} 获取算子详情
	engine.GET("/operator/info/:operator_id", o.OperatorHandler.OperatorQueryByOperatorID)
	// GET /api/agent-operator-integration/internal-v1/operator/info/list 获取算子分页列表
	engine.GET("/operator/info/list", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.OperatorQueryPage)
	// POST /api/agent-operator-integration/internal-v1/operator/info/update 更新算子信息(目前仅Dataflow调用)
	engine.POST("/operator/info/update", o.OperatorHandler.OperatorUpdateByOpenAPI)
	// POST /api/agent-operator-integration/internal-v1/operator/proxy/:operator_id 执行算子
	engine.POST("/operator/proxy/:operator_id", o.OperatorHandler.ExecuteOperator)

	/*已发布版本详情*/
	// GET /api/agent-operator-integration/internal-v1/operator/history/:operator_id/:version 获取已发布算子指定版本详情
	engine.GET("/operator/history/:operator_id/:version", o.OperatorHandler.QueryOperatorHistoryDetail)
	// GET /api/agent-operator-integration/internal-v1/operator/history/:operator_id
	engine.GET("/operator/history/:operator_id", o.OperatorHandler.QueryOperatorHistoryList)

	/*算子市场相关接口*/
	// GET /api/agent-operator-integration/internal-v1/operator/market 获取算子市场列表，支持分页，排序，及query过滤条件
	engine.GET("/operator/market", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.QueryOperatorMarketList)
	// GET /api/agent-operator-integration/internal-v1/operator/market/:operator_id 在算子市场查看详情
	engine.GET("/operator/market/:operator_id", o.OperatorHandler.QueryOperatorMarketDetail)

	/*内置算子管理*/
	// POST /api/agent-operator-integration/internal-v1/operator/intcomp 注册(更新)内置算子
	engine.POST("/operator/intcomp", middlewareBusinessDomain(true, true, o.businessDomainService), o.OperatorHandler.RegisterInternalOperator)

	/*算子分类管理*/
	// GET /api/agent-operator-integration/internal-v1/operator/category //获取算子分类列表
	engine.GET("/operator/category", o.CategoryHandler.CategoryList)
	// POST /api/agent-operator-integration/internal-v1/operator/category // 新建算子分类
	engine.POST("/operator/category", o.CategoryHandler.CategoryCreate)
	// PUT /api/agent-operator-integration/internal-v1/operator/category/:category_type // 更新算子分类
	engine.PUT("/operator/category/:category_type", o.CategoryHandler.CategoryUpdate)
	// DELETE /api/agent-operator-integration/internal-v1/operator/category/:category_type // 删除算子分类
	engine.DELETE("/operator/category/:category_type", o.CategoryHandler.CategoryDelete)
}

// RegisterPublic 注册外部API
func (o *operatorRestHandler) RegisterPublic(engine *gin.RouterGroup) {
	// POST /api/agent-operator-integration/v1/operator/register
	engine.POST("/operator/register", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.OperatorRegister)
	// 查询算子相关接口
	// GET /api/agent-operator-integration/v1/operator/info/{operator_id}
	engine.GET("/operator/info/:operator_id", o.OperatorHandler.OperatorQueryByOperatorID)
	// GET /api/agent-operator-integration/v1/operator/info/list
	engine.GET("/operator/info/list", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.OperatorQueryPage)
	// DELETE /api/agent-operator-integration/v1/operator/delete
	engine.DELETE("/operator/delete", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.OperatorDelete)
	// POST /api/agent-operator-integration/v1/operator/status
	engine.POST("/operator/status", o.OperatorHandler.OperatorStatusUpdate)
	// POST /api/agent-operator-integration/v1/operator/info
	engine.POST("/operator/info", o.OperatorHandler.OperatorEdit)
	// POST /api/agent-operator-integration/v1/operator/info/update
	engine.POST("/operator/info/update", o.OperatorHandler.OperatorUpdateByOpenAPI)
	// POST /api/agent-operator-integration/v1/operator/debug
	engine.POST("/operator/debug", o.OperatorHandler.DebugOperator)

	/*已发布版本详情*/
	// GET /api/agent-operator-integration/v1/operator/history/:operator_id/:version 获取已发布算子指定版本详情
	engine.GET("/operator/history/:operator_id/:version", o.OperatorHandler.QueryOperatorHistoryDetail)
	// GET /api/agent-operator-integration/v1/operator/history/:operator_id
	engine.GET("/operator/history/:operator_id", o.OperatorHandler.QueryOperatorHistoryList)

	/*算子市场相关接口*/
	// GET /api/agent-operator-integration/v1/operator/market 获取算子市场列表，支持分页，排序，及query过滤条件
	engine.GET("/operator/market", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.QueryOperatorMarketList)
	// GET /api/agent-operator-integration/v1/operator/market/:operator_id 在算子市场查看详情
	engine.GET("/operator/market/:operator_id", middlewareBusinessDomain(true, false, o.businessDomainService), o.OperatorHandler.QueryOperatorMarketDetail)

	/*内置算子管理*/
	// POST /api/agent-operator-integration/v1/operator/intcomp
	engine.POST("/operator/intcomp", middlewareBusinessDomain(true, true, o.businessDomainService), o.OperatorHandler.RegisterInternalOperator)

	/*算子分类管理*/
	// GET /api/agent-operator-integration/v1/operator/category //获取算子分类列表
	engine.GET("/operator/category", o.CategoryHandler.CategoryList)
}
