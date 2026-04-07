package bdagenttpldbacc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type testLogger struct{}

func (testLogger) Infof(string, ...interface{})  {}
func (testLogger) Infoln(...interface{})         {}
func (testLogger) Debugf(string, ...interface{}) {}
func (testLogger) Debugln(...interface{})        {}
func (testLogger) Errorf(string, ...interface{}) {}
func (testLogger) Errorln(...interface{})        {}
func (testLogger) Warnf(string, ...interface{})  {}
func (testLogger) Warnln(...interface{})         {}
func (testLogger) Panicf(string, ...interface{}) {}
func (testLogger) Panicln(...interface{})        {}
func (testLogger) Fatalf(string, ...interface{}) {}
func (testLogger) Fatalln(...interface{})        {}

func newRepoWithMock(t *testing.T) (*BizDomainAgentTplRelRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	return &BizDomainAgentTplRelRepo{db: db, logger: testLogger{}, IDBAccBaseRepo: dbaccess.NewDBAccBase()}, db, mock
}

// ==================== Singleton ====================

func TestNewBizDomainAgentTplRelRepo_Singleton(t *testing.T) {
	oldImpl := bizDomainAgentTplRelRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		bizDomainAgentTplRelRepoOnce = sync.Once{}
		if oldImpl != nil {
			bizDomainAgentTplRelRepoOnce.Do(func() {})
		}

		bizDomainAgentTplRelRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	bizDomainAgentTplRelRepoOnce = sync.Once{}
	bizDomainAgentTplRelRepoImpl = nil

	r1 := NewBizDomainAgentTplRelRepo()
	r2 := NewBizDomainAgentTplRelRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// ==================== BatchCreate ====================

func TestBatchCreate_Empty(t *testing.T) {
	t.Parallel()

	repo, db, _ := newRepoWithMock(t)
	defer db.Close()
	assert.NoError(t, repo.BatchCreate(context.Background(), nil, []*dapo.BizDomainAgentTplRelPo{}))
}

func TestBatchCreate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_biz_domain_agent_tpl_rel`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.BatchCreate(context.Background(), nil, []*dapo.BizDomainAgentTplRelPo{
		{BizDomainID: "bd-1", AgentTplID: 1},
	})
	assert.NoError(t, err)
}

func TestBatchCreate_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_biz_domain_agent_tpl_rel`).WillReturnError(errors.New("insert err"))

	err := repo.BatchCreate(context.Background(), nil, []*dapo.BizDomainAgentTplRelPo{
		{BizDomainID: "bd-1", AgentTplID: 1},
	})
	assert.Error(t, err)
}

// ==================== DeleteByAgentTplID ====================

func TestDeleteByAgentTplID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_tpl_rel`).WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteByAgentTplID(context.Background(), nil, 1))
}

func TestDeleteByAgentTplID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_tpl_rel`).WillReturnError(errors.New("del err"))
	assert.Error(t, repo.DeleteByAgentTplID(context.Background(), nil, 1))
}

// ==================== DeleteByBizDomainID ====================

func TestDeleteByBizDomainID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_tpl_rel`).WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteByBizDomainID(context.Background(), nil, "bd-1"))
}

func TestDeleteByBizDomainID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_tpl_rel`).WillReturnError(errors.New("del err"))
	assert.Error(t, repo.DeleteByBizDomainID(context.Background(), nil, "bd-1"))
}

// ==================== GetByAgentTplID ====================

func TestGetByAgentTplID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_biz_domain_agent_tpl_rel`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByAgentTplID(context.Background(), nil, 1)
	assert.Error(t, err)
}

// ==================== GetByBizDomainID ====================

func TestGetByBizDomainID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_biz_domain_agent_tpl_rel`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByBizDomainID(context.Background(), nil, "bd-1")
	assert.Error(t, err)
}
