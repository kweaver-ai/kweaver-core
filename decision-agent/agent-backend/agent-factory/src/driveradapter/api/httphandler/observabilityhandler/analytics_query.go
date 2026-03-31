package observabilityhandler

import (
	"fmt"
	"net/http"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

func (h *observabilityHTTPHandler) AnalyticsQuery(c *gin.Context) {
	// 1. 获取请求参数
	var req observabilityreq.AnalyticsQueryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Errorf("[AnalyticsQuery] should bind json err: %v", err)
		o11y.Error(c, fmt.Sprintf("[AnalyticsQuery] should bind json err: %v", err))
		httpErr := capierr.New400Err(c, fmt.Sprintf("[AnalyticsQuery] should bind json err: %v", err))
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 参数验证
	if req.AnalysisLevel == "" {
		err := capierr.New400Err(c, "[AnalyticsQuery] analysis_level is required")
		h.logger.Errorf("[AnalyticsQuery] analysis_level is empty: %v", err)
		o11y.Error(c, "[AnalyticsQuery] analysis_level is empty")
		rest.ReplyError(c, err)

		return
	}

	if req.ID == "" {
		err := capierr.New400Err(c, "[AnalyticsQuery] id is required")
		h.logger.Errorf("[AnalyticsQuery] id is empty: %v", err)
		o11y.Error(c, "[AnalyticsQuery] id is empty")
		rest.ReplyError(c, err)

		return
	}

	if req.StartTime == 0 || req.EndTime == 0 {
		err := capierr.New400Err(c, "[AnalyticsQuery] start_time and end_time are required")
		h.logger.Errorf("[AnalyticsQuery] time range is invalid: %v", err)
		o11y.Error(c, "[AnalyticsQuery] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	if req.StartTime > req.EndTime {
		err := capierr.New400Err(c, "[AnalyticsQuery] start_time cannot be greater than end_time")
		h.logger.Errorf("[AnalyticsQuery] time range is invalid: %v", err)
		o11y.Error(c, "[AnalyticsQuery] time range is invalid")
		rest.ReplyError(c, err)

		return
	}

	// 3. 获取用户信息
	req.XAccountID = c.Request.Header.Get("x-account-id")
	req.XAccountType = cenum.AccountType(c.Request.Header.Get("x-account-type"))

	// 4. 调用服务
	ctx := c.Request.Context()
	data, httpErr := h.observabilitySvc.AnalyticsQuery(ctx, &req)

	if httpErr != nil {
		o11y.Error(ctx, fmt.Sprintf("[AnalyticsQuery] analytics query failed: %v", httpErr.Error()))
		h.logger.Errorf("[AnalyticsQuery] analytics query failed: %v", httpErr.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	// 5. 返回响应
	c.JSON(http.StatusOK, data)
}
