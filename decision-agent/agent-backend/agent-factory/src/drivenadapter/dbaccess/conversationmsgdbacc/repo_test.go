package conversationmsgdbacc

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
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation_message/conversationmsgreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type msgTestLogger struct{}

func (msgTestLogger) Infof(string, ...interface{})  {}
func (msgTestLogger) Infoln(...interface{})         {}
func (msgTestLogger) Debugf(string, ...interface{}) {}
func (msgTestLogger) Debugln(...interface{})        {}
func (msgTestLogger) Errorf(string, ...interface{}) {}
func (msgTestLogger) Errorln(...interface{})        {}
func (msgTestLogger) Warnf(string, ...interface{})  {}
func (msgTestLogger) Warnln(...interface{})         {}
func (msgTestLogger) Panicf(string, ...interface{}) {}
func (msgTestLogger) Panicln(...interface{})        {}
func (msgTestLogger) Fatalf(string, ...interface{}) {}
func (msgTestLogger) Fatalln(...interface{})        {}

func newMsgRepoWithMock(t *testing.T) (*ConversationMsgRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &ConversationMsgRepo{
		db:             db,
		logger:         msgTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func mockConversationMsgRows() *sqlmock.Rows {
	content := "hello"

	return sqlmock.NewRows([]string{
		"f_id",
		"f_agent_app_key",
		"f_conversation_id",
		"f_agent_id",
		"f_agent_version",
		"f_reply_id",
		"f_index",
		"f_role",
		"f_content",
		"f_content_type",
		"f_status",
		"f_ext",
		"f_create_time",
		"f_update_time",
		"f_create_by",
		"f_update_by",
		"f_is_deleted",
	}).AddRow(
		"msg-1",
		"app-1",
		"conv-1",
		"agent-1",
		"v1",
		"",
		1,
		"",
		content,
		"",
		"",
		nil,
		int64(1),
		int64(1),
		"u1",
		"u1",
		0,
	)
}

func TestNewConversationMsgRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := conversationMsgRepoOnce //nolint:govet
	oldImpl := conversationMsgRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		conversationMsgRepoOnce = oldOnce //nolint:govet
		conversationMsgRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	conversationMsgRepoOnce = sync.Once{}
	conversationMsgRepoImpl = nil

	r1 := NewConversationMsgRepo()
	r2 := NewConversationMsgRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestConversationMsgRepo_Create(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_conversation_message")).
		WillReturnResult(sqlmock.NewResult(1, 1))

	id, err := repo.Create(context.Background(), &dapo.ConversationMsgPO{ConversationID: "conv-1", AgentAPPKey: "app-1", CreateBy: "u1", UpdateBy: "u1"})
	require.NoError(t, err)
	assert.NotEmpty(t, id)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_Create_InsertError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_conversation_message")).
		WillReturnError(errors.New("insert failed"))

	_, err := repo.Create(context.Background(), &dapo.ConversationMsgPO{ConversationID: "conv-1", AgentAPPKey: "app-1"})
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_GetByID(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_id = \?`).
		WithArgs("msg-1").
		WillReturnRows(mockConversationMsgRows())

	po, err := repo.GetByID(context.Background(), "msg-1")
	require.NoError(t, err)
	assert.Equal(t, "msg-1", po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_GetByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_id = \?`).
		WithArgs("msg-1").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByID(context.Background(), "msg-1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_GetMaxIndexByID(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_conversation_id = \? order by f_index DESC limit 1`).
		WithArgs("conv-1").
		WillReturnRows(mockConversationMsgRows())

	idx, err := repo.GetMaxIndexByID(context.Background(), "conv-1")
	require.NoError(t, err)
	assert.Equal(t, 1, idx)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_GetMaxIndexByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_conversation_id = \? order by f_index DESC limit 1`).
		WithArgs("conv-1").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetMaxIndexByID(context.Background(), "conv-1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get max index by id")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_GetLatestMsgByConversationID(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_conversation_id = \? order by f_index DESC limit 1`).
		WithArgs("conv-1").
		WillReturnRows(mockConversationMsgRows())

	po, err := repo.GetLatestMsgByConversationID(context.Background(), "conv-1")
	require.NoError(t, err)
	assert.Equal(t, "msg-1", po.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_GetLatestMsgByConversationID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_conversation_id = \? order by f_index DESC limit 1`).
		WithArgs("conv-1").
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetLatestMsgByConversationID(context.Background(), "conv-1")
	assert.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_List(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_conversation_id = \? and f_is_deleted = \? order by f_index ASC`).
		WithArgs("conv-1", 0).
		WillReturnRows(mockConversationMsgRows())

	list, err := repo.List(context.Background(), conversationmsgreq.ListReq{ConversationID: "conv-1"})
	require.NoError(t, err)
	assert.Len(t, list, 1)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_List_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`select .* from t_data_agent_conversation_message where f_conversation_id = \? and f_is_deleted = \? order by f_index ASC`).
		WithArgs("conv-1", 0).
		WillReturnError(errors.New("query failed"))

	_, err := repo.List(context.Background(), conversationmsgreq.ListReq{ConversationID: "conv-1"})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get conversation message list")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_Update(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`update t_data_agent_conversation_message set .* where f_id = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), &dapo.ConversationMsgPO{ID: "msg-1", ConversationID: "conv-1", UpdateBy: "u2", UpdateTime: 2})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_Delete(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from t_data_agent_conversation_message where f_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(context.Background(), "msg-1")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_DeleteByConversationID_And_APPKey(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation_message set f_is_deleted = ? where f_conversation_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation_message set f_is_deleted = ? where f_agent_app_key = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DeleteByConversationID(context.Background(), nil, "conv-1")
	require.NoError(t, err)
	err = repo.DeleteByAPPKey(context.Background(), nil, "app-1")
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationMsgRepo_DeleteByConversationID_And_APPKey_WithTxBranch(t *testing.T) {
	t.Parallel()

	repo, db, mock := newMsgRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation_message set f_is_deleted = ? where f_conversation_id = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(regexp.QuoteMeta("update t_data_agent_conversation_message set f_is_deleted = ? where f_agent_app_key = ?")).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	err = repo.DeleteByConversationID(context.Background(), tx, "conv-1")
	require.NoError(t, err)
	err = repo.DeleteByAPPKey(context.Background(), tx, "app-1")
	require.NoError(t, err)
	require.NoError(t, tx.Rollback())
	require.NoError(t, mock.ExpectationsWereMet())
}
