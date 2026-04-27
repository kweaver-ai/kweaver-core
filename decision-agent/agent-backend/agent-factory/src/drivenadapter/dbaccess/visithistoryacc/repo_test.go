package visithistoryacc

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type visitTestLogger struct{}

func (visitTestLogger) Infof(string, ...interface{})  {}
func (visitTestLogger) Infoln(...interface{})         {}
func (visitTestLogger) Debugf(string, ...interface{}) {}
func (visitTestLogger) Debugln(...interface{})        {}
func (visitTestLogger) Errorf(string, ...interface{}) {}
func (visitTestLogger) Errorln(...interface{})        {}
func (visitTestLogger) Warnf(string, ...interface{})  {}
func (visitTestLogger) Warnln(...interface{})         {}
func (visitTestLogger) Panicf(string, ...interface{}) {}
func (visitTestLogger) Panicln(...interface{})        {}
func (visitTestLogger) Fatalf(string, ...interface{}) {}
func (visitTestLogger) Fatalln(...interface{})        {}

func newVisitRepoWithMock(t *testing.T) (*visitHistoryRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &visitHistoryRepo{
		db:     db,
		logger: visitTestLogger{},
	}

	return repo, db, mock
}

func visitPO() *dapo.VisitHistoryPO {
	return &dapo.VisitHistoryPO{
		AgentID:      "agent-1",
		AgentVersion: "latest",
		CreateBy:     "u1",
		UpdateTime:   123,
	}
}

func TestNewVisitHistoryRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := visitHistoryRepoOnce //nolint:govet
	oldImpl := visitHistoryRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		visitHistoryRepoOnce = oldOnce //nolint:govet
		visitHistoryRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	visitHistoryRepoOnce = sync.Once{}
	visitHistoryRepoImpl = nil

	r1 := NewVisitHistoryRepo()
	r2 := NewVisitHistoryRepo()

	require.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestVisitHistoryRepo_IncVisitCount(t *testing.T) {
	t.Parallel()

	t.Run("invalid input", func(t *testing.T) {
		t.Parallel()

		repo, db, _ := newVisitRepoWithMock(t)
		defer db.Close()

		err := repo.IncVisitCount(context.Background(), &dapo.VisitHistoryPO{})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "agentID or agentVersion or userID is empty")
	})

	t.Run("exists query error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newVisitRepoWithMock(t)
		defer db.Close()

		po := visitPO()
		mock.ExpectQuery(`select 1 from t_data_agent_visit_history where f_agent_id = \? and f_agent_version = \? and f_create_by = \? limit 1`).
			WithArgs(po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnError(errors.New("query failed"))

		err := repo.IncVisitCount(context.Background(), po)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "check agent id")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert when not exists success", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newVisitRepoWithMock(t)
		defer db.Close()

		po := visitPO()
		mock.ExpectQuery(`select 1 from t_data_agent_visit_history where f_agent_id = \? and f_agent_version = \? and f_create_by = \? limit 1`).
			WithArgs(po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnRows(sqlmock.NewRows([]string{"1"}))
		mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_visit_history")).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.IncVisitCount(context.Background(), po)
		require.NoError(t, err)
		assert.NotEmpty(t, po.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("insert when not exists error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newVisitRepoWithMock(t)
		defer db.Close()

		po := visitPO()
		mock.ExpectQuery(`select 1 from t_data_agent_visit_history where f_agent_id = \? and f_agent_version = \? and f_create_by = \? limit 1`).
			WithArgs(po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnRows(sqlmock.NewRows([]string{"1"}))
		mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_visit_history")).
			WillReturnError(errors.New("insert failed"))

		err := repo.IncVisitCount(context.Background(), po)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "insert agent id")
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update when exists", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newVisitRepoWithMock(t)
		defer db.Close()

		po := visitPO()
		mock.ExpectQuery(`select 1 from t_data_agent_visit_history where f_agent_id = \? and f_agent_version = \? and f_create_by = \? limit 1`).
			WithArgs(po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
		mock.ExpectExec(`(?i)update t_data_agent_visit_history set f_visit_count = f_visit_count \+ 1, f_update_time= \? where f_agent_id = \? and f_agent_version = \? and f_create_by = \?`).
			WithArgs(po.UpdateTime, po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.IncVisitCount(context.Background(), po)
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("update when exists error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newVisitRepoWithMock(t)
		defer db.Close()

		po := visitPO()
		mock.ExpectQuery(`select 1 from t_data_agent_visit_history where f_agent_id = \? and f_agent_version = \? and f_create_by = \? limit 1`).
			WithArgs(po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))
		mock.ExpectExec(`(?i)update t_data_agent_visit_history set f_visit_count = f_visit_count \+ 1, f_update_time= \? where f_agent_id = \? and f_agent_version = \? and f_create_by = \?`).
			WithArgs(po.UpdateTime, po.AgentID, po.AgentVersion, po.CreateBy).
			WillReturnError(errors.New("update failed"))

		err := repo.IncVisitCount(context.Background(), po)
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
