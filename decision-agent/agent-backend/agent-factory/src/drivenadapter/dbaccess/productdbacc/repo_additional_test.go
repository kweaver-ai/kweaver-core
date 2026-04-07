package productdbacc

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func ctxWithUser(uid string) context.Context {
	visitor := &rest.Visitor{ID: uid}
	return context.WithValue(context.Background(), cenum.VisitUserInfoCtxKey.String(), visitor) //nolint:staticcheck
}

// ==================== Delete ====================

func TestDelete_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_product`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Delete(ctxWithUser("u1"), 1)
	assert.NoError(t, err)
}

func TestDelete_ExecError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_product`).
		WillReturnError(errors.New("exec failed"))

	err := repo.Delete(ctxWithUser("u1"), 1)
	assert.Error(t, err)
}

// ==================== GetByKeys ====================

func TestGetByKeys_EmptyKeys(t *testing.T) {
	t.Parallel()

	repo, db, _ := newProductRepoWithMock(t)
	defer db.Close()

	pos, err := repo.GetByKeys(context.Background(), []string{})
	assert.NoError(t, err)
	assert.Empty(t, pos)
}

func TestGetByKeys_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(mockProductRows())

	pos, err := repo.GetByKeys(context.Background(), []string{"prod-key"})
	require.NoError(t, err)
	assert.Len(t, pos, 1)
	assert.Equal(t, "prod-key", pos[0].Key)
}

func TestGetByKeys_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("db error"))

	pos, err := repo.GetByKeys(context.Background(), []string{"key1"})
	assert.Error(t, err)
	assert.Nil(t, pos)
}

// ==================== GetByNameMapByKeys ====================

func TestGetByNameMapByKeys_EmptyKeys(t *testing.T) {
	t.Parallel()

	repo, db, _ := newProductRepoWithMock(t)
	defer db.Close()

	m, err := repo.GetByNameMapByKeys(context.Background(), []string{})
	assert.NoError(t, err)
	assert.Empty(t, m)
}

func TestGetByNameMapByKeys_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(mockProductRows())

	m, err := repo.GetByNameMapByKeys(context.Background(), []string{"prod-key"})
	require.NoError(t, err)
	assert.Equal(t, "prod-key", m["prod"])
}

func TestGetByNameMapByKeys_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("db error"))

	m, err := repo.GetByNameMapByKeys(context.Background(), []string{"key1"})
	assert.Error(t, err)
	assert.Nil(t, m)
}

// ==================== List ====================

func TestList_CountError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("count error"))

	_, _, err := repo.List(context.Background(), 0, 10)
	assert.Error(t, err)
}

func TestList_CountZero(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(0)))

	pos, total, err := repo.List(context.Background(), 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Nil(t, pos)
}

func TestList_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	// Count query
	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1)))

	// Find query
	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(mockProductRows())

	pos, total, err := repo.List(context.Background(), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, pos, 1)
}

func TestList_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(1)))

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("find error"))

	_, _, err := repo.List(context.Background(), 0, 10)
	assert.Error(t, err)
}

// ==================== ExistsByName (success path) ====================

func TestExistsByName_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.ExistsByName(context.Background(), "test")
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestExistsByName_NotExists(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"1"})) // no rows

	exists, err := repo.ExistsByName(context.Background(), "test")
	assert.NoError(t, err)
	assert.False(t, exists)
}

// ==================== ExistsByKey (success path) ====================

func TestExistsByKey_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.ExistsByKey(context.Background(), "key")
	assert.NoError(t, err)
	assert.True(t, exists)
}

// ==================== ExistsByID (success path) ====================

func TestExistsByID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.ExistsByID(context.Background(), 1)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// ==================== ExistsByNameExcludeID (success path) ====================

func TestExistsByNameExcludeID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.ExistsByNameExcludeID(context.Background(), "test", 1)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// ==================== ExistsByKeyExcludeID (success path) ====================

func TestExistsByKeyExcludeID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.ExistsByKeyExcludeID(context.Background(), "key", 1)
	assert.NoError(t, err)
	assert.True(t, exists)
}

// ==================== GetByKey (error path) ====================

func TestGetByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByKey(context.Background(), "key")
	assert.Error(t, err)
}
