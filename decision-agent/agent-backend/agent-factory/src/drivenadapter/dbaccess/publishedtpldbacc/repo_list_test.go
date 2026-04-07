package publishedtpldbacc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/published/pubedreq"
)

// ==================== GetPubTplList ====================

func TestGetPubTplList_EmptyTplIDsDoesNotPanic(t *testing.T) {
	t.Parallel()

	repo, db, _ := newPubedTplRepoWithMock(t)
	defer db.Close()

	_, err := repo.GetPubTplList(context.Background(), &pubedreq.PubedTplListReq{Size: 10})
	assert.Error(t, err)
}

func TestGetPubTplList_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).
		WillReturnError(errors.New("find error"))

	req := &pubedreq.PubedTplListReq{
		TplIDsByBd: []string{"tpl-1"},
	}
	req.Size = 10

	_, err := repo.GetPubTplList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubTplList_WithNameFilter_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).
		WillReturnError(errors.New("find error"))

	req := &pubedreq.PubedTplListReq{
		Name:       "test",
		TplIDsByBd: []string{"tpl-1"},
	}
	req.Size = 10

	_, err := repo.GetPubTplList(context.Background(), req)
	assert.Error(t, err)
}

func TestGetPubTplList_WithCategoryFilter_Error(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).
		WillReturnError(errors.New("find error"))

	req := &pubedreq.PubedTplListReq{
		CategoryID: "cat-1",
		TplIDsByBd: []string{"tpl-1"},
	}
	req.Size = 10

	_, err := repo.GetPubTplList(context.Background(), req)
	assert.Error(t, err)
}

// ==================== GetCategoryJoinPosByTplID ====================

func TestGetCategoryJoinPosByTplID_QueryError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newPubedTplRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .*`).
		WillReturnError(errors.New("query err"))

	_, err := repo.GetCategoryJoinPosByTplID(context.Background(), nil, 100)
	assert.Error(t, err)
}
