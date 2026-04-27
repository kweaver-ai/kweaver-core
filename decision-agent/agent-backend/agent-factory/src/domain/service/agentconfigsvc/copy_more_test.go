package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataAgentConfigSvc_Copy_CreateAgentError(t *testing.T) {
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
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{
		ID:     "agent-1",
		Name:   "Agent",
		Key:    "key-1",
		Config: "{}",
	}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(errors.New("create failed"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, _, err = svc.Copy(context.Background(), "agent-1", &agentconfigreq.CopyReq{Name: ""})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "保存新Agent失败")
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDataAgentConfigSvc_Copy_BdRelBatchCreateError(t *testing.T) {
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
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		logger:         mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{
		ID: "agent-1", Name: "Agent", Key: "key-1", Config: "{}",
	}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(errors.New("batch create failed"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, _, err = svc.Copy(context.Background(), "agent-1", &agentconfigreq.CopyReq{Name: ""})
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDataAgentConfigSvc_Copy_AssociateResourceError(t *testing.T) {
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
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{
		ID: "agent-1", Name: "Agent", Key: "key-1", Config: "{}",
	}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(errors.New("associate failed"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, _, err = svc.Copy(context.Background(), "agent-1", &agentconfigreq.CopyReq{Name: ""})
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDataAgentConfigSvc_Copy_Success(t *testing.T) {
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
	mockBdAgentRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
	mockBizDomainHttp := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{
		ID: "agent-1", Name: "Agent", Key: "key-1", Config: "{}",
	}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)
	mockBdAgentRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
	mockBizDomainHttp.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(nil)

	resp, auditLog, err := svc.Copy(context.Background(), "agent-1", &agentconfigreq.CopyReq{Name: ""})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "Agent_副本", resp.Name)
	assert.Equal(t, "agent-1", auditLog.ID)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}
