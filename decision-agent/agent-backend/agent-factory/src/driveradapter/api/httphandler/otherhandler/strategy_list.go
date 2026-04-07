package otherhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
)

// 内置策略列表,key为category,value为策略列表
var strategyMap = map[string][]skillvalobj.Strategy{
	// "llm": {
	// 	{
	// 		ID:          "summary",
	// 		Name:        "summary",
	// 		Description: "摘要",
	// 	},
	// },
	// "frontend": {
	// 	{
	// 		ID:          "default",
	// 		Name:        "default",
	// 		Description: "默认",
	// 	},
	// },
}

// @Summary      根据分类获取结果处理策略列表
// @Description  根据分类获取结果处理策略列表
// @Tags         工具结果处理策略
// @Accept       json
// @Produce      json
// @Param        category_id  path      string  true  "category_id"
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/tool-result-process-strategy/category/{category_id}/strategy [get]
func (o *otherHTTPHandler) StrategyList(c *gin.Context) {
	category := c.Param("category_id")
	if category == "" {
		err := capierr.New400Err(c, "category_id is required")
		_ = c.Error(err)

		return
	}

	if _, ok := strategyMap[category]; !ok {
		err := capierr.New400Err(c, "category_id is invalid")
		_ = c.Error(err)

		return
	}

	response := map[string]interface{}{
		"entries": strategyMap[category],
		"total":   len(strategyMap[category]),
	}
	c.JSON(http.StatusOK, response)
}
