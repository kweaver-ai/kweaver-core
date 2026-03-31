package observabilitysvc

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/apierr"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func (s *observabilitySvc) AgentDetail(ctx context.Context, req *observabilityreq.AgentDetailReq) (*observabilityresp.AgentResp, error) {
	// 1. 获取agent信息
	agentInfo, err := s.squareSvc.GetAgentInfoByIDOrKey(ctx, &squarereq.AgentInfoReq{
		AgentID:      req.AgentID,
		AgentVersion: req.AgentVersion,
	})
	if err != nil {
		s.logger.Errorf("[AgentDetail] get agent config failed: %v", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_Agent_GetAgentFailed).WithErrorDetails(err.Error())
	}

	// 2. 通过uniquery查询agent-app的日志数据
	uniqueryReq := uniquerydto.ReqDataView{
		Limit:        10000,
		Offset:       0,
		Start:        req.StartTime,
		End:          req.EndTime,
		Format:       "original",
		XAccountID:   req.XAccountID,
		XAccountType: string(req.XAccountType),
		Filters: &uniquerydto.GlobalFilter{
			Operation: "and",
			SubConditions: []uniquerydto.Condition{
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
					Value:     req.AgentID,
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
			},
		},
	}
	// 通过GetDataView 查询出来的是run的详情列表，根据run信息计算指标
	viewResults, err := s.uniquery.GetDataView(ctx, "__dip_o11y_log", uniqueryReq)
	// viewResults, err := s.uniquery.GetDataViewMock(ctx, "__dip_o11y_log", uniqueryReq, "session_list")
	if err != nil {
		s.logger.Errorf("[AgentDetail] get data view failed: %v", err.Error())
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, apierr.AgentAPP_InternalError).WithErrorDetails(err.Error())
	}

	entries := viewResults.Entries

	// 3. 计算Agent级别指标
	var (
		totalRequests     = len(entries)
		sessionSet        = make(map[string]bool)
		successRunCount   = 0
		totalExecuteTime  = 0.0
		totalTTFTTime     = 0.0
		totalToolCalls    = 0
		totalToolFailures = 0
		validExecuteCount = 0 // 有效执行时间的数据数量
		validTTFTCount    = 0 // 有效ttft的数据数量
	)

	// NOTE: 计算指标时，应该剔除一些数据，比如响应时间为0和ttft为0的数据
	for _, entry := range entries {
		entryMap := entry.(map[string]interface{})
		attributes := entryMap["Attributes"].(map[string]interface{})

		// 统计唯一session
		if sessionID, ok := attributes["session_id"].(map[string]interface{})["Data"].(string); ok {
			sessionSet[sessionID] = true
		}

		// 统计成功run数量
		if status, ok := attributes["status"].(map[string]interface{})["Data"].(string); ok && status == "success" {
			successRunCount++
		}

		// 累加执行时间（只累加不为0的数据）
		if totalTime, ok := attributes["total_time"].(map[string]interface{})["Data"].(float64); ok && totalTime > 0 {
			totalExecuteTime += totalTime
			validExecuteCount++
		}

		// 累加首token响应时间（只累加不为0的数据）
		if ttft, ok := attributes["ttft"].(map[string]interface{})["Data"].(float64); ok && ttft > 0 {
			totalTTFTTime += ttft
			validTTFTCount++
		}

		// 统计工具调用情况
		if toolCallCount, ok := attributes["tool_call_count"].(map[string]interface{})["Data"].(float64); ok {
			totalToolCalls += int(toolCallCount)
		}

		if toolCallFailedCount, ok := attributes["tool_call_failed_count"].(map[string]interface{})["Data"].(float64); ok {
			totalToolFailures += int(toolCallFailedCount)
		}
	}

	// 4. 计算各项指标
	totalSessions := len(sessionSet)

	var avgSessionRounds int

	if totalSessions > 0 {
		avgSessionRounds = totalRequests / totalSessions
	}

	var runSuccessRate float32
	if totalRequests > 0 {
		runSuccessRate = float32(successRunCount) / float32(totalRequests) * 100
	}

	var avgExecuteDuration int
	if validExecuteCount > 0 {
		avgExecuteDuration = int(totalExecuteTime / float64(validExecuteCount))
	}

	var avgTTFTDuration int
	if validTTFTCount > 0 {
		avgTTFTDuration = int(totalTTFTTime / float64(validTTFTCount))
	}

	var toolSuccessRate float32
	if totalToolCalls > 0 {
		toolSuccessRate = float32(totalToolCalls-totalToolFailures) / float32(totalToolCalls) * 100
	}

	// 5. 返回结果
	return &observabilityresp.AgentResp{
		Agent: observabilityresp.Agent{
			ID:      req.AgentID,
			Version: req.AgentVersion,
			Name:    agentInfo.DataAgent.Name,
			Config:  agentInfo.Config,
		},
		TotalRequests:      totalRequests,
		TotalSessions:      totalSessions,
		AvgSessionRounds:   avgSessionRounds,
		RunSuccessRate:     runSuccessRate,
		AvgExecuteDuration: avgExecuteDuration,
		AvgTTFTDuration:    avgTTFTDuration,
		ToolSuccessRate:    toolSuccessRate,
	}, nil
}
