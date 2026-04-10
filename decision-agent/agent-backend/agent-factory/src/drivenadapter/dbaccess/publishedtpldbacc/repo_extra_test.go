package publishedtpldbacc

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

var mockCreatePo = dapo.PublishedTplPo{
	Name: "tpl-name",
	Key:  "tpl-key",
}

var mockCatAssocPos = []*dapo.PubTplCatAssocPo{
	{PublishedTplID: 1, CategoryID: "cat-1"},
}

// ==================== Create ====================

func TestCreate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	// InsertStruct
	mock.ExpectExec(`(?i)insert into t_data_agent_config_tpl_published`).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// GetByKeyWithTx
	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_key = \?`).
		WithArgs("tpl-key").
		WillReturnRows(mockPubedTplRows())

	mock.ExpectRollback()

	id, err := repo.Create(context.Background(), tx, &mockCreatePo)
	require.NoError(t, err)
	assert.Equal(t, int64(1), id)

	_ = tx.Rollback()
}

func TestCreate_InsertError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectExec(`(?i)insert into t_data_agent_config_tpl_published`).
		WillReturnError(errors.New("insert failed"))
	mock.ExpectRollback()

	_, err = repo.Create(context.Background(), tx, &mockCreatePo)
	assert.Error(t, err)

	_ = tx.Rollback()
}

// ==================== BatchCreateCategoryAssoc ====================

func TestBatchCreateCategoryAssoc_Empty(t *testing.T) {
	t.Parallel()

	repo, db, _ := newPubedTplRepoWithMock(t)
	defer db.Close()

	err := repo.BatchCreateCategoryAssoc(context.Background(), nil, nil)
	assert.NoError(t, err)
}

func TestBatchCreateCategoryAssoc_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into`).
		WillReturnError(errors.New("insert failed"))

	err := repo.BatchCreateCategoryAssoc(context.Background(), nil, mockCatAssocPos)
	assert.Error(t, err)
}

// ==================== DelCategoryAssocByTplID ====================

func TestDelCategoryAssocByTplID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DelCategoryAssocByTplID(context.Background(), nil, 1)
	assert.NoError(t, err)
}

func TestDelCategoryAssocByTplID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from")).
		WillReturnError(errors.New("delete failed"))

	err := repo.DelCategoryAssocByTplID(context.Background(), nil, 1)
	assert.Error(t, err)
}

// ==================== GetCategoryAssocByTplID ====================

// NOTE: TestGetCategoryAssocByTplID_Happy 跳过，因 dbhelper2.Find 对 []*PubTplCatAssocPo 指针切片 scan 不兼容 sqlmock

func TestGetCategoryAssocByTplID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetCategoryAssocByTplID(context.Background(), nil, 100)
	assert.Error(t, err)
}

// ==================== GetCategoryJoinPosByTplID ====================

func TestGetCategoryJoinPosByTplID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	rows := sqlmock.NewRows([]string{"f_id", "f_published_tpl_id", "f_category_id", "f_category_name"}).
		AddRow(int64(1), int64(100), "cat-1", "分类1")

	mock.ExpectQuery(`(?i)SELECT .* FROM`).
		WillReturnRows(rows)

	pos, err := repo.GetCategoryJoinPosByTplID(context.Background(), nil, 100)
	require.NoError(t, err)
	assert.Len(t, pos, 1)
}

func TestGetCategoryJoinPosByTplID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)SELECT .* FROM`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetCategoryJoinPosByTplID(context.Background(), nil, 100)
	assert.Error(t, err)
}

// ==================== GetByIDWithTx ====================

func TestGetByIDWithTx_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_id = \?`).
		WithArgs(int64(1)).
		WillReturnRows(mockPubedTplRows())
	mock.ExpectRollback()

	po, err := repo.GetByIDWithTx(context.Background(), tx, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), po.ID)

	_ = tx.Rollback()
}

// ==================== GetByKeyWithTx ====================

func TestGetByKeyWithTx_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectQuery(`select .* from t_data_agent_config_tpl_published where f_key = \?`).
		WithArgs("tpl-key").
		WillReturnRows(mockPubedTplRows())
	mock.ExpectRollback()

	po, err := repo.GetByKeyWithTx(context.Background(), tx, "tpl-key")
	require.NoError(t, err)
	assert.Equal(t, "tpl-key", po.Key)

	_ = tx.Rollback()
}

// ==================== GetByCategoryID ====================

func TestGetByCategoryID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)SELECT .* FROM`).
		WillReturnRows(mockPubedTplRows())

	pos, err := repo.GetByCategoryID(context.Background(), "cat-1")
	require.NoError(t, err)
	assert.Len(t, pos, 1)
}

func TestGetByCategoryID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)SELECT .* FROM`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByCategoryID(context.Background(), "cat-1")
	assert.Error(t, err)
}
