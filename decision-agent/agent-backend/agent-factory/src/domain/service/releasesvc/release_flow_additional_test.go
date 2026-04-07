package releasesvc

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopReleaseLogger struct{}

func (noopReleaseLogger) Infof(string, ...interface{})  {}
func (noopReleaseLogger) Infoln(...interface{})         {}
func (noopReleaseLogger) Debugf(string, ...interface{}) {}
func (noopReleaseLogger) Debugln(...interface{})        {}
func (noopReleaseLogger) Errorf(string, ...interface{}) {}
func (noopReleaseLogger) Errorln(...interface{})        {}
func (noopReleaseLogger) Warnf(string, ...interface{})  {}
func (noopReleaseLogger) Warnln(...interface{})         {}
func (noopReleaseLogger) Panicf(string, ...interface{}) {}
func (noopReleaseLogger) Panicln(...interface{})        {}
func (noopReleaseLogger) Fatalf(string, ...interface{}) {}
func (noopReleaseLogger) Fatalln(...interface{})        {}

const validPublishConfigJSON = `{
  "input":{"fields":[{"name":"q","type":"string"}]},
  "llms":[{"is_default":true,"llm_config":{"name":"m1","model_type":"llm","max_tokens":100}}],
  "output":{"default_format":"markdown"}
}`

func newReleaseTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock, func()) {
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

func createReleaseCtx(userID string) context.Context {
	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), &rest.Visitor{ID: userID}) //nolint:staticcheck
}

func TestReleaseSvc_UpdatePublishInfo_SuccessAndSpaceDeleteError(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newReleaseTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)

		svc := &releaseSvc{
			SvcBase:                service.NewSvcBase(),
			agentConfigRepo:        mockAgentRepo,
			releaseRepo:            mockReleaseRepo,
			pmsSvc:                 mockPms,
			umHttp:                 mockUm,
			releaseCategoryRelRepo: mockCategoryRel,
			releasePermissionRepo:  mockPermRepo,
			authZHttp:              mockAuthz,
		}

		builtInNo := cdaenum.BuiltInNo
		agentPo := &dapo.DataAgentPo{
			ID:        "a1",
			Name:      "agent1",
			CreatedBy: "u1",
			IsBuiltIn: &builtInNo,
		}
		releasePo := &dapo.ReleasePO{
			ID:           "r1",
			AgentID:      "a1",
			AgentVersion: "v2",
			UpdateBy:     "u1",
			UpdateTime:   123,
		}

		req := &releasereq.UpdatePublishInfoReq{}
		req.Description = "desc"
		req.PublishToWhere = []daenum.PublishToWhere{daenum.PublishToWhereSquare}
		req.PublishToBes = []cdaenum.PublishToBe{cdaenum.PublishToBeAPIAgent}
		req.PmsControl = nil

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(agentPo, nil)
		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(releasePo, nil)
		mockReleaseRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockUm.EXPECT().GetSingleUserName(gomock.Any(), "u1").Return("U1", nil)
		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, "r1").Return(nil)
		mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), []string{"a1"}).Return(map[string]string{"a1": "agent1"}, nil)
		mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)

		resp, _, err := svc.UpdatePublishInfo(createReleaseCtx("u1"), "a1", req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "r1", resp.ReleaseId)
	})
}

func TestReleaseSvc_UnPublish_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectCommit()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase:                service.NewSvcBase(),
		agentConfigRepo:        mockAgentRepo,
		releaseRepo:            mockReleaseRepo,
		pmsSvc:                 mockPms,
		releaseCategoryRelRepo: mockCategoryRel,
		releasePermissionRepo:  mockPermRepo,
		authZHttp:              mockAuthz,
	}

	agentPo := &dapo.DataAgentPo{ID: "a1", Name: "agent1", CreatedBy: "u1"}
	releasePo := &dapo.ReleasePO{ID: "r1", AgentID: "a1"}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(agentPo, nil)
	mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(releasePo, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, "a1").Return(nil)
	mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "r1").Return(nil)
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, "r1").Return(nil)
	mockAgentRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusUnpublished, "a1", "").Return(nil)
	mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)

	_, err := svc.UnPublish(createReleaseCtx("u1"), "a1")
	assert.NoError(t, err)
}

