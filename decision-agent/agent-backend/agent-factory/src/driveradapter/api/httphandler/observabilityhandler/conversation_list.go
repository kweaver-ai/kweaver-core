package observabilityhandler

import (
	"fmt"
	"net/http"

	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/capierr"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"github.com/gin-gonic/gin"
)

var (
	_ interface{}
	_ observabilityresp.ObservabilityConversationDetail
)

// @Summary      查询Agent下的所有对话列表
// @Description  查询Agent下的所有对话列表，支持分页和根据对话title模糊查询
// @Tags         可观测性
// @Accept       json
// @Produce      json
// @Param        agent_key  path      string  true  "agent_key"
// @Param        request  body      object  true  "请求体"
// @Success      200  {object}  object  "成功返回对话列表"
// @Failure      400  {object}  object  "失败"
// @Failure      404  {object}  object  "失败"
// @Failure      500  {object}  object  "失败"
// @Security     BearerAuth
// @Router       /v1/observability/agent/{agent_key}/conversation [post]
func (h *observabilityHTTPHandler) ConversationList(c *gin.Context) {
	// 1. 获取路径参数
	agentID := c.Param("agent_id")
	if agentID == "" {
		h.logger.Errorf("[ConversationList] agent_id is required")
		err := capierr.New400Err(c, "[ConversationList] agent_id is required")
		otellog.LogError(c.Request.Context(), "[ConversationList] agent_id is required", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// 2. 获取请求参数
	var conversationListReq observabilityreq.ConversationListReq
	if err := c.ShouldBindJSON(&conversationListReq); err != nil {
		h.logger.Errorf("[ConversationList] should bind json err: %v", err)
		httpErr := capierr.New400Err(c, fmt.Sprintf("[ConversationList] should bind json err: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[ConversationList] should bind json err: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 3. 设置路径参数到请求中
	// conversationListReq.AgentID = agentID
	// 这里不需要设置，因为路径参数中是agent_key ，通过agent_key 获取到 对话的 记录列表，而不是debug 的列表

	// 4. 参数验证
	if conversationListReq.Size <= 0 {
		conversationListReq.Size = 10
	}

	if conversationListReq.Page <= 0 {
		conversationListReq.Page = 1
	}

	if conversationListReq.StartTime > conversationListReq.EndTime {
		h.logger.Errorf("[ConversationList] start_time must be less than end_time")
		err := capierr.New400Err(c, "[ConversationList] start_time must be less than end_time")
		otellog.LogError(c.Request.Context(), "[ConversationList] start_time must be less than end_time", err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, err)

		return
	}

	// // 设置时间范围默认值
	// if conversationListReq.StartTime == 0 || conversationListReq.EndTime == 0 {
	// 	// 如果未设置时间范围，默认查询最近30天
	// 	now := time.Now().UnixMilli()
	// 	if conversationListReq.StartTime == 0 {
	// 		conversationListReq.StartTime = now - 30*24*60*60*1000 // 30天前
	// 	}
	// 	if conversationListReq.EndTime == 0 {
	// 		conversationListReq.EndTime = now
	// 	}
	// }

	// 4. 获取用户信息
	user := chelper.GetVisitorFromCtx(c)
	if user == nil {
		err := capierr.New404Err(c, "[ConversationList] user not found")
		otellog.LogError(c.Request.Context(), "[ConversationList] user not found", err)
		h.logger.Errorf("[ConversationList] user not found: %v", err)
		rest.ReplyError(c, err)

		return
	}

	// 5. 调用服务
	// agentID 值实际为 agent key
	// conversationListReq.AgentID 值实际为 agent ID
	data, totalCount, err := h.conversationSvc.ListByAgentID(c.Request.Context(), agentID, conversationListReq.Title, conversationListReq.Page, conversationListReq.Size, conversationListReq.StartTime, conversationListReq.EndTime)
	if err != nil {
		h.logger.Errorf("[ConversationList] call conversation service error: %v", err)
		httpErr := capierr.New500Err(c, fmt.Sprintf("[ConversationList] call conversation service error: %v", err))
		otellog.LogError(c.Request.Context(), fmt.Sprintf("[ConversationList] call conversation service error: %v", err), err)
		oteltrace.EndSpan(c.Request.Context(), err)
		rest.ReplyError(c, httpErr)

		return
	}

	// 6. 批量查询所有conversation的session数量
	entries := make([]observabilityresp.ObservabilityConversationDetail, len(data))

	// 收集conversation IDs
	conversationIDs := make([]string, len(data))
	for i, conv := range data {
		conversationIDs[i] = conv.ID
	}

	// 批量获取session数量
	sessionCounts := make(map[string]int)

	if len(conversationIDs) > 0 {
		// 转换账户类型
		var xAccountType cenum.AccountType

		xAccountType.LoadFromMDLVisitorType(user.Type)

		counts, err := h.observabilitySvc.GetSessionCountsByConversationIDs(c.Request.Context(), conversationListReq.AgentID, conversationIDs, conversationListReq.StartTime, conversationListReq.EndTime, user.ID, string(xAccountType))
		if err != nil {
			h.logger.Errorf("[ConversationList] batch get session counts error: %v", err)
			// 不返回错误，继续处理，所有sessionCount保持为0
		} else {
			sessionCounts = counts
		}
	}

	// 构建响应条目
	for i, conv := range data {
		sessionCount := sessionCounts[conv.ID]
		entries[i] = observabilityresp.ObservabilityConversationDetail{
			Conversation: conv,
			SessionCount: sessionCount,
		}
	}

	conversationListResp := observabilityresp.ConversationListResp{
		Entries:    entries,
		TotalCount: totalCount,
	}

	c.JSON(http.StatusOK, conversationListResp)
}