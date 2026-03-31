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

func (s *observabilitySvc) SessionDetail(ctx context.Context, req *observabilityreq.SessionDetailReq) (*observabilityresp.SessionDetailResp, error) {
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
			Field:     "Attributes.agent_version.Data",
			Operation: "!=",
			Value:     "v0",
			ValueFrom: "const",
		},
		{
			Field:     "Attributes.session_id.Data",
			Operation: "==",
			Value:     req.SessionID,
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

	// 2. 构建uniquery请求
	uniqueryReq := uniquerydto.ReqDataView{
		Limit:        10000,
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
	// viewResults, err := s.uniquery.GetDataViewMock(ctx, "__dip_o11y_log", uniqueryReq, "run_list")
	if err != nil {
		s.logger.Errorf("[SessionDetail] GetDataView failed: %v", err)
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
	}

	entries := viewResults.Entries

	// 4. 计算session级别的指标
	var (
		sessionRunCount   = len(entries)
		sessionStartTime  int64
		sessionEndTime    int64
		totalExecuteTime  = 0.0
		totalTTFTTime     = 0.0
		runErrorCount     = 0
		toolFailCount     = 0
		firstRun          = true
		validExecuteCount = 0 // 有效执行时间的数据数量
		validTTFTCount    = 0 // 有效ttft的数据数量
	)

	for _, entry := range entries {
		entryMap := entry.(map[string]any)
		attributes := entryMap["Attributes"].(map[string]any)

		startTime := int64(attributes["start_time"].(map[string]any)["Data"].(float64))
		endTime := int64(attributes["end_time"].(map[string]any)["Data"].(float64))
		status := attributes["status"].(map[string]any)["Data"].(string)
		totalTime := attributes["total_time"].(map[string]any)["Data"].(float64)
		ttft := attributes["ttft"].(map[string]any)["Data"].(float64)
		toolCallFailedCount := int(attributes["tool_call_failed_count"].(map[string]any)["Data"].(float64))

		// 确定session的时间范围
		if firstRun {
			sessionStartTime = startTime
			sessionEndTime = endTime
			firstRun = false
		} else {
			if startTime < sessionStartTime {
				sessionStartTime = startTime
			}

			if endTime > sessionEndTime {
				sessionEndTime = endTime
			}
		}

		// 累加执行时间（只累加不为0的数据）
		if totalTime > 0 {
			totalExecuteTime += totalTime
			validExecuteCount++
		}

		// 累加首token响应时间（只累加不为0的数据）
		if ttft > 0 {
			totalTTFTTime += ttft
			validTTFTCount++
		}

		toolFailCount += toolCallFailedCount

		// 统计错误run
		if status != "success" {
			runErrorCount++
		}
	}

	// 计算平均值 单位是ms
	var avgRunExecuteDuration int

	// 计算平均值 单位是毫秒
	var avgRunTTFTDuration int

	if validExecuteCount > 0 {
		avgRunExecuteDuration = int(totalExecuteTime / float64(validExecuteCount))
	}

	if validTTFTCount > 0 {
		avgRunTTFTDuration = int(totalTTFTTime / float64(validTTFTCount))
	}

	sessionDuration := int(sessionEndTime - sessionStartTime)

	// 5. 返回响应
	return &observabilityresp.SessionDetailResp{
		SessionID:             req.SessionID,
		StartTime:             sessionStartTime,
		EndTime:               sessionEndTime,
		SessionRunCount:       sessionRunCount,
		SessionDuration:       sessionDuration,
		AvgRunExecuteDuration: avgRunExecuteDuration,
		AvgRunTTFTDuration:    avgRunTTFTDuration,
		RunErrorCount:         runErrorCount,
		ToolFailCount:         toolFailCount,
	}, nil
}
