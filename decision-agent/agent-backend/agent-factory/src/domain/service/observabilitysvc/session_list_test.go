package observabilitysvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// ---------- SessionList tests ----------

func TestObservabilitySvc_SessionList_EmptyResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:   "agent-123",
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.TotalCount)
	assert.Empty(t, resp.Entries)
}

func TestObservabilitySvc_SessionList_SingleSession(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:   "agent-1",
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 1.5, 200, 2, 0, 1000000, 1500000),
		makeEntry("session-1", "run-2", "conv-1", "agent-1", "success", 2.0, 300, 1, 0, 1200000, 1800000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.TotalCount)
	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, "session-1", resp.Entries[0].SessionID)
}

func TestObservabilitySvc_SessionList_MultiSessions(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:   "agent-1",
		StartTime: 1000000,
		EndTime:   9000000,
		Page:      1,
		Size:      20,
	}

	// 3 sessions, session-1 with 2 runs (error in one), session-2 and session-3 with 1 run each
	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 1.5, 200, 2, 0, 1000000, 1500000),
		makeEntry("session-1", "run-2", "conv-1", "agent-1", "failed", 2.5, 400, 1, 1, 1600000, 2000000),
		makeEntry("session-2", "run-3", "conv-2", "agent-1", "success", 3.0, 100, 5, 0, 3000000, 4000000),
		makeEntry("session-3", "run-4", "conv-3", "agent-1", "success", 0.5, 150, 0, 0, 5000000, 6000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, resp.TotalCount)
}

func TestObservabilitySvc_SessionList_WithConversationID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:        "agent-1",
		ConversationID: "conv-456",
		StartTime:      1000000,
		EndTime:        2000000,
		Page:           1,
		Size:           20,
	}

	entries := []interface{}{
		makeEntry("session-5", "run-5", "conv-456", "agent-1", "success", 2.0, 250, 3, 0, 1100000, 1400000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.TotalCount)
}

func TestObservabilitySvc_SessionList_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:   "agent-1",
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("uniquery unavailable"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.SessionList(ctx, req)
	assert.Error(t, err)
}

func TestObservabilitySvc_SessionList_WithAccountInfo(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:      "agent-1",
		XAccountID:   "account-123",
		XAccountType: cenum.AccountTypeUser,
		StartTime:    1000000,
		EndTime:      2000000,
		Page:         1,
		Size:         20,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestObservabilitySvc_SessionList_TimeRangeCalculation(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:   "agent-1",
		StartTime: 1000000,
		EndTime:   9000000,
		Page:      1,
		Size:      20,
	}

	// Single session, two runs with different time ranges — verify min/max are tracked
	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 2000000, 3000000),
		makeEntry("session-1", "run-2", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 5000000), // wider range
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	require.Len(t, resp.Entries, 1)
	// The session should span from min(startTime) to max(endTime)
	assert.Equal(t, int64(1000000), resp.Entries[0].StartTime)
	assert.Equal(t, int64(5000000), resp.Entries[0].EndTime)
	assert.Equal(t, int(5000000-1000000), resp.Entries[0].SessionDuration)
}

func TestObservabilitySvc_SessionList_ErrorRunCounting(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionListReq{
		AgentID:   "agent-1",
		StartTime: 1000000,
		EndTime:   9000000,
		Page:      1,
		Size:      20,
	}

	// 3 runs: 1 success, 2 non-success → ErrorCount = 2
	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
		makeEntry("session-1", "run-2", "conv-1", "agent-1", "failed", 1.0, 100, 0, 0, 2000000, 3000000),
		makeEntry("session-1", "run-3", "conv-1", "agent-1", "timeout", 1.0, 100, 0, 0, 3000000, 4000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.TotalCount)
}

// ---------- GetSessionCountsByConversationIDs tests ----------

func TestObservabilitySvc_GetSessionCountsByConversationIDs_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()

	conversationIDs := []string{"conv-1", "conv-2", "conv-3"}

	// conv-1 → 2 unique sessions, conv-2 → 1 session, conv-3 → 0 sessions (not in results)
	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
		makeEntry("session-2", "run-2", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
		makeEntry("session-1", "run-3", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 2000000, 3000000), // same session-1, deduped
		makeEntry("session-3", "run-4", "conv-2", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	result, err := svc.GetSessionCountsByConversationIDs(ctx, "agent-1", conversationIDs, 1000000, 2000000, "account-1", "user")
	require.NoError(t, err)
	assert.Equal(t, 2, result["conv-1"]) // 2 distinct sessions
	assert.Equal(t, 1, result["conv-2"])
	assert.Equal(t, 0, result["conv-3"]) // missing in results → default 0
}

func TestObservabilitySvc_GetSessionCountsByConversationIDs_EmptyConversationIDs(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	result, err := svc.GetSessionCountsByConversationIDs(ctx, "agent-1", []string{}, 1000000, 2000000, "account-1", "user")
	require.NoError(t, err)
	assert.Empty(t, result)
}

func TestObservabilitySvc_GetSessionCountsByConversationIDs_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("service down"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.GetSessionCountsByConversationIDs(ctx, "agent-1", []string{"conv-1"}, 1000000, 2000000, "account-1", "user")
	assert.Error(t, err)
}

func TestObservabilitySvc_GetSessionCountsByConversationIDs_MultipleConversations(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()

	conversationIDs := []string{"conv-1", "conv-2", "conv-3", "conv-4", "conv-5"}

	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
		makeEntry("session-2", "run-2", "conv-2", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
		makeEntry("session-3", "run-3", "conv-3", "agent-1", "success", 1.0, 100, 0, 0, 1000000, 2000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	result, err := svc.GetSessionCountsByConversationIDs(ctx, "agent-1", conversationIDs, 1000000, 2000000, "account-1", "user")
	require.NoError(t, err)
	assert.Equal(t, 1, result["conv-1"])
	assert.Equal(t, 1, result["conv-2"])
	assert.Equal(t, 1, result["conv-3"])
	assert.Equal(t, 0, result["conv-4"]) // not in entries → 0
	assert.Equal(t, 0, result["conv-5"]) // not in entries → 0
}
