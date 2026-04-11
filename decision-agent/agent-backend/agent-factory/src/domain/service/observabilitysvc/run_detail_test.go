package observabilitysvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// makeRunDetailEntry returns an entry map with all fields needed by RunDetail
func makeRunDetailEntry(agentID, runID, sessionID, convID, userID, callType, status string,
	totalTime, ttft, totalTokens, toolCallCount, toolCallFailedCount float64,
	startTime, endTime float64,
) map[string]any {
	return map[string]any{
		"Attributes": map[string]any{
			"agent_id":               map[string]any{"Data": agentID},
			"agent_version":          map[string]any{"Data": "v1"},
			"run_id":                 map[string]any{"Data": runID},
			"session_id":             map[string]any{"Data": sessionID},
			"conversation_id":        map[string]any{"Data": convID},
			"user_id":                map[string]any{"Data": userID},
			"call_type":              map[string]any{"Data": callType},
			"status":                 map[string]any{"Data": status},
			"total_time":             map[string]any{"Data": totalTime},
			"ttft":                   map[string]any{"Data": ttft},
			"total_tokens":           map[string]any{"Data": totalTokens},
			"tool_call_count":        map[string]any{"Data": toolCallCount},
			"tool_call_failed_count": map[string]any{"Data": toolCallFailedCount},
			"start_time":             map[string]any{"Data": startTime},
			"end_time":               map[string]any{"Data": endTime},
			"input_message":          map[string]any{"Data": "hello"},
		},
	}
}

func TestObservabilitySvc_RunDetail_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:     "run-abc",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	entry := makeRunDetailEntry("agent-1", "run-abc", "session-1", "conv-1", "user-1", "chat", "success",
		1.5, 200, 500, 3, 0, 1100000, 1600000)

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{entry}}, nil)

	resp, err := svc.RunDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "run-abc", resp.RunID)
	assert.Equal(t, "agent-1", resp.AgentID)
	assert.Equal(t, "session-1", resp.SessionID)
	assert.Equal(t, "success", resp.Status)
	assert.Equal(t, int64(1100000), resp.StartTime)
	assert.Equal(t, int64(1600000), resp.EndTime)
	assert.Equal(t, int(200), resp.TTFT)
	assert.Equal(t, 3, resp.ToolCallCount)
	assert.Equal(t, 0, resp.ToolCallFailedCount)
}

func TestObservabilitySvc_RunDetail_NotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:     "run-unknown",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := svc.RunDetail(ctx, req)
	require.NoError(t, err)
	assert.Nil(t, resp)
}

func TestObservabilitySvc_RunDetail_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:     "run-abc",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("db error"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.RunDetail(ctx, req)
	assert.Error(t, err)
}

func TestObservabilitySvc_RunDetail_WithAgentIDAndSessionID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:          "run-abc",
		AgentID:        "agent-1",
		ConversationID: "conv-1",
		SessionID:      "session-1",
		StartTime:      1000000,
		EndTime:        2000000,
	}

	entry := makeRunDetailEntry("agent-1", "run-abc", "session-1", "conv-1", "user-1", "chat", "success",
		2.0, 150, 400, 2, 1, 1200000, 1700000)

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{entry}}, nil)

	resp, err := svc.RunDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "run-abc", resp.RunID)
	assert.Equal(t, 1, resp.ToolCallFailedCount)
}

