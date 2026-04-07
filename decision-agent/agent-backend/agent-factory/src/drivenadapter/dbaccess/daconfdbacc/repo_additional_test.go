package daconfdbacc

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
)

// ==================== Create / CreateBatch error ====================

func TestCreate_InsertError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_data_agent_config`).WillReturnError(errors.New("insert err"))
	assert.Error(t, repo.Create(context.Background(), nil, "a1", &dapo.DataAgentPo{}))
}

func TestCreateBatch_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)insert into t_data_agent_config`).WillReturnError(errors.New("batch err"))
	assert.Error(t, repo.CreateBatch(context.Background(), nil, []*dapo.DataAgentPo{{ID: "a1"}}))
}

// ==================== Update / UpdateByKey error ====================

func TestUpdate_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config`).WillReturnError(errors.New("update err"))
	assert.Error(t, repo.Update(context.Background(), nil, &dapo.DataAgentPo{ID: "a1"}))
}

func TestUpdateByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config`).WillReturnError(errors.New("update err"))
	assert.Error(t, repo.UpdateByKey(context.Background(), nil, &dapo.DataAgentPo{Key: "k1"}))
}

// ==================== Delete error ====================

func TestDelete_ExecError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config`).WillReturnError(errors.New("exec err"))
	assert.Error(t, repo.Delete(userCtx("u1"), nil, "a1"))
}

// ==================== GetAllIDs error ====================

func TestGetAllIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetAllIDs(context.Background())
	assert.Error(t, err)
}

// ==================== GetByKeys / GetByIDS error ====================

func TestGetByKeys_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByKeys(context.Background(), []string{"k1"})
	assert.Error(t, err)
}

func TestGetByIDS_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByIDS(context.Background(), []string{"a1"})
	assert.Error(t, err)
}

// ==================== GetMapByIDs error ====================

func TestGetMapByIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetMapByIDs(context.Background(), []string{"a1"})
	assert.Error(t, err)
}

// ==================== GetIDNameMapByID error ====================

func TestGetIDNameMapByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetIDNameMapByID(context.Background(), []string{"a1"})
	assert.Error(t, err)
}

// ==================== GetByIDsAndCreatedBy error ====================

func TestGetByIDsAndCreatedBy_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByIDsAndCreatedBy(context.Background(), []string{"a1"}, "u1")
	assert.Error(t, err)
}

// ==================== UpdateStatus error ====================

func TestUpdateStatus_ExecError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(`(?i)update t_data_agent_config`).WillReturnError(errors.New("exec err"))
	assert.Error(t, repo.UpdateStatus(context.Background(), nil, cdaenum.StatusPublished, "a1", "u1"))
}

// ==================== ExistsByName error ====================

func TestExistsByName_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.ExistsByName(context.Background(), "n1")
	assert.Error(t, err)
}

func TestExistsByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.ExistsByID(context.Background(), "a1")
	assert.Error(t, err)
}

// ==================== Exists success (true) paths ====================

func TestExistsByNameExcludeID_True(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).
		WillReturnRows(sqlmock.NewRows([]string{"1"}).AddRow(1))

	exists, err := repo.ExistsByNameExcludeID(context.Background(), "n1", "a1")
	assert.NoError(t, err)
	assert.True(t, exists)
}

// ==================== GetByID / GetByKey error ====================

func TestGetByID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByID(context.Background(), "a1")
	assert.Error(t, err)
}

func TestGetByKey_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newDARepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_config`).WillReturnError(errors.New("query err"))

	_, err := repo.GetByKey(context.Background(), "k1")
	assert.Error(t, err)
}
