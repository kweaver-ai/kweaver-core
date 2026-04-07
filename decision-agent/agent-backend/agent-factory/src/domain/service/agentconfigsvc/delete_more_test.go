package v3agentconfigsvc

import (
	"context"
	"errors"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== Delete — deeper paths ====================

func TestDelete_RepoDeleteError(t *testing.T) {
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

	builtInNo := cdaenum.BuiltInNo
	po := &dapo.DataAgentPo{
		ID: "a1", Name: "Agent", Status: cdaenum.StatusUnpublished,
		CreatedBy: "u1", IsBuiltIn: &builtInNo,
	}

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Delete(gomock.Any(), tx, "a1").Return(errors.New("delete err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Delete(context.Background(), "a1", "u1", false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDelete_BdRelDeleteError(t *testing.T) {
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

	builtInNo := cdaenum.BuiltInNo
	po := &dapo.DataAgentPo{
		ID: "a1", Name: "Agent", Status: cdaenum.StatusUnpublished,
		CreatedBy: "u1", IsBuiltIn: &builtInNo,
	}

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		logger:         mockLogger,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Delete(gomock.Any(), tx, "a1").Return(nil)
	mockBdAgentRelRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, "a1").Return(errors.New("bd rel err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Delete(context.Background(), "a1", "u1", false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDelete_DisassociateResourceError(t *testing.T) {
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

	builtInNo := cdaenum.BuiltInNo
	po := &dapo.DataAgentPo{
		ID: "a1", Name: "Agent", Status: cdaenum.StatusUnpublished,
		CreatedBy: "u1", IsBuiltIn: &builtInNo,
	}

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Delete(gomock.Any(), tx, "a1").Return(nil)
	mockBdAgentRelRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, "a1").Return(nil)
	mockBizDomainHttp.EXPECT().DisassociateResource(gomock.Any(), gomock.Any()).Return(errors.New("http err"))
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	_, err = svc.Delete(context.Background(), "a1", "u1", false)
	assert.Error(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDelete_Success(t *testing.T) {
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

	builtInNo := cdaenum.BuiltInNo
	po := &dapo.DataAgentPo{
		ID: "a1", Name: "Agent", Status: cdaenum.StatusUnpublished,
		CreatedBy: "u1", IsBuiltIn: &builtInNo,
	}

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Delete(gomock.Any(), tx, "a1").Return(nil)
	mockBdAgentRelRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, "a1").Return(nil)
	mockBizDomainHttp.EXPECT().DisassociateResource(gomock.Any(), gomock.Any()).Return(nil)

	auditLog, err := svc.Delete(context.Background(), "a1", "u1", false)
	assert.NoError(t, err)
	assert.Equal(t, "a1", auditLog.ID)
	assert.Equal(t, "Agent", auditLog.Name)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDelete_PrivateAPI_Success(t *testing.T) {
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

	builtInNo := cdaenum.BuiltInNo
	po := &dapo.DataAgentPo{
		ID: "a1", Name: "Agent", Status: cdaenum.StatusUnpublished,
		CreatedBy: "other-user", IsBuiltIn: &builtInNo,
	}

	svc := &dataAgentConfigSvc{
		SvcBase:        service.NewSvcBase(),
		agentConfRepo:  mockAgentConfRepo,
		bdAgentRelRepo: mockBdAgentRelRepo,
		bizDomainHttp:  mockBizDomainHttp,
		logger:         mockLogger,
	}

	// isPrivate=true → skips owner/builtIn checks
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(po, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Delete(gomock.Any(), tx, "a1").Return(nil)
	mockBdAgentRelRepo.EXPECT().DeleteByAgentID(gomock.Any(), tx, "a1").Return(nil)
	mockBizDomainHttp.EXPECT().DisassociateResource(gomock.Any(), gomock.Any()).Return(nil)

	_, err = svc.Delete(context.Background(), "a1", "u1", true)
	assert.NoError(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}
