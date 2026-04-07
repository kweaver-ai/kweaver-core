package agentsvc

import (
	"testing"

	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/stretchr/testify/assert"
)

func TestBuildUserQuery(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		originalQuery  string
		conversationID string
		selectedFiles  []agentreq.SelectedFile
		wantContains   []string
	}{
		{
			name:           "empty files",
			originalQuery:  "What is the weather?",
			conversationID: "conv-123",
			selectedFiles:  []agentreq.SelectedFile{},
			wantContains:   []string{"What is the weather?"}, // Just returns original query
		},
		{
			name:           "with files",
			originalQuery:  "Analyze the data",
			conversationID: "conv-123",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-123/uploads/data.csv"},
			},
			wantContains: []string{
				"/workspace/conv-123/uploads/",
				"data.csv",
				"Analyze the data",
			},
		},
		{
			name:           "multiple files",
			originalQuery:  "Compare these files",
			conversationID: "conv-456",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-456/uploads/file1.csv"},
				{FileName: "/workspace/conv-456/uploads/file2.csv"},
			},
			wantContains: []string{
				"Compare these files",
				"file1.csv",
				"file2.csv",
			},
		},
		{
			name:           "empty original query with files",
			originalQuery:  "",
			conversationID: "conv-789",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-789/uploads/data.csv"},
			},
			wantContains: []string{"data.csv"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := buildUserQuery(tt.originalQuery, tt.conversationID, tt.selectedFiles)

			for _, substr := range tt.wantContains {
				assert.Contains(t, result, substr)
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
			name:           "with files",
			conversationID: "conv-123",
			userID:         "user-456",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-123/uploads/data.csv"},
			},
			wantContains: []string{
				"/workspace/conv-123/uploads/",
				"data.csv",
				cutil.GetSandboxSessionID(),
			},
		},
		{
			name:           "multiple files",
			conversationID: "conv-456",
			userID:         "user-789",
			selectedFiles: []agentreq.SelectedFile{
				{FileName: "/workspace/conv-456/uploads/file1.txt"},
				{FileName: "/workspace/conv-456/uploads/file2.txt"},
			},
			wantContains: []string{
				"file1.txt",
				"file2.txt",
				cutil.GetSandboxSessionID(),
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
