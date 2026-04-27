package bizdomain

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/capierr"
)

// QueryResourceAssociationsTestHandler 测试关联关系查询
func (h *BizDomainTestHandler) QueryResourceAssociationsTestHandler(c *gin.Context) {
	// 解析请求参数
	var req TestBizDomainHttpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		err := capierr.New400Err(c, "Invalid request parameters: "+err.Error())
		_ = c.Error(err)

		return
	}
	// 调用关联关系查询测试方法
	// res, err := h.bizDomainSvc.QueryResourceAssociationsTest(c.Request.Context(), req.AgentID)
	//
	//	if err != nil {
	//		err := capierr.New500Err(c, "QueryResourceAssociationsTest failed: "+err.Error())
	//		_ = c.Error(err)
	//
	//		return
	//	}
	//
	//	response := map[string]interface{}{
	//		"message":   "Resource associations query test completed successfully",
	//		"status":    "success",
	//		"agent_id":  req.AgentID,
	//		"operation": "query_resource_associations",
	//		"data":      res, // 返回查询到的关联关系数据
	//	}
	//
	// c.JSON(http.StatusOK, response)
}
