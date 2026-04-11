package tplsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/cconf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/conf"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_tpl/agenttplreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driven/ihttpaccess/iumacc/httpaccmock"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopTplLogger struct{}

func (noopTplLogger) Infof(string, ...interface{})  {}
func (noopTplLogger) Infoln(...interface{})         {}
func (noopTplLogger) Debugf(string, ...interface{}) {}
func (noopTplLogger) Debugln(...interface{})        {}
func (noopTplLogger) Errorf(string, ...interface{}) {}
func (noopTplLogger) Errorln(...interface{})        {}
func (noopTplLogger) Warnf(string, ...interface{})  {}
func (noopTplLogger) Warnln(...interface{})         {}
func (noopTplLogger) Panicf(string, ...interface{}) {}
func (noopTplLogger) Panicln(...interface{})        {}
func (noopTplLogger) Fatalf(string, ...interface{}) {}
func (noopTplLogger) Fatalln(...interface{})        {}

func newTplTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock, func()) {
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

func setTplDisableBizDomain(t *testing.T, disable bool) {
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

func TestDataAgentTplSvc_HelperFunctions(t *testing.T) {
	t.Run("handleCategory category map query error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCategory := idbaccessmock.NewMockICategoryRepo(ctrl)
		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			categoryRepo:     mockCategory,
			publishedTplRepo: mockPubed,
		}

		mockCategory.EXPECT().GetIDNameMap(gomock.Any(), []string{"c1"}).Return(nil, assert.AnError)

		err := svc.handleCategory(context.Background(), []string{"c1"}, 1, nil)
		assert.Error(t, err)
	})

	t.Run("handleCategory category not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCategory := idbaccessmock.NewMockICategoryRepo(ctrl)
		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			categoryRepo:     mockCategory,
			publishedTplRepo: mockPubed,
		}

		mockCategory.EXPECT().GetIDNameMap(gomock.Any(), []string{"c1"}).Return(map[string]string{}, nil)

		err := svc.handleCategory(context.Background(), []string{"c1"}, 1, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "不存在")
	})

	t.Run("handleCategory success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockCategory := idbaccessmock.NewMockICategoryRepo(ctrl)
		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			categoryRepo:     mockCategory,
			publishedTplRepo: mockPubed,
		}

		mockCategory.EXPECT().GetIDNameMap(gomock.Any(), []string{" c1 ", "c2"}).Return(map[string]string{
			" c1 ": "C1", "c2": "C2",
		}, nil)
		mockPubed.EXPECT().DelCategoryAssocByTplID(gomock.Any(), nil, int64(1)).Return(nil)
		mockPubed.EXPECT().BatchCreateCategoryAssoc(gomock.Any(), nil, gomock.Any()).Return(nil)

		err := svc.handleCategory(context.Background(), []string{" c1 ", "c2"}, 1, nil)
		assert.NoError(t, err)
	})

	t.Run("genPublishedPo delete failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		svc := &dataAgentTplSvc{publishedTplRepo: mockPubed}
		po := &dapo.DataAgentTplPo{ID: 1, Name: "n1", Key: "k1"}

		mockPubed.EXPECT().DeleteByTplID(gomock.Any(), nil, int64(1)).Return(assert.AnError)

		_, err := svc.genPublishedPo(context.Background(), nil, po, 1, "u1")
		assert.Error(t, err)
	})

	t.Run("genPublishedPo success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		svc := &dataAgentTplSvc{publishedTplRepo: mockPubed}
		po := &dapo.DataAgentTplPo{ID: 1, Name: "n1", Key: "k1"}

		mockPubed.EXPECT().DeleteByTplID(gomock.Any(), nil, int64(1)).Return(nil)
		mockPubed.EXPECT().Create(gomock.Any(), nil, gomock.Any()).Return(int64(9), nil)

		id, err := svc.genPublishedPo(context.Background(), nil, po, 1, "u1")
		assert.NoError(t, err)
		assert.Equal(t, int64(9), id)
	})

	t.Run("copyPo and updatePo", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		svc := &dataAgentTplSvc{
			agentTplRepo: mockTplRepo,
		}
		source := &dapo.DataAgentTplPo{ID: 1, Name: "src", Key: "src-key"}
		newPo := &dapo.DataAgentTplPo{}

		ctx := createTplCtxWithUserID("u1")

		mockTplRepo.EXPECT().Create(gomock.Any(), nil, gomock.Any()).Return(nil)
		mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), nil, gomock.Any()).Return(&dapo.DataAgentTplPo{ID: 10}, nil)

		id, err := svc.copyPo(ctx, nil, newPo, source, "copy")
		assert.NoError(t, err)
		assert.Equal(t, int64(10), id)
		assert.Equal(t, "copy", newPo.Name)
		assert.Equal(t, cdaenum.StatusUnpublished, newPo.Status)

		mockTplRepo.EXPECT().Update(gomock.Any(), nil, gomock.Any()).Return(nil)
		err = svc.updatePo(ctx, nil, &dapo.DataAgentTplPo{ID: 10})
		assert.NoError(t, err)
	})
}

