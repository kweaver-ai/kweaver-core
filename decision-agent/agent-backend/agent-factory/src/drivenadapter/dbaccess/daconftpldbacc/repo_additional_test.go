package daconftpldbacc

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
)

// ==================== Create / Update error paths ====================

func TestCreate_InsertError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_data_agent_config_tpl`).
		WillReturnError(errors.New("insert err"))

	err := repo.Create(context.Background(), nil, &dapo.DataAgentTplPo{Name: "t1"})
	assert.Error(t, err)
}

func TestUpdate_ExecError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config_tpl`).
		WillReturnError(errors.New("update err"))

	err := repo.Update(context.Background(), nil, &dapo.DataAgentTplPo{ID: 1})
	assert.Error(t, err)
}

// ==================== Exists error paths ====================

func TestExistsByName_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.ExistsByName(context.Background(), "t1")
	assert.Error(t, err)
}

func TestExistsByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.ExistsByKey(context.Background(), "k1")
	assert.Error(t, err)
}

func TestExistsByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.ExistsByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestExistsByNameExcludeID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.ExistsByNameExcludeID(context.Background(), "t1", 1)
	assert.Error(t, err)
}

func TestExistsByNameExcludeID_True(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByNameExcludeID(context.Background(), "t1", 1)
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestExistsByKeyExcludeID_True(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	exists, err := repo.ExistsByKeyExcludeID(context.Background(), "k1", 1)
	require.NoError(t, err)
	assert.True(t, exists)
}

// ==================== GetByKeys / GetByIDS error paths ====================

func TestGetByKeys_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByKeys(context.Background(), []string{"k1"})
	assert.Error(t, err)
}

func TestGetByIDS_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByIDS(context.Background(), []int64{1})
	assert.Error(t, err)
}

// ==================== GetAllIDs error ====================

func TestGetAllIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetAllIDs(context.Background())
	assert.Error(t, err)
}

// ==================== GetByCategoryID error ====================

func TestGetByCategoryID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByCategoryID(context.Background(), "c1")
	assert.Error(t, err)
}

// ==================== GetMapByIDs error ====================

func TestGetMapByIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetMapByIDs(context.Background(), []int64{1})
	assert.Error(t, err)
}

// ==================== Delete error path ====================

func TestDelete_ExecError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config_tpl`).
		WillReturnError(errors.New("exec err"))

	err := repo.Delete(tplUserCtx("u1"), nil, 1)
	assert.Error(t, err)
}

// ==================== UpdateStatus error path ====================

func TestUpdateStatus_ExecError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config_tpl`).
		WillReturnError(errors.New("exec err"))

	err := repo.UpdateStatus(context.Background(), nil, "published", 1, "u1", 100)
	assert.Error(t, err)
}

// ==================== GetByKeyWithTx error ====================

func TestGetByKeyWithTx_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectBegin()

	tx, err := db.Begin()
	require.NoError(t, err)

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))
	mock.ExpectRollback()

	_, err = repo.GetByKeyWithTx(context.Background(), tx, "k1")
	assert.Error(t, err)

	_ = tx.Rollback()
}

// ==================== GetByID / GetByKey error ====================

func TestGetByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestGetByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config_tpl`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByKey(context.Background(), "k1")
	assert.Error(t, err)
}
