package conversationhistoryacc

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type historyTestLogger struct{}

func (historyTestLogger) Infof(string, ...interface{})  {}
func (historyTestLogger) Infoln(...interface{})         {}
func (historyTestLogger) Debugf(string, ...interface{}) {}
func (historyTestLogger) Debugln(...interface{})        {}
func (historyTestLogger) Errorf(string, ...interface{}) {}
func (historyTestLogger) Errorln(...interface{})        {}
func (historyTestLogger) Warnf(string, ...interface{})  {}
func (historyTestLogger) Warnln(...interface{})         {}
func (historyTestLogger) Panicf(string, ...interface{}) {}
func (historyTestLogger) Panicln(...interface{})        {}
func (historyTestLogger) Fatalf(string, ...interface{}) {}
func (historyTestLogger) Fatalln(...interface{})        {}

func newHistoryRepoWithMock(t *testing.T) (*conversationHistoryRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &conversationHistoryRepo{
		RepoBase: drivenadapter.NewRepoBase(),
		db:       db,
		logger:   historyTestLogger{},
	}

	return repo, db, mock
}

func TestNewConversationHistoryRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := conversationHistoryRepoOnce //nolint:govet
	oldImpl := conversationHistoryRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		conversationHistoryRepoOnce = oldOnce //nolint:govet
		conversationHistoryRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	conversationHistoryRepoOnce = sync.Once{}
	conversationHistoryRepoImpl = nil

	r1 := NewConversationHistoryRepo()
	r2 := NewConversationHistoryRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestConversationHistoryRepo_GetLatestVisitAgentIds(t *testing.T) {
	t.Parallel()

	repo, db, mock := newHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT f_bot_id, MAX(f_modified_at) as last_modified_at FROM tb_conversation_history_v2 WHERE f_user_id=? AND f_deleted=0 GROUP BY f_bot_id ORDER BY last_modified_at DESC")).
		WithArgs("u1").
		WillReturnRows(sqlmock.NewRows([]string{"f_bot_id", "last_modified_at"}).AddRow("agent-1", int64(100)).AddRow("agent-2", int64(99)))

	rt, err := repo.GetLatestVisitAgentIds(context.Background(), "u1")
	require.NoError(t, err)
	require.Len(t, rt, 2)
	assert.Equal(t, "agent-1", rt[0].AgentId)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestConversationHistoryRepo_GetLatestVisitAgentIds_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT f_bot_id, MAX(f_modified_at) as last_modified_at FROM tb_conversation_history_v2 WHERE f_user_id=? AND f_deleted=0 GROUP BY f_bot_id ORDER BY last_modified_at DESC")).
		WithArgs("u1").
		WillReturnError(errors.New("query failed"))

	rt, err := repo.GetLatestVisitAgentIds(context.Background(), "u1")
	assert.Nil(t, rt)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get latest visit agent")
	require.NoError(t, mock.ExpectationsWereMet())
}
