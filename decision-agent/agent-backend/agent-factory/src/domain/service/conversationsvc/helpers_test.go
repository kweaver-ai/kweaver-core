package conversationsvc

import (
	"context"
	"testing"

	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/persistence/dapo"
	"github.com/stretchr/testify/assert"
)

func TestGetID(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	tests := []struct {
		name                     string
		messages                 []*dapo.ConversationMsgPO
		regenerateUserMsgID      string
		regenerateAssistantMsgID string
		expectUserMsgID          string
		expectAssistantMsgID     string
		shouldPanic              bool
	}{
		{
			name:                     "both empty - return empty",
			messages:                 []*dapo.ConversationMsgPO{},
			regenerateUserMsgID:      "",
			regenerateAssistantMsgID: "",
			expectUserMsgID:          "",
			expectAssistantMsgID:     "",
			shouldPanic:              false,
		},
		{
			name: "find user message with assistant following",
			messages: []*dapo.ConversationMsgPO{
				{ID: "msg1", Role: "user"},
				{ID: "msg2", Role: "assistant"},
				{ID: "msg3", Role: "user"},
			},
			regenerateUserMsgID:      "msg1",
			regenerateAssistantMsgID: "",
			expectUserMsgID:          "msg1",
			expectAssistantMsgID:     "msg2",
			shouldPanic:              false,
		},
		{
			name: "find assistant message",
			messages: []*dapo.ConversationMsgPO{
				{ID: "msg1", Role: "user", ReplyID: ""},
				{ID: "msg2", Role: "assistant"},
			},
			regenerateUserMsgID:      "",
			regenerateAssistantMsgID: "msg2",
			expectUserMsgID:          "", // ReplyID is empty
			expectAssistantMsgID:     "msg2",
			shouldPanic:              false,
		},
		{
			name: "user message not found - return empty",
			messages: []*dapo.ConversationMsgPO{
				{ID: "msg1", Role: "user"},
				{ID: "msg2", Role: "assistant"},
			},
			regenerateUserMsgID:      "msg999",
			regenerateAssistantMsgID: "",
			expectUserMsgID:          "",
			expectAssistantMsgID:     "",
			shouldPanic:              false,
		},
		{
			name: "only user ID provided",
			messages: []*dapo.ConversationMsgPO{
				{ID: "msg1", Role: "user"},
			},
			regenerateUserMsgID:      "msg1",
			regenerateAssistantMsgID: "",
			// This will panic because messages[index+1] doesn't exist
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.shouldPanic {
				assert.Panics(t, func() {
					GetID(ctx, tt.messages, tt.regenerateUserMsgID, tt.regenerateAssistantMsgID)
				})
			} else {
				userMsgID, assistantMsgID := GetID(ctx, tt.messages, tt.regenerateUserMsgID, tt.regenerateAssistantMsgID)

				assert.Equal(t, tt.expectUserMsgID, userMsgID)
				assert.Equal(t, tt.expectAssistantMsgID, assistantMsgID)
			}
		})
	}
}

func TestBuildWorkspaceContextMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		conversationID string
		userID         string
		selectedFiles  []agentreq.SelectedFile
		wantContains   []string
	}{
		{
			name:           "empty files",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles:  []agentreq.SelectedFile{},
			wantContains:   []string{},
		},
		{
			name:           "single file",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-123/uploads/temparea/data.csv"},
			},
			wantContains: []string{
				cutil.GetSandboxSessionID(),
				"data.csv",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildWorkspaceContextMessage(tt.conversationID, tt.userID, tt.selectedFiles)

			if len(tt.wantContains) == 0 {
				assert.Empty(t, result)
			} else {
				for _, substr := range tt.wantContains {
					assert.Contains(t, result, substr)
				}
			}
		})
	}
}
