package v3agentconfigsvc

import (
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/imodelfactoryacc/modelfactoryaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
)

// newTestGinCtx creates a gin.Context for testing.
func newTestGinCtx() *gin.Context {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	return c
}

// ---- createTemplateFromAgent ----

func TestDataAgentConfigSvc_CreateTemplateFromAgent_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	ctx := context.Background()
	sourcePo := &dapo.DataAgentPo{
		ID:     "agent-1",
		Name:   "Test Agent",
		Key:    "test-key",
		Config: `{}`,
	}

	mockTplRepo.EXPECT().Create(gomock.Any(), nil, gomock.Any()).Return(errors.New("db error"))

	res, err := svc.createTemplateFromAgent(ctx, sourcePo, "New Template", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "保存模板失败")
	assert.Nil(t, res)
}

func TestDataAgentConfigSvc_CreateTemplateFromAgent_GetByKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	ctx := context.Background()
	sourcePo := &dapo.DataAgentPo{
		ID:     "agent-1",
		Name:   "Test Agent",
		Key:    "test-key",
		Config: `{}`,
	}

	mockTplRepo.EXPECT().Create(gomock.Any(), nil, gomock.Any()).Return(nil)
	mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), nil, gomock.Any()).Return(nil, errors.New("not found"))

	res, err := svc.createTemplateFromAgent(ctx, sourcePo, "New Template", nil)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "获取模板PO失败")
	assert.Nil(t, res)
}

func TestDataAgentConfigSvc_CreateTemplateFromAgent_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	ctx := context.Background()
	sourcePo := &dapo.DataAgentPo{
		ID:     "agent-1",
		Name:   "Test Agent",
		Key:    "test-key",
		Config: `{}`,
	}

	returnedPo := &dapo.DataAgentTplPo{}
	returnedPo.ID = 42

	mockTplRepo.EXPECT().Create(gomock.Any(), nil, gomock.Any()).Return(nil)
	mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), nil, gomock.Any()).Return(returnedPo, nil)

	res, err := svc.createTemplateFromAgent(ctx, sourcePo, "New Template", nil)

	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, int64(42), res.ID)
	assert.Equal(t, "New Template", res.Name)
}

// ---- Copy2TplAndPublish ----

func TestDataAgentConfigSvc_Copy2TplAndPublish_NoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	_, _, err := svc.Copy2TplAndPublish(ctx, "agent-1", &agenttplreq.PublishReq{})

	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Copy2TplAndPublish_PermissionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("pms error"))

	_, _, err := svc.Copy2TplAndPublish(ctx, "agent-1", &agenttplreq.PublishReq{})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check tpl publish permission failed")
}

func TestDataAgentConfigSvc_Copy2TplAndPublish_GetTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		pmsSvc:       mockPmsSvc,
		agentTplRepo: mockAgentTplRepo,
		logger:       mockLogger,
	}

	// has permission
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	// getTx uses agentTplRepo.BeginTx
	mockAgentTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, _, err := svc.Copy2TplAndPublish(context.Background(), "agent-1", &agenttplreq.PublishReq{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "开启事务失败")
}

func TestDataAgentConfigSvc_AIAutogenV3_SystemPrompt_NilAcc(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase:         service.NewSvcBase(),
		modelFactoryAcc: nil,
	}

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From: daenum.AiAutogenFromSystemPrompt,
	}

	assert.Panics(t, func() {
		_, _, _ = svc.AIAutogenV3(ginCtx, req)
	})
}

func TestDataAgentConfigSvc_AIAutogenV3_OpeningRemarks_NilAcc(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase:         service.NewSvcBase(),
		modelFactoryAcc: nil,
	}

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From: daenum.AiAutogenFromOpeningRemarks,
	}

	assert.Panics(t, func() {
		_, _, _ = svc.AIAutogenV3(ginCtx, req)
	})
}

// ---- doAiChat ----

func TestDataAgentConfigSvc_DoAiChat_UnsupportedFrom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From: daenum.AiAutogenFrom("unsupported"),
	}

	_, err := svc.doAiChat(ginCtx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的来源类型")
}