func TestDataAgentTplSvc_PublishUnpublish_Success(t *testing.T) {
	t.Run("publish success with external tx", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, _, done := newTplTx(t)
		defer done()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubed,
			umHttp:           mockUm,
			logger:           noopTplLogger{},
		}

		req := agenttplreq.NewPublishReq()
		ctx := createTplCtxWithUserID("u1")
		po := &dapo.DataAgentTplPo{ID: 1, Name: "tpl1", CreatedBy: "u1"}

		mockTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(1)).Return(po, nil)
		mockPubed.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(1)).Return(nil)
		mockPubed.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(100), nil)
		mockTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, int64(1), "u1", gomock.Any()).Return(nil)
		mockUm.EXPECT().GetSingleUserName(gomock.Any(), "u1").Return("U1", nil)

		resp, _, err := svc.Publish(ctx, tx, req, 1, true)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, int64(100), resp.AgentTplId)
	})

	t.Run("unpublish success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubed := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubed,
			pmsSvc:           mockPms,
		}

		ctx := createTplCtxWithUserID("u1")

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(&dapo.DataAgentTplPo{ID: 1, Name: "tpl1", CreatedBy: "u1"}, nil)
		mockPubed.EXPECT().GetByTplID(gomock.Any(), int64(1)).Return(&dapo.PublishedTplPo{ID: 100, TplID: 1}, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockPubed.EXPECT().DelCategoryAssocByTplID(gomock.Any(), tx, int64(1)).Return(nil)
		mockTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusUnpublished, int64(1), "", int64(0)).Return(nil)
		mockPubed.EXPECT().Delete(gomock.Any(), tx, int64(100)).Return(nil)

		_, err := svc.Unpublish(ctx, 1)
		assert.NoError(t, err)
	})
}

