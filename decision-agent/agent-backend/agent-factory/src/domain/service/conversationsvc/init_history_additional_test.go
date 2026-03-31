package conversationsvc

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/conf"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/sandboxplatformhttp/sandboxplatformdto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
)

type noopConversationLogger struct{}

func (noopConversationLogger) Infof(string, ...interface{})  {}
func (noopConversationLogger) Infoln(...interface{})         {}
func (noopConversationLogger) Debugf(string, ...interface{}) {}
func (noopConversationLogger) Debugln(...interface{})        {}
func (noopConversationLogger) Errorf(string, ...interface{}) {}
func (noopConversationLogger) Errorln(...interface{})        {}
func (noopConversationLogger) Warnf(string, ...interface{})  {}
func (noopConversationLogger) Warnln(...interface{})         {}
func (noopConversationLogger) Panicf(string, ...interface{}) {}
func (noopConversationLogger) Panicln(...interface{})        {}
func (noopConversationLogger) Fatalf(string, ...interface{}) {}
func (noopConversationLogger) Fatalln(...interface{})        {}

type fakeSandboxPlatform struct {
	getFn    func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error)
	createFn func(context.Context, sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error)
}

func (f *fakeSandboxPlatform) CreateSession(ctx context.Context, req sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
	return f.createFn(ctx, req)
}

func (f *fakeSandboxPlatform) GetSession(ctx context.Context, sessionID string) (*sandboxplatformdto.GetSessionResp, error) {
	return f.getFn(ctx, sessionID)
}

func (f *fakeSandboxPlatform) DeleteSession(context.Context, string) error { return nil }
func (f *fakeSandboxPlatform) ListFiles(context.Context, string, int) ([]string, error) {
	return nil, nil
}

func TestConversationSvc_Init_MoreBranches(t *testing.T) {
	t.Parallel()

	t.Run("create conversation failed", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:          service.NewSvcBase(),
			logger:           noopConversationLogger{},
			conversationRepo: mockRepo,
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				Enable: false,
			},
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, errors.New("db failed"))

		_, err := svc.Init(context.Background(), conversationreq.InitReq{UserID: "u1"})
		assert.Error(t, err)
	})

	t.Run("sandbox disabled", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:          service.NewSvcBase(),
			logger:           noopConversationLogger{},
			conversationRepo: mockRepo,
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				Enable: false,
			},
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "c1"}, nil)

		resp, err := svc.Init(context.Background(), conversationreq.InitReq{UserID: "u1"})
		assert.NoError(t, err)
		assert.Equal(t, "c1", resp.ID)
		assert.Empty(t, resp.SandboxSessionID)
	})

	t.Run("sandbox enabled and existing session running", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		fakeSandbox := &fakeSandboxPlatform{
			getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
				return &sandboxplatformdto.GetSessionResp{Status: "running"}, nil
			},
			createFn: func(context.Context, sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
				return nil, errors.New("should not create")
			},
		}
		svc := &conversationSvc{
			SvcBase:          service.NewSvcBase(),
			logger:           noopConversationLogger{},
			conversationRepo: mockRepo,
			sandboxPlatform:  fakeSandbox,
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				Enable:        true,
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "c1"}, nil)

		resp, err := svc.Init(context.Background(), conversationreq.InitReq{UserID: "u1"})
		assert.NoError(t, err)
		assert.Equal(t, cutil.GetSandboxSessionID(), resp.SandboxSessionID)
	})

	t.Run("sandbox create failed but init still succeeds", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		notFoundErr := rest.NewHTTPError(context.Background(), http.StatusNotFound, rest.PublicError_NotFound)
		fakeSandbox := &fakeSandboxPlatform{
			getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
				return nil, notFoundErr
			},
			createFn: func(context.Context, sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
				return nil, errors.New("create failed")
			},
		}
		svc := &conversationSvc{
			SvcBase:          service.NewSvcBase(),
			logger:           noopConversationLogger{},
			conversationRepo: mockRepo,
			sandboxPlatform:  fakeSandbox,
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				Enable:        true,
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "c1"}, nil)

		resp, err := svc.Init(context.Background(), conversationreq.InitReq{UserID: "u1"})
		assert.NoError(t, err)
		assert.Equal(t, "c1", resp.ID)
		assert.Empty(t, resp.SandboxSessionID)
	})

	t.Run("sandbox get non-404 and recreate success", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		call := 0
		fakeSandbox := &fakeSandboxPlatform{
			getFn: func(context.Context, string) (*sandboxplatformdto.GetSessionResp, error) {
				call++
				if call == 1 {
					return nil, errors.New("temporary error")
				}
				return &sandboxplatformdto.GetSessionResp{Status: "running"}, nil
			},
			createFn: func(context.Context, sandboxplatformdto.CreateSessionReq) (*sandboxplatformdto.CreateSessionResp, error) {
				return &sandboxplatformdto.CreateSessionResp{ID: "sess-u1"}, nil
			},
		}
		svc := &conversationSvc{
			SvcBase:          service.NewSvcBase(),
			logger:           noopConversationLogger{},
			conversationRepo: mockRepo,
			sandboxPlatform:  fakeSandbox,
			sandboxPlatformConf: &conf.SandboxPlatformConf{
				Enable:        true,
				MaxRetries:    1,
				RetryInterval: "1ms",
			},
		}

		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(&dapo.ConversationPO{ID: "c1"}, nil)

		resp, err := svc.Init(context.Background(), conversationreq.InitReq{UserID: "u1"})
		assert.NoError(t, err)
		assert.Equal(t, "sess-u1", resp.SandboxSessionID)
	})
}

