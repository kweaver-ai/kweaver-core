package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_Copy2Tpl_PanicsWithoutAgentConfRepo(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	agentID := "agent-123"
	req := &agentconfigreq.Copy2TplReq{}

	assert.Panics(t, func() {
		_, _, _ = svc.Copy2Tpl(ctx, agentID, req, nil)
	})
}

// ---- removeDataSourceFromConfig ----

func TestDataAgentConfigSvc_RemoveDataSourceFromConfig_EmptyConfig(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	po := &dapo.DataAgentTplPo{}
	po.Config = `{}`

	err := svc.removeDataSourceFromConfig(po)
	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_RemoveDataSourceFromConfig_WithDataSource(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	po := &dapo.DataAgentTplPo{}
	po.Config = `{"dataSource":{"doc":{"docDataSources":[{"id":"ds-1"}]}}}`

	err := svc.removeDataSourceFromConfig(po)
	assert.NoError(t, err)
}

// ---- getNewNameForAgentCopy2Tpl ----

func TestDataAgentConfigSvc_GetNewNameForAgentCopy2Tpl_WithProvidedName_NotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.Copy2TplReq{Name: "我的模板"}
	sourcePo := &dapo.DataAgentPo{Name: "原始Agent"}

	mockTplRepo.EXPECT().ExistsByName(gomock.Any(), "我的模板").Return(false, nil)

	// 函数逻辑：req.Name 只做存在性校验，最终 newName 总是由 sourcePo.Name+"_模板" 生成
	name, err := svc.getNewNameForAgentCopy2Tpl(ctx, req, sourcePo)
	assert.NoError(t, err)
	assert.Contains(t, name, "_模板")
}

func TestDataAgentConfigSvc_GetNewNameForAgentCopy2Tpl_WithProvidedName_AlreadyExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.Copy2TplReq{Name: "已存在的模板"}
	sourcePo := &dapo.DataAgentPo{Name: "原始Agent"}

	mockTplRepo.EXPECT().ExistsByName(gomock.Any(), "已存在的模板").Return(true, nil)

	_, err := svc.getNewNameForAgentCopy2Tpl(ctx, req, sourcePo)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_GetNewNameForAgentCopy2Tpl_WithProvidedName_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.Copy2TplReq{Name: "某模板"}
	sourcePo := &dapo.DataAgentPo{Name: "原始Agent"}

	mockTplRepo.EXPECT().ExistsByName(gomock.Any(), "某模板").Return(false, errors.New("db error"))

	_, err := svc.getNewNameForAgentCopy2Tpl(ctx, req, sourcePo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "检查模板名称是否存在失败")
}

func TestDataAgentConfigSvc_GetNewNameForAgentCopy2Tpl_EmptyName_UsesSourceName(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	ctx := context.Background()
	req := &agentconfigreq.Copy2TplReq{Name: ""}
	sourcePo := &dapo.DataAgentPo{Name: "原始Agent"}

	name, err := svc.getNewNameForAgentCopy2Tpl(ctx, req, sourcePo)
	assert.NoError(t, err)
	assert.Contains(t, name, "_模板")
}

func TestDataAgentConfigSvc_GetNewNameForAgentCopy2Tpl_LongSourceName_Truncated(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	ctx := context.Background()
	req := &agentconfigreq.Copy2TplReq{Name: ""}
	longName := strings.Repeat("A", cconstant.NameMaxLength+10)
	sourcePo := &dapo.DataAgentPo{Name: longName}

	name, err := svc.getNewNameForAgentCopy2Tpl(ctx, req, sourcePo)
	assert.NoError(t, err)
	assert.LessOrEqual(t, len([]rune(name)), cconstant.NameMaxLength)
	assert.Contains(t, name, "_模板")
}

// ---- Copy2Tpl 函数分支 ----

func TestDataAgentConfigSvc_Copy2Tpl_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(nil, errors.New("db err"))

	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, nil)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Copy2Tpl_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	// uid="" != createdBy="other-user", IsBuiltIn=nil → not built-in → direct 403
	builtInNo := cdaenum.BuiltInNo
	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: "other-user", IsBuiltIn: &builtInNo}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)

	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不是owner")
}

func TestDataAgentConfigSvc_Copy2Tpl_NameConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		agentTplRepo:  mockTplRepo,
		logger:        mockLogger,
	}

	// CreatedBy="" matches ctx uid="" → owner check passes
	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: ""}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().ExistsByName(gomock.Any(), "conflicting-name").Return(true, nil)

	req := &agentconfigreq.Copy2TplReq{Name: "conflicting-name"}
	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", req, nil)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Copy2Tpl_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		agentTplRepo:  mockTplRepo,
		logger:        mockLogger,
	}

	// CreatedBy="" matches ctx uid="" → owner check passes
	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: ""}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "开启事务失败")
}

func TestDataAgentConfigSvc_Copy2Tpl_WithExistingTx_CreateTplError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		agentTplRepo:  mockTplRepo,
		logger:        mockLogger,
	}

	// CreatedBy="" matches ctx uid="" → owner check passes
	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: "", Config: "{}"}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("create err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	existingTx := &sql.Tx{}
	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, existingTx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "创建模板失败")
}

func TestDataAgentConfigSvc_Copy2Tpl_WithExistingTx_BdRelBatchCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBdTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:           service.NewSvcBase(),
		agentConfRepo:     mockAgentConfRepo,
		agentTplRepo:      mockTplRepo,
		bdAgentTplRelRepo: mockBdTplRelRepo,
		logger:            mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: "", Config: "{}"}
	tplPo := &dapo.DataAgentTplPo{ID: 101}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(tplPo, nil)
	mockBdTplRelRepo.EXPECT().BatchCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("batch create err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	existingTx := &sql.Tx{}
	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, existingTx)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Copy2Tpl_WithExistingTx_BizDomainHttpError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBdTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:           service.NewSvcBase(),
		agentConfRepo:     mockAgentConfRepo,
		agentTplRepo:      mockTplRepo,
		bdAgentTplRelRepo: mockBdTplRelRepo,
		bizDomainHttp:     mockBizDomainHttp,
		logger:            mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: "", Config: "{}"}
	tplPo := &dapo.DataAgentTplPo{ID: 101}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(tplPo, nil)
	mockBdTplRelRepo.EXPECT().BatchCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(errors.New("http err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	existingTx := &sql.Tx{}
	_, _, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, existingTx)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_CreateTemplateFromAgent_RemoveDataSourceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:      service.NewSvcBase(),
		agentTplRepo: mockTplRepo,
	}

	// invalid JSON in Config → RemoveDataSourceFromConfig fails → removeDataSourceFromConfig error
	sourcePo := &dapo.DataAgentPo{
		ID:     "src",
		Config: "{invalid-json",
	}

	_, err := svc.createTemplateFromAgent(context.Background(), sourcePo, "tpl", (*sql.Tx)(nil))
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Copy2Tpl_WithExistingTx_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockBdTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:           service.NewSvcBase(),
		agentConfRepo:     mockAgentConfRepo,
		agentTplRepo:      mockTplRepo,
		bdAgentTplRelRepo: mockBdTplRelRepo,
		bizDomainHttp:     mockBizDomainHttp,
		logger:            mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: "", Config: "{}"}
	tplPo := &dapo.DataAgentTplPo{ID: 101}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(tplPo, nil)
	mockBdTplRelRepo.EXPECT().BatchCreate(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(nil)

	existingTx := &sql.Tx{}
	res, auditInfo, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, existingTx)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "agent-1", auditInfo.ID)
}
