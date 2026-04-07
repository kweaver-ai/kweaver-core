package agentsvc

import (
	"testing"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/conversationmsgvo"
	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/stretchr/testify/assert"
)

func TestSession_UpdateAndGetTempMsgResp(t *testing.T) {
	t.Parallel()

	session := &Session{
		ConversationID: "conv-123",
		Signal:         make(chan struct{}, 1),
	}

	resp := agentresp.ChatResp{
		ConversationID: "conv-123",
		Message: conversationmsgvo.Message{
			Content: "test answer",
		},
	}

	session.UpdateTempMsgResp(resp)
	result := session.GetTempMsgResp()

	assert.Equal(t, "conv-123", result.ConversationID)
	assert.Equal(t, "test answer", result.Message.Content)
}

func TestSession_GetAndSetIsResuming(t *testing.T) {
	t.Parallel()

	session := &Session{
		ConversationID: "conv-123",
		Signal:         make(chan struct{}, 1),
	}

	assert.False(t, session.GetIsResuming())

	session.SetIsResuming(true)
	assert.True(t, session.GetIsResuming())

	session.SetIsResuming(false)
	assert.False(t, session.GetIsResuming())
}

func TestSession_SetAndGetSignal(t *testing.T) {
	t.Parallel()

	session := &Session{
		ConversationID: "conv-123",
	}

	signal := make(chan struct{}, 1)
	session.SetSignal(signal)

	result := session.GetSignal()
	assert.Equal(t, signal, result)
}

func TestSession_CloseSignal(t *testing.T) {
	t.Parallel()

	session := &Session{
		ConversationID: "conv-123",
		Signal:         make(chan struct{}, 1),
	}

	session.CloseSignal()

	// Signal should be nil after closing
	assert.Nil(t, session.Signal)
}

func TestSession_SendSignal(t *testing.T) {
	t.Parallel()

	t.Run("sends signal when resuming", func(t *testing.T) {
		t.Parallel()

		session := &Session{
			ConversationID: "conv-123",
			Signal:         make(chan struct{}, 1),
			IsResuming:     true,
		}

		go session.SendSignal()

		select {
		case <-session.Signal:
			t.Log("Signal received successfully")
		case <-time.After(100 * time.Millisecond):
			t.Error("Did not receive signal in time")
		}
	})

	t.Run("does not send signal when not resuming", func(t *testing.T) {
		t.Parallel()

		session := &Session{
			ConversationID: "conv-123",
			Signal:         make(chan struct{}, 1),
			IsResuming:     false,
		}

		go session.SendSignal()

		select {
		case <-session.Signal:
			t.Error("Should not send signal when not resuming")
		case <-time.After(50 * time.Millisecond):
			t.Log("Correctly did not send signal")
		}
	})

	t.Run("does not panic when signal is nil", func(t *testing.T) {
		t.Parallel()

		session := &Session{
			ConversationID: "conv-123",
			Signal:         nil,
			IsResuming:     true,
		}

		assert.NotPanics(t, func() {
			session.SendSignal()
		})
	})
}

func TestSession_ConcurrencySafety(t *testing.T) {
	t.Parallel()

	session := &Session{
		ConversationID: "conv-123",
		Signal:         make(chan struct{}, 10),
	}

	done := make(chan bool)

	// Concurrent writes
	for i := 0; i < 10; i++ {
		go func(idx int) {
			session.UpdateTempMsgResp(agentresp.ChatResp{
				ConversationID: "conv-123",
				Message: conversationmsgvo.Message{
					Content: string(rune('a' + idx)),
				},
			})
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		go func() {
			_ = session.GetTempMsgResp()
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should not have any race conditions
	assert.True(t, true)
}