func TestConversationSvc_GetHistory_MoreBranches(t *testing.T) {
	t.Parallel()

	t.Run("detail error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(nil, errors.New("not found"))

		history, err := svc.GetHistory(context.Background(), "c1", 10, "", "")
		assert.Error(t, err)
		assert.Nil(t, history)
	})

	t.Run("assistant content unmarshal error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		bad := "{"

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1", CreateBy: "u1"}, nil)
		mockMsgRepo.EXPECT().GetRecentMessages(gomock.Any(), "c1", 10).Return([]*dapo.ConversationMsgPO{
			{ID: "m1", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &bad},
		}, nil)

		history, err := svc.GetHistory(context.Background(), "c1", 10, "", "")
		assert.Error(t, err)
		assert.Nil(t, history)
	})

	t.Run("user content unmarshal error", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		bad := "{"

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1", CreateBy: "u1"}, nil)
		mockMsgRepo.EXPECT().GetRecentMessages(gomock.Any(), "c1", 10).Return([]*dapo.ConversationMsgPO{
			{ID: "m1", ConversationID: "c1", Role: cdaenum.MsgRoleUser, Content: &bad},
		}, nil)

		history, err := svc.GetHistory(context.Background(), "c1", 10, "", "")
		assert.Error(t, err)
		assert.Nil(t, history)
	})

	t.Run("success includes workspace context and respects limit", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		user1 := `{"text":"hello","selected_files":[{"file_name":"a.txt"}]}`
		assistant1 := `{"final_answer":{"answer":{"text":"hi"}}}`
		user2 := `{"text":"next"}`
		assistant2 := `{"final_answer":{"skill_process":[{"text":"skill answer"}]}}`

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1", CreateBy: "u1"}, nil)
		mockMsgRepo.EXPECT().GetRecentMessages(gomock.Any(), "c1", 2).Return([]*dapo.ConversationMsgPO{
			{ID: "m1", ConversationID: "c1", Role: cdaenum.MsgRoleUser, Content: &user1},
			{ID: "m2", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &assistant1},
			{ID: "m3", ConversationID: "c1", Role: cdaenum.MsgRoleUser, Content: &user2},
			{ID: "m4", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &assistant2},
		}, nil)

		history, err := svc.GetHistory(context.Background(), "c1", 2, "", "")
		assert.NoError(t, err)
		assert.Len(t, history, 2)
		assert.Equal(t, "user", history[0].Role)
		assert.Equal(t, "next", history[0].Content)
		assert.Equal(t, "assistant", history[1].Role)
		assert.Equal(t, "skill answer", history[1].Content)

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1", CreateBy: "u1"}, nil)
		mockMsgRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return([]*dapo.ConversationMsgPO{
			{ID: "m1", ConversationID: "c1", Role: cdaenum.MsgRoleUser, Content: &user1},
			{ID: "m2", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &assistant1},
			{ID: "m3", ConversationID: "c1", Role: cdaenum.MsgRoleUser, Content: &user2},
			{ID: "m4", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &assistant2},
		}, nil)

		full, err := svc.GetHistory(context.Background(), "c1", -1, "", "")
		assert.NoError(t, err)
		assert.True(t, strings.Contains(full[0].Content, "a.txt"))
	})

	t.Run("assistant answer_type_other branches", func(t *testing.T) {
		t.Parallel()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockRepo := idbaccessmock.NewMockIConversationRepo(ctrl)
		mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)
		svc := &conversationSvc{
			SvcBase:             service.NewSvcBase(),
			conversationRepo:    mockRepo,
			conversationMsgRepo: mockMsgRepo,
		}

		assistantOtherStr := `{"final_answer":{"answer_type_other":"plain-other"}}`
		assistantOtherObj := `{"final_answer":{"answer_type_other":{"k":"v"}}}`

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1", CreateBy: "u1"}, nil)
		mockMsgRepo.EXPECT().GetRecentMessages(gomock.Any(), "c1", 10).Return([]*dapo.ConversationMsgPO{
			{ID: "m1", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &assistantOtherStr},
		}, nil)

		h1, err := svc.GetHistory(context.Background(), "c1", 10, "", "")
		assert.NoError(t, err)
		assert.Equal(t, "plain-other", h1[0].Content)

		mockRepo.EXPECT().GetByID(gomock.Any(), "c1").Return(&dapo.ConversationPO{ID: "c1", CreateBy: "u1"}, nil)
		mockMsgRepo.EXPECT().GetRecentMessages(gomock.Any(), "c1", 10).Return([]*dapo.ConversationMsgPO{
			{ID: "m2", ConversationID: "c1", Role: cdaenum.MsgRoleAssistant, Content: &assistantOtherObj},
		}, nil)

		h2, err := svc.GetHistory(context.Background(), "c1", 10, "", "")
		assert.NoError(t, err)
		assert.True(t, strings.Contains(h2[0].Content, "\"k\":\"v\""))
	})
}
