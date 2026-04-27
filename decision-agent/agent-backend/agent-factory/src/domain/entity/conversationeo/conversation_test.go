package conversationeo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestConversation_New(t *testing.T) {
	t.Parallel()

	conversation := &Conversation{
		ConversationPO: &dapo.ConversationPO{},
		Messages:       []*dapo.ConversationMsgPO{},
	}

	assert.NotNil(t, conversation)
	assert.NotNil(t, conversation.ConversationPO)
	assert.NotNil(t, conversation.Messages)
}

func TestConversation_NilPO(t *testing.T) {
	t.Parallel()

	conversation := &Conversation{
		ConversationPO: nil,
		Messages:       nil,
	}

	assert.NotNil(t, conversation)
	assert.Nil(t, conversation.ConversationPO)
	assert.Nil(t, conversation.Messages)
}

func TestConversation_EmptyMessages(t *testing.T) {
	t.Parallel()

	conversation := &Conversation{
		ConversationPO: &dapo.ConversationPO{},
		Messages:       []*dapo.ConversationMsgPO{},
	}

	assert.NotNil(t, conversation.Messages)
	assert.Empty(t, conversation.Messages)
}
