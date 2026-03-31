package observabilitysvc

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// sessionMetric 用于临时存储session级别的计算指标
type sessionMetric struct {
	SessionID        string
	StartTime        int64
	EndTime          int64
	RunCount         int
	TotalExecuteTime float64
	TotalTTFTTime    float64
	ErrorCount       int
	ToolFailCount    int
}

func (s *observabilitySvc) SessionList(ctx context.Context, req *observabilityreq.SessionListReq) (*observabilityresp.SessionListResp, error) {
	// 1. 构建过滤条件
	conditions := []uniquerydto.Condition{
		{
			Field:     "Resource.service.name",
			Operation: "==",
			Value:     "agent-app",
			ValueFrom: "const",
		},
		{
			Field:     "SeverityText",
			Operation: "==",
			Value:     "Info",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.call_type.Data",
			Operation: "!=",
			Value:     "api_chat",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.agent_id.Data",
			Operation: "==",
			Value:     req.AgentID,
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.agent_version.Data",
			Operation: "!=",
			Value:     "v0",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.start_time.Data",
			Operation: ">=",
			Value:     req.StartTime,
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.end_time.Data",
			Operation: "<",
			Value:     req.EndTime,
			ValueFrom: "const",
		},
	}

	// 1.1 如果conversation_id不为空，添加conversation_id过滤条件
	if req.ConversationID != "" {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.conversation_id.Data",
			Operation: "==",
			Value:     req.ConversationID,
			ValueFrom: "const",
		})
	}

	// 2. 构建uniquery请求
	uniqueryReq := uniquerydto.ReqDataView{
		Limit:        req.Size,
		Offset:       (req.Page - 1) * req.Size,
		Start:        req.StartTime,
		End:          req.EndTime,
		Format:       "original",
		XAccountID:   req.XAccountID,
		XAccountType: string(req.XAccountType),
		Filters: &uniquerydto.GlobalFilter{
			Operation:     "and",
			SubConditions: conditions,
		},
	}

	// 3. 调用uniquery GetDataView接口
	viewResults, err := s.uniquery.GetDataView(ctx, "__dip_o11y_log", uniqueryReq)
	// viewResults, err := s.uniquery.GetDataViewMock(ctx, "__dip_o11y_log", uniqueryReq, "session_list")
	if err != nil {
		s.logger.Errorf("[SessionList] GetDataView failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
	}

	entries := viewResults.Entries

	// 4. 按session_id分组，计算每个session的指标
	sessionMetrics := make(map[string]*sessionMetric)

	for _, entry := range entries {
		entryMap := entry.(map[string]any)
		attributes := entryMap["Attributes"].(map[string]any)

		sessionID := attributes["session_id"].(map[string]any)["Data"].(string)
		startTime := int64(attributes["start_time"].(map[string]any)["Data"].(float64))
		endTime := int64(attributes["end_time"].(map[string]any)["Data"].(float64))
		status := attributes["status"].(map[string]any)["Data"].(string)
		totalTime := attributes["total_time"].(map[string]any)["Data"].(float64)
		ttft := attributes["ttft"].(map[string]any)["Data"].(float64)
		toolCallFailedCount := int(attributes["tool_call_failed_count"].(map[string]any)["Data"].(float64))

		if _, exists := sessionMetrics[sessionID]; !exists {
			sessionMetrics[sessionID] = &sessionMetric{
				SessionID:        sessionID,
				StartTime:        startTime,
				EndTime:          endTime,
				RunCount:         0,
				TotalExecuteTime: 0,
				TotalTTFTTime:    0,
				ErrorCount:       0,
				ToolFailCount:    0,
			}
		}

		metric := sessionMetrics[sessionID]
		metric.RunCount++
		metric.TotalExecuteTime += totalTime
		metric.TotalTTFTTime += ttft
		metric.ToolFailCount += toolCallFailedCount

		// 更新session的时间范围
		if startTime < metric.StartTime {
			metric.StartTime = startTime
		}

		if endTime > metric.EndTime {
			metric.EndTime = endTime
		}

		// 统计错误run
		if status != "success" {
			metric.ErrorCount++
		}
	}

	// 5. 转换为响应格式
	var sessionListItems []observabilityresp.SessionListItem

	for _, metric := range sessionMetrics {
		sessionDuration := metric.EndTime - metric.StartTime

		sessionListItems = append(sessionListItems, observabilityresp.SessionListItem{
			SessionID:       metric.SessionID,
			StartTime:       metric.StartTime,
			EndTime:         metric.EndTime,
			SessionDuration: int(sessionDuration),
		})
	}

	// 6. 返回响应
	return &observabilityresp.SessionListResp{
		Entries:    sessionListItems,
		TotalCount: len(sessionListItems),
	}, nil
}

