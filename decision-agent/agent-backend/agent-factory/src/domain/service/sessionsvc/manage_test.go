package sessionsvc

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/session/sessionreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/ctype"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/iredisaccess/isessionredis/isessionredismock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestManage_UnsupportedAction(t *testing.T) {
	t.Parallel()

	svc := &sessionSvc{}

	req := sessionreq.ManageReq{
		Action: "unsupported_action",
	}

	resp, err := svc.Manage(context.Background(), req, nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported action")
	assert.Empty(t, resp.ConversationSessionID)
}

func TestManage_ValidActionsWithoutDeps(t *testing.T) {
	t.Parallel()

	// Test that valid actions are properly routed
	// Even though they will fail due to missing dependencies,
	// this verifies the routing logic works correctly
	svc := &sessionSvc{}

	testCases := []struct {
		name   string
		action sessionreq.SessionManageActionType
	}{
		{
			name:   "GetInfoOrCreate action",
			action: sessionreq.SessionManageActionGetInfoOrCreate,
		},
		{
			name:   "RecoverLifetimeOrCreate action",
			action: sessionreq.SessionManageActionRecoverLifetimeOrCreate,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := sessionreq.ManageReq{
				Action:         tc.action,
				ConversationID: "conv-123",
				AgentID:        "agent-123",
				AgentVersion:   "v1.0.0",
			}

			// Since HandleGetInfoOrCreate and HandleRecoverLifetimeOrCreate
			// require sessionRedisAcc to be set up, and we're testing without it,
			// these will cause a panic or error.
			// We catch the panic to verify the routing works.
			assert.Panics(t, func() {
				_, _ = svc.Manage(context.Background(), req, nil)
			})
		})
	}
}

func TestTriggerAgentCacheUpsert_NoAgentExecutor(t *testing.T) {
	t.Parallel()

	svc := &sessionSvc{
		// agentExecutorV1 is nil
	}

	req := sessionreq.ManageReq{
		AgentID:      "agent-123",
		AgentVersion: "v1.0.0",
	}

	// This should panic because agentExecutorV1 is nil
	assert.Panics(t, func() {
		_ = svc.triggerAgentCacheUpsert(context.Background(), req, nil)
	})
}

func TestManage_HandleGetInfoOrCreate_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	ctx := context.Background()
	req := sessionreq.ManageReq{
		Action:         sessionreq.SessionManageActionGetInfoOrCreate,
		ConversationID: "conv-123",
	}
	visitorInfo := &ctype.VisitorInfo{}

	// Mock Redis to return error
	expectedErr := errors.New("redis connection failed")
	mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(false, int64(0), 0, expectedErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	resp, err := svc.Manage(ctx, req, visitorInfo)

	assert.Error(t, err)
	assert.Empty(t, resp.ConversationSessionID)
	assert.Equal(t, int64(0), resp.StartTime)
	assert.Equal(t, 0, resp.TTL)
}

func TestManage_HandleRecoverLifetimeOrCreate_Error(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	ctx := context.Background()
	req := sessionreq.ManageReq{
		Action:         sessionreq.SessionManageActionRecoverLifetimeOrCreate,
		ConversationID: "conv-456",
	}
	visitorInfo := &ctype.VisitorInfo{}

	// Mock Redis to return error
	expectedErr := errors.New("redis connection failed")
	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-456", gomock.Any()).Return(false, int64(0), expectedErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	resp, err := svc.Manage(ctx, req, visitorInfo)

	assert.Error(t, err)
	assert.Empty(t, resp.ConversationSessionID)
	assert.Equal(t, int64(0), resp.StartTime)
	assert.Equal(t, 0, resp.TTL)
}

func TestManage_HandleGetInfoOrCreate_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	ctx := context.Background()
	req := sessionreq.ManageReq{
		Action:         sessionreq.SessionManageActionGetInfoOrCreate,
		ConversationID: "conv-789",
	}
	visitorInfo := &ctype.VisitorInfo{}

	existingStartTime := int64(1234567890)
	existingTTL := 3600

	mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-789").Return(true, existingStartTime, existingTTL, nil)

	resp, err := svc.Manage(ctx, req, visitorInfo)

	assert.NoError(t, err)
	assert.Equal(t, "conv-789-1234567890", resp.ConversationSessionID)
	assert.Equal(t, existingStartTime, resp.StartTime)
	assert.Equal(t, existingTTL, resp.TTL)
}

func TestManage_HandleRecoverLifetimeOrCreate_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	ctx := context.Background()
	req := sessionreq.ManageReq{
		Action:         sessionreq.SessionManageActionRecoverLifetimeOrCreate,
		ConversationID: "conv-999",
	}
	visitorInfo := &ctype.VisitorInfo{}

	existingStartTime := int64(9876543210)
	existingTTL := 7200

	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-999", gomock.Any()).Return(true, existingStartTime, nil)
	mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-999").Return(existingTTL, nil)
	// Allow any number of Errorln calls from panic recovery in the goroutine
	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

	resp, err := svc.Manage(ctx, req, visitorInfo)

	assert.NoError(t, err)
	assert.Equal(t, "conv-999-9876543210", resp.ConversationSessionID)
	assert.Equal(t, existingStartTime, resp.StartTime)
	assert.Equal(t, existingTTL, resp.TTL)
}
