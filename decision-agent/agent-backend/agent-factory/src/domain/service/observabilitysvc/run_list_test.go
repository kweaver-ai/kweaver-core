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

func TestObservabilitySvc_RunList_EmptyResult(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunListReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.RunList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 0, resp.TotalCount)
	assert.Empty(t, resp.Entries)
}

func TestObservabilitySvc_RunList_MultipleRuns(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunListReq{
		SessionID: "session-1",
		AgentID:   "agent-1",
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	entries := []interface{}{
		makeRunEntry("agent-1", "run-1", "session-1", "success", 1.5),
		makeRunEntry("agent-1", "run-2", "session-1", "failed", 2.5),
		makeRunEntry("agent-1", "run-3", "session-1", "success", 0.8),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.RunList(ctx, req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 3, resp.TotalCount)
	assert.Len(t, resp.Entries, 3)
	assert.Equal(t, "run-1", resp.Entries[0].RunID)
	assert.Equal(t, "success", resp.Entries[0].Status)
	assert.Equal(t, "run-2", resp.Entries[1].RunID)
	assert.Equal(t, "failed", resp.Entries[1].Status)
}

func TestObservabilitySvc_RunList_WithConversationID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunListReq{
		SessionID:      "session-1",
		AgentID:        "agent-1",
		ConversationID: "conv-456",
		StartTime:      1000000,
		EndTime:        2000000,
		Page:           1,
		Size:           20,
	}

	entries := []interface{}{
		makeRunEntry("agent-1", "run-1", "session-1", "success", 1.5),
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: entries}, nil)

	resp, err := svc.RunList(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 1, resp.TotalCount)
}

func TestObservabilitySvc_RunList_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, mockLogger := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunListReq{
		SessionID: "session-1",
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{}, errors.New("connection failed"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.RunList(ctx, req)
	assert.Error(t, err)
}

func TestObservabilitySvc_RunList_NoAgentIDNoConversationID(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	svc, mockUniquery, _ := newTestSvc(ctrl)
	ctx := context.Background()
	req := &observabilityreq.RunListReq{
		SessionID: "session-1",
		// no AgentID, no ConversationID → these conditions are skipped
		StartTime: 1000000,
		EndTime:   2000000,
		Page:      1,
		Size:      20,
	}

	mockUniquery.EXPECT().GetDataView(ctx, "__dip_o11y_log", gomock.Any()).
		Return(uniquerydto.ViewResults{Entries: []interface{}{}}, nil)

	resp, err := svc.RunList(ctx, req)
	require.NoError(t, err)
	assert.Equal(t, 0, resp.TotalCount)
}
