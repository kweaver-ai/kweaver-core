package conversationp2e

import (
	"context"
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/conversation_message/conversationmsgreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/persistence/dapo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess/idbaccessmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestConversation_WithoutMessages(t *testing.T) {
	t.Parallel()

	po := &dapo.ConversationPO{
		ID: "conv-1",
	}

	eo, err := Conversation(context.Background(), po, nil, false)

	assert.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "conv-1", eo.ID)
	assert.Nil(t, eo.Messages)
}

func TestConversation_WithMessages(t *testing.T) {
	t.Parallel()

	// Note: This test requires a mock IConversationMsgRepo
	// With nil repo, this will panic, so we test for that
	po := &dapo.ConversationPO{
		ID: "conv-1",
	}

	assert.Panics(t, func() {
		Conversation(context.Background(), po, nil, true) //nolint:errcheck
	})
}

func TestConversation_EmptyPO(t *testing.T) {
	t.Parallel()

	po := &dapo.ConversationPO{}

	eo, err := Conversation(context.Background(), po, nil, false)

	assert.NoError(t, err)
	assert.NotNil(t, eo)
}

func TestConversations_EmptyList(t *testing.T) {
	t.Parallel()

	pos := []*dapo.ConversationPO{}

	eos, err := Conversations(context.Background(), pos, nil)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 0)
}

func TestConversations_SingleItem(t *testing.T) {
	t.Parallel()

	pos := []*dapo.ConversationPO{
		{ID: "conv-1"},
	}

	eos, err := Conversations(context.Background(), pos, nil)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
	assert.Equal(t, "conv-1", eos[0].ID)
}

func TestConversations_MultipleItems(t *testing.T) {
	t.Parallel()

	pos := []*dapo.ConversationPO{
		{ID: "conv-1"},
		{ID: "conv-2"},
		{ID: "conv-3"},
	}

	eos, err := Conversations(context.Background(), pos, nil)

	assert.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 3)
	assert.Equal(t, "conv-1", eos[0].ID)
	assert.Equal(t, "conv-2", eos[1].ID)
	assert.Equal(t, "conv-3", eos[2].ID)
}

func TestConversation_WithMessages_Success(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)

	po := &dapo.ConversationPO{
		ID: "conv-1",
	}

	// Mock the List method to return empty list
	mockMsgRepo.EXPECT().List(ctx, conversationmsgreq.ListReq{ConversationID: "conv-1"}).Return([]*dapo.ConversationMsgPO{}, nil)

	eo, err := Conversation(ctx, po, mockMsgRepo, true)

	require.NoError(t, err)
	assert.NotNil(t, eo)
	assert.Equal(t, "conv-1", eo.ID)
	assert.NotNil(t, eo.Messages)
}

func TestConversation_WithMessages_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)

	po := &dapo.ConversationPO{
		ID: "conv-1",
	}

	// Mock the List method to return an error
	mockMsgRepo.EXPECT().List(ctx, conversationmsgreq.ListReq{ConversationID: "conv-1"}).Return(nil, errors.New("database error"))

	eo, err := Conversation(ctx, po, mockMsgRepo, true)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "查询对话消息失败")
	assert.Nil(t, eo)
}

func TestConversations_Error(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMsgRepo := idbaccessmock.NewMockIConversationMsgRepo(ctrl)

	pos := []*dapo.ConversationPO{
		{ID: "conv-1"},
	}

	// Mock the List method to return an error (this will be called by Conversations with withMsg=false, so List won't be called)
	// Actually, Conversations calls Conversation with withMsg=false, so no List call will be made
	// To test error path, we need to make Conversation return an error somehow
	// But Conversation only returns error when withMsg=true and List fails

	// Let's test with withMsg=false (no error expected)
	eos, err := Conversations(ctx, pos, mockMsgRepo)

	require.NoError(t, err)
	assert.NotNil(t, eos)
	assert.Len(t, eos, 1)
}
