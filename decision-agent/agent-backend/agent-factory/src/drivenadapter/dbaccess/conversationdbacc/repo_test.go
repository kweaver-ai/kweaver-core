package conversationdbacc

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/common"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type dbTestLogger struct{}

func (dbTestLogger) Infof(string, ...interface{})  {}
func (dbTestLogger) Infoln(...interface{})         {}
func (dbTestLogger) Debugf(string, ...interface{}) {}
func (dbTestLogger) Debugln(...interface{})        {}
func (dbTestLogger) Errorf(string, ...interface{}) {}
func (dbTestLogger) Errorln(...interface{})        {}
func (dbTestLogger) Warnf(string, ...interface{})  {}
func (dbTestLogger) Warnln(...interface{})         {}
func (dbTestLogger) Panicf(string, ...interface{}) {}
func (dbTestLogger) Panicln(...interface{})        {}
func (dbTestLogger) Fatalf(string, ...interface{}) {}
func (dbTestLogger) Fatalln(...interface{})        {}

func newRepoWithMock(t *testing.T) (*ConversationRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &ConversationRepo{
		db:             db,
		logger:         dbTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func mockConversationRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id",
		"f_agent_app_key",
		"f_title",
		"f_origin",
		"f_message_index",
		"f_read_message_index",
		"f_ext",
		"f_create_time",
		"f_update_time",
		"f_create_by",
		"f_update_by",
		"f_is_deleted",
	}).AddRow(
		"conv-1",
		"app-1",
		"title",
		"",
		1,
		0,
		nil,
		int64(1),
		int64(1),
		"u1",
		"u1",
		0,
	)
}

func TestNewConversationRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := conversationRepoOnce //nolint:govet
	oldImpl := conversationRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		conversationRepoOnce = oldOnce //nolint:govet
		conversationRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	conversationRepoOnce = sync.Once{}
	conversationRepoImpl = nil

	r1 := NewConversationRepo()
	r2 := NewConversationRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestConversationRepo_Create(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_conversation")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	po := &dapo.ConversationPO{AgentAPPKey: "app-1", Title: "t1", CreateBy: "u1", UpdateBy: "u1"}
	rt, err := repo.Create(context.Background(), po)
	require.NoError(t, err)
	assert.NotNil(t, rt)
	assert.NotEmpty(t, rt.ID)
	assert.NotZero(t, rt.CreateTime)
	assert.Equal(t, rt.CreateTime, rt.UpdateTime)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_Create_InsertError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_conversation")).
		WillReturnError(errors.New("insert failed"))

	po := &dapo.ConversationPO{AgentAPPKey: "app-1", Title: "t1", CreateBy: "u1", UpdateBy: "u1"}
	_, err := repo.Create(context.Background(), po)
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_GetByID(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation where f_id = \? and f_is_deleted = \?`).
		WithArgs("conv-1", 0).
		WillReturnRows(mockConversationRows())

	po, err := repo.GetByID(context.Background(), "conv-1")
	require.NoError(t, err)
	assert.Equal(t, "conv-1", po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_GetByID_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation where f_id = \? and f_is_deleted = \?`).
		WithArgs("conv-1", 0).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByID(context.Background(), "conv-1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_Update(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`update t_data_agent_conversation set .* where f_id = \? and f_is_deleted = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), &dapo.ConversationPO{ID: "conv-1", Title: "t2", MessageIndex: 2, ReadMessageIndex: 1, UpdateTime: 2, UpdateBy: "u2"})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_Delete_And_DeleteByAPPKey(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation set f_is_deleted = ? where f_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation set f_is_deleted = ? where f_agent_app_key = ?")).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err := repo.Delete(context.Background(), nil, "conv-1")
	require.NoError(t, err)
	err = repo.DeleteByAPPKey(context.Background(), nil, "app-1")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_Delete_WithTx(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation set f_is_deleted = ? where f_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation set f_is_deleted = ? where f_agent_app_key = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	err = repo.Delete(context.Background(), tx, "conv-1")
	require.NoError(t, err)
	err = repo.DeleteByAPPKey(context.Background(), tx, "app-1")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_List(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation where f_agent_app_key = \? and f_create_by = \? and f_is_deleted = \? order by f_update_time DESC limit 10 offset 0`).
		WithArgs("app-1", "u1", 0).
		WillReturnRows(mockConversationRows())
	mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_conversation where f_agent_app_key = ? and f_create_by = ? and f_is_deleted = ?")).
		WithArgs("app-1", "u1", 0).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rt, count, err := repo.List(context.Background(), conversationreq.ListReq{AgentAPPKey: "app-1", UserId: "u1", PageSize: common.PageSize{Page: 1, Size: 10}})
	require.NoError(t, err)
	assert.Len(t, rt, 1)
	assert.Equal(t, int64(1), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_List_FindOrCountError(t *testing.T) {
	t.Parallel()

	t.Run("find error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_conversation where f_agent_app_key = \? and f_create_by = \? and f_is_deleted = \? order by f_update_time DESC limit 10 offset 0`).
			WithArgs("app-1", "u1", 0).
			WillReturnError(errors.New("find failed"))

		_, _, err := repo.List(context.Background(), conversationreq.ListReq{AgentAPPKey: "app-1", UserId: "u1", PageSize: common.PageSize{Page: 1, Size: 10}})
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_conversation where f_agent_app_key = \? and f_create_by = \? and f_is_deleted = \? order by f_update_time DESC limit 10 offset 0`).
			WithArgs("app-1", "u1", 0).
			WillReturnRows(mockConversationRows())
		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_conversation where f_agent_app_key = ? and f_create_by = ? and f_is_deleted = ?")).
			WithArgs("app-1", "u1", 0).
			WillReturnError(errors.New("count failed"))

		_, _, err := repo.List(context.Background(), conversationreq.ListReq{AgentAPPKey: "app-1", UserId: "u1", PageSize: common.PageSize{Page: 1, Size: 10}})
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestConversationRepo_ListByAgentID(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation where f_agent_app_key = \? and f_is_deleted = \? and f_message_index <> \? order by f_update_time DESC limit 10 offset 0`).
		WithArgs("agent-1", 0, 0).
		WillReturnRows(mockConversationRows())
	mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_conversation where f_agent_app_key = ? and f_is_deleted = ? and f_message_index <> ?")).
		WithArgs("agent-1", 0, 0).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	rt, count, err := repo.ListByAgentID(context.Background(), "agent-1", "", 1, 10)
	require.NoError(t, err)
	assert.Len(t, rt, 1)
	assert.Equal(t, int64(1), count)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationRepo_ListByAgentID_Errors(t *testing.T) {
	t.Parallel()

	t.Run("find error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_conversation where f_agent_app_key = \? and f_is_deleted = \? and f_message_index <> \? order by f_update_time DESC limit 10 offset 0`).
			WithArgs("agent-1", 0, 0).
			WillReturnError(errors.New("find failed"))

		_, _, err := repo.ListByAgentID(context.Background(), "agent-1", "", 1, 10)
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newRepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_conversation where f_agent_app_key = \? and f_is_deleted = \? and f_message_index <> \? order by f_update_time DESC limit 10 offset 0`).
			WithArgs("agent-1", 0, 0).
			WillReturnRows(mockConversationRows())
		mock.ExpectQuery(regexp.QuoteMeta("select count(*) from t_data_agent_conversation where f_agent_app_key = ? and f_is_deleted = ? and f_message_index <> ?")).
			WithArgs("agent-1", 0, 0).
			WillReturnError(errors.New("count failed"))

		_, _, err := repo.ListByAgentID(context.Background(), "agent-1", "", 1, 10)
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
