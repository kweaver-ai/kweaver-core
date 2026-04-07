package util

import (
	"testing"

	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/stretchr/testify/assert"
)

func TestBuildWorkspaceContextMessage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		conversationID string
		userID         string
		selectedFiles  []agentreq.SelectedFile
		wantEmpty      bool
		wantContains   []string
	}{
		{
			name:           "empty selected files returns empty string",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles:  []agentreq.SelectedFile{},
			wantEmpty:      true,
		},
		{
			name:           "nil selected files returns empty string",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles:  nil,
			wantEmpty:      true,
		},
		{
			name:           "single file with simple name",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-123/uploads/temparea/data.csv"},
			},
			wantEmpty: false,
			wantContains: []string{
				"Current workspace path: /workspace/conv-123/uploads/temparea/",
				"Sandbox Session ID: " + cutil.GetSandboxSessionID(),
				"- data.csv (/workspace/conv-123/uploads/temparea/data.csv)",
				"System auto-generated context - not user query",
			},
		},
		{
			name:           "multiple files",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-123/uploads/temparea/data.csv"},
				{FileName: "/workspace/conv-123/uploads/temparea/config.json"},
				{FileName: "/workspace/conv-123/uploads/temparea/output.txt"},
			},
			wantEmpty: false,
			wantContains: []string{
				"- data.csv (/workspace/conv-123/uploads/temparea/data.csv)",
				"- config.json (/workspace/conv-123/uploads/temparea/config.json)",
				"- output.txt (/workspace/conv-123/uploads/temparea/output.txt)",
			},
		},
		{
			name:           "file with complex path",
			conversationID: "conv-abc",
			userID:         "user-xyz",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-abc/uploads/temparea/subdir/nested/file.pdf"},
			},
			wantEmpty: false,
			wantContains: []string{
				"- file.pdf (/workspace/conv-abc/uploads/temparea/subdir/nested/file.pdf)",
				cutil.GetSandboxSessionID(),
			},
		},
		{
			name:           "file with special characters in name",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-123/uploads/temparea/file-with-dashes.txt"},
			},
			wantEmpty: false,
			wantContains: []string{
				"- file-with-dashes.txt",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := BuildWorkspaceContextMessage(tt.conversationID, tt.userID, tt.selectedFiles)

			if tt.wantEmpty {
				assert.Empty(t, got)
			} else {
				assert.NotEmpty(t, got)

				for _, substr := range tt.wantContains {
					assert.Contains(t, got, substr)
				}
			}
		})
	}
}
