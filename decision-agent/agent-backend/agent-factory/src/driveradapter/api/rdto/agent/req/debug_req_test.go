package agentreq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDebugReq_StructFields(t *testing.T) {
	t.Parallel()

	input := DebugInput{
		Query: "Test query",
	}

	req := DebugReq{
		AgentID:         "agent-123",
		AgentVersion:    "v1.0.0",
		Input:           input,
		ConversationID:  "conv-456",
		ChatMode:        "debug",
		Stream:          true,
		IncStream:       false,
		UserID:          "user-789",
		Token:           "token-101",
		AgentAPPKey:     "app-202",
		ExecutorVersion: "v2",
	}

	assert.Equal(t, "agent-123", req.AgentID)
	assert.Equal(t, "v1.0.0", req.AgentVersion)
	assert.Equal(t, "Test query", req.Input.Query)
	assert.Equal(t, "conv-456", req.ConversationID)
	assert.Equal(t, "debug", req.ChatMode)
	assert.True(t, req.Stream)
	assert.False(t, req.IncStream)
	assert.Equal(t, "user-789", req.UserID)
	assert.Equal(t, "token-101", req.Token)
	assert.Equal(t, "app-202", req.AgentAPPKey)
	assert.Equal(t, "v2", req.ExecutorVersion)
}

func TestDebugReq_Empty(t *testing.T) {
	t.Parallel()

	req := DebugReq{}

	assert.Empty(t, req.AgentID)
	assert.Empty(t, req.AgentVersion)
	assert.Empty(t, req.ConversationID)
	assert.Empty(t, req.ChatMode)
	assert.False(t, req.Stream)
	assert.False(t, req.IncStream)
	assert.Empty(t, req.UserID)
	assert.Empty(t, req.Token)
	assert.Empty(t, req.AgentAPPKey)
	assert.Empty(t, req.ExecutorVersion)
}

func TestDebugInput_StructFields(t *testing.T) {
	t.Parallel()

	input := DebugInput{
		Query: "Test query",
		CustomQuerys: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	assert.Equal(t, "Test query", input.Query)
	assert.NotNil(t, input.CustomQuerys)
	assert.Equal(t, "value1", input.CustomQuerys["key1"])
	assert.Equal(t, 123, input.CustomQuerys["key2"])
}

func TestDebugInput_Empty(t *testing.T) {
	t.Parallel()

	input := DebugInput{}

	assert.Empty(t, input.Query)
	assert.Nil(t, input.CustomQuerys)
	assert.Nil(t, input.History)
}

func TestDebugReq_WithStreamOptions(t *testing.T) {
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

			req := DebugReq{
				Stream:    tt.stream,
				IncStream: tt.incStream,
			}
			assert.Equal(t, tt.stream, req.Stream)
			assert.Equal(t, tt.incStream, req.IncStream)
		})
	}
}

func TestDebugReq_WithSelectedFiles(t *testing.T) {
	t.Parallel()

	files := []SelectedFile{
		{FileName: "file1.pdf"},
		{FileName: "file2.txt"},
	}

	req := DebugReq{
		SelectedFiles: files,
	}

	assert.Len(t, req.SelectedFiles, 2)
	assert.Equal(t, "file1.pdf", req.SelectedFiles[0].FileName)
	assert.Equal(t, "file2.txt", req.SelectedFiles[1].FileName)
}

func TestDebugReq_WithExecutorVersion(t *testing.T) {
	t.Parallel()

	versions := []string{
		"v1",
		"v2",
		"",
	}

	for _, version := range versions {
		req := DebugReq{
			ExecutorVersion: version,
		}
		assert.Equal(t, version, req.ExecutorVersion)
	}
}

func TestDebugReq_WithAgentID(t *testing.T) {
	t.Parallel()

	agentIDs := []string{
		"agent-001",
		"test-agent",
		"AGENT-ABC-123",
		"",
	}

	for _, agentID := range agentIDs {
		req := DebugReq{
			AgentID: agentID,
		}
		assert.Equal(t, agentID, req.AgentID)
	}
}

func TestDebugInput_WithQuery(t *testing.T) {
	t.Parallel()

	queries := []string{
		"Test query 1",
		"Test query 2",
		"",
		"Complex query with special chars: @#$%",
	}

	for _, query := range queries {
		input := DebugInput{
			Query: query,
		}
		assert.Equal(t, query, input.Query)
	}
}

func TestDebugInput_WithCustomQuerys(t *testing.T) {
	t.Parallel()

	input := DebugInput{
		CustomQuerys: map[string]interface{}{
			"string_key": "string_value",
			"number_key": 42,
			"bool_key":   true,
			"null_key":   nil,
			"object_key": map[string]string{"nested": "value"},
		},
	}

	assert.Equal(t, "string_value", input.CustomQuerys["string_key"])
	assert.Equal(t, 42, input.CustomQuerys["number_key"])
	assert.True(t, input.CustomQuerys["bool_key"].(bool))
	assert.Nil(t, input.CustomQuerys["null_key"])
}

func TestDebugReq_WithConversationID(t *testing.T) {
	t.Parallel()

	convIDs := []string{
		"conv-001",
		"test-conversation",
		"",
	}

	for _, convID := range convIDs {
		req := DebugReq{
			ConversationID: convID,
		}
		assert.Equal(t, convID, req.ConversationID)
	}
}
