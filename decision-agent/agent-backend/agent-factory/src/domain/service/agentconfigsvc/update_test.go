package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_Update_PanicsWithoutAgentConfRepo(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
	}

	ctx := context.Background()
	agentID := "agent-123"
	req := &agentconfigreq.UpdateReq{}

	assert.Panics(t, func() {
		_, _ = svc.Update(ctx, req, agentID)
	})
}

func TestDataAgentConfigSvc_Update_ProductNotFoundError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
		logger:      mockLogger,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "p-1"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(false, errors.New("db err"))

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Update_ProductNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
		logger:      mockLogger,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "p-1"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(false, nil)

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品不存在")
}

func TestDataAgentConfigSvc_Update_AgentNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "p-1"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(nil, sql.ErrNoRows)

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据智能体配置不存在")
}

func TestDataAgentConfigSvc_Update_GetByIDError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "p-1"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(nil, errors.New("db err"))

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Update_NoChange_EarlyReturn(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	// Use IsInternalAPI=true: skips IsChanged check → continues to BeginTx
	// But we test the NoChange path: IsInternalAPI=false, same data so IsChanged=false → early return
	// To make IsChanged=false, ensure config serializes identically
	cfg := daconfvalobj.NewConfig()
	cfgStr, _ := cutil.JSON().MarshalToString(cfg)

	req := &agentconfigreq.UpdateReq{
		Name:          "Test Agent",
		Profile:       "profile",
		ProductKey:    "p-1",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "avatar",
		Config:        cfg,
		IsInternalAPI: false,
	}

	profile := "profile"
	oldPo := &dapo.DataAgentPo{
		ID:         "agent-1",
		Name:       "Test Agent",
		Profile:    &profile,
		ProductKey: "p-1",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar",
		Config:     cfgStr,
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(oldPo, nil)

	// No change → early return, no further mock calls expected
	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_Update_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
		Name:          "New Name",
		Profile:       "profile",
		ProductKey:    "p-1",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "avatar",
		Config:        cfg,
		IsInternalAPI: false,
	}

	profile := "profile"
	oldPo := &dapo.DataAgentPo{
		ID:         "agent-1",
		Name:       "Old Name",
		Profile:    &profile,
		ProductKey: "p-1",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar",
		Config:     "{}",
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(oldPo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx error"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Update_InternalAPI_IsChanged_True_BeginTxError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

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
		Name:          "New Name",
		Profile:       "profile",
		ProductKey:    "p-1",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "avatar",
		Config:        cfg,
		IsInternalAPI: true,
		UpdatedBy:     "sys-user",
	}

	profile := "profile"
	oldPo := &dapo.DataAgentPo{
		ID:         "agent-1",
		Name:       "Old Name",
		Profile:    &profile,
		ProductKey: "p-1",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar",
		Config:     "{}",
		CreatedBy:  "user-1",
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(oldPo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()
	mockLogger.EXPECT().Errorln(gomock.Any()).AnyTimes()

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Update_IsNotChanged_Returns(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
	}

	cfg := daconfvalobj.NewConfig()
	profile := "profile"
	cfgStr, _ := cutil.JSON().MarshalToString(cfg)
	req := &agentconfigreq.UpdateReq{
		Name:          "Same Name",
		Profile:       profile,
		ProductKey:    "prod-1",
		AvatarType:    cdaenum.AvatarTypeBuiltIn,
		Avatar:        "avatar",
		Config:        cfg,
		IsInternalAPI: false,
	}
	oldPo := &dapo.DataAgentPo{
		ID:         "agent-1",
		Name:       "Same Name",
		Profile:    &profile,
		ProductKey: "prod-1",
		AvatarType: cdaenum.AvatarTypeBuiltIn,
		Avatar:     "avatar",
		Config:     cfgStr,
		CreatedBy:  "user-1",
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "prod-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(oldPo, nil)
	// No BeginTx: isChanged=false → returns early

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.NoError(t, err)
}

func TestDataAgentConfigSvc_Update_D2eNilConfigPanic(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
	}

	// Config=nil → D2e panics; use same Name as oldPo to avoid MQ goroutine trigger
	req := &agentconfigreq.UpdateReq{
		Name:          "SameName",
		ProductKey:    "prod-1",
		IsInternalAPI: true,
		Config:        nil,
	}
	profile := ""
	oldPo := &dapo.DataAgentPo{
		ID:         "agent-1",
		Name:       "SameName",
		Profile:    &profile,
		ProductKey: "prod-1",
		Config:     "{}",
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "prod-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(oldPo, nil)

	// Config=nil → D2e panics
	assert.Panics(t, func() {
		_, _ = svc.Update(context.Background(), req, "agent-1")
	})
}

func TestDataAgentConfigSvc_Update_GetByIDGenericError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "prod-1"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "prod-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-err").Return(nil, errors.New("db error"))

	_, err := svc.Update(context.Background(), req, "agent-err")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Update_ProductNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "bad-product"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "bad-product").Return(false, nil)

	_, err := svc.Update(context.Background(), req, "agent-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品不存在")
}

func TestDataAgentConfigSvc_Update_GetByIDNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
	}

	req := &agentconfigreq.UpdateReq{ProductKey: "prod-1"}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "prod-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-missing").Return(nil, sql.ErrNoRows)

	_, err := svc.Update(context.Background(), req, "agent-missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "数据智能体配置不存在")
}
