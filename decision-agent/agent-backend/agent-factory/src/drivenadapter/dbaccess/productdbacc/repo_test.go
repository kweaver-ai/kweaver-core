package productdbacc

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

type prodTestLogger struct{}

func (prodTestLogger) Infof(string, ...interface{})  {}
func (prodTestLogger) Infoln(...interface{})         {}
func (prodTestLogger) Debugf(string, ...interface{}) {}
func (prodTestLogger) Debugln(...interface{})        {}
func (prodTestLogger) Errorf(string, ...interface{}) {}
func (prodTestLogger) Errorln(...interface{})        {}
func (prodTestLogger) Warnf(string, ...interface{})  {}
func (prodTestLogger) Warnln(...interface{})         {}
func (prodTestLogger) Panicf(string, ...interface{}) {}
func (prodTestLogger) Panicln(...interface{})        {}
func (prodTestLogger) Fatalf(string, ...interface{}) {}
func (prodTestLogger) Fatalln(...interface{})        {}

func newProductRepoWithMock(t *testing.T) (*ProductRepo, *sqlx.DB, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlx.New()
	require.NoError(t, err)

	repo := &ProductRepo{
		db:             db,
		logger:         prodTestLogger{},
		IDBAccBaseRepo: dbaccess.NewDBAccBase(),
	}

	return repo, db, mock
}

func mockProductRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id", "f_name", "f_key", "f_profile",
		"f_created_at", "f_updated_at", "f_created_by", "f_updated_by",
		"f_deleted_at", "f_deleted_by",
	}).AddRow(
		int64(1), "prod", "prod-key", "描述",
		int64(1000), int64(1000), "u1", "u1",
		int64(0), "",
	)
}

// ==================== Singleton ====================

func TestNewProductRepo_Singleton(t *testing.T) {
	old := productRepoOnce //nolint:govet
	oldImpl := productRepoImpl
	oldGDB := global.GDB

	t.Cleanup(func() {
		productRepoOnce = old //nolint:govet
		productRepoImpl = oldImpl
		global.GDB = oldGDB
	})

	db, _, err := sqlx.New()
	require.NoError(t, err)

	global.GDB = db
	productRepoOnce = sync.Once{}
	productRepoImpl = nil

	r1 := NewProductRepo()
	r2 := NewProductRepo()

	assert.NotNil(t, r1)
	assert.Same(t, r1, r2)
}

// ==================== Create ====================

func TestCreate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_product`).
		WillReturnResult(sqlmock.NewResult(1, 1))

	key, err := repo.Create(context.Background(), &dapo.ProductPo{Name: "test"})
	require.NoError(t, err)
	assert.NotEmpty(t, key) // Key 自动生成
}

func TestCreate_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_product`).
		WillReturnError(errors.New("insert failed"))

	_, err := repo.Create(context.Background(), &dapo.ProductPo{Name: "test"})
	assert.Error(t, err)
}

// ==================== Delete ====================

func TestDelete_EmptyUID(t *testing.T) {
	t.Parallel()

	repo, db, _ := newProductRepoWithMock(t)
	defer db.Close()

	// context 没有 visitor，uid 为空 → 应返回错误
	err := repo.Delete(context.Background(), 1)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "uid is empty")
}

// ==================== ExistsByName ====================

func TestExistsByName_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ExistsByName(context.Background(), "test")
	assert.Error(t, err)
}

// ==================== ExistsByKey ====================

func TestExistsByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ExistsByKey(context.Background(), "key")
	assert.Error(t, err)
}

// ==================== ExistsByID ====================

func TestExistsByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ExistsByID(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== ExistsByNameExcludeID ====================

func TestExistsByNameExcludeID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ExistsByNameExcludeID(context.Background(), "test", 1)
	assert.Error(t, err)
}

// ==================== ExistsByKeyExcludeID ====================

func TestExistsByKeyExcludeID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.ExistsByKeyExcludeID(context.Background(), "key", 1)
	assert.Error(t, err)
}

// ==================== GetByID ====================

func TestGetByID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(mockProductRows())

	po, err := repo.GetByID(context.Background(), 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), po.ID)
}

func TestGetByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

// ==================== GetByKey ====================

func TestGetByKey_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_product`).
		WillReturnRows(mockProductRows())

	po, err := repo.GetByKey(context.Background(), "prod-key")
	require.NoError(t, err)
	assert.Equal(t, "prod-key", po.Key)
}

// ==================== Update ====================

func TestUpdate_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_product`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.Update(context.Background(), &dapo.ProductPo{ID: 1, Name: "new"})
	assert.NoError(t, err)
}

func TestUpdate_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newProductRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_product`).
		WillReturnError(errors.New("update failed"))

	err := repo.Update(context.Background(), &dapo.ProductPo{ID: 1, Name: "new"})
	assert.Error(t, err)
}
