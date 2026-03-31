package releaseacc

import (
	"context"
	"errors"
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

// ==================== releasePermissionRepo.DelByReleaseID ====================

func TestReleasePermission_DelByReleaseID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DelByReleaseID(context.Background(), nil, "release-1")
	assert.NoError(t, err)
}

func TestReleasePermission_DelByReleaseID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from")).
		WillReturnError(errors.New("delete failed"))

	err := repo.DelByReleaseID(context.Background(), nil, "release-1")
	assert.Error(t, err)
}

// ==================== releaseCategoryRelRepo.DelByReleaseID ====================

func TestReleaseCategoryRel_DelByReleaseID_Happy(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from")).
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repo.DelByReleaseID(context.Background(), nil, "release-1")
	assert.NoError(t, err)
}

func TestReleaseCategoryRel_DelByReleaseID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta("delete from")).
		WillReturnError(errors.New("delete failed"))

	err := repo.DelByReleaseID(context.Background(), nil, "release-1")
	assert.Error(t, err)
}

// ==================== releasePermissionRepo.GetByReleaseID ====================

func TestReleasePermission_GetByReleaseID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleasePermissionRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByReleaseID(context.Background(), "release-1")
	assert.Error(t, err)
}

// ==================== releaseCategoryRelRepo.GetByReleaseID ====================

func TestReleaseCategoryRel_GetByReleaseID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByReleaseID(context.Background(), "release-1")
	assert.Error(t, err)
}

// ==================== releaseCategoryRelRepo.GetByCategoryID ====================

func TestReleaseCategoryRel_GetByCategoryID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseCategoryRelRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetByCategoryID(context.Background(), "cat-1")
	assert.Error(t, err)
}

// ==================== releaseHistoryRepo.ListByAgentID ====================

func TestReleaseHistory_ListByAgentID_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseHistoryRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from`).
		WillReturnError(errors.New("query failed"))

	_, _, err := repo.ListByAgentID(context.Background(), "agent-1")
	assert.Error(t, err)
}

// ==================== releaseRepo.GetMapByAgentIDs ====================

func TestReleaseRepo_GetMapByAgentIDs_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newReleaseRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from`).
		WillReturnError(errors.New("query failed"))

	_, err := repo.GetMapByAgentIDs(context.Background(), []string{"agent-1"})
	assert.Error(t, err)
}
