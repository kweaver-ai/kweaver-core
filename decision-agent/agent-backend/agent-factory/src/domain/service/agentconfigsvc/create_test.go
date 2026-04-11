package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_Create_PanicsWithoutAgentConfRepo(t *testing.T) {
	svc := &dataAgentConfigSvc{
		SvcBase: service.NewSvcBase(),
		// agentConfRepo is nil
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{}

	assert.Panics(t, func() {
		_, _ = svc.Create(ctx, req)
	})
}

func TestDataAgentConfigSvc_Create_ProductExistsByKeyError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
		logger:      mockLogger,
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.ProductKey = "product-123"

	dbErr := errors.New("database connection failed")
	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.ProductKey).Return(false, dbErr)

	_, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestDataAgentConfigSvc_Create_ProductNotExists(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:     service.NewSvcBase(),
		productRepo: mockProductRepo,
		logger:      mockLogger,
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.ProductKey = "non-existent-product"

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.ProductKey).Return(false, nil)

	_, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "产品不存在")
}

func TestDataAgentConfigSvc_Create_D2ePanicOnNilConfig(t *testing.T) {
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

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.ProductKey = "product-123"

	// D2e conversion will panic with nil Config
	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.ProductKey).Return(true, nil)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	assert.Panics(t, func() {
		_, _ = svc.Create(ctx, req)
	})
}

func TestDataAgentConfigSvc_Create_SystemAgentPermissionCheckError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		pmsSvc:        mockPmsSvc,
		logger:        mockLogger,
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.ProductKey = "product-123"

	builtInYes := cenum.YesNoInt8(1)
	req.IsSystemAgent = &builtInYes

	pmsErr := errors.New("permission check failed")

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.ProductKey).Return(true, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, pmsErr)

	_, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "check system agent create permission failed")
}

func TestDataAgentConfigSvc_Create_SystemAgentNoPermission(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockPmsSvc := v3portdrivermock.NewMockIPermissionSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		productRepo:   mockProductRepo,
		agentConfRepo: mockAgentConfRepo,
		pmsSvc:        mockPmsSvc,
		logger:        mockLogger,
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.ProductKey = "product-123"

	builtInYes := cenum.YesNoInt8(1)
	req.IsSystemAgent = &builtInYes

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), req.ProductKey).Return(true, nil)
	mockPmsSvc.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

	_, err := svc.Create(ctx, req)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "do not have system agent create permission")
}

func TestDataAgentConfigSvc_createPo_InternalAPISetsCreatedBy(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.IsInternalAPI = true
	req.CreatedBy = "internal-user-123"

	po := &dapo.DataAgentPo{}

	mockAgentConfRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_ context.Context, _ interface{}, _ string, actualPo *dapo.DataAgentPo) {
		// Verify CreatedBy is set from request
		assert.Equal(t, "internal-user-123", actualPo.CreatedBy)
		assert.Equal(t, "internal-user-123", actualPo.UpdatedBy)
	}).Return(nil)

	err := svc.createPo(ctx, nil, req, po, "agent-123")

	assert.NoError(t, err)
	assert.Equal(t, cdaenum.StatusUnpublished, po.Status)
}

func TestDataAgentConfigSvc_CreatePo_RepoCreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
	}

	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			IsInternalAPI: true,
			UpdatedBy:     "sys",
		},
	}
	po := &dapo.DataAgentPo{}

	mockAgentConfRepo.EXPECT().Create(gomock.Any(), gomock.Any(), "agent-id", po).
		Return(errors.New("db create error"))

	err := svc.createPo(context.Background(), nil, req, po, "agent-id")
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_Create_BeginTxError(t *testing.T) {
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

	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			Name:       "BxAgent",
			ProductKey: "prod-key",
			Config:     daconfvalobj.NewConfig(),
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "prod-key").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err := svc.Create(context.Background(), req)
	assert.Error(t, err)
}

func TestDataAgentConfigSvc_createPo_NonInternalAPISetsCreatedByFromCtx(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	ctx := context.Background()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{},
	}
	req.IsInternalAPI = false

	po := &dapo.DataAgentPo{}

	mockAgentConfRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Do(func(_ context.Context, _ interface{}, _ string, actualPo *dapo.DataAgentPo) {
		// Verify CreatedBy is set from context (empty string in test)
		assert.Equal(t, "", actualPo.CreatedBy)
		assert.Equal(t, "", actualPo.UpdatedBy)
	}).Return(nil)

	err := svc.createPo(ctx, nil, req, po, "agent-123")

	assert.NoError(t, err)
	assert.Equal(t, cdaenum.StatusUnpublished, po.Status)
}