func TestDataAgentTplSvc_Copy_Update_Delete_UpdatePublishInfo(t *testing.T) {
	t.Run("copy success when biz domain disabled skips association", func(t *testing.T) {
		setTplDisableBizDomain(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo: mockTplRepo,
		}

		ctx := createTplCtxWithUserID("u1")
		sourcePo := &dapo.DataAgentTplPo{ID: 7, Name: "tpl_src", Key: "tpl_src_key", CreatedBy: "u1"}

		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(sourcePo, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), tx, gomock.Any()).Return(&dapo.DataAgentTplPo{ID: 88}, nil)

		resp, _, err := svc.Copy(ctx, 7)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(88), resp.ID)
	})

	t.Run("copy success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:           &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:      mockTplRepo,
			bdAgentTplRelRepo: mockBdRelRepo,
			bizDomainHttp:     mockBizDomain,
		}

		ctx := createTplCtxWithUserID("u1")
		ctx = context.WithValue(ctx, cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029
		sourcePo := &dapo.DataAgentTplPo{ID: 7, Name: "tpl_src", Key: "tpl_src_key", CreatedBy: "u1"}

		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(sourcePo, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), tx, gomock.Any()).Return(&dapo.DataAgentTplPo{ID: 88}, nil)
		mockBdRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockBizDomain.EXPECT().AssociateResource(gomock.Any(), gomock.Any()).Return(nil)

		resp, auditInfo, err := svc.Copy(ctx, 7)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(88), resp.ID)
		assert.Equal(t, "tpl_src_副本", resp.Name)
		assert.Equal(t, "7", auditInfo.ID)
	})

	t.Run("copy rollback when rel create failed", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:           &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:      mockTplRepo,
			bdAgentTplRelRepo: mockBdRelRepo,
		}

		ctx := createTplCtxWithUserID("u1")
		sourcePo := &dapo.DataAgentTplPo{ID: 7, Name: "tpl_src", Key: "tpl_src_key", CreatedBy: "u1"}

		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(7)).Return(sourcePo, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockTplRepo.EXPECT().GetByKeyWithTx(gomock.Any(), tx, gomock.Any()).Return(&dapo.DataAgentTplPo{ID: 88}, nil)
		mockBdRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(errors.New("rel insert failed"))

		resp, _, err := svc.Copy(ctx, 7)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("update success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo: mockTplRepo,
		}

		ctx := createTplCtxWithUserID("u1")
		req := &agenttplreq.UpdateReq{
			Name:   "updated_tpl",
			Config: &daconfvalobj.Config{},
		}
		oldPo := &dapo.DataAgentTplPo{ID: 9, Name: "old", CreatedBy: "u1", ProductKey: "chat"}

		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(9)).Return(oldPo, nil)
		mockTplRepo.EXPECT().ExistsByNameExcludeID(gomock.Any(), "updated_tpl", int64(9)).Return(false, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().Update(gomock.Any(), tx, gomock.Any()).Return(nil)

		auditInfo, err := svc.Update(ctx, req, 9)
		require.NoError(t, err)
		assert.Equal(t, "9", auditInfo.ID)
	})

	t.Run("delete success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockBdRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockBizDomain := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		builtInNo := cdaenum.BuiltInNo
		po := &dapo.DataAgentTplPo{ID: 10, Name: "tpl10", CreatedBy: "u1", Status: cdaenum.StatusUnpublished, IsBuiltIn: &builtInNo}

		svc := &dataAgentTplSvc{
			SvcBase:           &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:      mockTplRepo,
			publishedTplRepo:  mockPubedRepo,
			bdAgentTplRelRepo: mockBdRelRepo,
			bizDomainHttp:     mockBizDomain,
		}

		ctx := context.WithValue(context.Background(), cenum.BizDomainIDCtxKey.String(), "bd-1") //nolint:staticcheck // SA1029

		mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(10)).Return(true, nil)
		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(po, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(10)).Return(nil)
		mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(10)).Return(nil)
		mockBdRelRepo.EXPECT().DeleteByAgentTplID(gomock.Any(), tx, int64(10)).Return(nil)
		mockBizDomain.EXPECT().DisassociateResource(gomock.Any(), gomock.Any()).Return(nil)

		auditInfo, err := svc.Delete(ctx, 10, "u1", false)
		require.NoError(t, err)
		assert.Equal(t, "10", auditInfo.ID)
	})

	t.Run("delete success when biz domain disabled skips disassociation", func(t *testing.T) {
		setTplDisableBizDomain(t, true)

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		builtInNo := cdaenum.BuiltInNo
		po := &dapo.DataAgentTplPo{ID: 10, Name: "tpl10", CreatedBy: "u1", Status: cdaenum.StatusUnpublished, IsBuiltIn: &builtInNo}

		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
		}

		ctx := context.Background()

		mockTplRepo.EXPECT().ExistsByID(gomock.Any(), int64(10)).Return(true, nil)
		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(10)).Return(po, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().Delete(gomock.Any(), tx, int64(10)).Return(nil)
		mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(10)).Return(nil)

		auditInfo, err := svc.Delete(ctx, 10, "u1", false)
		require.NoError(t, err)
		assert.Equal(t, "10", auditInfo.ID)
	})

	t.Run("update publish info success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockCategoryRepo := idbaccessmock.NewMockICategoryRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)

		pubBy := "u1"
		pubAt := int64(123456)
		po := &dapo.DataAgentTplPo{ID: 11, Name: "tpl11", CreatedBy: "u1", PublishedBy: &pubBy, PublishedAt: &pubAt}
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
			categoryRepo:     mockCategoryRepo,
			pmsSvc:           mockPms,
			umHttp:           mockUm,
		}

		req := &agenttplreq.UpdatePublishInfoReq{CategoryIDs: []string{"cat1"}}
		ctx := createTplCtxWithUserID("u1")

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(11)).Return(po, nil)
		mockPubedRepo.EXPECT().GetByTplID(gomock.Any(), int64(11)).Return(&dapo.PublishedTplPo{ID: 101, TplID: 11}, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockCategoryRepo.EXPECT().GetIDNameMap(gomock.Any(), []string{"cat1"}).Return(map[string]string{"cat1": "分类1"}, nil)
		mockPubedRepo.EXPECT().DelCategoryAssocByTplID(gomock.Any(), tx, int64(101)).Return(nil)
		mockPubedRepo.EXPECT().BatchCreateCategoryAssoc(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockUm.EXPECT().GetSingleUserName(gomock.Any(), "u1").Return("U1", nil)

		resp, auditInfo, err := svc.UpdatePublishInfo(ctx, req, 11)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(11), resp.AgentTplId)
		assert.Equal(t, "11", auditInfo.ID)
	})
}