func TestReleaseSvc_Publish_MoreBranches(t *testing.T) {
	t.Run("latest version query error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		svc := &releaseSvc{
			SvcBase:            service.NewSvcBase(),
			agentConfigRepo:    mockAgentRepo,
			releaseHistoryRepo: mockHistory,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
			ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
		}, nil)
		mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, assert.AnError)

		req := &releasereq.PublishReq{
			UserID:               "u1",
			AgentID:              "a1",
			UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
			IsInternalAPI:        true,
		}
		_, _, err := svc.Publish(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get latest version from release history failed")
	})

	t.Run("invalid old version format", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		svc := &releaseSvc{
			SvcBase:            service.NewSvcBase(),
			agentConfigRepo:    mockAgentRepo,
			releaseHistoryRepo: mockHistory,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
			ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
		}, nil)
		mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(&dapo.ReleaseHistoryPO{AgentVersion: "bad"}, nil)

		req := &releasereq.PublishReq{
			UserID:               "u1",
			AgentID:              "a1",
			UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
			IsInternalAPI:        true,
		}
		_, _, err := svc.Publish(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "generate agent version failed")
	})

	t.Run("invalid config json", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		svc := &releaseSvc{
			SvcBase:            service.NewSvcBase(),
			agentConfigRepo:    mockAgentRepo,
			releaseHistoryRepo: mockHistory,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
			ID: "a1", Name: "agent1", Config: "{bad-json",
		}, nil)
		mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, nil)

		req := &releasereq.PublishReq{
			UserID:               "u1",
			AgentID:              "a1",
			UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
			IsInternalAPI:        true,
		}
		_, _, err := svc.Publish(context.Background(), req)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unmarshal agent config failed")
	})

	t.Run("success update existing release", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newReleaseTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

		svc := &releaseSvc{
			SvcBase:                &service.SvcBase{Logger: noopReleaseLogger{}},
			agentConfigRepo:        mockAgentRepo,
			releaseRepo:            mockReleaseRepo,
			releaseHistoryRepo:     mockHistory,
			releaseCategoryRelRepo: mockCategoryRel,
			releasePermissionRepo:  mockPermRepo,
			authZHttp:              mockAuthz,
			umHttp:                 mockUm,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
			ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
		}, nil)
		mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(&dapo.ReleaseHistoryPO{AgentVersion: "v1"}, nil)
		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{ID: "r1", AgentID: "a1"}, nil)
		mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockReleaseRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "r1").Return(nil)
		mockUm.EXPECT().GetSingleUserName(gomock.Any(), "u1").Return("U1", nil)
		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, "r1").Return(nil)
		mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), []string{"a1"}).Return(map[string]string{"a1": "agent1"}, nil)
		mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)
		mockHistory.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("h1", nil)
		mockAgentRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, "a1", "u1").Return(nil)

		req := &releasereq.PublishReq{
			UserID: "u1", AgentID: "a1", IsInternalAPI: true,
			UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		}
		req.PublishToWhere = []daenum.PublishToWhere{daenum.PublishToWhereSquare}
		req.PublishToBes = []cdaenum.PublishToBe{cdaenum.PublishToBeAPIAgent}
		resp, _, err := svc.Publish(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "r1", resp.ReleaseId)
		assert.Equal(t, "v2", resp.Version)
	})

	t.Run("success create new release", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newReleaseTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
		mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
		mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
		mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

		svc := &releaseSvc{
			SvcBase:               &service.SvcBase{Logger: noopReleaseLogger{}},
			agentConfigRepo:       mockAgentRepo,
			releaseRepo:           mockReleaseRepo,
			releaseHistoryRepo:    mockHistory,
			releasePermissionRepo: mockPermRepo,
			authZHttp:             mockAuthz,
			umHttp:                mockUm,
		}

		mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
			ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
		}, nil)
		mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, nil)
		mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(nil, nil)
		mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockReleaseRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("r-new", nil)
		mockUm.EXPECT().GetSingleUserName(gomock.Any(), "u1").Return("U1", nil)
		mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), []string{"a1"}).Return(map[string]string{"a1": "agent1"}, nil)
		mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)
		mockHistory.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("h1", nil)
		mockAgentRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, "a1", "u1").Return(nil)

		req := &releasereq.PublishReq{
			UserID: "u1", AgentID: "a1", IsInternalAPI: true,
			UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		}
		req.PublishToWhere = []daenum.PublishToWhere{daenum.PublishToWhereSquare}
		req.PublishToBes = []cdaenum.PublishToBe{cdaenum.PublishToBeAPIAgent}
		resp, _, err := svc.Publish(context.Background(), req)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "v1", resp.Version)
	})
}
