package conversationdbacc

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/conversation/conversationreq"
)

// ==================== List with Title filter ====================

func TestList_WithTitle_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_conversation`).
		WillReturnError(errors.New("find err"))

	req := conversationreq.ListReq{
		AgentAPPKey: "app-1",
		UserId:      "u1",
		Title:       "test",
	}
	req.Page = 1
	req.Size = 10
	_, _, err := repo.List(context.Background(), req)
	assert.Error(t, err)
}

func TestList_WithTitle_CountError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_conversation`).
		WillReturnRows(mockConversationRows())
	mock.ExpectQuery(`(?i)select count\(\*\) from t_conversation`).
		WillReturnError(errors.New("count err"))

	req := conversationreq.ListReq{
		AgentAPPKey: "app-1",
		UserId:      "u1",
		Title:       "test",
	}
	req.Page = 1
	req.Size = 10
	_, _, err := repo.List(context.Background(), req)
	assert.Error(t, err)
}

// ==================== ListByAgentID with Title filter ====================

func TestListByAgentID_WithTitle_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_conversation`).
		WillReturnError(errors.New("find err"))

	_, _, err := repo.ListByAgentID(context.Background(), "agent-1", "keyword", 1, 10)
	assert.Error(t, err)
}

func TestListByAgentID_NoPagination_FindError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_conversation`).
		WillReturnError(errors.New("find err"))

	_, _, err := repo.ListByAgentID(context.Background(), "agent-1", "", 0, 0)
	assert.Error(t, err)
}

func TestListByAgentID_WithTitle_CountError(t *testing.T) {
	t.Parallel()

	repo, db, mock := newRepoWithMock(t)
	defer db.Close()

	mock.ExpectQuery(`(?i)select .* from t_conversation`).
		WillReturnRows(mockConversationRows())
	mock.ExpectQuery(`(?i)select count\(\*\) from t_conversation`).
		WillReturnError(errors.New("count err"))

	_, _, err := repo.ListByAgentID(context.Background(), "agent-1", "keyword", 1, 10)
	assert.Error(t, err)
}
