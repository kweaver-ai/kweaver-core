package v3agentconfigsvc

import (
	"context"
	"database/sql"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setAgentConfigDisableBizDomain(t *testing.T, disable bool) {
	t.Helper()

	oldCfg := global.GConfig
	global.GConfig = &conf.Config{
		Config:       cconf.BaseDefConfig(),
		SwitchFields: conf.NewSwitchFields(),
	}
	global.GConfig.SwitchFields.DisableBizDomain = disable

	t.Cleanup(func() {
		global.GConfig = oldCfg
	})
}

func TestCreate_DisableBizDomain_SuccessSkipsAssociation(t *testing.T) {
	setAgentConfigDisableBizDomain(t, true)

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

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		productRepo:   mockProductRepo,
	}

	req := &agentconfigreq.CreateReq{
		UpdateReq: &agentconfigreq.UpdateReq{
			ProductKey: "p-1",
			Config:     daconfvalobj.NewConfig(),
		},
	}

	mockProductRepo.EXPECT().ExistsByKey(gomock.Any(), "p-1").Return(true, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)

	id, err := svc.Create(context.Background(), req)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestCopy_DisableBizDomain_SuccessSkipsAssociation(t *testing.T) {
	setAgentConfigDisableBizDomain(t, true)

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
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{
		ID: "agent-1", Name: "Agent", Key: "key-1", Config: "{}",
	}
	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any(), gomock.Any()).Return(nil)

	resp, auditLog, err := svc.Copy(context.Background(), "agent-1", &agentconfigreq.CopyReq{Name: ""})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "agent-1", auditLog.ID)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestDelete_DisableBizDomain_SuccessSkipsAssociation(t *testing.T) {
	setAgentConfigDisableBizDomain(t, true)

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
	logger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		logger:        logger,
	}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "a1").Return(&dapo.DataAgentPo{
		ID:        "a1",
		Name:      "Agent",
		Status:    "draft",
		CreatedBy: "u1",
	}, nil)
	mockAgentConfRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
	mockAgentConfRepo.EXPECT().Delete(gomock.Any(), tx, "a1").Return(nil)

	_, err = svc.Delete(context.Background(), "a1", "u1", true)
	assert.NoError(t, err)
	require.NoError(t, sqlMk.ExpectationsWereMet())
}

func TestCopy2Tpl_DisableBizDomain_SuccessSkipsAssociation(t *testing.T) {
	setAgentConfigDisableBizDomain(t, true)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentConfRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)

	svc := &dataAgentConfigSvc{
		SvcBase:       service.NewSvcBase(),
		agentConfRepo: mockAgentConfRepo,
		agentTplRepo:  mockTplRepo,
		logger:        mockLogger,
	}

	sourcePo := &dapo.DataAgentPo{ID: "agent-1", Name: "Agent", CreatedBy: "", Config: "{}"}
	tplPo := &dapo.DataAgentTplPo{ID: 101}

	mockAgentConfRepo.EXPECT().GetByID(gomock.Any(), "agent-1").Return(sourcePo, nil)
	mockTplRepo.EXPECT().Create(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), gomock.Any(), gomock.Any()).Return(tplPo, nil)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	existingTx := &sql.Tx{}
	res, auditInfo, err := svc.Copy2Tpl(context.Background(), "agent-1", &agentconfigreq.Copy2TplReq{}, existingTx)
	assert.NoError(t, err)
	assert.NotNil(t, res)
	assert.Equal(t, "agent-1", auditInfo.ID)
}
