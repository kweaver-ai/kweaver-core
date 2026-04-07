package agentreq

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
)

func TestSelectedFile_StructFields(t *testing.T) {
	t.Parallel()

	file := SelectedFile{
		FileName: "test-file.pdf",
	}

	assert.Equal(t, "test-file.pdf", file.FileName)
}

func TestSelectedFile_Empty(t *testing.T) {
	t.Parallel()

	file := SelectedFile{}

	assert.Empty(t, file.FileName)
}

func TestChatReq_StructFields(t *testing.T) {
	t.Parallel()

	req := &ChatReq{
		AgentAPPKey:    "app-123",
		AgentID:        "agent-456",
		AgentKey:       "key-789",
		AgentVersion:   "v1.0.0",
		ConversationID: "conv-101",
		Query:          "Test query",
		ChatMode:       "normal",
		Stream:         true,
		IncStream:      false,
	}

	assert.Equal(t, "app-123", req.AgentAPPKey)
	assert.Equal(t, "agent-456", req.AgentID)
	assert.Equal(t, "key-789", req.AgentKey)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "conv-101", req.ConversationID)
	assert.Equal(t, "Test query", req.Query)
	assert.Equal(t, "normal", req.ChatMode)
	assert.True(t, req.Stream)
	assert.False(t, req.IncStream)
}

func TestChatReq_Empty(t *testing.T) {
	t.Parallel()

	req := &ChatReq{}

	assert.Empty(t, req.AgentAPPKey)
	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentKey)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.Query)
	assert.Empty(t, req.ChatMode)
	assert.False(t, req.Stream)
	assert.False(t, req.IncStream)
}

func TestChatReq_WithStreamOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		stream    bool
		incStream bool
	}{
		{
			name:      "stream only",
			stream:    true,
			incStream: false,
		},
		{
			name:      "incremental stream",
			stream:    true,
			incStream: true,
		},
		{
			name:      "no stream",
			stream:    false,
			incStream: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &ChatReq{
				Stream:    tt.stream,
				IncStream: tt.incStream,
			}
			assert.Equal(t, tt.stream, req.Stream)
			assert.Equal(t, tt.incStream, req.IncStream)
		})
	}
}

func TestChatReq_WithSelectedFiles(t *testing.T) {
	t.Parallel()

	files := []SelectedFile{
		{FileName: "file1.pdf"},
		{FileName: "file2.txt"},
	}

	req := &ChatReq{
		SelectedFiles: files,
	}

	assert.Len(t, req.SelectedFiles, 2)
	assert.Equal(t, "file1.pdf", req.SelectedFiles[0].FileName)
	assert.Equal(t, "file2.txt", req.SelectedFiles[1].FileName)
}

func TestChatReq_WithEmptySelectedFiles(t *testing.T) {
	t.Parallel()

	req := &ChatReq{
		SelectedFiles: []SelectedFile{},
	}

	assert.NotNil(t, req.SelectedFiles)
	assert.Len(t, req.SelectedFiles, 0)
}

func TestChatReq_WithNilSelectedFiles(t *testing.T) {
	t.Parallel()

	req := &ChatReq{
		SelectedFiles: nil,
	}

	assert.Nil(t, req.SelectedFiles)
}

func TestInternalParam_StructFields(t *testing.T) {
	t.Parallel()

	param := InternalParam{
		UserID:                "user-123",
		Token:                 "token-456",
		UserMessageID:         "msg-789",
		AssistantMessageID:    "assistant-101",
		AssistantMessageIndex: 0,
		VisitorType:           constant.RealName,
		CallType:              constant.DebugChat,
		ReqStartTime:          1234567890,
		TTFT:                  1000,
		XAccountID:            "account-202",
		XAccountType:          cenum.AccountTypeUser,
		XBusinessDomainID:     "domain-303",
		SandboxSessionID:      "sandbox-404",
	}

	assert.Equal(t, "user-123", param.UserID)
	assert.Equal(t, "token-456", param.Token)
	assert.Equal(t, "msg-789", param.UserMessageID)
	assert.Equal(t, "assistant-101", param.AssistantMessageID)
	assert.Equal(t, 0, param.AssistantMessageIndex)
	assert.Equal(t, constant.RealName, param.VisitorType)
	assert.Equal(t, constant.DebugChat, param.CallType)
	assert.Equal(t, int64(1234567890), param.ReqStartTime)
	assert.Equal(t, int64(1000), param.TTFT)
	assert.Equal(t, "account-202", param.XAccountID)
	assert.Equal(t, cenum.AccountTypeUser, param.XAccountType)
	assert.Equal(t, "domain-303", param.XBusinessDomainID)
	assert.Equal(t, "sandbox-404", param.SandboxSessionID)
}

func TestInternalParam_Empty(t *testing.T) {
	t.Parallel()

	param := InternalParam{}

	assert.Empty(t, param.UserID)
	assert.Empty(t, param.Token)
	assert.Empty(t, param.UserMessageID)
	assert.Empty(t, param.AssistantMessageID)
	assert.Equal(t, 0, param.AssistantMessageIndex)
	assert.Equal(t, int64(0), param.ReqStartTime)
	assert.Equal(t, int64(0), param.TTFT)
	assert.Empty(t, param.XAccountID)
	assert.Empty(t, param.XBusinessDomainID)
	assert.Empty(t, param.SandboxSessionID)
}

func TestChatReq_WithExecutorVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"v1",
		"v2",
		"",
	}

	for _, version := range versions {
		req := &ChatReq{
			ExecutorVersion: version,
		}
		assert.Equal(t, version, req.ExecutorVersion)
	}
}

func TestChatReq_WithModelName(t *testing.T) {
	t.Parallel()

	models := []string{
		"gpt-4",
		"gpt-3.5-turbo",
		"claude-3",
		"",
	}

	for _, model := range models {
		req := &ChatReq{
			ModelName: model,
		}
		assert.Equal(t, model, req.ModelName)
	}
}
