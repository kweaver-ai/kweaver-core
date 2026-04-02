package observabilityhandler

import (
	"fmt"
	"net/http"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

func (h *observabilityHTTPHandler) AnalyticsQuery(c *gin.Context) {
	// 1. 获取请求参数
	var analyticsQueryReq observabilityreq.AnalyticsQueryReq
	if err := c.ShouldBindJSON(&analyticsQueryReq); err != nil {
		h.logger.Errorf("[AnalyticsQuery] should bind json err: %v", err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[AnalyticsQuery] should bind json err: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[AnalyticsQuery] should bind json err: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 2. 参数验证
	if analyticsQueryReq.AnalysisLevel == "" {
		err := capierr.New400Err(c, "[AnalyticsQuery] analysis_level is required")
		h.logger.Errorf("[AnalyticsQuery] analysis_level is empty: %v", err)
		otellog.LogError(c.Request.Context(), "[AnalyticsQuery] analysis_level is empty", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if analyticsQueryReq.ID == "" {
		err := capierr.New400Err(c, "[AnalyticsQuery] id is required")
		h.logger.Errorf("[AnalyticsQuery] id is empty: %v", err)
		otellog.LogError(c.Request.Context(), "[AnalyticsQuery] id is empty", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if analyticsQueryReq.StartTime == 0 || analyticsQueryReq.EndTime == 0 {
		err := capierr.New400Err(c, "[AnalyticsQuery] start_time and end_time are required")
		h.logger.Errorf("[AnalyticsQuery] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[AnalyticsQuery] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	if analyticsQueryReq.StartTime > analyticsQueryReq.EndTime {
		err := capierr.New400Err(c, "[AnalyticsQuery] start_time cannot be greater than end_time")
		h.logger.Errorf("[AnalyticsQuery] time range is invalid: %v", err)
		otellog.LogError(c.Request.Context(), "[AnalyticsQuery] time range is invalid", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 3. 获取用户信息
	analyticsQueryReq.XAccountID = c.Request.Header.Get("x-account-id")
	analyticsQueryReq.XAccountType = cenum.AccountType(c.Request.Header.Get("x-account-type"))

	// 4. 调用服务
	ctx := c.Request.Context()
	data, httpErr := h.observabilitySvc.AnalyticsQuery(ctx, &analyticsQueryReq)

	if httpErr != nil {
		otellog.LogError(ctx, fmt.Sprintf("[AnalyticsQuery] analytics query failed: %v", httpErr.Error()), httpErr)
		h.logger.Errorf("[AnalyticsQuery] analytics query failed: %v", httpErr.Error())
		rest.ReplyError(c, httpErr)

		return
	}

	// 5. 返回响应
	c.JSON(http.StatusOK, data)
}