// ---- AIAutogenV3 ----

func TestDataAgentConfigSvc_AIAutogenV3_UnsupportedFrom(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From: daenum.AiAutogenFrom("unsupported_type"),
	}

	_, _, err := svc.AIAutogenV3(ginCtx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不支持的类型")
}

func TestDataAgentConfigSvc_DoAiChat_PreSetQuestion_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModelAcc := modelfactoryaccmock.NewMockIModelApiAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:         service.NewSvcBase(),
		modelFactoryAcc: mockModelAcc,
		logger:          mockLogger,
	}

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From:     daenum.AiAutogenFromPreSetQuestion,
		Language: "zh-CN",
		Params: &agentconfigreq.Params{
			Name:    "TestAgent",
			Profile: "Test",
		},
	}

	choice := openai.ChatCompletionChoice{
		Message: openai.ChatCompletionMessage{Content: `["Q1", "Q2"]`},
	}
	mockModelAcc.EXPECT().ChatCompletion(gomock.Any(), gomock.Any()).
		Return(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{choice}}, nil)

	chatRes, err := svc.doAiChat(ginCtx, req)
	assert.NoError(t, err)
	assert.Len(t, chatRes.Choices, 1)
}

func TestDataAgentConfigSvc_DoAiChat_PreSetQuestion_ModelError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModelAcc := modelfactoryaccmock.NewMockIModelApiAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:         service.NewSvcBase(),
		modelFactoryAcc: mockModelAcc,
		logger:          mockLogger,
	}

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From:     daenum.AiAutogenFromPreSetQuestion,
		Language: "en-US",
		Params: &agentconfigreq.Params{
			Name: "TestAgent",
		},
	}

	mockModelAcc.EXPECT().ChatCompletion(gomock.Any(), gomock.Any()).
		Return(openai.ChatCompletionResponse{}, errors.New("model error"))
	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

	_, err := svc.doAiChat(ginCtx, req)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_AIAutogenNotStream_PreSetQuestion_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockModelAcc := modelfactoryaccmock.NewMockIModelApiAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:         service.NewSvcBase(),
		modelFactoryAcc: mockModelAcc,
		logger:          mockLogger,
	}

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From:     daenum.AiAutogenFromPreSetQuestion,
		Language: "zh-TW",
		Params: &agentconfigreq.Params{
			Name: "TestAgent",
		},
	}

	choice := openai.ChatCompletionChoice{
		Message: openai.ChatCompletionMessage{Content: `["Q1", "Q2", "Q3"]`},
	}
	mockModelAcc.EXPECT().ChatCompletion(gomock.Any(), gomock.Any()).
		Return(openai.ChatCompletionResponse{Choices: []openai.ChatCompletionChoice{choice}}, nil)

	questions, err := svc.AIAutogenNotStream(ginCtx, req)
	assert.NoError(t, err)
	assert.Len(t, questions, 3)
}

func TestDataAgentConfigSvc_CheckPreSetQuestionResFormat_ValidJSON(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	questions, ok := svc.checkPreSetQuestionResFormat(`["Q1", "Q2", "Q3"]`)
	assert.True(t, ok)
	assert.Len(t, questions, 3)
}

func TestDataAgentConfigSvc_CheckPreSetQuestionResFormat_EmptyArray(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	_, ok := svc.checkPreSetQuestionResFormat(`[]`)
	assert.False(t, ok)
}

func TestDataAgentConfigSvc_CheckPreSetQuestionResFormat_InvalidJSON(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	_, ok := svc.checkPreSetQuestionResFormat(`not json`)
	assert.False(t, ok)
}

// ---- AIAutogenNotStream ----

func TestDataAgentConfigSvc_AIAutogenNotStream_DoAiChatError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

	ginCtx := newTestGinCtx()
	req := &agentconfigreq.AiAutogenReq{
		From: daenum.AiAutogenFrom("unsupported"),
	}

	questions, err := svc.AIAutogenNotStream(ginCtx, req)

	assert.Error(t, err)
	assert.Nil(t, questions)
}
