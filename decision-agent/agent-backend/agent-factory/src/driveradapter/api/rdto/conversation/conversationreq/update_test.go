package conversationreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateReq_StructFields(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		ID:         "conv-123",
		Title:      "Test Conversation",
		TempareaId: "temp-456",
	}

	assert.Equal(t, "conv-123", req.ID)
	assert.Equal(t, "Test Conversation", req.Title)
	assert.Equal(t, "temp-456", req.TempareaId)
}

func TestUpdateReq_Empty(t *testing.T) {
	t.Parallel()

	req := UpdateReq{}

	assert.Empty(t, req.ID)
	assert.Empty(t, req.Title)
	assert.Empty(t, req.TempareaId)
}

func TestUpdateReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := UpdateReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, `"name"不能为空`, errMsgMap["ID.required"])
	assert.Equal(t, `"title"不能为空`, errMsgMap["Title.required"])
}

func TestUpdateReq_ReqCheck(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		ID:    "conv-123",
		Title: "Test Conversation",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestUpdateReq_ReqCheck_Empty(t *testing.T) {
	t.Parallel()

	req := UpdateReq{}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestUpdateReq_WithDifferentIDs(t *testing.T) {
	t.Parallel()

	ids := []string{
		"conv-001",
		"conversation-xyz",
		"会话-123",
		"12345",
	}

	for _, id := range ids {
		req := UpdateReq{
			ID: id,
		}
		assert.Equal(t, id, req.ID)
	}
}

func TestUpdateReq_WithDifferentTitles(t *testing.T) {
	t.Parallel()

	titles := []string{
		"Test Conversation",
		"中文标题",
		"Conversation with numbers 123",
		"Title with special chars !@#$%",
	}

	for _, title := range titles {
		req := UpdateReq{
			Title: title,
		}
		assert.Equal(t, title, req.Title)
	}
}

func TestUpdateReq_WithTempareaId(t *testing.T) {
	t.Parallel()

	tempareaIds := []string{
		"temp-001",
		"temp-xyz",
		"",
		"临时-123",
	}

	for _, tempareaId := range tempareaIds {
		req := UpdateReq{
			TempareaId: tempareaId,
		}
		assert.Equal(t, tempareaId, req.TempareaId)
	}
}

func TestUpdateReq_WithAllFields(t *testing.T) {
	t.Parallel()

	req := UpdateReq{
		ID:         "conv-abc-123",
		Title:      "Complete Conversation Title",
		TempareaId: "temp-def-456",
	}

	assert.Equal(t, "conv-abc-123", req.ID)
	assert.Equal(t, "Complete Conversation Title", req.Title)
	assert.Equal(t, "temp-def-456", req.TempareaId)

	err := req.ReqCheck()
	assert.NoError(t, err)
}
