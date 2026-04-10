package bdagentdbacc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
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

func newRepoWithMock(t *testing.T) (*BizDomainAgentRelRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	return &BizDomainAgentRelRepo{db: db, logger: testLogger{}, IDBAccBaseRepo: dbaccess.NewDBAccBase()}, db, mock
}

// ==================== Singleton ====================

func TestNewBizDomainAgentRelRepo_Singleton(t *testing.T) {
	old := bizDomainAgentRelRepoOnce //nolint:govet
	oldImpl := bizDomainAgentRelRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() { bizDomainAgentRelRepoOnce = old; bizDomainAgentRelRepoImpl = oldImpl; global.GDB = oldGDB }) //nolint:govet

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	bizDomainAgentRelRepoOnce = sync.Once{}
	bizDomainAgentRelRepoImpl = nil

	r1 := NewBizDomainAgentRelRepo()
	r2 := NewBizDomainAgentRelRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// ==================== BatchCreate ====================

func TestBatchCreate_Empty(t *testing.T) {
	t.Parallel()

	repo, db, _ := newRepoWithMock(t)
	defer db.Close()
	assert.NoError(t, repo.BatchCreate(context.Background(), nil, []*dapo.BizDomainAgentRelPo{}))
}

func TestBatchCreate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_biz_domain_agent_rel`).WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.BatchCreate(context.Background(), nil, []*dapo.BizDomainAgentRelPo{
		{BizDomainID: "bd-1", AgentID: "a-1"},
	})
	assert.NoError(t, err)
}

func TestBatchCreate_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_biz_domain_agent_rel`).WillReturnError(errors.New("insert err"))

	err := repo.BatchCreate(context.Background(), nil, []*dapo.BizDomainAgentRelPo{
		{BizDomainID: "bd-1", AgentID: "a-1"},
	})
	assert.Error(t, err)
}

// ==================== DeleteByAgentID ====================

func TestDeleteByAgentID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_rel`).WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteByAgentID(context.Background(), nil, "a-1"))
}

func TestDeleteByAgentID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_rel`).WillReturnError(errors.New("del err"))
	assert.Error(t, repo.DeleteByAgentID(context.Background(), nil, "a-1"))
}

// ==================== DeleteByBizDomainID ====================

func TestDeleteByBizDomainID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_rel`).WillReturnResult(sqlmock.NewResult(0, 1))
	assert.NoError(t, repo.DeleteByBizDomainID(context.Background(), nil, "bd-1"))
}

func TestDeleteByBizDomainID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)delete from t_biz_domain_agent_rel`).WillReturnError(errors.New("del err"))
	assert.Error(t, repo.DeleteByBizDomainID(context.Background(), nil, "bd-1"))
}

// ==================== GetByAgentID ====================

func TestGetByAgentID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_biz_domain_agent_rel`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByAgentID(context.Background(), nil, "a-1")
	assert.Error(t, err)
}

// ==================== GetByBizDomainID ====================

func TestGetByBizDomainID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_biz_domain_agent_rel`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByBizDomainID(context.Background(), nil, "bd-1")
	assert.Error(t, err)
}
