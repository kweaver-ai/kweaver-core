package observabilitysvc

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	observabilityreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/observability/req"
	observabilityresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/observability/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
)

func (s *observabilitySvc) AnalyticsQuery(ctx context.Context, req *observabilityreq.AnalyticsQueryReq) (*observabilityresp.AnalyticsQueryResp, error) {
	// s.logger.Debugf("[AnalyticsQuery] start, analysis_level: %s, id: %s, time_range: [%d, %d]",
	// 	req.AnalysisLevel, req.ID, req.StartTime, req.EndTime)
	resp := &observabilityresp.AnalyticsQueryResp{
		Success: true,
		Error:   "",
	}

	switch req.AnalysisLevel {
	case "agent":
		// 获取Agent级别的指标数据
		agentMetrics, err := s.getAgentMetrics(ctx, req)
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
		} else {
			resp.Data = agentMetrics
		}
	case "session":
		// 获取Session级别的指标数据
		sessionMetrics, err := s.getSessionMetrics(ctx, req)
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
		} else {
			resp.Data = sessionMetrics
		}
	case "run":
		// 获取Run级别的指标数据
		runMetrics, err := s.getRunMetrics(ctx, req)
		if err != nil {
			resp.Success = false
			resp.Error = err.Error()
		} else {
			resp.Data = runMetrics
		}
	default:
		s.logger.Warnf("[AnalyticsQuery] unknown analysis level: %s", req.AnalysisLevel)

		resp.Success = false
		resp.Error = "unknown analysis level: " + req.AnalysisLevel
	}

	return resp, nil
}

// getAgentMetrics 获取Agent级别的指标数据
func (s *observabilitySvc) getAgentMetrics(ctx context.Context, req *observabilityreq.AnalyticsQueryReq) (*observabilityresp.AgentMetrics, error) {
	// 1. 调用AgentDetail方法获取Agent级别的指标
	agentDetailReq := &observabilityreq.AgentDetailReq{
		AgentID:      req.ID,
		AgentVersion: "latest", // 如果version为空，使用latest
		StartTime:    req.StartTime,
		EndTime:      req.EndTime,
		XAccountID:   req.XAccountID,
		XAccountType: req.XAccountType,
	}

	agentResp, err := s.AgentDetail(ctx, agentDetailReq)
	if err != nil {
		return nil, err
	}

	// 2. 获取Agent配置
	var agentConfig daconfvalobj.Config

	agentInfo, agentErr := s.squareSvc.GetAgentInfoByIDOrKey(ctx, &squarereq.AgentInfoReq{
		AgentID:      req.ID,
		AgentVersion: "latest",
	})
	if agentErr == nil && agentInfo != nil {
		agentConfig = agentInfo.Config
	}

	// 3. 获取Session列表 - 使用分页参数获取前10000个session
	sessionListReq := &observabilityreq.SessionListReq{
		AgentID:        req.ID,
		ConversationID: "",
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Page:           1,
		Size:           10000, // 分页大小，获取前10000个session
		XAccountID:     req.XAccountID,
		XAccountType:   req.XAccountType,
	}
	sessionListResp, err := s.SessionList(ctx, sessionListReq)

	var sessionList []observabilityresp.Session

	if err == nil && sessionListResp != nil {
		// 将SessionDetailResp转换为Session
		for _, sessionDetail := range sessionListResp.Entries {
			sessionList = append(sessionList, observabilityresp.Session{
				SessionID:        sessionDetail.SessionID,
				SessionStartTime: formatTimeToISO8601(sessionDetail.StartTime),
				SessionEndTime:   formatTimeToISO8601(sessionDetail.EndTime),
				SessionDuration:  int64(sessionDetail.SessionDuration),
			})
		}
	}

	// 4. 转换为AgentMetrics结构
	return &observabilityresp.AgentMetrics{
		AgentConfig: agentConfig,
		AgentMetrics: observabilityresp.AgentMetric{
			TotalRequests:      agentResp.TotalRequests,
			TotalSessions:      agentResp.TotalSessions,
			AvgSessionRounds:   agentResp.AvgSessionRounds,
			RunSuccessRate:     agentResp.RunSuccessRate,
			AvgExecuteDuration: agentResp.AvgExecuteDuration,
			AvgTTFTDuration:    agentResp.AvgTTFTDuration,
			ToolSuccessRate:    agentResp.ToolSuccessRate,
		},
		SessionList: sessionList,
		TrendData: observabilityresp.TrendData{
			Last7Days:   []any{},
			Last24Hours: []any{},
		},
	}, nil
}

