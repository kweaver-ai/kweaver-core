package agentinoutsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_inout/agentinoutresp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopAgentLogger struct{}

func (noopAgentLogger) Infof(string, ...interface{})  {}
func (noopAgentLogger) Infoln(...interface{})         {}
func (noopAgentLogger) Debugf(string, ...interface{}) {}
func (noopAgentLogger) Debugln(...interface{})        {}
func (noopAgentLogger) Errorf(string, ...interface{}) {}
func (noopAgentLogger) Errorln(...interface{})        {}
func (noopAgentLogger) Warnf(string, ...interface{})  {}
func (noopAgentLogger) Warnln(...interface{})         {}
func (noopAgentLogger) Panicf(string, ...interface{}) {}
func (noopAgentLogger) Panicln(...interface{})        {}
func (noopAgentLogger) Fatalf(string, ...interface{}) {}
func (noopAgentLogger) Fatalln(...interface{})        {}

func newAgentSQLTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	cleanup := func() {
		require.NoError(t, mock.ExpectationsWereMet())

		_ = db.Close()
	}

	return tx, mock, cleanup
}

func makeExportData(keys ...string) *agentinoutresp.ExportResp {
	agents := make([]*agentinoutresp.ExportAgentItem, 0, len(keys))
	for _, key := range keys {
		agents = append(agents, &agentinoutresp.ExportAgentItem{
			DataAgentPo: &dapo.DataAgentPo{
				Key:  key,
				Name: "name-" + key,
			},
		})
	}

	return &agentinoutresp.ExportResp{Agents: agents}
}

func setAgentInOutDisableBizDomain(t *testing.T, disable bool) {
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

func TestAgentInOutSvc_checkBizDomainConflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
	mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
	svc := &agentInOutSvc{
		agentConfRepo: mockAgentRepo,
		bizDomainHttp: mockBiz,
	}

	ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029

	t.Run("biz domain query error", func(t *testing.T) {
		resp := agentinoutresp.NewImportResp()

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).
			Return(nil, nil, errors.New("http failed"))

		err := svc.checkBizDomainConflict(ctx, makeExportData("k1"), resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get all agent id list by biz domain id failed")
	})

	t.Run("agent repo query error", func(t *testing.T) {
		resp := agentinoutresp.NewImportResp()

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).
			Return(nil, errors.New("db failed"))

		err := svc.checkBizDomainConflict(ctx, makeExportData("k1"), resp)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get agent by keys failed")
	})

	t.Run("conflict found", func(t *testing.T) {
		resp := agentinoutresp.NewImportResp()

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).
			Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).
			Return([]*dapo.DataAgentPo{{ID: "a2", Key: "k1", Name: "name-k1"}}, nil)

		err := svc.checkBizDomainConflict(ctx, makeExportData("k1"), resp)
		assert.NoError(t, err)
		assert.True(t, resp.HasFail())
		assert.Len(t, resp.BizDomainConflict, 1)
	})
}

func TestAgentInOutSvc_importByCreate(t *testing.T) {
	t.Run("success when biz domain disabled skips association", func(t *testing.T) {
		setAgentInOutDisableBizDomain(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newAgentSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo: mockAgentRepo,
			logger:        noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockAgentRepo.EXPECT().CreateBatch(gomock.Any(), tx, gomock.Any()).Return(nil)

		err := svc.importByCreate(context.Background(), exportData, "u1", resp)
		assert.NoError(t, err)
	})

	t.Run("begin tx error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo:  mockAgentRepo,
			bdAgentRelRepo: mockBdRelRepo,
			bizDomainHttp:  mockBiz,
			logger:         noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx failed"))

		err := svc.importByCreate(context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1"), exportData, "u1", resp) //nolint:staticcheck // SA1029
		assert.Error(t, err)
	})

	t.Run("associate batch error rolls back", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newAgentSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo:  mockAgentRepo,
			bdAgentRelRepo: mockBdRelRepo,
			bizDomainHttp:  mockBiz,
			logger:         noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockAgentRepo.EXPECT().CreateBatch(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBdRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBiz.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(errors.New("http failed"))

		err := svc.importByCreate(context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1"), exportData, "u1", resp) //nolint:staticcheck // SA1029
		assert.Error(t, err)
	})

	t.Run("success commits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newAgentSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo:  mockAgentRepo,
			bdAgentRelRepo: mockBdRelRepo,
			bizDomainHttp:  mockBiz,
			logger:         noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockAgentRepo.EXPECT().CreateBatch(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBdRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBiz.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.importByCreate(context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1"), exportData, "u1", resp) //nolint:staticcheck // SA1029
		assert.NoError(t, err)
	})
}

func TestAgentInOutSvc_importByUpsert(t *testing.T) {
	t.Run("success when biz domain disabled skips conflict check and association", func(t *testing.T) {
		setAgentInOutDisableBizDomain(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newAgentSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo: mockAgentRepo,
			logger:        noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()

		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockAgentRepo.EXPECT().CreateBatch(gomock.Any(), tx, gomock.Any()).Return(nil)

		err := svc.importByUpsert(context.Background(), exportData, "u1", resp)
		assert.NoError(t, err)
	})

	t.Run("begin tx error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo:  mockAgentRepo,
			bdAgentRelRepo: mockBdRelRepo,
			bizDomainHttp:  mockBiz,
			logger:         noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()
		ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).Return([]string{}, map[string]string{}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil).Times(2)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(nil, errors.New("tx failed"))

		err := svc.importByUpsert(ctx, exportData, "u1", resp)
		assert.Error(t, err)
	})

	t.Run("update path error rolls back", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newAgentSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo:  mockAgentRepo,
			bdAgentRelRepo: mockBdRelRepo,
			bizDomainHttp:  mockBiz,
			logger:         noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()
		ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029

		existing := &dapo.DataAgentPo{ID: "a1", Key: "k1", CreatedBy: "u1"}

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).Return([]string{"a1"}, map[string]string{"a1": "bd-1"}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{existing}, nil).Times(2)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockAgentRepo.EXPECT().UpdateByKey(gomock.Any(), tx, gomock.Any()).Return(errors.New("update failed"))

		err := svc.importByUpsert(ctx, exportData, "u1", resp)
		assert.Error(t, err)
	})

	t.Run("create path success commits", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newAgentSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockBiz := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &agentInOutSvc{
			agentConfRepo:  mockAgentRepo,
			bdAgentRelRepo: mockBdRelRepo,
			bizDomainHttp:  mockBiz,
			logger:         noopAgentLogger{},
		}

		exportData := makeExportData("k1")
		resp := agentinoutresp.NewImportResp()
		ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029

		mockBiz.EXPECT().GetAllAgentIDList(gomock.Any(), []string{"bd-1"}).Return([]string{}, map[string]string{}, nil)
		mockAgentRepo.EXPECT().GetByKeys(gomock.Any(), []string{"k1"}).Return([]*dapo.DataAgentPo{}, nil).Times(2)
		mockAgentRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockAgentRepo.EXPECT().CreateBatch(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBdRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBiz.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.importByUpsert(ctx, exportData, "u1", resp)
		assert.NoError(t, err)
	})
}
