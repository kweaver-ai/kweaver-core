package agentresp

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationSessionInitResp_StructFields(t *testing.T) {
	t.Parallel()

	resp := ConversationSessionInitResp{
		ConversationSessionID: "session-123",
		TTL:                   3600,
	}

	assert.Equal(t, "session-123", resp.ConversationSessionID)
	assert.Equal(t, 3600, resp.TTL)
}

func TestConversationSessionInitResp_Empty(t *testing.T) {
	t.Parallel()

	resp := ConversationSessionInitResp{}

	assert.Empty(t, resp.ConversationSessionID)
	assert.Zero(t, resp.TTL)
}

func TestConversationSessionInitResp_WithSessionID(t *testing.T) {
	t.Parallel()

	sessionIDs := []string{
		"session-001",
		"test-session",
		"uuid-1234-5678-9012",
		"",
	}

	for _, sessionID := range sessionIDs {
		resp := ConversationSessionInitResp{
			ConversationSessionID: sessionID,
		}
		assert.Equal(t, sessionID, resp.ConversationSessionID)
	}
}

func TestConversationSessionInitResp_WithTTL(t *testing.T) {
	t.Parallel()

	ttls := []int{
		0,
		60,
		300,
		600,
		1800,
		3600,
		7200,
		86400,
	}

	for _, ttl := range ttls {
		resp := ConversationSessionInitResp{
			TTL: ttl,
		}
		assert.Equal(t, ttl, resp.TTL)
	}
}

func TestConversationSessionInitResp_WithBothFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		conversationSessionID string
		ttl                   int
	}{
		{
			name:                  "short TTL",
			conversationSessionID: "session-1",
			ttl:                   60,
		},
		{
			name:                  "long TTL",
			conversationSessionID: "session-2",
			ttl:                   86400,
		},
		{
			name:                  "zero TTL",
			conversationSessionID: "session-3",
			ttl:                   0,
		},
		{
			name:                  "hour TTL",
			conversationSessionID: "session-4",
			ttl:                   3600,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			resp := ConversationSessionInitResp{
				ConversationSessionID: tt.conversationSessionID,
				TTL:                   tt.ttl,
			}

			assert.Equal(t, tt.conversationSessionID, resp.ConversationSessionID)
			assert.Equal(t, tt.ttl, resp.TTL)
		})
	}
}
