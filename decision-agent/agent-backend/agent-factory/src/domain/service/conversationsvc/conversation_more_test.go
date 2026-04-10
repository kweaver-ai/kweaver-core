package conversationsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newConversationTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())

		_ = db.Close()
	}

	return tx, mock, cleanup
}

func TestConversationSvc_waitForSessionReady(t *testing.T) {
	t.Run("running session", func(t *testing.T) {
		svc := &conversationSvc{
			logger: noopConversationLogger{},
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
			sandboxPlatform: &fakeSandboxPlatform{
				getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
					return &sandboxplatformdto.GetSessionResp{Status: "running"}, nil
				},
			},
		}

		sid, err := svc.waitForSessionReady(context.Background(), "sess-1")
		assert.NoError(t, err)
		assert.Equal(t, "sess-1", sid)
	})

	t.Run("invalid session status", func(t *testing.T) {
		svc := &conversationSvc{
			logger: noopConversationLogger{},
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
			sandboxPlatform: &fakeSandboxPlatform{
				getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
					return &sandboxplatformdto.GetSessionResp{Status: "stopped"}, nil
				},
			},
		}

		_, err := svc.waitForSessionReady(context.Background(), "sess-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid state")
	})

	t.Run("timeout after retries", func(t *testing.T) {
		svc := &conversationSvc{
			logger: noopConversationLogger{},
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
			sandboxPlatform: &fakeSandboxPlatform{
				getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
					return &sandboxplatformdto.GetSessionResp{Status: "starting"}, nil
				},
			},
		}

		_, err := svc.waitForSessionReady(context.Background(), "sess-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for session ready")
	})

	t.Run("get session error retries then timeout", func(t *testing.T) {
		svc := &conversationSvc{
			logger: noopConversationLogger{},
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
			sandboxPlatform: &fakeSandboxPlatform{
				getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
					return nil, errors.New("network failed")
				},
			},
		}

		_, err := svc.waitForSessionReady(context.Background(), "sess-1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "timeout waiting for session ready")
	})
}

