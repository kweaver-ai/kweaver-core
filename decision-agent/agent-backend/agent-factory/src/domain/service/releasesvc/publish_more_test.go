package releasesvc

import (
	"context"
	"database/sql"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/pmsvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/publishvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/release/releasereq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iauthzacc/authzaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReleaseSvc_Publish_PermissionDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentRepo,
		pmsSvc:          mockPms,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{ID: "a1", Name: "agent1"}, nil)
	mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	req := &releasereq.PublishReq{
		AgentID:              "a1",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        false,
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "permission")
}

func TestReleaseSvc_Publish_PermissionCheckError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentRepo,
		pmsSvc:          mockPms,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{ID: "a1", Name: "agent1"}, nil)
	mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, assert.AnError)

	req := &releasereq.PublishReq{
		AgentID:              "a1",
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
		IsInternalAPI:        false,
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check publish permission failed")
}

func TestReleaseSvc_Publish_GetReleaseByAgentIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		agentConfigRepo:    mockAgentRepo,
		releaseHistoryRepo: mockHistory,
		releaseRepo:        mockReleaseRepo,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
	}, nil)
	mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(nil, assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get release by agent id failed")
}

func TestReleaseSvc_Publish_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            service.NewSvcBase(),
		agentConfigRepo:    mockAgentRepo,
		releaseHistoryRepo: mockHistory,
		releaseRepo:        mockReleaseRepo,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
	}, nil)
	mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(nil, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "begin transaction failed")
}

func TestReleaseSvc_Publish_UpdateReleaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:    mockAgentRepo,
		releaseHistoryRepo: mockHistory,
		releaseRepo:        mockReleaseRepo,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
	}, nil)
	mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(&dapo.ReleaseHistoryPO{AgentVersion: "v1"}, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{ID: "r1", AgentID: "a1"}, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update release failed")
}

func TestReleaseSvc_Publish_DelCategoryRelError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:                &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:        mockAgentRepo,
		releaseHistoryRepo:     mockHistory,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRel,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
	}, nil)
	mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(&dapo.ReleaseHistoryPO{AgentVersion: "v1"}, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(&dapo.ReleasePO{ID: "r1", AgentID: "a1"}, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockCategoryRel.EXPECT().DelByReleaseID(gomock.Any(), tx, "r1").Return(assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "delete release category rel by release id failed")
}

func TestReleaseSvc_Publish_FillPublishedByNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockCategoryRel := idbaccessmock.NewMockIReleaseCategoryRelRepo(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase:                &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:        mockAgentRepo,
		releaseHistoryRepo:     mockHistory,
		releaseRepo:            mockReleaseRepo,
		releaseCategoryRelRepo: mockCategoryRel,
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
	mockUm.EXPECT().GetSingleUserName(gomock.Any(), gomock.Any()).Return("", assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "fill published by name failed")
}

func TestReleaseSvc_Publish_CreateReleaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:            &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:    mockAgentRepo,
		releaseHistoryRepo: mockHistory,
		releaseRepo:        mockReleaseRepo,
	}

	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID: "a1", Name: "agent1", Config: validPublishConfigJSON,
	}, nil)
	mockHistory.EXPECT().GetLatestVersionByAgentID(gomock.Any(), "a1").Return(nil, nil)
	mockReleaseRepo.EXPECT().GetByAgentID(gomock.Any(), "a1").Return(nil, nil)
	mockReleaseRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockReleaseRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("", assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create release failed")
}

func TestReleaseSvc_Publish_CreateHistoryError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase:               &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:       mockAgentRepo,
		releaseHistoryRepo:    mockHistory,
		releaseRepo:           mockReleaseRepo,
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
	mockUm.EXPECT().GetSingleUserName(gomock.Any(), gomock.Any()).Return("U1", nil)
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), gomock.Any()).Return(map[string]string{"a1": "agent1"}, nil)
	mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)
	mockHistory.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("", assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	req.PublishToWhere = []daenum.PublishToWhere{daenum.PublishToWhereSquare}
	req.PublishToBes = []cdaenum.PublishToBe{cdaenum.PublishToBeAPIAgent}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "create release history failed")
}

