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
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestHandleGetInfoOrCreate(t *testing.T) {
	t.Parallel()

	t.Run("nil service causes panic", func(t *testing.T) {
		t.Parallel()

		var svc *sessionSvc

		ctx := context.Background()
		req := sessionreq.ManageReq{
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		// This will panic when trying to use sessionRedisAcc
		assert.Panics(t, func() {
			svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false) //nolint:errcheck
		})
	})

	t.Run("nil session redis causes panic", func(t *testing.T) {
		t.Parallel()

		svc := &sessionSvc{
			// sessionRedisAcc is nil, will panic
		}
		ctx := context.Background()
		req := sessionreq.ManageReq{
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		// This will panic when trying to use sessionRedisAcc
		assert.Panics(t, func() {
			svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false) //nolint:errcheck
		})
	})

	t.Run("session exists returns existing data", func(t *testing.T) {
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
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		existingStartTime := int64(1234567890)
		existingTTL := 3600

		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(true, existingStartTime, existingTTL, nil)

		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false)

		require.NoError(t, err)
		assert.Equal(t, existingStartTime, startTime)
		assert.Equal(t, existingTTL, ttl)
	})

	t.Run("session does not exist creates new session", func(t *testing.T) {
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
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(false, int64(0), 0, nil)
		mockSessionRedis.EXPECT().SetSession(gomock.Any(), "conv-123", gomock.Any(), gomock.Any()).Return(true, nil)
		mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-123").Return(3600, nil)

		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false)

		require.NoError(t, err)
		assert.NotZero(t, startTime)
		assert.Equal(t, 3600, ttl)
	})

	t.Run("get session error returns error", func(t *testing.T) {
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
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		expectedErr := errors.New("redis connection failed")
		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(false, int64(0), 0, expectedErr)
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false)

		assert.Error(t, err)
		assert.Zero(t, startTime)
		assert.Zero(t, ttl)
	})

	t.Run("set session error returns error", func(t *testing.T) {
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
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		expectedErr := errors.New("failed to set session")

		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(false, int64(0), 0, nil)
		mockSessionRedis.EXPECT().SetSession(gomock.Any(), "conv-123", gomock.Any(), gomock.Any()).Return(false, expectedErr)
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false)

		assert.Error(t, err)
		assert.Zero(t, startTime)
		assert.Zero(t, ttl)
	})

	t.Run("get TTL error after create returns error", func(t *testing.T) {
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
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		expectedErr := errors.New("failed to get TTL")

		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(false, int64(0), 0, nil)
		mockSessionRedis.EXPECT().SetSession(gomock.Any(), "conv-123", gomock.Any(), gomock.Any()).Return(true, nil)
		mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-123").Return(0, expectedErr)
		mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).Times(1)

		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, false)

		assert.Error(t, err)
		assert.Zero(t, startTime)
		assert.Zero(t, ttl)
	})

	t.Run("with cache trigger enabled", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := cmpmock.NewMockLogger(ctrl)
		mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

		svc := &sessionSvc{
			logger:          mockLogger,
			sessionRedisAcc: mockSessionRedis,
			// agentExecutorV1 would be needed for full testing
		}

		ctx := context.Background()
		req := sessionreq.ManageReq{
			ConversationID: "conv-123",
		}
		visitorInfo := &ctype.VisitorInfo{}

		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-123").Return(true, int64(1234567890), 3600, nil)

		// isTriggerCacheUpsert = true, but agentExecutorV1 is nil
		// The GoSafe function will handle the panic from nil agentExecutorV1
		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, true)

		// Despite the nil agentExecutorV1, the main function should succeed
		// because the cache upsert is in a separate goroutine
		require.NoError(t, err)
		assert.Equal(t, int64(1234567890), startTime)
		assert.Equal(t, 3600, ttl)
	})

	t.Run("with cache trigger enabled and new session", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockLogger := cmpmock.NewMockLogger(ctrl)
		mockSessionRedis := isessionredismock.NewMockISessionRedisAcc(ctrl)

		svc := &sessionSvc{
			logger:          mockLogger,
			sessionRedisAcc: mockSessionRedis,
			// agentExecutorV1 is nil, GoSafe will handle it
		}

		ctx := context.Background()
		req := sessionreq.ManageReq{
			ConversationID: "conv-new",
		}
		visitorInfo := &ctype.VisitorInfo{}

		// Session does not exist, create new one
		mockSessionRedis.EXPECT().GetSessionWithTTL(gomock.Any(), "conv-new").Return(false, int64(0), 0, nil)
		mockSessionRedis.EXPECT().SetSession(gomock.Any(), "conv-new", gomock.Any(), gomock.Any()).Return(true, nil)
		mockSessionRedis.EXPECT().GetSessionTTL(gomock.Any(), "conv-new").Return(3600, nil)

		// Expect Errorln to be called because triggerAgentCacheUpsert will panic with nil agentExecutorV1
		// The panic is caught by GoSafe and logged
		mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

		// isTriggerCacheUpsert = true, agentExecutorV1 is nil
		// The GoSafe function will handle the panic from nil agentExecutorV1
		startTime, ttl, err := svc.HandleGetInfoOrCreate(ctx, req, visitorInfo, true)

		// Main function should succeed despite nil agentExecutorV1
		require.NoError(t, err)
		assert.NotZero(t, startTime)
		assert.Equal(t, 3600, ttl)
	})
}
