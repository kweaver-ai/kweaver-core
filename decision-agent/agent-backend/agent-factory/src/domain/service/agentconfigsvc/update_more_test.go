package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Update — deeper paths ====================

func TestUpdate_RepoUpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.UpdateReq{
		Name:          "NewName",
		ProductKey:    "p-1",
		Config:        cfg,
		IsInternalAPI: true,
		UpdatedBy:     "sys-user",
	}

	profile := ""
	oldPo := &dapo.DataAgentPo{
		ID: "a1", Name: "OldName", Profile: &profile,
		ProductKey: "p-1", Config: "{}",
		CreatedBy: "u1",
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(oldPo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(errors.New("update err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

	_, err = svc.Update(context.Background(), req, "a1")
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestUpdate_Success_WithNameChange(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockMqAccess := cmpmock.NewMockLogger(ctrl) // placeholder - not actually used

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}
	_ = mockMqAccess // suppress unused

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.UpdateReq{
		Name:          "NewName",
		ProductKey:    "p-1",
		Config:        cfg,
		IsInternalAPI: true,
		UpdatedBy:     "sys-user",
	}

	profile := ""
	oldPo := &dapo.DataAgentPo{
		ID: "a1", Name: "OldName", Profile: &profile,
		ProductKey: "p-1", Config: "{}",
		CreatedBy: "u1",
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(oldPo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes() // MQ goroutine may log
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	auditLog, err := svc.Update(context.Background(), req, "a1")
	assert.NoError(t, err)
	assert.Equal(t, "OldName", auditLog.OldName)
	assert.Equal(t, "NewName", auditLog.NewName)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestUpdatePo_NotOwner_NotBuiltIn_Returns403(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	builtInNo := cdaenum.BuiltInNo
	oldPo := &dapo.DataAgentPo{
		ID: "a1", CreatedBy: "other-user",
		IsBuiltIn: &builtInNo,
	}
	newPo := &dapo.DataAgentPo{}
	req := &agentconfigreq.UpdateReq{IsInternalAPI: false}

	// UpdatedBy from ctx will be "" → not matching "other-user" → not owner
	err := svc.updatePo(context.Background(), nil, req, newPo, oldPo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不是owner")
}

func TestUpdatePo_NotOwner_BuiltIn_PmsCheckError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		pmsSvc:        mockPmsSvc,
	}

	builtInYes := cdaenum.BuiltInYes
	oldPo := &dapo.DataAgentPo{
		ID: "a1", CreatedBy: "other-user",
		IsBuiltIn: &builtInYes,
	}
	newPo := &dapo.DataAgentPo{}
	req := &agentconfigreq.UpdateReq{IsInternalAPI: false}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, errors.New("pms err"))

	err := svc.updatePo(context.Background(), nil, req, newPo, oldPo)
	assert.Error(t, err)
}

func TestUpdatePo_NotOwner_BuiltIn_NoPms_Returns403(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		pmsSvc:        mockPmsSvc,
	}

	builtInYes := cdaenum.BuiltInYes
	oldPo := &dapo.DataAgentPo{
		ID: "a1", CreatedBy: "other-user",
		IsBuiltIn: &builtInYes,
	}
	newPo := &dapo.DataAgentPo{}
	req := &agentconfigreq.UpdateReq{IsInternalAPI: false}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	err := svc.updatePo(context.Background(), nil, req, newPo, oldPo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "不是owner")
}
