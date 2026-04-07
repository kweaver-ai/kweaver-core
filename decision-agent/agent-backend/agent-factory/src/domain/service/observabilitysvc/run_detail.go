package observabilitysvc

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func (s *observabilitySvc) RunDetail(ctx context.Context, req *observabilityreq.RunDetailReq) (*observabilityresp.RunDetailResp, error) {
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
			Field:     "Attributes.run_id.Data",
			Operation: "==",
			Value:     req.RunID,
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

	// 如果agent_id不为空，添加agent_id过滤条件
	if req.AgentID != "" {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.agent_id.Data",
			Operation: "==",
			Value:     req.AgentID,
			ValueFrom: "const",
		})
	}

	// 如果conversation_id不为空，添加conversation_id过滤条件
	if req.ConversationID != "" {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.conversation_id.Data",
			Operation: "==",
			Value:     req.ConversationID,
			ValueFrom: "const",
		})
	}

	// 如果session_id不为空，添加session_id过滤条件
	if req.SessionID != "" {
		conditions = append(conditions, uniquerydto.Condition{
			Field:     "Attributes.session_id.Data",
			Operation: "==",
			Value:     req.SessionID,
			ValueFrom: "const",
		})
	}

	// 2. 构建uniquery请求
	uniqueryReq := uniquerydto.ReqDataView{
		Limit:        1,
		Offset:       0,
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
	// viewResults, err := s.uniquery.GetDataViewMock(ctx, "__dip_o11y_log", uniqueryReq, "run_detail")
	if err != nil {
		s.logger.Errorf("[RunDetail] GetDataView failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
	}

	if len(viewResults.Entries) == 0 {
		s.logger.Warnf("[RunDetail] run not found, run_id: %s", req.RunID)
		return nil, nil
	}

	entry := viewResults.Entries[0]
	entryMap := entry.(map[string]any)
	attributes := entryMap["Attributes"].(map[string]any)

	// 4. 解析progress字段
	var progressList []agentrespvo.Progress

	if progressData, exists := attributes["progress"]; exists && progressData != nil {
		if progressMap, ok := progressData.(map[string]any); ok {
			if dataField, dataExists := progressMap["Data"]; dataExists && dataField != nil {
				// 尝试将dataField转换为字符串（JSON字符串）
				if jsonStr, ok := dataField.(string); ok {
					// 使用json.Unmarshal解析JSON字符串
					if err := json.Unmarshal([]byte(jsonStr), &progressList); err != nil {
						s.logger.Warnf("[RunDetail] Failed to unmarshal progress JSON: %v", err)
					}
				} else {
					s.logger.Warnf("[RunDetail] progress from opensearch is not a string")
				}
			}
		}
	}

	// 5. 返回响应
	return &observabilityresp.RunDetailResp{
		RunID:          attributes["run_id"].(map[string]any)["Data"].(string),
		AgentID:        attributes["agent_id"].(map[string]any)["Data"].(string),
		AgentVersion:   attributes["agent_version"].(map[string]any)["Data"].(string),
		ConversationID: attributes["conversation_id"].(map[string]any)["Data"].(string),
		SessionID:      attributes["session_id"].(map[string]any)["Data"].(string),
		UserID:         attributes["user_id"].(map[string]any)["Data"].(string),
		CallType:       attributes["call_type"].(map[string]any)["Data"].(string),

		StartTime:           int64(attributes["start_time"].(map[string]any)["Data"].(float64)),
		EndTime:             int64(attributes["end_time"].(map[string]any)["Data"].(float64)),
		TTFT:                int(attributes["ttft"].(map[string]any)["Data"].(float64)),
		TotalTime:           int64(attributes["total_time"].(map[string]any)["Data"].(float64)),
		TotalTokens:         int64(attributes["total_tokens"].(map[string]any)["Data"].(float64)),
		InputMessage:        attributes["input_message"].(map[string]any)["Data"].(string),
		ToolCallCount:       int(attributes["tool_call_count"].(map[string]any)["Data"].(float64)),
		ToolCallFailedCount: int(attributes["tool_call_failed_count"].(map[string]any)["Data"].(float64)),
		Progress:            progressList,
		Status:              attributes["status"].(map[string]any)["Data"].(string),
	}, nil
}

// parseProgressManually 手动解析progress数据（兼容旧数据格式）
// func parseProgressManually(dataField any) []agentrespvo.Progress {
// 	var progressList []agentrespvo.Progress

// 	if progressArray, ok := dataField.([]any); ok {
// 		for _, item := range progressArray {
// 			if progressItem, ok := item.(map[string]any); ok {
// 				progress := agentrespvo.Progress{
// 					ID:                    safeGetString(progressItem, "id"),
// 					AgentName:             safeGetString(progressItem, "agent_name"),
// 					Stage:                 safeGetString(progressItem, "stage"),
// 					Answer:                progressItem["answer"],
// 					Think:                 progressItem["think"],
// 					Status:                safeGetString(progressItem, "status"),
// 					SkillInfo:             safeParseSkillInfo(progressItem["skill_info"]),
// 					InputMessage:          progressItem["input_message"],
// 					Interrupted:           safeGetBool(progressItem, "interrupted"),
// 					Flags:                 progressItem["flags"],
// 					StartTime:             safeGetFloat64(progressItem, "start_time"),
// 					EndTime:               safeGetFloat64(progressItem, "end_time"),
// 					EstimatedInputTokens:  safeGetInt64(progressItem, "estimated_input_tokens"),
// 					EstimatedOutputTokens: safeGetInt64(progressItem, "estimated_output_tokens"),
// 					EstimatedRatioTokens:  safeGetFloat64(progressItem, "estimated_ratio_tokens"),
// 					TokenUsage:            safeParseTokenUsage(progressItem["token_usage"]),
// 				}
// 				progressList = append(progressList, progress)
// 			}
// 		}
// 	}

// 	return progressList
// }
