package conversationeo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestMessagePO_NewMessage(t *testing.T) {
	t.Parallel()

	content := "Hello World"
	msg := &dapo.ConversationMsgPO{
		ID:      "msg-123",
		Content: &content,
	}

	assert.NotNil(t, msg)
	assert.Equal(t, "msg-123", msg.ID)
	assert.Equal(t, "Hello World", *msg.Content)
}

func TestMessagePO_EmptyMessage(t *testing.T) {
	t.Parallel()

	msg := &dapo.ConversationMsgPO{}

	assert.NotNil(t, msg)
	assert.Empty(t, msg.ID)
	assert.Nil(t, msg.Content)
}
