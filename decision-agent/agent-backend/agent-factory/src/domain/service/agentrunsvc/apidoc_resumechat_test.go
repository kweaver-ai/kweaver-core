package agentsvc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/service"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp/cmpmock"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
)

// ---------- GetAPIDoc tests ----------

func TestAgentSvc_GetAPIDoc_SquareSvcError(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase:   service.NewSvcBase(),
		squareSvc: mockSquare,
		logger:    mockLogger,
	}

	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(nil, errors.New("not found"))

	ctx := context.Background()
	req := &agentreq.GetAPIDocReq{AgentID: "a1", AgentVersion: "v1"}
	_, err := svc.GetAPIDoc(ctx, req)
	assert.Error(t, err)
}

func TestAgentSvc_GetAPIDoc_Success(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase:   service.NewSvcBase(),
		squareSvc: mockSquare,
		logger:    mockLogger,
	}

	agentInfoResp := newTestAgent()
	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfoResp, nil)

	ctx := context.Background()
	req := &agentreq.GetAPIDocReq{AgentID: "a1", AgentVersion: "v1"}
	result, err := svc.GetAPIDoc(ctx, req)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestAgentSvc_GetAPIDoc_CustomFieldsAndProfile(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase:   service.NewSvcBase(),
		squareSvc: mockSquare,
		logger:    mockLogger,
	}

	agentInfoResp := newTestAgent()
	profile := "这是一个测试 Agent"
	agentInfoResp.DataAgent.Name = "Agent-Doc-Test"
	agentInfoResp.DataAgent.Key = "agent-key-001"
	agentInfoResp.DataAgent.Profile = &profile
	agentInfoResp.Version = "v9"
	agentInfoResp.Config.Input.Fields = daconfvalobj.Fields{
		&daconfvalobj.Field{Name: "query", Type: cdaenum.InputFieldTypeString},
		&daconfvalobj.Field{Name: "history", Type: cdaenum.InputFieldTypeJSONObject},
		&daconfvalobj.Field{Name: "custom_obj", Type: cdaenum.InputFieldTypeJSONObject},
		&daconfvalobj.Field{Name: "custom_str", Type: cdaenum.InputFieldTypeString},
		&daconfvalobj.Field{Name: "tool", Type: cdaenum.InputFieldTypeJSONObject},
	}

	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfoResp, nil)

	ctx := context.Background()
	req := &agentreq.GetAPIDocReq{AgentID: "a1", AgentVersion: "v1"}
	result, err := svc.GetAPIDoc(ctx, req)
	require.NoError(t, err)

	apiDoc, ok := result.(*openapi3.T)
	require.True(t, ok)

	pathItem := apiDoc.Paths.Value("/api/agent-app/v1/app/{app_key}/api/chat/completion")
	require.NotNil(t, pathItem)
	require.NotNil(t, pathItem.Post)

	assert.Equal(t, "Agent-Doc-Test", pathItem.Post.Summary)
	assert.Equal(t, profile, pathItem.Post.Description)

	content := pathItem.Post.RequestBody.Value.Content["application/json"]
	require.NotNil(t, content)

	example, ok := content.Example.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, false, example["stream"])
	assert.Equal(t, "v9", example["agent_version"])
	assert.Equal(t, "agent-key-001", example["agent_key"])

	historyVal, hasHistory := example["history"]
	assert.True(t, hasHistory)
	assert.IsType(t, []map[string]string{}, historyVal)

	customQuerys, ok := example["custom_querys"].(map[string]interface{})
	require.True(t, ok)
	assert.Contains(t, customQuerys, "custom_obj")
	assert.Contains(t, customQuerys, "custom_str")
	assert.NotContains(t, customQuerys, "tool")

	require.NotNil(t, content.Schema)
	require.NotNil(t, content.Schema.Value)
	assert.Contains(t, content.Schema.Value.Properties, "query")
	assert.Contains(t, content.Schema.Value.Properties, "custom_obj")
	assert.Contains(t, content.Schema.Value.Properties, "custom_str")
}

