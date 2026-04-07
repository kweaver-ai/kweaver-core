package observabilitysvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/uniqueryaccess/uniquerydto"
	observabilityreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/observability/req"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestObservabilitySvc_SessionDetail_EmptyResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionDetailReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.SessionDetail(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "session-1", resp.SessionID)
	assert.Equal(t, 0, resp.SessionRunCount)
	assert.Equal(t, 0, resp.AvgRunExecuteDuration)
	assert.Equal(t, 0, resp.AvgRunTTFTDuration)
}

func TestObservabilitySvc_SessionDetail_SingleRun(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionDetailReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   5000000,
	}

	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 2.0, 300, 3, 1, 1000000, 2000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "session-1", resp.SessionID)
	assert.Equal(t, 1, resp.SessionRunCount)
	assert.Equal(t, int64(1000000), resp.StartTime)
	assert.Equal(t, int64(2000000), resp.EndTime)
	assert.Equal(t, int(1000000), resp.SessionDuration)
	assert.Equal(t, int(2), resp.AvgRunExecuteDuration) // int(2.0/1)
	assert.Equal(t, int(300), resp.AvgRunTTFTDuration)
	assert.Equal(t, 0, resp.RunErrorCount)
	assert.Equal(t, 1, resp.ToolFailCount)
}

func TestObservabilitySvc_SessionDetail_MultipleRuns(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionDetailReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   9000000,
	}

	// 3 runs: 1 success (totalTime=2.0), 1 failed (totalTime=3.0), 1 success (totalTime=0 → skipped from avg)
	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "success", 2.0, 200, 2, 0, 2000000, 4000000),
		makeEntry("session-1", "run-2", "conv-1", "agent-1", "failed", 3.0, 400, 1, 1, 1000000, 5000000),
		makeEntry("session-1", "run-3", "conv-1", "agent-1", "success", 0, 0, 0, 0, 3000000, 6000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionDetail(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 3, resp.SessionRunCount)
	// time range: min start = 1000000, max end = 6000000
	assert.Equal(t, int64(1000000), resp.StartTime)
	assert.Equal(t, int64(6000000), resp.EndTime)
	assert.Equal(t, int(5000000), resp.SessionDuration)
	// avgRunExecuteDuration: (2+3)/2 = 2 (only 2 valid runs with totalTime>0)
	assert.Equal(t, int(2), resp.AvgRunExecuteDuration)
	// avgRunTTFTDuration: (200+400)/2 = 300
	assert.Equal(t, int(300), resp.AvgRunTTFTDuration)
	// 1 failed run
	assert.Equal(t, 1, resp.RunErrorCount)
	// tool fail from run-2
	assert.Equal(t, 1, resp.ToolFailCount)
}

func TestObservabilitySvc_SessionDetail_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionDetailReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   2000000,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("network error"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.SessionDetail(ctx, req)
	assert.Error(t, err)
}

func TestObservabilitySvc_SessionDetail_WithAgentIDAndConversationID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionDetailReq{
		SessionID:      "session-1",
		AgentID:        "agent-1",
		ConversationID: "conv-456",
		StartTime:      1000000,
		EndTime:        5000000,
	}

	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-456", "agent-1", "success", 1.5, 150, 2, 0, 1000000, 2000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionDetail(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 1, resp.SessionRunCount)
}

func TestObservabilitySvc_SessionDetail_AllRunsFailed(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.SessionDetailReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   9000000,
	}

	entries := []interface{}{
		makeEntry("session-1", "run-1", "conv-1", "agent-1", "failed", 1.0, 100, 0, 0, 1000000, 2000000),
		makeEntry("session-1", "run-2", "conv-1", "agent-1", "timeout", 2.0, 200, 0, 0, 2000000, 3000000),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.SessionDetail(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 2, resp.RunErrorCount)
}