func TestDataAgentTplSvc_Publish_Unpublish_And_Detail_Branches(t *testing.T) {
	t.Run("publish internal tx success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		mockUm := httpaccmock.NewMockUmHttpAcc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
			pmsSvc:           mockPms,
			umHttp:           mockUm,
		}

		ctx := createTplCtxWithUserID("u1")
		req := agenttplreq.NewPublishReq()
		req.CategoryIDs = nil
		po := &dapo.DataAgentTplPo{ID: 21, Name: "tpl21", CreatedBy: "u1"}

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(21)).Return(po, nil)
		mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(21)).Return(nil)
		mockPubedRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(220), nil)
		mockTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, int64(21), "u1", gomock.Any()).Return(nil)
		mockUm.EXPECT().GetSingleUserName(gomock.Any(), "u1").Return("U1", nil)

		resp, _, err := svc.Publish(ctx, nil, req, 21, false)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, int64(220), resp.AgentTplId)
	})

	t.Run("publish not owner", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
			pmsSvc:           mockPms,
		}

		ctx := createTplCtxWithUserID("u1")
		req := agenttplreq.NewPublishReq()

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(21)).Return(&dapo.DataAgentTplPo{ID: 21, Name: "tpl21", CreatedBy: "u2"}, nil)

		resp, _, err := svc.Publish(ctx, nil, req, 21, false)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("publish update status error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newTplTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
			pmsSvc:           mockPms,
		}

		ctx := createTplCtxWithUserID("u1")
		req := agenttplreq.NewPublishReq()
		req.CategoryIDs = nil
		po := &dapo.DataAgentTplPo{ID: 21, Name: "tpl21", CreatedBy: "u1"}

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRepo.EXPECT().GetByIDWithTx(gomock.Any(), tx, int64(21)).Return(po, nil)
		mockPubedRepo.EXPECT().DeleteByTplID(gomock.Any(), tx, int64(21)).Return(nil)
		mockPubedRepo.EXPECT().Create(gomock.Any(), tx, gomock.Any()).Return(int64(220), nil)
		mockTplRepo.EXPECT().UpdateStatus(gomock.Any(), tx, cdaenum.StatusPublished, int64(21), "u1", gomock.Any()).Return(errors.New("update status failed"))

		resp, _, err := svc.Publish(ctx, nil, req, 21, false)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("unpublish published tpl not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
			pmsSvc:           mockPms,
		}

		ctx := createTplCtxWithUserID("u1")

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(30)).Return(&dapo.DataAgentTplPo{ID: 30, Name: "tpl30", CreatedBy: "u1"}, nil)
		mockPubedRepo.EXPECT().GetByTplID(gomock.Any(), int64(30)).Return(nil, sql.ErrNoRows)

		_, err := svc.Unpublish(ctx, 30)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "此已发布模板不存在")
	})

	t.Run("unpublish other owner without extra permission", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockPubedRepo := idbaccessmock.NewMockIPublishedTplRepo(ctrl)
		mockPms := v3portdrivermock.NewMockIPermissionSvc(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:          &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo:     mockTplRepo,
			publishedTplRepo: mockPubedRepo,
			pmsSvc:           mockPms,
		}

		ctx := createTplCtxWithUserID("u1")

		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(true, nil)
		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(31)).Return(&dapo.DataAgentTplPo{ID: 31, Name: "tpl31", CreatedBy: "u2"}, nil)
		mockPubedRepo.EXPECT().GetByTplID(gomock.Any(), int64(31)).Return(&dapo.PublishedTplPo{ID: 301, TplID: 31}, nil)
		mockPms.EXPECT().GetSingleMgmtPermission(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)

		_, err := svc.Unpublish(ctx, 31)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "无取消发布的权限")
	})

	t.Run("detail invalid config", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo: mockTplRepo,
			productRepo:  mockProductRepo,
		}

		mockTplRepo.EXPECT().GetByID(gomock.Any(), int64(50)).Return(&dapo.DataAgentTplPo{
			ID:     50,
			Name:   "tpl50",
			Config: "{bad-json",
		}, nil)

		resp, err := svc.Detail(context.Background(), 50)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "convert po to eo")
	})

	t.Run("detail by key product repo error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockProductRepo := idbaccessmock.NewMockIProductRepo(ctrl)
		svc := &dataAgentTplSvc{
			SvcBase:      &service.SvcBase{Logger: noopTplLogger{}},
			agentTplRepo: mockTplRepo,
			productRepo:  mockProductRepo,
		}

		mockTplRepo.EXPECT().GetByKey(gomock.Any(), "k60").Return(&dapo.DataAgentTplPo{
			ID:         60,
			Name:       "tpl60",
			ProductKey: "p60",
			Config:     "{}",
		}, nil)
		mockProductRepo.EXPECT().GetByKey(gomock.Any(), "p60").Return(nil, errors.New("product repo down"))

		resp, err := svc.DetailByKey(context.Background(), "k60")
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "convert po to eo")
	})
}