func TestAgentSvc_GetAPIDoc_RemoveEmptyCustomQuerys(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockSquare := v3portdrivermock.NewMockISquareSvc(ctrl)
	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase:   service.NewSvcBase(),
		squareSvc: mockSquare,
		logger:    mockLogger,
	}

	agentInfoResp := newTestAgent()
	agentInfoResp.DataAgent.Name = "Agent-No-Custom"
	agentInfoResp.DataAgent.Key = "agent-key-002"
	agentInfoResp.Version = "v2"
	agentInfoResp.Config.Input.Fields = daconfvalobj.Fields{
		&daconfvalobj.Field{Name: "query", Type: cdaenum.InputFieldTypeString},
		&daconfvalobj.Field{Name: "history", Type: cdaenum.InputFieldTypeJSONObject},
		&daconfvalobj.Field{Name: "tool", Type: cdaenum.InputFieldTypeJSONObject},
	}

	mockSquare.EXPECT().GetAgentInfoByIDOrKey(gomock.Any(), gomock.Any()).Return(agentInfoResp, nil)

	ctx := context.Background()
	req := &agentreq.GetAPIDocReq{AgentID: "a2", AgentVersion: "v2"}
	result, err := svc.GetAPIDoc(ctx, req)
	require.NoError(t, err)

	apiDoc, ok := result.(*openapi3.T)
	require.True(t, ok)

	pathItem := apiDoc.Paths.Value("/api/agent-app/v1/app/{app_key}/api/chat/completion")
	require.NotNil(t, pathItem)
	assert.Equal(t, "Agent-No-Custom", pathItem.Post.Summary)
	assert.Equal(t, "", pathItem.Post.Description)

	content := pathItem.Post.RequestBody.Value.Content["application/json"]
	require.NotNil(t, content)

	example, ok := content.Example.(map[string]interface{})
	require.True(t, ok)
	assert.NotContains(t, example, "custom_querys")
	assert.Equal(t, false, example["stream"])
	assert.Equal(t, "v2", example["agent_version"])
	assert.Equal(t, "agent-key-002", example["agent_key"])
}

// ---------- ResumeChat tests ----------

func TestAgentSvc_ResumeChat_SessionNotFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	ctx := context.Background()
	// Use a conversation ID that is not stored in SessionMap
	_, err := svc.ResumeChat(ctx, "conv-not-exist-999")
	assert.Error(t, err)
}

func TestAgentSvc_ResumeChat_SessionFound(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	// Register a session
	session := &Session{ConversationID: "conv-resume-test"}
	SessionMap.Store("conv-resume-test", session)

	defer SessionMap.Delete("conv-resume-test")

	// Close the signal after a short time so the goroutine exits
	go func() {
		time.Sleep(10 * time.Millisecond)
		session.CloseSignal()
	}()

	ctx := context.Background()
	ch, err := svc.ResumeChat(ctx, "conv-resume-test")
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Drain the channel
	for range ch {
	}
}

func TestAgentSvc_ResumeChat_SessionFoundWithExistingSignal(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	// Register a session with pre-existing signal
	existingSignal := make(chan struct{})
	session := &Session{ConversationID: "conv-resume-existing", Signal: existingSignal}
	SessionMap.Store("conv-resume-existing", session)

	defer SessionMap.Delete("conv-resume-existing")

	// Close the signal immediately
	close(existingSignal)

	ctx := context.Background()
	ch, err := svc.ResumeChat(ctx, "conv-resume-existing")
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	// Drain channel with timeout
	done := make(chan struct{})
	go func() {
		for range ch {
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Log("channel drain timeout (acceptable)")
	}
}

func TestAgentSvc_ResumeChat_WithSignalUpdates(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := cmpmock.NewMockLogger(ctrl)
	mockLogger.EXPECT().Errorf(gomock.Any(), gomock.Any()).AnyTimes()

	svc := &agentSvc{
		SvcBase: service.NewSvcBase(),
		logger:  mockLogger,
	}

	session := &Session{
		ConversationID: "conv-resume-signal",
		TempMsgResp:    agentresp.ChatResp{ConversationID: "conv-resume-signal"},
	}
	SessionMap.Store("conv-resume-signal", session)

	defer SessionMap.Delete("conv-resume-signal")

	ch, err := svc.ResumeChat(context.Background(), "conv-resume-signal")
	assert.NoError(t, err)
	assert.NotNil(t, ch)

	go func() {
		time.Sleep(10 * time.Millisecond)
		session.UpdateTempMsgResp(agentresp.ChatResp{ConversationID: "conv-resume-signal", AgentRunID: "run-updated"})
		session.SendSignal()
		time.Sleep(10 * time.Millisecond)
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
	case <-time.After(2 * time.Second):
		t.Fatal("resume channel did not close in time")
	}

	assert.Greater(t, received, 0)
}
