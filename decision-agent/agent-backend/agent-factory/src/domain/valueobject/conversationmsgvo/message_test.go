package conversationmsgvo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentconfigvo"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/stretchr/testify/assert"
)

func TestMessage_IsInterrupted(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		message  *Message
		expected bool
	}{
		{
			name:     "no ext",
			message:  &Message{},
			expected: false,
		},
		{
			name:     "ext is nil",
			message:  &Message{Ext: nil},
			expected: false,
		},
		{
			name: "ext without interrupt info",
			message: &Message{
				Ext: &MessageExt{
					TotalTime: 1.5,
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.message.IsInterrupted())
		})
	}
}

func TestMessage_Fields(t *testing.T) {
	t.Parallel()

	msg := &Message{
		ID:             "msg-1",
		ConversationID: "conv-1",
		Role:           cdaenum.ConversationMsgRole("user"),
		Content:        "test content",
		ContentType:    chatresenum.AnswerTypePrompt,
		Status:         cdaenum.ConversationMsgStatus("received"),
		ReplyID:        "reply-1",
		AgentInfo:      valueobject.AgentInfo{AgentID: "agent-1", AgentName: "Test Agent"},
		Index:          1,
	}

	assert.Equal(t, "msg-1", msg.ID)
	assert.Equal(t, "conv-1", msg.ConversationID)
	assert.Equal(t, cdaenum.ConversationMsgRole("user"), msg.Role)
	assert.Equal(t, "test content", msg.Content)
	assert.Equal(t, chatresenum.AnswerTypePrompt, msg.ContentType)
	assert.Equal(t, cdaenum.ConversationMsgStatus("received"), msg.Status)
	assert.Equal(t, "reply-1", msg.ReplyID)
	assert.Equal(t, "agent-1", msg.AgentInfo.AgentID)
	assert.Equal(t, 1, msg.Index)
}

func TestUserContent_Fields(t *testing.T) {
	t.Parallel()

	content := UserContent{
		Text: "Hello, AI!",
		SelectedFiles: []agentreq.SelectedFile{
			{FileName: "/workspace/file1.csv"},
		},
	}

	assert.Equal(t, "Hello, AI!", content.Text)
	assert.Len(t, content.SelectedFiles, 1)
	assert.Equal(t, "/workspace/file1.csv", content.SelectedFiles[0].FileName)
}

func TestAssistantContent_Fields(t *testing.T) {
	t.Parallel()

	content := AssistantContent{
		FinalAnswer: FinalAnswer{
			Query:  "What is AI?",
			Answer: Answer{Text: "AI is artificial intelligence."},
		},
		MiddleAnswer: &MiddleAnswer{},
	}

	assert.Equal(t, "What is AI?", content.FinalAnswer.Query)
	assert.Equal(t, "AI is artificial intelligence.", content.FinalAnswer.Answer.Text)
	assert.NotNil(t, content.MiddleAnswer)
}

func TestFinalAnswer_Fields(t *testing.T) {
	t.Parallel()

	answer := FinalAnswer{
		Query: "Test query",
		Answer: Answer{
			Text: "Test answer",
		},
		Thinking: "Thinking process",
		SelectedFiles: []agentreq.SelectedFile{
			{FileName: "/workspace/data.csv"},
		},
		SkillProcess: []*SkillsProcessItem{
			{
				AgentName: "SearchAgent",
				Text:      "Searching...",
				Status:    "completed",
			},
		},
		OutputVariablesConfig: &agentconfigvo.Variable{
			AnswerVar: "answer",
			OtherVars: []string{"var1", "var2"},
		},
	}

	assert.Equal(t, "Test query", answer.Query)
	assert.Equal(t, "Test answer", answer.Answer.Text)
	assert.Equal(t, "Thinking process", answer.Thinking)
	assert.Len(t, answer.SelectedFiles, 1)
	assert.Len(t, answer.SkillProcess, 1)
	assert.Equal(t, "SearchAgent", answer.SkillProcess[0].AgentName)
	assert.NotNil(t, answer.OutputVariablesConfig)
	assert.Equal(t, "answer", answer.OutputVariablesConfig.AnswerVar)
}

func TestSkillsProcessItem_Fields(t *testing.T) {
	t.Parallel()

	item := &SkillsProcessItem{
		AgentName:    "TestAgent",
		Text:         "Agent response",
		Status:       "completed",
		Type:         "skill",
		Thinking:     "Thinking...",
		InputMessage: "Input",
		Interrupted:  false,
		RelatedQueries: []*RelatedQuestion{
			{Query: "Related question 1"},
			{Query: "Related question 2"},
		},
	}

	assert.Equal(t, "TestAgent", item.AgentName)
	assert.Equal(t, "Agent response", item.Text)
	assert.Equal(t, "completed", item.Status)
	assert.Equal(t, "skill", item.Type)
	assert.Equal(t, "Thinking...", item.Thinking)
	assert.Equal(t, "Input", item.InputMessage)
	assert.False(t, item.Interrupted)
	assert.Len(t, item.RelatedQueries, 2)
	assert.Equal(t, "Related question 1", item.RelatedQueries[0].Query)
}

func TestRelatedQuestion_Fields(t *testing.T) {
	t.Parallel()

	q := RelatedQuestion{Query: "What is machine learning?"}
	assert.Equal(t, "What is machine learning?", q.Query)
}

func TestMiddleAnswer_Fields(t *testing.T) {
	t.Parallel()

	answer := MiddleAnswer{
		Progress: []*agentrespvo.Progress{
			{Stage: "llm"},
		},
		DocRetrieval:   &agentrespvo.DocRetrievalField{},
		GraphRetrieval: nil,
		OtherVariables: map[string]interface{}{
			"var1": "value1",
		},
	}

	assert.Len(t, answer.Progress, 1)
	assert.NotNil(t, answer.DocRetrieval)
	assert.Nil(t, answer.GraphRetrieval)
	assert.NotNil(t, answer.OtherVariables)
	assert.Equal(t, "value1", answer.OtherVariables["var1"])
}

func TestAnswer_Fields(t *testing.T) {
	t.Parallel()

	answer := Answer{
		Text:  "This is the answer",
		Cites: []string{"cite1", "cite2"},
		Ask:   "Follow-up question?",
	}

	assert.Equal(t, "This is the answer", answer.Text)
	assert.NotNil(t, answer.Cites)
	assert.NotNil(t, answer.Ask)
}