func TestConversationSvc_ListAndDelete_MoreBranches(t *testing.T) {
	t.Run("list success with latest message status mapping", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		req := conversationreq.ListReq{AgentAPPKey: "app1"}
		po := &dapo.ConversationPO{ID: "c1", AgentAPPKey: "app1", Title: "t1"}
		mockConvRepo.EXPECT().List(gomock.Any(), req).Return([]*dapo.ConversationPO{po}, int64(1), nil)
		mockMsgRepo.EXPECT().GetLatestMsgByConversationID(gomock.Any(), "c1").Return(&dapo.ConversationMsgPO{
			Status: cdaenum.MsgStatusProcessing,
		}, nil)

		resp, count, err := svc.List(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Len(t, resp, 1)
		assert.Equal(t, cdaenum.ConvStatusProcessing, resp[0].Status)
	})

	t.Run("list latest message sql not found mapped to completed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		req := conversationreq.ListReq{AgentAPPKey: "app1"}
		po := &dapo.ConversationPO{ID: "c1", AgentAPPKey: "app1", Title: "t1"}
		mockConvRepo.EXPECT().List(gomock.Any(), req).Return([]*dapo.ConversationPO{po}, int64(1), nil)
		mockMsgRepo.EXPECT().GetLatestMsgByConversationID(gomock.Any(), "c1").Return(nil, sql.ErrNoRows)

		resp, _, err := svc.List(context.Background(), req)
		assert.NoError(t, err)
		assert.Equal(t, cdaenum.ConvStatusCompleted, resp[0].Status)
	})

	t.Run("list latest message query error returns error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		req := conversationreq.ListReq{AgentAPPKey: "app1"}
		po := &dapo.ConversationPO{ID: "c1", AgentAPPKey: "app1", Title: "t1"}
		mockConvRepo.EXPECT().List(gomock.Any(), req).Return([]*dapo.ConversationPO{po}, int64(1), nil)
		mockMsgRepo.EXPECT().GetLatestMsgByConversationID(gomock.Any(), "c1").Return(nil, errors.New("db failed"))

		resp, _, err := svc.List(context.Background(), req)
		assert.Error(t, err)
		assert.Empty(t, resp)
	})

	t.Run("list latest message status mappings", func(t *testing.T) {
		tests := []struct {
			msgStatus cdaenum.ConversationMsgStatus
			want      cdaenum.ConversationStatus
		}{
			{msgStatus: cdaenum.MsgStatusSucceded, want: cdaenum.ConvStatusCompleted},
			{msgStatus: cdaenum.MsgStatusCancelled, want: cdaenum.ConvStatusCancelled},
			{msgStatus: cdaenum.MsgStatusFailed, want: cdaenum.ConvStatusFailed},
		}

		for _, tt := range tests {
			ctrl := gomock.NewController(t)
			mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
			mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
			svc := &conversationSvc{
				SvcBase:             service.NewSvcBase(),
				conversationRepo:    mockConvRepo,
				conversationMsgRepo: mockMsgRepo,
			}
			req := conversationreq.ListReq{AgentAPPKey: "app1"}
			po := &dapo.ConversationPO{ID: "c1", AgentAPPKey: "app1", Title: "t1"}
			mockConvRepo.EXPECT().List(gomock.Any(), req).Return([]*dapo.ConversationPO{po}, int64(1), nil)
			mockMsgRepo.EXPECT().GetLatestMsgByConversationID(gomock.Any(), "c1").Return(&dapo.ConversationMsgPO{
				Status: tt.msgStatus,
			}, nil)

			resp, _, err := svc.List(context.Background(), req)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, resp[0].Status)
			ctrl.Finish()
		}
	})

	t.Run("list by agent id success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		po := &dapo.ConversationPO{ID: "c1", AgentAPPKey: "app1", Title: "t1"}
		mockConvRepo.EXPECT().ListByAgentID(gomock.Any(), "a1", "", 1, 10).Return([]*dapo.ConversationPO{po}, int64(1), nil)

		resp, count, err := svc.ListByAgentID(context.Background(), "a1", "", 1, 10, 0, 0)
		assert.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Len(t, resp, 1)
	})

	t.Run("delete by app key success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newConversationTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
			logger:              noopConversationLogger{},
		}

		mockConvRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockConvRepo.EXPECT().DeleteByAPPKey(gomock.Any(), tx, "app1").Return(nil)
		mockMsgRepo.EXPECT().DeleteByAPPKey(gomock.Any(), tx, "app1").Return(nil)

		err := svc.DeleteByAppKey(context.Background(), "app1")
		assert.NoError(t, err)
	})

	t.Run("delete by app key wrap delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newConversationTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
			logger:              noopConversationLogger{},
		}

		mockConvRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockConvRepo.EXPECT().DeleteByAPPKey(gomock.Any(), tx, "app1").Return(errors.New("conv delete failed"))

		err := svc.DeleteByAppKey(context.Background(), "app1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "删除对话数据失败")
	})

	t.Run("delete success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newConversationTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
			logger:              noopConversationLogger{},
		}

		mockConvRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1"}, nil)
		mockConvRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockConvRepo.EXPECT().Delete(gomock.Any(), tx, "c1").Return(nil)
		mockMsgRepo.EXPECT().DeleteByConversationID(gomock.Any(), tx, "c1").Return(nil)

		err := svc.Delete(context.Background(), "c1")
		assert.NoError(t, err)
	})

	t.Run("delete wraps conversation msg delete error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newConversationTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockConvRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockConvRepo,
			conversationMsgRepo: mockMsgRepo,
			logger:              noopConversationLogger{},
		}

		mockConvRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1"}, nil)
		mockConvRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockConvRepo.EXPECT().Delete(gomock.Any(), tx, "c1").Return(nil)
		mockMsgRepo.EXPECT().DeleteByConversationID(gomock.Any(), tx, "c1").Return(errors.New("msg delete failed"))

		err := svc.Delete(context.Background(), "c1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "删除对话消息数据失败")
	})
}
