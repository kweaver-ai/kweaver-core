package conversationreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListReq_StructFields(t *testing.T) {
	t.Parallel()

	req := ListReq{
		AgentAPPKey: "app-key-123",
		UserId:      "user-456",
		Title:       "Test Conversation",
	}
	req.Size = 10
	req.Page = 1

	assert.Equal(t, "app-key-123", req.AgentAPPKey)
	assert.Equal(t, "user-456", req.UserId)
	assert.Equal(t, "Test Conversation", req.Title)
	assert.Equal(t, 10, req.Size)
	assert.Equal(t, 1, req.Page)
}

func TestListReq_Empty(t *testing.T) {
	t.Parallel()

	req := ListReq{}

	assert.Empty(t, req.AgentAPPKey)
	assert.Empty(t, req.UserId)
	assert.Empty(t, req.Title)
	assert.Equal(t, 0, req.Size)
	assert.Equal(t, 0, req.Page)
}

func TestListReq_WithPagination(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	req.Size = 20
	req.Page = 2

	offset := req.GetOffset()
	assert.Equal(t, 20, offset)
}

func TestListReq_WithDefaultPagination(t *testing.T) {
	t.Parallel()

	req := ListReq{}
	// PageSize has default values when Size is 0
	req.Size = 0
	req.Page = 0

	offset := req.GetOffset()
	assert.Equal(t, 0, offset)
}

func TestListReq_WithTitle(t *testing.T) {
	t.Parallel()

	titles := []string{
		"Test Conversation",
		"中文会话",
		"Conversation with numbers 123",
		"",
	}

	for _, title := range titles {
		req := ListReq{
			Title: title,
		}
		assert.Equal(t, title, req.Title)
	}
}

func TestListReq_WithAgentAPPKey(t *testing.T) {
	t.Parallel()

	keys := []string{
		"app-key-001",
		"agent-app-xyz",
		"key-中文-123",
	}

	for _, key := range keys {
		req := ListReq{
			AgentAPPKey: key,
		}
		assert.Equal(t, key, req.AgentAPPKey)
	}
}

func TestListReq_WithUserId(t *testing.T) {
	t.Parallel()

	userIDs := []string{
		"user-001",
		"user-xyz",
		"用户123",
	}

	for _, userID := range userIDs {
		req := ListReq{
			UserId: userID,
		}
		assert.Equal(t, userID, req.UserId)
	}
}

func TestListReq_PaginationEdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		page     int
		size     int
		expected int
	}{
		{
			name:     "first page",
			page:     1,
			size:     10,
			expected: 0,
		},
		{
			name:     "second page",
			page:     2,
			size:     10,
			expected: 10,
		},
		{
			name:     "large page number",
			page:     100,
			size:     20,
			expected: 1980,
		},
		{
			name:     "zero page",
			page:     0,
			size:     10,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := ListReq{}
			req.Page = tt.page
			req.Size = tt.size

			offset := req.GetOffset()
			assert.Equal(t, tt.expected, offset)
		})
	}
}

func TestListReq_EmbeddedPageSize(t *testing.T) {
	t.Parallel()

	req := ListReq{}

	// Verify that PageSize is embedded
	assert.IsType(t, req.Size, 0)
	assert.IsType(t, req.Page, 0)

	// Set and verify pagination values
	req.Size = 15
	req.Page = 3

	assert.Equal(t, 15, req.Size)
	assert.Equal(t, 3, req.Page)
	assert.Equal(t, 30, req.GetOffset())
}
