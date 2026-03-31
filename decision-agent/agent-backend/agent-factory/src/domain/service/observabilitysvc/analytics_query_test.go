package observabilitysvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/square/squareresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iuniqueryhttp/uniquerymock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// newFullSvc creates svc with all three dependencies mocked
func newFullSvc(ctrl *gomock.Controller) (*observabilitySvc, *uniquerymock.MockIUniquery, *cmpmock.MockLogger, *v3portdrivermock.MockISquareSvc) {
	mu := uniquerymock.NewMockIUniquery(ctrl)
	ml := cmpmock.NewMockLogger(ctrl)
	ms := v3portdrivermock.NewMockISquareSvc(ctrl)
	svc := &observabilitySvc{logger: ml, uniquery: mu, squareSvc: ms}

	return svc, mu, ml, ms
}

// ---------- AnalyticsQuery: unknown / default case ----------

func TestObservabilitySvc_AnalyticsQuery_UnknownLevel(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, ml, _ := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "unknown_level",
		ID:            "some-id",
	}

	ml.EXPECT().Warnf(gomock.Any(), gomock.Any()).Times(1)

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.False(t, resp.Success)
	assert.Contains(t, resp.Error, "unknown analysis level")
}

func TestObservabilitySvc_AnalyticsQuery_EmptyLevel(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, ml, _ := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{AnalysisLevel: "", ID: "x"}

	ml.EXPECT().Warnf(gomock.Any(), gomock.Any()).Times(1)

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

// ---------- AnalyticsQuery: agent level ----------

func TestObservabilitySvc_AnalyticsQuery_AgentLevel_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "agent",
		ID:            "agent-1",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	// AgentDetail calls squareSvc.GetAgentInfo + uniquery.GetDataView
	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil).Times(1)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil).Times(1)

	// getAgentMetrics also calls squareSvc.GetAgentInfo again for agent config
	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil).Times(1)

	// SessionList calls uniquery.GetDataView
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil).Times(1)

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
	assert.NotNil(t, resp.Data)
}

func TestObservabilitySvc_AnalyticsQuery_AgentLevel_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, _, ml, ms := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "agent",
		ID:            "agent-1",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(nil, errors.New("square down"))
	ml.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err) // AnalyticsQuery itself doesn't error, sets resp.Success=false
	assert.False(t, resp.Success)
	assert.NotEmpty(t, resp.Error)
}

// ---------- AnalyticsQuery: session level ----------

func TestObservabilitySvc_AnalyticsQuery_SessionLevel_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, ms := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "session",
		ID:            "session-1",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	// SessionDetail call
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{
			makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 2.0, 200, 2, 0, 1000000, 2000000),
		}}, nil).Times(1)

	// RunList call
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{
			makeRunEntry("agent-1", "run-1", "session-1", "success", 2.0),
		}}, nil).Times(1)

	// getAgentInfo for agent config (from RunList entries[0].AgentID)
	ms.EXPECT().GetAgentInfoByIDOrKey(ctx, gomock.Any()).Return(&squareresp.AgentMarketAgentInfoResp{}, nil).Times(1)

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestObservabilitySvc_AnalyticsQuery_SessionLevel_EmptyRuns(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, _ := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "session",
		ID:            "session-1",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	// SessionDetail returns empty
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil).Times(1)

	// RunList returns empty (no squareSvc call since no runs)
	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil).Times(1)

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestObservabilitySvc_AnalyticsQuery_SessionLevel_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, ml, _ := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "session",
		ID:            "session-1",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("db error"))
	ml.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}

// ---------- AnalyticsQuery: run level ----------

func TestObservabilitySvc_AnalyticsQuery_RunLevel_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, _, _ := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "run",
		ID:            "run-abc",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	entry := makeRunDetailEntry("agent-1", "run-abc", "session-1", "conv-1", "user-1", "chat", "success",
		2.0, 300, 500, 3, 0, 1100000, 1700000)

	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{entry}}, nil)

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.True(t, resp.Success)
}

func TestObservabilitySvc_AnalyticsQuery_RunLevel_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mu, ml, _ := newFullSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.AnalyticsQueryReq{
		AnalysisLevel: "run",
		ID:            "run-abc",
		StartTime:     1000000,
		EndTime:       2000000,
	}

	mu.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("db error"))
	ml.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	resp, err := svc.AnalyticsQuery(ctx, req)
	require.NoError(t, err)
	assert.False(t, resp.Success)
}
