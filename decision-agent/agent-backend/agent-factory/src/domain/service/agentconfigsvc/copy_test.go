package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_Copy_PanicsWithoutAgentConfRepo(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		// agentConfRepo is nil
	}

	ctx := context.Background()
	agentID := "agent-123"
	req := &agentconfigreq.CopyReq{}

	assert.Panics(t, func() {
		_, _, _ = svc.Copy(ctx, agentID, req)
	})
}

func TestDataAgentConfigSvc_Copy_AgentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	agentID := "non-existent-agent"
	req := &agentconfigreq.CopyReq{
		Name: "New Agent Copy",
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(nil, errors.New("record not found"))

	resp, auditLogInfo, err := svc.Copy(ctx, agentID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Empty(t, auditLogInfo.ID)
	assert.Empty(t, auditLogInfo.Name)
}

func TestDataAgentConfigSvc_Copy_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	req := &agentconfigreq.CopyReq{
		Name: "", // Empty name to skip ExistsByName check
	}

	sourcePo := &dapo.DataAgentPo{
		ID:   "agent-123",
		Name: "Original Agent",
		Key:  "original-key",
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(sourcePo, nil)

	txErr := errors.New("transaction begin failed")
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, txErr)

	resp, auditLogInfo, err := svc.Copy(ctx, agentID, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "开启事务失败")
	assert.Nil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.NotEmpty(t, auditLogInfo.Name)
}

func TestDataAgentConfigSvc_Copy_NameConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	ctx := context.Background()
	agentID := "agent-123"
	req := &agentconfigreq.CopyReq{
		Name: "Existing Agent Name",
	}

	sourcePo := &dapo.DataAgentPo{
		ID:   "agent-123",
		Name: "Original Agent",
		Key:  "original-key",
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), agentID).Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().ExistsByName(gomock.Any(), "Existing Agent Name").Return(true, nil)

	resp, auditLogInfo, err := svc.Copy(ctx, agentID, req)

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.NotEmpty(t, auditLogInfo.ID)
	assert.NotEmpty(t, auditLogInfo.Name)
}

func TestDataAgentConfigSvc_Copy_ExistsByNameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	sourcePo := &dapo.DataAgentPo{
		ID:   "agent-1",
		Name: "Agent",
		Key:  "key-1",
	}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().ExistsByName(gomock.Any(), "NewName").Return(false, errors.New("db err"))

	_, _, err := svc.Copy(context.Background(), "agent-1", &agentconfigreq.CopyReq{Name: "NewName"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "检查Agent名称是否存在失败")
}

func TestDataAgentConfigSvc_copyAgentPo_Success(t *testing.T) {
	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	builtInNo := cdaenum.BuiltInNo
	sourcePo := &dapo.DataAgentPo{
		ID:        "agent-1",
		Name:      "Agent",
		Key:       "key-1",
		Config:    "{}",
		IsBuiltIn: &builtInNo,
	}
	newPo := &dapo.DataAgentPo{}

	err := svc.copyAgentPo(context.Background(), newPo, sourcePo, "new-id", "new-key", "New Agent")
	assert.NoError(t, err)
	assert.Equal(t, "new-id", newPo.ID)
	assert.Equal(t, "new-key", newPo.Key)
	assert.Equal(t, "New Agent", newPo.Name)
	assert.Equal(t, cdaenum.StatusUnpublished, newPo.Status)
}