// getSessionMetrics 获取Session级别的指标数据
func (s *observabilitySvc) getSessionMetrics(ctx context.Context, req *observabilityreq.AnalyticsQueryReq) (*observabilityresp.SessionMetrics, error) {
	// 1. 调用SessionDetail方法获取Session级别的指标
	sessionDetailReq := &observabilityreq.SessionDetailReq{
		AgentID:        "", // 这里可能需要从其他地方获取AgentID
		AgentVersion:   "",
		ConversationID: "", // 这里可能需要从其他地方获取ConversationID
		SessionID:      req.ID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		XAccountID:     req.XAccountID,
		XAccountType:   req.XAccountType,
	}

	sessionDetailResp, err := s.SessionDetail(ctx, sessionDetailReq)
	if err != nil {
		return nil, err
	}

	if sessionDetailResp == nil {
		return nil, nil
	}

	// 2. 获取Run列表 - 使用分页参数获取前10000个run
	var runList []observabilityresp.RunItem

	var agentConfig daconfvalobj.Config

	runListReq := &observabilityreq.RunListReq{
		AgentID:        "",
		ConversationID: "",
		SessionID:      req.ID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		Page:           1,
		Size:           10000, // 分页大小，获取前10000个run
		XAccountID:     req.XAccountID,
		XAccountType:   req.XAccountType,
	}

	runListResp, err := s.RunList(ctx, runListReq)
	if err == nil && runListResp != nil {
		for _, runDetail := range runListResp.Entries {
			runList = append(runList, observabilityresp.RunItem{
				RunID:        runDetail.RunID,
				ResponseTime: runDetail.TotalTime, // 使用TotalTime作为响应时间
				Status:       runDetail.Status,
			})
		}
		// 获取Run的AgentID，获取agent配置
		if len(runListResp.Entries) > 0 {
			agentID := runListResp.Entries[0].AgentID

			agentInfo, agentErr := s.squareSvc.GetAgentInfoByIDOrKey(ctx, &squarereq.AgentInfoReq{
				AgentID:      agentID,
				AgentVersion: "latest",
			})
			if agentErr == nil && agentInfo != nil {
				agentConfig = agentInfo.Config
			}
		}
	}

	// 3. 转换为SessionMetrics结构
	return &observabilityresp.SessionMetrics{
		SessionMetrics: observabilityresp.SessionMetric{
			SessionRunCount:       sessionDetailResp.SessionRunCount,
			SessionDuration:       sessionDetailResp.SessionDuration,
			AvgRunExecuteDuration: sessionDetailResp.AvgRunExecuteDuration,
			AvgRunTTFTDuration:    sessionDetailResp.AvgRunTTFTDuration,
			RunErrorCount:         sessionDetailResp.RunErrorCount,
			ToolFailCount:         sessionDetailResp.ToolFailCount,
		},
		AgentConfig: agentConfig,
		RunList:     runList,
	}, nil
}

// getRunMetrics 获取Run级别的指标数据
func (s *observabilitySvc) getRunMetrics(ctx context.Context, req *observabilityreq.AnalyticsQueryReq) (*observabilityresp.RunMetrics, error) {
	// 1. 调用RunDetail方法获取Run级别的详情
	runDetailReq := &observabilityreq.RunDetailReq{
		AgentID:        "", // 这里可能需要从其他地方获取AgentID
		AgentVersion:   "",
		ConversationID: "", // 这里可能需要从其他地方获取ConversationID
		SessionID:      "", // 这里可能需要从其他地方获取SessionID
		RunID:          req.ID,
		StartTime:      req.StartTime,
		EndTime:        req.EndTime,
		XAccountID:     req.XAccountID,
		XAccountType:   req.XAccountType,
	}

	runDetailResp, err := s.RunDetail(ctx, runDetailReq)
	if err != nil {
		return nil, err
	}

	// 2. 转换为RunMetrics结构
	return &observabilityresp.RunMetrics{
		RunID:               runDetailResp.RunID,
		InputMessage:        runDetailResp.InputMessage,
		StartTime:           runDetailResp.StartTime,
		EndTime:             runDetailResp.EndTime,
		TTFT:                runDetailResp.TTFT,
		TotalTime:           runDetailResp.TotalTime,
		TotalTokens:         runDetailResp.TotalTokens,
		ToolCallCount:       runDetailResp.ToolCallCount,
		ToolCallFailedCount: runDetailResp.ToolCallFailedCount,
		Progress:            runDetailResp.Progress,
		Status:              runDetailResp.Status,
	}, nil
}
