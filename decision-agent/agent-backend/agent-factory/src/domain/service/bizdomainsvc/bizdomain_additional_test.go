package bizdomainsvc

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/ihttpaccess/ibizdomainacc/bizdomainaccmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type noopLogger struct{}

func (noopLogger) Infof(string, ...interface{})  {}
func (noopLogger) Infoln(...interface{})         {}
func (noopLogger) Debugf(string, ...interface{}) {}
func (noopLogger) Debugln(...interface{})        {}
func (noopLogger) Errorf(string, ...interface{}) {}
func (noopLogger) Errorln(...interface{})        {}
func (noopLogger) Warnf(string, ...interface{})  {}
func (noopLogger) Warnln(...interface{})         {}
func (noopLogger) Panicf(string, ...interface{}) {}
func (noopLogger) Panicln(...interface{})        {}
func (noopLogger) Fatalf(string, ...interface{}) {}
func (noopLogger) Fatalln(...interface{})        {}

func newSQLTx(t *testing.T) (*sql.Tx, sqlmock.Sqlmock, func()) {
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

func TestBizDomainSvc_InitBizDomainAgentRel_Additional(t *testing.T) {
	t.Run("existing relations skip with rollback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)

		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{{AgentID: "a1"}}, nil)

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.NoError(t, err)
	})

	t.Run("get existing rels error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return(nil, errors.New("query rel failed"))

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get existing agent rels failed")
	})

	t.Run("get all agent ids error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{}, nil)
		mockAgentRepo.EXPECT().GetAllIDs(gomock.Any()).Return(nil, errors.New("query agent ids failed"))

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get all agent ids failed")
	})

	t.Run("empty agents skip with rollback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{}, nil)
		mockAgentRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]string{}, nil)

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.NoError(t, err)
	})

	t.Run("batch create error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{}, nil)
		mockAgentRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]string{"a1"}, nil)
		mockRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).
			Return(errors.New("batch insert failed"))

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "batch create agent rels failed")
	})

	t.Run("associate batch error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{}, nil)
		mockAgentRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]string{"a1"}, nil)
		mockRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).
			Return(errors.New("http failed"))

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "associate resource batch failed")
	})

	t.Run("commit error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectCommit().WillReturnError(errors.New("commit failed"))

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{}, nil)
		mockAgentRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]string{"a1"}, nil)
		mockRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "commit tx failed")
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentRepo := idbaccessmock.NewMockIDataAgentConfigRepo(ctrl)
		mockRelRepo := idbaccessmock.NewMockIBizDomainAgentRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentRelPo{}, nil)
		mockAgentRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]string{"a1", "a2"}, nil)
		mockRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.InitBizDomainAgentRel(context.Background(), mockAgentRepo, mockRelRepo)
		assert.NoError(t, err)
	})
}

func TestBizDomainSvc_InitBizDomainAgentTplRel_Additional(t *testing.T) {
	t.Run("existing relations skip with rollback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{{AgentTplID: 1}}, nil)

		err := svc.InitBizDomainAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.NoError(t, err)
	})

	t.Run("get all tpl ids error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{}, nil)
		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return(nil, errors.New("query tpl ids failed"))

		err := svc.InitBizDomainAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "get all agent tpl ids failed")
	})

	t.Run("associate batch error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{}, nil)
		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).
			Return(errors.New("http failed"))

		err := svc.InitBizDomainAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "associate resource batch failed")
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{}, nil)
		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(nil)

		err := svc.InitBizDomainAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.NoError(t, err)
	})
}

func TestBizDomainSvc_FixMissingAgentTplRel_Additional(t *testing.T) {
	t.Run("get existing rels error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return(nil, errors.New("query existing rel failed"))

		resp, err := svc.FixMissingAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("no missing ids skip with rollback", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{{AgentTplID: 1}, {AgentTplID: 2}}, nil)

		resp, err := svc.FixMissingAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 0, resp.FixedCount)
	})

	t.Run("batch create error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{{AgentTplID: 1}}, nil)
		mockTplRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).
			Return(errors.New("batch failed"))

		resp, err := svc.FixMissingAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "batch create agent tpl rels failed")
	})

	t.Run("associate batch error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectRollback()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{{AgentTplID: 1}}, nil)
		mockTplRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).
			Return(errors.New("http failed"))

		resp, err := svc.FixMissingAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Contains(t, err.Error(), "associate resource batch failed")
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tx, sqlMock, done := newSQLTx(t)
		defer done()
		sqlMock.ExpectCommit()

		mockAgentTplRepo := idbaccessmock.NewMockIDataAgentTplRepo(ctrl)
		mockTplRelRepo := idbaccessmock.NewMockIBizDomainAgentTplRelRepo(ctrl)
		mockHTTP := bizdomainaccmock.NewMockBizDomainHttpAcc(ctrl)
		svc := &BizDomainSvc{logger: noopLogger{}, bizDomainHttp: mockHTTP}

		mockAgentTplRepo.EXPECT().GetAllIDs(gomock.Any()).Return([]int64{1, 2}, nil)
		mockTplRelRepo.EXPECT().BeginTx(gomock.Any()).Return(tx, nil)
		mockTplRelRepo.EXPECT().GetByBizDomainID(gomock.Any(), tx, gomock.Any()).
			Return([]*dapo.BizDomainAgentTplRelPo{{AgentTplID: 1}}, nil)
		mockTplRelRepo.EXPECT().BatchCreate(gomock.Any(), tx, gomock.Any()).Return(nil)
		mockHTTP.EXPECT().AssociateResourceBatch(gomock.Any(), gomock.Any()).Return(nil)

		resp, err := svc.FixMissingAgentTplRel(context.Background(), mockAgentTplRepo, mockTplRelRepo)
		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, 1, resp.FixedCount)
		assert.Equal(t, []int64{2}, resp.FixedIDs)
	})
}
