package daconfdbacc

import (
	"context"
	"errors"
	"regexp"
	"sync"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/dbaccess"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent_config/agentconfigreq"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/global"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/kweaver-ai/proton-rds-sdk-go/sqlx"
)

type daTestLogger struct{}

func (daTestLogger) Infof(string, ...interface{})  {}
func (daTestLogger) Infoln(...interface{})         {}
func (daTestLogger) Debugf(string, ...interface{}) {}
func (daTestLogger) Debugln(...interface{})        {}
func (daTestLogger) Errorf(string, ...interface{}) {}
func (daTestLogger) Errorln(...interface{})        {}
func (daTestLogger) Warnf(string, ...interface{})  {}
func (daTestLogger) Warnln(...interface{})         {}
func (daTestLogger) Panicf(string, ...interface{}) {}
func (daTestLogger) Panicln(...interface{})        {}
func (daTestLogger) Fatalf(string, ...interface{}) {}
func (daTestLogger) Fatalln(...interface{})        {}

func newDARepoWithMock(t *testing.T) (*DAConfigRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &DAConfigRepo{
		db:             db,
		logger:         daTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func dataAgentColumns() []string {
	return []string{
		"f_id",
		"f_name",
		"f_key",
		"f_profile",
		"f_product_key",
		"f_avatar_type",
		"f_avatar",
		"f_status",
		"f_is_built_in",
		"f_is_system_agent",
		"f_created_at",
		"f_updated_at",
		"f_created_by",
		"f_updated_by",
		"f_deleted_at",
		"f_deleted_by",
		"f_config",
		"f_created_type",
		"f_create_from",
	}
}

func dataAgentRows() *sqlmock.Rows {
	return sqlmock.NewRows(dataAgentColumns()).AddRow(
		"agent-1",
		"name-1",
		"key-1",
		nil,
		"pk",
		0,
		"",
		string(cdaenum.StatusUnpublished),
		nil,
		nil,
		int64(1),
		int64(1),
		"u1",
		"u1",
		int64(0),
		"",
		"{}",
		"",
		"",
	)
}

func userCtx(uid string) context.Context {
	visitor := &rest.Visitor{ID: uid}
	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck
}

func TestNewDataAgentRepo_Singleton(t *testing.T) {
	t.Parallel()

	oldOnce := agentRepoOnce //nolint:govet
	oldImpl := agentRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		agentRepoOnce = oldOnce //nolint:govet
		agentRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	agentRepoOnce = sync.Once{}
	agentRepoImpl = nil

	r1 := NewDataAgentRepo()
	r2 := NewDataAgentRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

func TestDAConfigRepo_ExistsMethods(t *testing.T) {
	t.Parallel()

	t.Run("exists by name true", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select 1 from t_data_agent_config where f_name = \? and f_deleted_at = \? limit 1`).
			WithArgs("n1", 0).
			WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

		exists, err := repo.ExistsByName(context.Background(), "n1")
		require.NoError(t, err)
		assert.True(t, exists)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exists by id false", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select 1 from t_data_agent_config where f_id = \? and f_deleted_at = \? limit 1`).
			WithArgs("id-1", 0).
			WillReturnRows(sqlmock.NewRows([]string{"1"}))

		exists, err := repo.ExistsByID(context.Background(), "id-1")
		require.NoError(t, err)
		assert.False(t, exists)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("exists by name exclude id error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select 1 from t_data_agent_config where f_deleted_at = \? and f_name = \? and f_id <> \? limit 1`).
			WithArgs(0, "n1", "id-1").
			WillReturnError(errors.New("query failed"))

		_, err := repo.ExistsByNameExcludeID(context.Background(), "n1", "id-1")
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigRepo_ListForBenchmark(t *testing.T) {
	t.Parallel()

	t.Run("count error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`(?s)select count\(\*\).*from t_data_agent_config c`).
			WithArgs("%foo%", "a1", "a2").
			WillReturnError(errors.New("count failed"))

		_, _, err := repo.ListForBenchmark(context.Background(), &agentconfigreq.ListForBenchmarkReq{
			Name:                "foo",
			PageSize:            agentconfigreq.ListForBenchmarkReq{}.PageSize,
			AgentIDsByBizDomain: []string{"a1", "a2"},
		})
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count zero", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`(?s)select count\(\*\).*from t_data_agent_config c`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		pos, count, err := repo.ListForBenchmark(context.Background(), &agentconfigreq.ListForBenchmarkReq{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.Empty(t, pos)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success with published and unpublished", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		req := &agentconfigreq.ListForBenchmarkReq{
			Name:                "foo",
			AgentIDsByBizDomain: []string{"a1", "a2"},
		}

		mock.ExpectQuery(`(?s)select count\(\*\).*from t_data_agent_config c.*c\.f_name LIKE \?.*c\.f_id IN \(\?,\?\)`).
			WithArgs("%foo%", "a1", "a2").
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{
			"f_id", "f_key", "f_name", "f_updated_at", "f_version", "f_status",
			"f_agent_id", "f_agent_key", "f_agent_name", "f_published_at", "f_release_version", "f_release_status",
		}).
			AddRow("a1", "k1", "n1", int64(100), "v0", "unpublished", "a1", "k1", "n1-pub", int64(200), "latest", "published").
			AddRow("a2", "k2", "n2", int64(300), "v0", "published", nil, nil, nil, nil, nil, nil)

		mock.ExpectQuery(`(?s)select c\.f_id,c\.f_key, c\.f_name, c\.f_updated_at,.*from t_data_agent_config c`).
			WithArgs("%foo%", "a1", "a2").
			WillReturnRows(rows)

		pos, count, err := repo.ListForBenchmark(context.Background(), req)
		require.NoError(t, err)
		assert.Equal(t, int64(2), count)
		assert.Len(t, pos, 3)
		assert.Equal(t, "latest", pos[0].Version)
		assert.Equal(t, "published", string(pos[0].Status))
		assert.Equal(t, "latest", pos[2].Version)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("list query error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`(?s)select count\(\*\).*from t_data_agent_config c`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`(?s)select c\.f_id,c\.f_key, c\.f_name, c\.f_updated_at,.*from t_data_agent_config c`).
			WillReturnError(errors.New("find failed"))

		_, _, err := repo.ListForBenchmark(context.Background(), &agentconfigreq.ListForBenchmarkReq{})
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigRepo_ListForBenchmark2(t *testing.T) {
	t.Parallel()

	t.Run("count zero", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(.*t_data_agent_config.*t_data_agent_release.*\) AS total_count`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))

		pos, count, err := repo.ListForBenchmark2(context.Background(), &agentconfigreq.ListForBenchmarkReq{})
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
		assert.Empty(t, pos)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(.*t_data_agent_config.*t_data_agent_release.*\) AS total_count`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"f_id", "f_key", "f_name", "f_version", "f_status", "updated_at"}).
			AddRow("a1", "k1", "n1", "latest", "published", int64(123))
		mock.ExpectQuery(`(?s)\(.*SELECT.*FROM\s+t_data_agent_config.*UNION ALL.*t_data_agent_release.*ORDER BY\s+updated_at DESC\s+LIMIT`).
			WillReturnRows(rows)

		pos, count, err := repo.ListForBenchmark2(context.Background(), &agentconfigreq.ListForBenchmarkReq{Name: "foo"})
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
		assert.Len(t, pos, 1)
		assert.Equal(t, "a1", pos[0].ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("count and list error", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(.*t_data_agent_config.*\) AS total_count`).
			WillReturnError(errors.New("count failed"))

		_, _, err := repo.ListForBenchmark2(context.Background(), &agentconfigreq.ListForBenchmarkReq{})
		assert.Error(t, err)

		mock.ExpectQuery(`(?s)SELECT COUNT\(\*\) FROM \(.*t_data_agent_config.*\) AS total_count`).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
		mock.ExpectQuery(`(?s)\(.*SELECT.*FROM\s+t_data_agent_config.*UNION ALL.*t_data_agent_release.*ORDER BY\s+updated_at DESC\s+LIMIT`).
			WillReturnError(errors.New("find failed"))

		_, _, err = repo.ListForBenchmark2(context.Background(), &agentconfigreq.ListForBenchmarkReq{})
		assert.Error(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigRepo_CreateAndCreateBatch(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_config")).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("insert into t_data_agent_config")).
		WillReturnResult(sqlmock.NewResult(1, 2))

	po := &dapo.DataAgentPo{Name: "name-1", Key: "key-1"}
	err := repo.Create(context.Background(), nil, "agent-1", po)
	require.NoError(t, err)
	assert.Equal(t, "agent-1", po.ID)

	err = repo.CreateBatch(context.Background(), nil, []*dapo.DataAgentPo{{ID: "a1"}, {ID: "a2"}})
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDAConfigRepo_GetAllIDs(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(regexp.QuoteMeta("select f_id from t_data_agent_config where f_deleted_at = ?")).
		WithArgs(0).
		WillReturnRows(sqlmock.NewRows([]string{"f_id"}).AddRow("a1").AddRow("a2"))

	ids, err := repo.GetAllIDs(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []string{"a1", "a2"}, ids)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDAConfigRepo_GetByBasicMethods(t *testing.T) {
	t.Parallel()

	t.Run("get by id and key", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_config where f_deleted_at = \? and f_id = \?`).
			WithArgs(0, "agent-1").
			WillReturnRows(dataAgentRows())
		mock.ExpectQuery(`select .* from t_data_agent_config where f_deleted_at = \? and f_key = \?`).
			WithArgs(0, "key-1").
			WillReturnRows(dataAgentRows())

		po1, err := repo.GetByID(context.Background(), "agent-1")
		require.NoError(t, err)
		assert.Equal(t, "agent-1", po1.ID)

		po2, err := repo.GetByKey(context.Background(), "key-1")
		require.NoError(t, err)
		assert.Equal(t, "agent-1", po2.ID)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("get by keys and ids", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_config where f_deleted_at = \? and f_key in \(\?,\?\)`).
			WithArgs(0, "k1", "k2").
			WillReturnRows(dataAgentRows())
		mock.ExpectQuery(`select .* from t_data_agent_config where f_deleted_at = \? and f_id in \(\?,\?\)`).
			WithArgs(0, "a1", "a2").
			WillReturnRows(dataAgentRows())

		posByKey, err := repo.GetByKeys(context.Background(), []string{"k1", "k2"})
		require.NoError(t, err)
		assert.Len(t, posByKey, 1)

		posByID, err := repo.GetByIDS(context.Background(), []string{"a1", "a2"})
		require.NoError(t, err)
		assert.Len(t, posByID, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigRepo_GetMapAndNameMapByIDs(t *testing.T) {
	t.Parallel()

	t.Run("empty ids", func(t *testing.T) {
		t.Parallel()

		repo, db, _ := newDARepoWithMock(t)
		defer db.Close()

		m1, err := repo.GetMapByIDs(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, m1)

		m2, err := repo.GetIDNameMapByID(context.Background(), nil)
		require.NoError(t, err)
		assert.Empty(t, m2)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_config where f_deleted_at = \? and f_id in \(\?,\?\)`).
			WithArgs(0, "a1", "a2").
			WillReturnRows(dataAgentRows())
		mock.ExpectQuery(regexp.QuoteMeta("select f_id,f_name from t_data_agent_config where f_deleted_at = ? and f_id in (?,?)")).
			WithArgs(0, "a1", "a2").
			WillReturnRows(sqlmock.NewRows([]string{"f_id", "f_name"}).AddRow("a1", "n1").AddRow("a2", "n2"))

		m1, err := repo.GetMapByIDs(context.Background(), []string{"a1", "a2"})
		require.NoError(t, err)
		assert.Len(t, m1, 1)

		m2, err := repo.GetIDNameMapByID(context.Background(), []string{"a1", "a2"})
		require.NoError(t, err)
		assert.Equal(t, "n1", m2["a1"])
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigRepo_GetByIDsAndCreatedBy(t *testing.T) {
	t.Parallel()

	t.Run("ids empty", func(t *testing.T) {
		t.Parallel()

		repo, db, _ := newDARepoWithMock(t)
		defer db.Close()

		res, err := repo.GetByIDsAndCreatedBy(context.Background(), nil, "u1")
		require.NoError(t, err)
		assert.Empty(t, res)
	})

	t.Run("success and dedupe", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectQuery(`select .* from t_data_agent_config where f_deleted_at = \? and f_created_by = \? and f_id in \(\?,\?\)`).
			WithArgs(0, "u1", "a1", "a2").
			WillReturnRows(dataAgentRows())

		res, err := repo.GetByIDsAndCreatedBy(context.Background(), []string{"a1", "a1", "a2"}, "u1")
		require.NoError(t, err)
		assert.Len(t, res, 1)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDAConfigRepo_UpdateAndUpdateByKeyAndStatus(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`update t_data_agent_config set .* where f_id = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update t_data_agent_config set .* where f_key = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`update t_data_agent_config set .* where f_id = \?`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	po := &dapo.DataAgentPo{ID: "a1", Key: "k1", Name: "n1"}
	err := repo.Update(context.Background(), nil, po)
	require.NoError(t, err)

	err = repo.UpdateByKey(context.Background(), nil, po)
	require.NoError(t, err)

	err = repo.UpdateStatus(context.Background(), nil, cdaenum.StatusPublished, "a1", "u1")
	require.NoError(t, err)

	err = repo.UpdateStatus(context.Background(), nil, cdaenum.Status("invalid"), "a1", "u1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDAConfigRepo_Delete(t *testing.T) {
	t.Parallel()

	t.Run("uid empty", func(t *testing.T) {
		t.Parallel()

		repo, db, _ := newDARepoWithMock(t)
		defer db.Close()

		err := repo.Delete(context.Background(), nil, "a1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "uid is empty")
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo, db, mock := newDARepoWithMock(t)
		defer db.Close()

		mock.ExpectExec(`update t_data_agent_config set .* where f_id = \?`).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(userCtx("u1"), nil, "a1")
		require.NoError(t, err)
		require.NoError(t, mock.ExpectationsWereMet())
	})
}
