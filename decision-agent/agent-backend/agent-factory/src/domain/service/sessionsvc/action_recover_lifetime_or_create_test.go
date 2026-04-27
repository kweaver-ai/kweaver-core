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

func TestHandleRecoverLifetimeOrCreate_SessionExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	req := sessionreq.ManageReq{
		ConversationID: "conv-123",
		AgentID:        "agent-123",
		AgentVersion:   "v1.0.0",
	}
	visitorInfo := &ctype.VisitorInfo{}

	existingStartTime := int64(1234567890)
	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-123", gomock.Any()).Return(true, existingStartTime, nil)
	mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-123").Return(3600, nil)

	startTime, ttl, err := svc.HandleRecoverLifetimeOrCreate(context.Background(), req, visitorInfo, false)

	assert.NoError(t, err)
	assert.Equal(t, existingStartTime, startTime)
	assert.Equal(t, 3600, ttl)
}

func TestHandleRecoverLifetimeOrCreate_SessionNotExists(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	req := sessionreq.ManageReq{
		ConversationID: "conv-123",
		AgentID:        "agent-123",
		AgentVersion:   "v1.0.0",
	}
	visitorInfo := &ctype.VisitorInfo{}

	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-123", gomock.Any()).Return(false, int64(0), nil)
	mockSessionRedis.EXPECT().SetSession(gomock.Any(), "conv-123", gomock.Any(), gomock.Any()).Return(true, nil)
	mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-123").Return(3600, nil)

	startTime, ttl, err := svc.HandleRecoverLifetimeOrCreate(context.Background(), req, visitorInfo, false)

	assert.NoError(t, err)
	assert.NotZero(t, startTime)
	assert.Equal(t, 3600, ttl)
}

func TestHandleRecoverLifetimeOrCreate_RefreshSessionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	req := sessionreq.ManageReq{
		ConversationID: "conv-123",
	}
	visitorInfo := &ctype.VisitorInfo{}

	expectedErr := errors.New("redis connection failed")
	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-123", gomock.Any()).Return(false, int64(0), expectedErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	startTime, ttl, err := svc.HandleRecoverLifetimeOrCreate(context.Background(), req, visitorInfo, false)

	assert.Error(t, err)
	assert.Zero(t, startTime)
	assert.Zero(t, ttl)
}

func TestHandleRecoverLifetimeOrCreate_SetSessionError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	req := sessionreq.ManageReq{
		ConversationID: "conv-123",
	}
	visitorInfo := &ctype.VisitorInfo{}

	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-123", gomock.Any()).Return(false, int64(0), nil)

	expectedErr := errors.New("failed to set session")
	mockSessionRedis.EXPECT().SetSession(gomock.Any(), "conv-123", gomock.Any(), gomock.Any()).Return(false, expectedErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	startTime, ttl, err := svc.HandleRecoverLifetimeOrCreate(context.Background(), req, visitorInfo, false)

	assert.Error(t, err)
	assert.Zero(t, startTime)
	assert.Zero(t, ttl)
}

func TestHandleRecoverLifetimeOrCreate_GetTTLError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	req := sessionreq.ManageReq{
		ConversationID: "conv-123",
	}
	visitorInfo := &ctype.VisitorInfo{}

	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-123", gomock.Any()).Return(true, int64(1234567890), nil)

	expectedErr := errors.New("failed to get ttl")
	mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-123").Return(0, expectedErr)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

	startTime, ttl, err := svc.HandleRecoverLifetimeOrCreate(context.Background(), req, visitorInfo, false)

	assert.Error(t, err)
	assert.Zero(t, startTime)
	assert.Zero(t, ttl)
}

func TestHandleRecoverLifetimeOrCreate_NoCacheTrigger(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

	svc := &sessionSvc{
		logger:          mockLogger,
		sessionRedisAcc: mockSessionRedis,
	}

	req := sessionreq.ManageReq{
		ConversationID: "conv-123",
	}
	visitorInfo := &ctype.VisitorInfo{}

	mockSessionRedis.EXPECT().RefreshSession(gomock.Any(), "conv-123", gomock.Any()).Return(true, int64(1234567890), nil)
	mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-123").Return(3600, nil)
	// No AgentCacheManage call expected when isTriggerCacheUpsert is false

	startTime, ttl, err := svc.HandleRecoverLifetimeOrCreate(context.Background(), req, visitorInfo, false)

	assert.NoError(t, err)
	assert.NotZero(t, startTime)
	assert.Equal(t, 3600, ttl)
}
