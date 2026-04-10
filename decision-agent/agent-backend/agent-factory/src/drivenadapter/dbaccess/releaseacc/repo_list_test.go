package releaseacc

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/square/squarereq"
)

// ==================== ListByAgentID ====================

func mockReleaseHistoryListRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id", "f_agent_id", "f_agent_config", "f_agent_version", "f_agent_desc",
		"f_create_time", "f_update_time", "f_create_by", "f_update_by",
	}).
		AddRow("hist-1", "agent-1", `{}`, "v2", "desc", int64(2), int64(2), "u1", "u1").
		AddRow("hist-2", "agent-1", `{}`, "v1", "desc", int64(1), int64(1), "u1", "u1")
}

func TestListByAgentID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	// Find query
	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_history`).
		WillReturnRows(mockReleaseHistoryListRows())
	// Count query
	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_history`).
		WillReturnRows(sqlmock.NewRows([]string{"count(*)"}).AddRow(int64(2)))

	rt, total, err := repo.ListByAgentID(context.Background(), "agent-1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, rt, 2)
}

func TestListByAgentID_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_history`).
		WillReturnError(errors.New("find error"))

	_, _, err := repo.ListByAgentID(context.Background(), "agent-1")
	assert.Error(t, err)
}

func TestListByAgentID_CountError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_history`).
		WillReturnRows(mockReleaseHistoryListRows())
	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_history`).
		WillReturnError(errors.New("count error"))

	_, _, err := repo.ListByAgentID(context.Background(), "agent-1")
	assert.Error(t, err)
}

// ==================== GetMapByAgentIDs ====================

func TestGetMapByAgentIDs_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release`).
		WillReturnRows(mockReleaseRows())

	m, err := repo.GetMapByAgentIDs(context.Background(), []string{"agent-1"})
	require.NoError(t, err)
	assert.Contains(t, m, "rel-1")
}

func TestGetMapByAgentIDs_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetMapByAgentIDs(context.Background(), []string{"agent-1"})
	assert.Error(t, err)
}

// ==================== GetMapByUniqFlags ====================

func TestGetMapByUniqFlags_EmptyFlags(t *testing.T) {
	t.Parallel()

	repo, db, _ := newReleaseRepoWithMock(t)
	defer db.Close()

	m, err := repo.GetMapByUniqFlags(context.Background(), []*comvalobj.DataAgentUniqFlag{})
	assert.NoError(t, err)
	assert.Nil(t, m)
}

func TestGetMapByUniqFlags_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release`).
		WillReturnRows(mockReleaseRows())

	flags := []*comvalobj.DataAgentUniqFlag{
		{AgentID: "agent-1", AgentVersion: "v1"},
	}

	m, err := repo.GetMapByUniqFlags(context.Background(), flags)
	require.NoError(t, err)
	assert.Contains(t, m, "rel-1")
}

func TestGetMapByUniqFlags_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release`).
		WillReturnError(errors.New("find err"))

	flags := []*comvalobj.DataAgentUniqFlag{
		{AgentID: "agent-1", AgentVersion: "v1"},
	}

	_, err := repo.GetMapByUniqFlags(context.Background(), flags)
	assert.Error(t, err)
}

// ==================== ListRecentAgentForMarket ====================

func TestListRecentAgentForMarket_UnpublishedError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	// listRecentUnpublishedAgent 的 Raw+Find 失败
	mock.ExpectQuery(`(?i)select .*`).
		WillReturnError(errors.New("unpublished query error"))

	req := squarereq.AgentSquareRecentAgentReq{
		UserID:    "u1",
		StartTime: 1000,
		EndTime:   2000,
	}
	req.Size = 10
	_, err := repo.ListRecentAgentForMarket(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listRecentUnpublishedAgent")
}

// ==================== GetByReleaseID (success + not found) ====================

func mockReleaseCategoryRelRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"f_id", "f_release_id", "f_category_id",
	}).AddRow(
		"1", "rel-1", "cat-1",
	)
}

func TestGetByReleaseID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_category_rel`).
		WillReturnRows(mockReleaseCategoryRelRows())

	poList, err := repo.GetByReleaseID(context.Background(), "rel-1")
	require.NoError(t, err)
	assert.Len(t, poList, 1)
}

func TestGetByReleaseID_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_category_rel`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByReleaseID(context.Background(), "rel-1")
	assert.Error(t, err)
}

// ==================== GetByCategoryID ====================

func TestGetByCategoryID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_category_rel`).
		WillReturnRows(mockReleaseCategoryRelRows())

	poList, err := repo.GetByCategoryID(context.Background(), "cat-1")
	require.NoError(t, err)
	assert.Len(t, poList, 1)
}

func TestGetByCategoryID_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_data_agent_release_category_rel`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetByCategoryID(context.Background(), "cat-1")
	assert.Error(t, err)
}
