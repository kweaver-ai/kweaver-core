package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Create — deeper paths ====================

func TestCreate_CreatePoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		productRepo:   mockProductRepo,
		logger:        mockLogger,
	}

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			ProductKey: "p-1",
			Config:     cfg,
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(errors.New("create err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Create(context.Background(), req)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestCreate_BatchCreateBdRelError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		productRepo:    mockProductRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		logger:         mockLogger,
	}

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			ProductKey: "p-1",
			Config:     cfg,
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(errors.New("batch err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Create(context.Background(), req)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestCreate_AssociateResourceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectRollback()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		productRepo:    mockProductRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			ProductKey: "p-1",
			Config:     cfg,
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(errors.New("http err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Create(context.Background(), req)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		productRepo:    mockProductRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			ProductKey: "p-1",
			Config:     cfg,
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(nil)

	id, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestCreate_InternalAPI_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	db, sqlMk, err := sqlmock.New()
	require.NoError(t, err)

	defer db.Close()
	sqlMk.ExpectBegin()
	sqlMk.ExpectCommit()

	tx, err := db.Begin()
	require.NoError(t, err)

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		productRepo:    mockProductRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	cfg := daconfvalobj.NewConfig()
	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			ProductKey:    "p-1",
			Config:        cfg,
			IsInternalAPI: true,
			CreatedBy:     "sys-user",
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(nil)

	id, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}