func TestObservabilitySvc_RunDetail_WithProgressJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:     "run-xyz",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	entry := map[string]any{
		"Attributes": map[string]any{
			"agent_id":               map[string]any{"Data": "agent-1"},
			"agent_version":          map[string]any{"Data": "v1"},
			"run_id":                 map[string]any{"Data": "run-xyz"},
			"session_id":             map[string]any{"Data": "session-1"},
			"conversation_id":        map[string]any{"Data": "conv-1"},
			"user_id":                map[string]any{"Data": "user-1"},
			"call_type":              map[string]any{"Data": "chat"},
			"status":                 map[string]any{"Data": "success"},
			"total_time":             map[string]any{"Data": 1.0},
			"ttft":                   map[string]any{"Data": 100.0},
			"total_tokens":           map[string]any{"Data": 300.0},
			"tool_call_count":        map[string]any{"Data": 0.0},
			"tool_call_failed_count": map[string]any{"Data": 0.0},
			"start_time":             map[string]any{"Data": 1100000.0},
			"end_time":               map[string]any{"Data": 1600000.0},
			"input_message":          map[string]any{"Data": "hello"},
			// progress as JSON string
			"progress": map[string]any{"Data": `[{"id":"prog-1","stage":"llm","status":"completed","answer":"hi","think":""}]`},
		},
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{entry}}, nil)

	resp, err := svc.RunDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Len(t, resp.Progress, 1)
	assert.Equal(t, "prog-1", resp.Progress[0].ID)
}

func TestObservabilitySvc_RunDetail_WithInvalidProgressJSON(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:     "run-xyz",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	// progress with invalid JSON string
	entry := map[string]any{
		"Attributes": map[string]any{
			"agent_id":               map[string]any{"Data": "agent-1"},
			"agent_version":          map[string]any{"Data": "v1"},
			"run_id":                 map[string]any{"Data": "run-xyz"},
			"session_id":             map[string]any{"Data": "session-1"},
			"conversation_id":        map[string]any{"Data": "conv-1"},
			"user_id":                map[string]any{"Data": "user-1"},
			"call_type":              map[string]any{"Data": "chat"},
			"status":                 map[string]any{"Data": "success"},
			"total_time":             map[string]any{"Data": 1.0},
			"ttft":                   map[string]any{"Data": 100.0},
			"total_tokens":           map[string]any{"Data": 300.0},
			"tool_call_count":        map[string]any{"Data": 0.0},
			"tool_call_failed_count": map[string]any{"Data": 0.0},
			"start_time":             map[string]any{"Data": 1100000.0},
			"end_time":               map[string]any{"Data": 1600000.0},
			"input_message":          map[string]any{"Data": "hello"},
			"progress":               map[string]any{"Data": "invalid-json{"},
		},
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{entry}}, nil)
	mockLogger.EXPECT().Warnf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := svc.RunDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	// progress should be empty due to parse error
	assert.Empty(t, resp.Progress)
}

func TestObservabilitySvc_RunDetail_WithNonStringProgress(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunDetailReq{
		RunID:     "run-xyz",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	// progress data is NOT a string (slice instead) → triggers Warnf "not a string"
	entry := map[string]any{
		"Attributes": map[string]any{
			"agent_id":               map[string]any{"Data": "agent-1"},
			"agent_version":          map[string]any{"Data": "v1"},
			"run_id":                 map[string]any{"Data": "run-xyz"},
			"session_id":             map[string]any{"Data": "session-1"},
			"conversation_id":        map[string]any{"Data": "conv-1"},
			"user_id":                map[string]any{"Data": "user-1"},
			"call_type":              map[string]any{"Data": "chat"},
			"status":                 map[string]any{"Data": "success"},
			"total_time":             map[string]any{"Data": 1.0},
			"ttft":                   map[string]any{"Data": 100.0},
			"total_tokens":           map[string]any{"Data": 300.0},
			"tool_call_count":        map[string]any{"Data": 0.0},
			"tool_call_failed_count": map[string]any{"Data": 0.0},
			"start_time":             map[string]any{"Data": 1100000.0},
			"end_time":               map[string]any{"Data": 1600000.0},
			"input_message":          map[string]any{"Data": "hello"},
			"progress":               map[string]any{"Data": []any{"item1", "item2"}}, // not a string
		},
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{entry}}, nil)
	mockLogger.EXPECT().Warnf(gomock.Any()).AnyTimes()

	resp, err := svc.RunDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Empty(t, resp.Progress)
}
