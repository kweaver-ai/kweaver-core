package pubedagentdbacc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/pubedagentdbacc/padbarg"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/global"
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

func newRepoWithMock(t *testing.T) (*pubedAgentRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	return &pubedAgentRepo{db: db, logger: testLogger{}, IDBAccBaseRepo: dbaccess.NewDBAccBase()}, db, mock
}

// ==================== Singleton ====================

func TestNewPubedAgentRepo_Singleton(t *testing.T) {
	old := pubedAgentRepoOnce //nolint:govet
	oldImpl := pubedAgentRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() { pubedAgentRepoOnce = old; pubedAgentRepoImpl = oldImpl; global.GDB = oldGDB }) //nolint:govet

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	pubedAgentRepoOnce = sync.Once{}
	pubedAgentRepoImpl = nil

	r1 := NewPubedAgentRepo()
	r2 := NewPubedAgentRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// ==================== GetPubedList ====================

func TestGetPubedList_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithNameFilter_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, Name: "test"}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithCategoryFilter_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, CategoryID: "cat-1"}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithToBeFlag_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, ToBeFlag: cdaenum.PublishToBeAPIAgent}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithToBeWebSDK_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, ToBeFlag: cdaenum.PublishToBeWebSDKAgent}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithToBeSkill_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, ToBeFlag: cdaenum.PublishToBeSkillAgent}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithToBeDataFlow_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, ToBeFlag: cdaenum.PublishToBeDataFlowAgent}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithIsToCustomSpace_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, IsToCustomSpace: 1}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithIsToSquare_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, IsToSquare: 1}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, IDs: []string{"id-1"}}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithAgentKeys_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, AgentKeys: []string{"key-1"}}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubedList_WithExcludeAgentKeys_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	req := &pubedreq.PubedAgentListReq{Size: 10, ExcludeAgentKeys: []string{"key-1"}}

	_, err := repo.GetPubedList(context.Background(), req)
	assert.Error(t, err)
}

// ==================== GetPubedListByXx ====================

func TestGetPubedListByXx_EmptyKeysAndIDs(t *testing.T) {
	t.Parallel()

	repo, db, _ := newRepoWithMock(t)
	defer db.Close()

	ret, err := repo.GetPubedListByXx(context.Background(), &padbarg.GetPaPoListByXxArg{})
	assert.NoError(t, err)
	assert.NotNil(t, ret)
	assert.Empty(t, ret.JoinPos)
}

func TestGetPubedListByXx_WithKeys_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	ret, err := repo.GetPubedListByXx(context.Background(), &padbarg.GetPaPoListByXxArg{
		AgentKeys: []string{"key-1"},
	})
	assert.Error(t, err)
	assert.NotNil(t, ret)
}

func TestGetPubedListByXx_WithIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	ret, err := repo.GetPubedListByXx(context.Background(), &padbarg.GetPaPoListByXxArg{
		AgentIDs: []string{"id-1"},
	})
	assert.Error(t, err)
	assert.NotNil(t, ret)
}

func TestGetPubedListByXx_WithPubToCond_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	ret, err := repo.GetPubedListByXx(context.Background(), &padbarg.GetPaPoListByXxArg{
		AgentKeys: []string{"key-1"},
		PubToWhereCond: &padbarg.PublishedToWhereCondition{
			IsToCustomSpace: true,
			IsToSquare:      true,
		},
	})
	assert.Error(t, err)
	assert.NotNil(t, ret)
}

// ==================== GetPubedPoMapByXx ====================

func TestGetPubedPoMapByXx_EmptyKeysAndIDs(t *testing.T) {
	t.Parallel()

	repo, db, _ := newRepoWithMock(t)
	defer db.Close()

	ret, err := repo.GetPubedPoMapByXx(context.Background(), &padbarg.GetPaPoListByXxArg{})
	assert.NoError(t, err)
	assert.NotNil(t, ret)
}

func TestGetPubedPoMapByXx_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	ret, err := repo.GetPubedPoMapByXx(context.Background(), &padbarg.GetPaPoListByXxArg{
		AgentKeys: []string{"key-1"},
	})
	assert.Error(t, err)
	assert.NotNil(t, ret)
}
