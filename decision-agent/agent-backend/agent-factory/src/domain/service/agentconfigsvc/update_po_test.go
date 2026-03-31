package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	idbaccessmock "github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
)

// ---- updatePo ----

func TestDataAgentConfigSvc_UpdatePo_InternalAPI_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	builtIn := cdaenum.BuiltInNo
	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.UpdateReq{
		IsInternalAPI: true,
		UpdatedBy:     "system-user",
	}
	po := &dapo.DataAgentPo{ID: "agent-1"}
	oldPo := &dapo.DataAgentPo{ID: "agent-1", CreatedBy: "owner", IsBuiltIn: &builtIn}

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := svc.updatePo(ctx, &sql.Tx{}, req, po, oldPo)

	assert.NoError(t, err)
	assert.Equal(t, "system-user", po.UpdatedBy)
}

func TestDataAgentConfigSvc_UpdatePo_IsOwner_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	builtIn := cdaenum.BuiltInNo
	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.UpdateReq{
		IsInternalAPI: false,
	}
	po := &dapo.DataAgentPo{ID: "agent-1", UpdatedBy: ""}
	oldPo := &dapo.DataAgentPo{ID: "agent-1", CreatedBy: "", IsBuiltIn: &builtIn}

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := svc.updatePo(ctx, &sql.Tx{}, req, po, oldPo)

	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_UpdatePo_NotOwner_NotBuiltIn_Forbidden(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	builtIn := cdaenum.BuiltInNo
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	req := &agentconfigreq.UpdateReq{
		IsInternalAPI: false,
	}
	po := &dapo.DataAgentPo{ID: "agent-1", UpdatedBy: ""}
	oldPo := &dapo.DataAgentPo{ID: "agent-1", CreatedBy: "real-owner", IsBuiltIn: &builtIn}

	err := svc.updatePo(ctx, &sql.Tx{}, req, po, oldPo)

	assert.Error(t, err)
}

func TestDataAgentConfigSvc_UpdatePo_NotOwner_BuiltIn_HasPermission_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	builtIn := cdaenum.BuiltInYes
	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockRepo,
		pmsSvc:        mockPmsSvc,
	}

	ctx := context.Background()
	req := &agentconfigreq.UpdateReq{
		IsInternalAPI: false,
	}
	po := &dapo.DataAgentPo{ID: "agent-1", UpdatedBy: ""}
	oldPo := &dapo.DataAgentPo{ID: "agent-1", CreatedBy: "real-owner", IsBuiltIn: &builtIn}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	err := svc.updatePo(ctx, &sql.Tx{}, req, po, oldPo)

	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_UpdatePo_NotOwner_BuiltIn_NoPermission_Forbidden(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	builtIn := cdaenum.BuiltInYes
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	ctx := context.Background()
	req := &agentconfigreq.UpdateReq{
		IsInternalAPI: false,
	}
	po := &dapo.DataAgentPo{ID: "agent-1", UpdatedBy: ""}
	oldPo := &dapo.DataAgentPo{ID: "agent-1", CreatedBy: "real-owner", IsBuiltIn: &builtIn}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	err := svc.updatePo(ctx, &sql.Tx{}, req, po, oldPo)

	assert.Error(t, err)
}

func TestDataAgentConfigSvc_UpdatePo_RepoUpdateError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	builtIn := cdaenum.BuiltInNo
	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockRepo,
	}

	ctx := context.Background()
	req := &agentconfigreq.UpdateReq{
		IsInternalAPI: true,
		UpdatedBy:     "system-user",
	}
	po := &dapo.DataAgentPo{ID: "agent-1"}
	oldPo := &dapo.DataAgentPo{ID: "agent-1", CreatedBy: "owner", IsBuiltIn: &builtIn}

	mockRepo.EXPECT().Update(gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("db error"))

	err := svc.updatePo(ctx, &sql.Tx{}, req, po, oldPo)

	assert.Error(t, err)
}

// ---- copyAgentPo ----

func TestDataAgentConfigSvc_UpdatePo_BuiltIn_PmsError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)

	builtInYes := cdaenum.BuiltInYes
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		pmsSvc:  mockPmsSvc,
	}

	oldPo := &dapo.DataAgentPo{
		CreatedBy: "other-user",
		IsBuiltIn: &builtInYes,
	}
	po := &dapo.DataAgentPo{}
	req := &agentconfigreq.UpdateReq{IsInternalAPI: false}

	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(false, errors.New("pms err"))

	err := svc.updatePo(context.Background(), (*sql.Tx)(nil), req, po, oldPo)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_UpdatePo_NotOwnerNotBuiltIn_403(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	builtInNo := cdaenum.BuiltInNo
	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockRepo,
	}

	oldPo := &dapo.DataAgentPo{
		CreatedBy: "other-user",
		IsBuiltIn: &builtInNo,
	}
	po := &dapo.DataAgentPo{}
	req := &agentconfigreq.UpdateReq{IsInternalAPI: false}

	err := svc.updatePo(context.Background(), (*sql.Tx)(nil), req, po, oldPo)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_CopyAgentPo_Success(t *testing.T) {
	t.Parallel()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	newPo := &dapo.DataAgentPo{}
	sourcePo := &dapo.DataAgentPo{
		ID:        "source-id",
		Name:      "Source Agent",
		Key:       "source-key",
		CreatedBy: "original-owner",
		Config:    `{}`,
	}

	err := svc.copyAgentPo(ctx, newPo, sourcePo, "new-id", "new-key", "New Agent")

	assert.NoError(t, err)
	assert.Equal(t, "new-id", newPo.ID)
	assert.Equal(t, "new-key", newPo.Key)
	assert.Equal(t, "New Agent", newPo.Name)
	assert.Equal(t, cdaenum.StatusUnpublished, newPo.Status)
	assert.Equal(t, int64(0), newPo.DeletedAt)
	assert.Equal(t, "", newPo.DeletedBy)
}

func TestDataAgentConfigSvc_CopyAgentPo_SetsTimestamps(t *testing.T) {
	t.Parallel()

	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	newPo := &dapo.DataAgentPo{}
	sourcePo := &dapo.DataAgentPo{
		ID:        "source-id",
		Name:      "Source",
		Key:       "source-key",
		CreatedAt: 1000,
		UpdatedAt: 2000,
		Config:    `{}`,
	}

	err := svc.copyAgentPo(ctx, newPo, sourcePo, "new-id", "new-key", "Copy")

	assert.NoError(t, err)
	assert.Greater(t, newPo.CreatedAt, int64(0))
	assert.Equal(t, newPo.CreatedAt, newPo.UpdatedAt)
}
