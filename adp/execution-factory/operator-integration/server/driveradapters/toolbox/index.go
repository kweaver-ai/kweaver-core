// Package toolbox 工具箱操作适配器
// @file index.go
// @description: 工具箱操作适配器
package toolbox

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	ltoolbox "github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/toolbox"
)

// ToolBoxHandler 工具箱操作接口
type ToolBoxHandler interface {
	// 工具箱操作接口
	CreateToolBox(c *gin.Context)
	UpdateToolBox(c *gin.Context)
	QueryToolBox(c *gin.Context)
	DeleteToolBox(c *gin.Context)
	QueryToolBoxPage(c *gin.Context)
	UpdateToolBoxStatus(c *gin.Context)
	// 工具操作接口
	CreateTool(c *gin.Context)
	UpdateTool(c *gin.Context)
	QueryTool(c *gin.Context)
	DeleteBoxTool(c *gin.Context)
	QueryBoxToolPage(c *gin.Context)
	UpdateToolStatus(c *gin.Context)
	GetMarketToolList(c *gin.Context)
	// 工具调试
	DebugTool(c *gin.Context)
	// 工具执行
	ExecuteTool(c *gin.Context)
	// 算子转换成工具
	OperatorToTool(c *gin.Context)
	// 添加或更新工具
	CreateInternalToolBox(c *gin.Context)
	// 查询工具箱信息
	GetReleaseToolBoxInfo(c *gin.Context)

	/*工具箱市场*/
	QueryMarketToolBoxPage(c *gin.Context)
	QueryMarketToolBox(c *gin.Context)
}

var (
	once sync.Once
	h    ToolBoxHandler
)

type toolBoxHandler struct {
	Logger      interfaces.Logger
	ToolService interfaces.IToolService
	Validator   interfaces.Validator
}

// NewToolBoxHandler 工具操作接口
func NewToolBoxHandler() ToolBoxHandler {
	once.Do(func() {
		confLoader := config.NewConfigLoader()
		h = &toolBoxHandler{
			Logger:      confLoader.GetLogger(),
			ToolService: ltoolbox.NewToolServiceImpl(),
			Validator:   validator.NewValidator(),
		}
	})
	return h
}
