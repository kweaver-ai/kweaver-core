package personalspacedbacc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess/personalspacedbacc/psdbarg"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/personal_space/personalspacereq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
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

func newRepoWithMock(t *testing.T) (*personalSpaceRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	return &personalSpaceRepo{db: db, logger: testLogger{}, IDBAccBaseRepo: dbaccess.NewDBAccBase()}, db, mock
}

// ==================== Singleton ====================

func TestNewPersonalSpaceRepo_Singleton(t *testing.T) {
	old := pubedAgentRepoOnce //nolint:govet
	oldImpl := pubedAgentRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() { pubedAgentRepoOnce = old; pubedAgentRepoImpl = oldImpl; global.GDB = oldGDB }) //nolint:govet

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	pubedAgentRepoOnce = sync.Once{}
	pubedAgentRepoImpl = nil

	r1 := NewPersonalSpaceRepo()
	r2 := NewPersonalSpaceRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// ==================== ListPersonalSpaceAgent ====================

func TestListPersonalSpaceAgent_PanicEmptyCreatedBy(t *testing.T) {
	t.Parallel()

	repo, db, _ := newRepoWithMock(t)
	defer db.Close()

	assert.Panics(t, func() {
		_, _ = repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
			ListReq:   &personalspacereq.AgentListReq{Size: 10},
			CreatedBy: "",
		})
	})
}

func TestListPersonalSpaceAgent_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:   &personalspacereq.AgentListReq{Size: 10},
		CreatedBy: "u1",
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithNameFilter_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:   &personalspacereq.AgentListReq{Size: 10, Name: "test"},
		CreatedBy: "u1",
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithBuiltInPermission_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:                       &personalspacereq.AgentListReq{Size: 10},
		CreatedBy:                     "u1",
		HasBuiltInAgentMgmtPermission: true,
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithPublishStatus_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:   &personalspacereq.AgentListReq{Size: 10, PublishStatus: cdaenum.StatusThreeStateUnpublished},
		CreatedBy: "u1",
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithPublishStatusPublished_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:   &personalspacereq.AgentListReq{Size: 10, PublishStatus: cdaenum.StatusThreeStatePublished},
		CreatedBy: "u1",
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithPublishStatusEdited_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:   &personalspacereq.AgentListReq{Size: 10, PublishStatus: cdaenum.StatusThreeStatePublishedEdited},
		CreatedBy: "u1",
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithPublishToBe_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:   &personalspacereq.AgentListReq{Size: 10, PublishToBe: cdaenum.PublishToBeAPIAgent},
		CreatedBy: "u1",
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceAgent_WithBizDomainIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceAgent(context.Background(), &psdbarg.AgentListArg{
		ListReq:             &personalspacereq.AgentListReq{Size: 10},
		CreatedBy:           "u1",
		AgentIDsByBizDomain: []string{"agent-1"},
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceTpl_NilTplIDsDoesNotPanic(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceTpl(context.Background(), &psdbarg.TplListArg{
		ListReq:    &personalspacereq.AgentTplListReq{Size: 10},
		CreatedBy:  "u1",
		TplIDsByBd: nil,
	})
	assert.Error(t, err)
}

// ==================== ListPersonalSpaceTpl ====================

func TestListPersonalSpaceTpl_EmptyTplIDsDoesNotPanic(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceTpl(context.Background(), &psdbarg.TplListArg{
		ListReq:    &personalspacereq.AgentTplListReq{Size: 10},
		TplIDsByBd: []string{},
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceTpl_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceTpl(context.Background(), &psdbarg.TplListArg{
		ListReq:    &personalspacereq.AgentTplListReq{Size: 10},
		CreatedBy:  "u1",
		TplIDsByBd: []string{"tpl-1"},
	})
	assert.Error(t, err)
}

func TestListPersonalSpaceTpl_WithNameFilter_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).WillReturnError(errors.New("find err"))

	_, err := repo.ListPersonalSpaceTpl(context.Background(), &psdbarg.TplListArg{
		ListReq:    &personalspacereq.AgentTplListReq{Size: 10, Name: "test"},
		CreatedBy:  "u1",
		TplIDsByBd: []string{"tpl-1"},
	})
	assert.Error(t, err)
}
