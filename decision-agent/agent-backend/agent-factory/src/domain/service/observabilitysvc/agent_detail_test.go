package observabilitysvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squarereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp/uniquerymock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// makeAgentDetailEntry returns an entry suitable for AgentDetail
func makeAgentDetailEntry(sessionID, agentID, status string, totalTime, ttft, toolCallCount, toolCallFailedCount float64, startTime, endTime float64) map[string]any {
	return map[string]any{
		"Attributes": map[string]any{
			"session_id":             map[string]any{"Data": sessionID},
			"agent_id":               map[string]any{"Data": agentID},
			"status":                 map[string]any{"Data": status},
			"total_time":             map[string]any{"Data": totalTime},
			"ttft":                   map[string]any{"Data": ttft},
			"tool_call_count":        map[string]any{"Data": toolCallCount},
			"tool_call_failed_count": map[string]any{"Data": toolCallFailedCount},
			"start_time":             map[string]any{"Data": startTime},
			"end_time":               map[string]any{"Data": endTime},
		},
	}
}

func newSvcWithSquare(ctrl *gomock.Controller) (*observabilitySvc, *v3portdrivermock.MockISquareSvc) { //nolint:unused
	mockUniquery := uniquerymock_new(ctrl)
	mockLogger := cmpmock_new(ctrl)
	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	svc := &observabilitySvc{
		logger:    mockLogger,
		uniquery:  mockUniquery,
		squareSvc: mockSquare,
	}

	return svc, mockSquare
}

func newSvcWithAll(ctrl *gomock.Controller) (*observabilitySvc, *uniquerymock.MockIUniquery, *cmpmock.MockLogger, *v3portdrivermock.MockISquareSvc) {
	mu := uniquerymock_new(ctrl)
	ml := cmpmock_new(ctrl)
	ms := v3portdrivermock.NewMockISquareSvc(ctrl)
	svc := &observabilitySvc{logger: ml, uniquery: mu, squareSvc: ms}

	return svc, mu, ml, ms
}

func TestObservabilitySvc_AgentDetail_Success_EmptyEntries(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{
		AgentID:      "agent-1",
		AgentVersion: "v1",
		StartTime:    1000000,
		EndTime:      2000000,
	}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.AgentDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "agent-1", resp.Agent.ID)
	assert.Equal(t, 0, resp.TotalRequests)
	assert.Equal(t, 0, resp.TotalSessions)
	assert.Equal(t, float32(0), resp.RunSuccessRate)
}

func TestObservabilitySvc_AgentDetail_AgentKeyLookupSuccess(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{
		AgentID:      "agent-key",
		AgentVersion: "v1",
		StartTime:    1000000,
		EndTime:      2000000,
	}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.AssignableToTypeOf(&squarereq.AgentInfoReq{})).
		DoAndReturn(func(_ context.Context, gotReq *squarereq.AgentInfoReq) (*squareresp.AgentMarketAgentInfoResp, error) {
			assert.Equal(t, "agent-key", gotReq.AgentID)
			return &squareresp.AgentMarketAgentInfoResp{}, nil
		})
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.AgentDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "agent-key", resp.Agent.ID)
	assert.Equal(t, 0, resp.TotalRequests)
}

func TestObservabilitySvc_AgentDetail_Success_WithEntries(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{
		AgentID:      "agent-1",
		AgentVersion: "v1",
		StartTime:    1000000,
		EndTime:      2000000,
	}

	agentResp := &squareresp.AgentMarketAgentInfoResp{}
	agentResp.Name = "Test Agent"

	// 4 runs: 3 success (2 different sessions), 1 failed
	entries := []interface{}{
		makeAgentDetailEntry("session-1", "agent-1", "success", 2.0, 200, 3, 0, 1000000, 1500000),
		makeAgentDetailEntry("session-1", "agent-1", "success", 3.0, 300, 2, 1, 1100000, 1600000),
		makeAgentDetailEntry("session-2", "agent-1", "success", 1.5, 150, 0, 0, 1200000, 1700000),
		makeAgentDetailEntry("session-2", "agent-1", "failed", 0, 0, 1, 1, 1300000, 1800000),
	}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(agentResp, nil)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.AgentDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 4, resp.TotalRequests)
	assert.Equal(t, 2, resp.TotalSessions)            // session-1, session-2
	assert.Equal(t, 2, resp.AvgSessionRounds)         // 4/2 = 2
	assert.Equal(t, float32(75), resp.RunSuccessRate) // 3/4 * 100
	// avgExecuteDuration: (2+3+1.5)/3 = 2.16... → int = 2
	assert.Equal(t, 2, resp.AvgExecuteDuration)
	// tool success rate: (3+2+0+1 - (0+1+0+1)) / (3+2+0+1) = 4/6 * 100 ≈ 66
	assert.InDelta(t, 66.66, float64(resp.ToolSuccessRate), 1.0)
}

func TestObservabilitySvc_AgentDetail_GetAgentInfoError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, ml, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{AgentID: "agent-1", AgentVersion: "v1"}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(nil, errors.New("square unavailable"))
	ml.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.AgentDetail(ctx, req)
	assert.Error(t, err)
}

func TestObservabilitySvc_AgentDetail_UniqueryError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, ml, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{AgentID: "agent-1", AgentVersion: "v1", StartTime: 1000000, EndTime: 2000000}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("db error"))
	ml.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.AgentDetail(ctx, req)
	assert.Error(t, err)
}

func TestObservabilitySvc_AgentDetail_ZeroValues(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{AgentID: "a", AgentVersion: "v1", StartTime: 0, EndTime: 999}

	entries := []interface{}{
		makeAgentDetailEntry("s1", "a", "success", 0, 0, 0, 0, 100, 200),
	}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.AgentDetail(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, float32(0), resp.ToolSuccessRate)
	assert.Equal(t, 0, resp.AvgExecuteDuration)
	assert.Equal(t, 0, resp.AvgTTFTDuration)
}

func TestObservabilitySvc_AgentDetail_AllRunsFailed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newSvcWithAll(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AgentDetailReq{AgentID: "agent-1", AgentVersion: "v1", StartTime: 1000000, EndTime: 5000000}

	entries := []interface{}{
		makeAgentDetailEntry("session-1", "agent-1", "failed", 2.0, 200, 1, 1, 1000000, 2000000),
		makeAgentDetailEntry("session-1", "agent-1", "failed", 3.0, 300, 1, 1, 2000000, 3000000),
	}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.AgentDetail(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, float32(0), resp.RunSuccessRate) // all failed
	assert.Equal(t, float32(0), resp.ToolSuccessRate)
}
