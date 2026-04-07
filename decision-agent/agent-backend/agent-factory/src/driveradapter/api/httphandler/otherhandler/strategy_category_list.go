package otherhandler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/skillvalobj"
)

var categoryList = []skillvalobj.Category{
	// {
	// 	ID:          "llm",
	// 	Name:        "llm",
	// 	Description: "传递给大模型的结果",
	// },
	// {
	// 	ID:          "frontend",
	// 	Name:        "前端",
	// 	Description: "传递给前端的结果",
	// },
}

// @Summary      获取结果处理策略分类列表
// @Description  获取结果处理策略分类列表
// @Tags         工具结果处理策略
// @Accept       json
// @Produce      json
// @Success      200  {object}  object  "成功"
// @Failure      400  {object}  object  "失败"
// @Failure      401  {object}  object  "失败"
// @Failure      403  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v3/tool-result-process-strategy/category [get]
func (o *otherHTTPHandler) CategoryList(c *gin.Context) {
	response := map[string]interface{}{
		"entries": categoryList,
		"total":   len(categoryList),
	}
	c.JSON(http.StatusOK, response)
}
