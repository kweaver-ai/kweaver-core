package agentsvc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"
)

// ResumeChat: session deleted between outer and goroutine inner load
func TestAgentSvc_ResumeChat_SessionDeletedBeforeGoroutine(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	convID := "conv-resume-del-goroutine"
	session := &Session{ConversationID: convID}
	SessionMap.Store(convID, session)

	ch, err := svc.ResumeChat(context.Background(), convID)
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Delete session immediately so goroutine's inner load fails
	SessionMap.Delete(convID)

	done := make(chan struct{})
	go func() {
		for range ch {
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close in time")
	}
}

// ResumeChat: signal loop - oldResp gets updated on each signal
func TestAgentSvc_ResumeChat_MultipleSignals(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	convID := "conv-resume-multi"
	session := &Session{
		ConversationID: convID,
		TempMsgResp:    agentresp.ChatResp{ConversationID: convID},
	}
	SessionMap.Store(convID, session)

	defer SessionMap.Delete(convID)

	ch, err := svc.ResumeChat(context.Background(), convID)
	assert.NoError(t, err)

	go func() {
		time.Sleep(5 * time.Millisecond)
		session.UpdateTempMsgResp(agentresp.ChatResp{ConversationID: convID, AgentRunID: "run-1"})
		session.SendSignal()
		time.Sleep(5 * time.Millisecond)
		session.UpdateTempMsgResp(agentresp.ChatResp{ConversationID: convID, AgentRunID: "run-2"})
		session.SendSignal()
		time.Sleep(5 * time.Millisecond)
		session.CloseSignal()
	}()

	received := 0
	done := make(chan struct{})

	go func() {
		for range ch {
			received++
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("channel did not close in time")
	}
	assert.Greater(t, received, 0)
}

// ResumeChat: initial StreamDiff with non-empty TempMsgResp
func TestAgentSvc_ResumeChat_InitialDiff(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	convID := "conv-resume-initdiff"
	session := &Session{
		ConversationID: convID,
		TempMsgResp:    agentresp.ChatResp{ConversationID: convID, AgentRunID: "run-init"},
	}
	SessionMap.Store(convID, session)

	defer SessionMap.Delete(convID)

	// Close signal immediately after a brief delay
	go func() {
		time.Sleep(20 * time.Millisecond)
		session.CloseSignal()
	}()

	ch, err := svc.ResumeChat(context.Background(), convID)
	assert.NoError(t, err)

	done := make(chan struct{})
	received := 0

	go func() {
		for range ch {
			received++
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close in time")
	}
	// Initial diff should have sent at least one message
	assert.Greater(t, received, 0)
}

// ResumeChat: signal loop with existing signal, sends multiple updates
func TestAgentSvc_ResumeChat_ExistingSignalWithUpdates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	convID := "conv-resume-exsig-upd"
	existingSignal := make(chan struct{}, 5)
	session := &Session{
		ConversationID: convID,
		Signal:         existingSignal,
		TempMsgResp:    agentresp.ChatResp{ConversationID: convID},
	}
	SessionMap.Store(convID, session)

	defer SessionMap.Delete(convID)

	ch, err := svc.ResumeChat(context.Background(), convID)
	assert.NoError(t, err)

	go func() {
		time.Sleep(10 * time.Millisecond)
		session.UpdateTempMsgResp(agentresp.ChatResp{ConversationID: convID, AgentRunID: "upd-1"})
		existingSignal <- struct{}{}

		time.Sleep(10 * time.Millisecond)
		close(existingSignal)
	}()

	done := make(chan struct{})
	go func() {
		for range ch {
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("channel did not close in time")
	}
}