// GetSessionCountsByConversationIDs 批量查询会话数量
func (s *observabilitySvc) GetSessionCountsByConversationIDs(ctx context.Context, agentID string, conversationIDs []string, startTime, endTime int64, xAccountID string, xAccountType string) (map[string]int, error) {
	// 构建过滤条件
	conditions := []uniquerydto.Condition{
		{
			Field:     "Resource.service.name",
			Operation: "==",
			Value:     "agent-app",
			ValueFrom: "const",
		},
		{
			Field:     "SeverityText",
			Operation: "==",
			Value:     "Info",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.agent_id.Data",
			Operation: "==",
			Value:     agentID,
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.agent_version.Data",
			Operation: "!=",
			Value:     "v0",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.start_time.Data",
			Operation: ">=",
			Value:     startTime,
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.end_time.Data",
			Operation: "<",
			Value:     endTime,
			ValueFrom: "const",
		},
	}

	// 添加conversation_id过滤条件：使用in操作符
	if len(conversationIDs) > 0 {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.conversation_id.Data",
			Operation: "in",
			Value:     conversationIDs,
			ValueFrom: "const",
		})
	}

	// 构建uniquery请求，不限制返回数量，因为我们只需要统计
	uniqueryReq := uniquerydto.ReqDataView{
		Limit:        10000, // 0表示不限制，获取所有匹配记录
		Offset:       0,
		Start:        startTime,
		End:          endTime,
		Format:       "original",
		XAccountID:   xAccountID,
		XAccountType: xAccountType,
		Filters: &uniquerydto.GlobalFilter{
			Operation:     "and",
			SubConditions: conditions,
		},
	}

	// 调用uniquery GetDataView接口
	viewResults, err := s.uniquery.GetDataView(ctx, "__dip_o11y_log", uniqueryReq)
	if err != nil {
		s.logger.Errorf("[GetSessionCountsByConversationIDs] GetDataView failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
	}

	entries := viewResults.Entries

	// 按conversation_id和session_id分组统计
	// 使用map[conversationID]map[sessionID]bool来去重
	conversationSessionMap := make(map[string]map[string]bool)

	for _, entry := range entries {
		entryMap := entry.(map[string]any)
		attributes := entryMap["Attributes"].(map[string]any)

		conversationID := attributes["conversation_id"].(map[string]any)["Data"].(string)
		sessionID := attributes["session_id"].(map[string]any)["Data"].(string)

		if _, ok := conversationSessionMap[conversationID]; !ok {
			conversationSessionMap[conversationID] = make(map[string]bool)
		}

		conversationSessionMap[conversationID][sessionID] = true
	}

	// 计算每个conversation的session数量
	result := make(map[string]int)
	for convID, sessionSet := range conversationSessionMap {
		result[convID] = len(sessionSet)
	}

	// 确保所有请求的conversationID都在结果中，即使数量为0
	for _, convID := range conversationIDs {
		if _, ok := result[convID]; !ok {
			result[convID] = 0
		}
	}

	return result, nil
}
