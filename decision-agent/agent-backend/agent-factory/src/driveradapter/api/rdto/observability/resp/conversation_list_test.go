package observabilityresp

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation/conversationresp"
	"github.com/stretchr/testify/assert"
)

func TestObservabilityConversationDetail_StructFields(t *testing.T) {
	t.Parallel()

	detail := ObservabilityConversationDetail{
		SessionCount: 5,
	}

	assert.Equal(t, 5, detail.SessionCount)
}

func TestObservabilityConversationDetail_Empty(t *testing.T) {
	t.Parallel()

	detail := ObservabilityConversationDetail{}

	assert.Equal(t, 0, detail.SessionCount)
}

func TestObservabilityConversationDetail_WithSessionCount(t *testing.T) {
	t.Parallel()

	counts := []int{0, 1, 5, 10, 100}

	for _, count := range counts {
		detail := ObservabilityConversationDetail{
			SessionCount: count,
		}
		assert.Equal(t, count, detail.SessionCount)
	}
}

func TestConversationListResp_StructFields(t *testing.T) {
	t.Parallel()

	entries := []ObservabilityConversationDetail{
		{SessionCount: 1},
		{SessionCount: 2},
	}

	resp := ConversationListResp{
		Entries:    entries,
		TotalCount: 10,
	}

	assert.Len(t, resp.Entries, 2)
	assert.Equal(t, int64(10), resp.TotalCount)
	assert.Equal(t, 1, resp.Entries[0].SessionCount)
	assert.Equal(t, 2, resp.Entries[1].SessionCount)
}

func TestConversationListResp_Empty(t *testing.T) {
	t.Parallel()

	resp := ConversationListResp{}

	assert.Nil(t, resp.Entries)
	assert.Equal(t, int64(0), resp.TotalCount)
}

func TestConversationListResp_WithEmptyEntries(t *testing.T) {
	t.Parallel()

	resp := ConversationListResp{
		Entries:    []ObservabilityConversationDetail{},
		TotalCount: 0,
	}

	assert.NotNil(t, resp.Entries)
	assert.Len(t, resp.Entries, 0)
	assert.Equal(t, int64(0), resp.TotalCount)
}

func TestConversationListResp_WithSingleEntry(t *testing.T) {
	t.Parallel()

	entries := []ObservabilityConversationDetail{
		{SessionCount: 42},
	}

	resp := ConversationListResp{
		Entries:    entries,
		TotalCount: 1,
	}

	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, 42, resp.Entries[0].SessionCount)
}

func TestConversationListResp_WithConversationDetail(t *testing.T) {
	t.Parallel()

	convDetail := conversationresp.ConversationDetail{}

	entries := []ObservabilityConversationDetail{
		{
			Conversation: convDetail,
			SessionCount: 3,
		},
	}

	resp := ConversationListResp{
		Entries:    entries,
		TotalCount: 1,
	}

	assert.Len(t, resp.Entries, 1)
	assert.Equal(t, 3, resp.Entries[0].SessionCount)
}

func TestConversationListResp_WithTotalCount(t *testing.T) {
	t.Parallel()

	totalCounts := []int64{0, 1, 10, 100, 1000}

	for _, totalCount := range totalCounts {
		resp := ConversationListResp{
			TotalCount: totalCount,
		}
		assert.Equal(t, totalCount, resp.TotalCount)
	}
}