func TestReleaseSvc_Publish_UpdateStatusError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectRollback()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase:               &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:       mockAgentRepo,
		releaseHistoryRepo:    mockHistory,
		releaseRepo:           mockReleaseRepo,
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
	mockUm.EXPECT().GetSingleUserName(gomock.Any(), gomock.Any()).Return("U1", nil)
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), gomock.Any()).Return(map[string]string{"a1": "agent1"}, nil)
	mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)
	mockHistory.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("h1", nil)
	mockAgentRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, "a1", gomock.Any()).Return(assert.AnError)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	req.PublishToWhere = []daenum.PublishToWhere{daenum.PublishToWhereSquare}
	req.PublishToBes = []cdaenum.PublishToBe{cdaenum.PublishToBeAPIAgent}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "update agent status to published failed")
}

func TestReleaseSvc_Publish_WithPmsControl(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tx, sqlMock, done := newReleaseTx(t)
	defer done()
	sqlMock.ExpectCommit()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockHistory := idbaccessmock.NewMockIReleaseHistoryRepo(ctrl)
	mockReleaseRepo := idbaccessmock.NewMockIReleaseRepo(ctrl)
	mockPermRepo := idbaccessmock.NewMockIReleasePermissionRepo(ctrl)
	mockAuthz := authzaccmock.NewMockAuthZHttpAcc(ctrl)
	mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

	svc := &releaseSvc{
		SvcBase:               &service.SvcBase{Logger: noopReleaseLogger{}},
		agentConfigRepo:       mockAgentRepo,
		releaseHistoryRepo:    mockHistory,
		releaseRepo:           mockReleaseRepo,
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
	mockUm.EXPECT().GetSingleUserName(gomock.Any(), gomock.Any()).Return("U1", nil)
	// handlePmsCtrl: DelByReleaseID, getAgentName, removeUsePmsByHTTPAcc, handlePmsCtrlRange
	mockPermRepo.EXPECT().DelByReleaseID(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockAgentRepo.EXPECT().GetIDNameMapByID(gomock.Any(), gomock.Any()).Return(map[string]string{"a1": "agent1"}, nil)
	mockAuthz.EXPECT().DeleteAgentPolicy(gomock.Any(), "a1").Return(nil)
	mockAuthz.EXPECT().GrantAgentUsePmsForAccessors(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	// handlePmsCtrlRange: BatchCreate x5 (role, user, usergroup, dept, appaccount)
	mockPermRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil).Times(5)
	mockHistory.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return("h1", nil)
	mockAgentRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, "a1", gomock.Any()).Return(nil)

	req := &releasereq.PublishReq{
		AgentID: "a1", UserID: "u1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{
			PublishInfo: publishvo.PublishInfo{
				PmsControl: &pmsvo.PmsControlObjS{
					RoleIDs:       []string{"r1"},
					UserIDs:       []string{"u1"},
					UserGroupIDs:  []string{"ug1"},
					DepartmentIDs: []string{"d1"},
					AppAccountIDs: []string{"aa1"},
				},
			},
		},
	}
	resp, _, err := svc.Publish(context.Background(), req)
	require.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestReleaseSvc_Publish_AgentNotFound404(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &releaseSvc{
		SvcBase:         service.NewSvcBase(),
		agentConfigRepo: mockAgentRepo,
	}

	// sql.ErrNoRows triggers IsSqlNotFound → 404
	mockAgentRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(nil, sql.ErrNoRows)

	req := &releasereq.PublishReq{
		AgentID: "a1", IsInternalAPI: true,
		UpdatePublishInfoReq: &releasereq.UpdatePublishInfoReq{},
	}
	_, _, err := svc.Publish(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent not found")
}
