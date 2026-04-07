package agenthandler

import (
	"context"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.uber.org/mock/gomock"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/otelconst"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iportdriver/iportdrivermock"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

func TestChat_BackfillsConversationIDOnRequestSpanAfterServiceReturns(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	oldTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())

		otel.SetTracerProvider(oldTP)
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	mockAgent.EXPECT().Chat(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, req *agentreq.ChatReq) (chan []byte, error) {
		req.ConversationID = "conv-created-in-chat"
		return jsonChannel(`{"ok":true}`), nil
	})

	c, recorderHTTP := newAgentCtx(http.MethodPost, "/", `{"agent_id":"a1","stream":false}`)
	c.Params = gin.Params{{Key: "app_key", Value: "app-1"}}
	setAgentVisitor(c, "u1", "Bearer tk", rest.VisitorType_User)

	spanCtx, _ := oteltrace.StartInternalSpan(c.Request.Context())
	c.Request = c.Request.WithContext(spanCtx)

	h.Chat(c)
	oteltrace.EndSpan(spanCtx, nil)

	assert.Equal(t, http.StatusOK, recorderHTTP.Code)

	spans := recorder.Ended()
	require.NotEmpty(t, spans)
	assert.Equal(t, "conv-created-in-chat", lastAttributeValue(spans, otelconst.AttrGenAIConversationID))
}

func TestResumeChat_SetsConversationIDOnRequestSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	oldTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())

		otel.SetTracerProvider(oldTP)
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	ch := make(chan []byte, 1)
	ch <- []byte(constant.DataEventEndStr)
	close(ch)
	mockAgent.EXPECT().ResumeChat(gomock.Any(), "conv-resume").Return(ch, nil)

	c, recorderHTTP := newAgentCtx(http.MethodPost, "/", `{"conversation_id":"conv-resume"}`)
	spanCtx, _ := oteltrace.StartInternalSpan(c.Request.Context())
	c.Request = c.Request.WithContext(spanCtx)

	h.ResumeChat(c)
	oteltrace.EndSpan(spanCtx, nil)

	assert.Equal(t, http.StatusOK, recorderHTTP.Code)

	spans := recorder.Ended()
	require.NotEmpty(t, spans)
	assert.Equal(t, "conv-resume", lastAttributeValue(spans, otelconst.AttrGenAIConversationID))
}

func TestTerminateChat_SetsConversationIDOnRequestSpan(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	oldTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())

		otel.SetTracerProvider(oldTP)
	})

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockAgent := iportdrivermock.NewMockIAgent(ctrl)
	h := &agentHTTPHandler{agentSvc: mockAgent, logger: agentTestLogger{}}

	mockAgent.EXPECT().TerminateChat(gomock.Any(), "conv-terminate", "run-1", "asst-1").Return(nil)

	c, recorderHTTP := newAgentCtx(http.MethodPost, "/", `{"conversation_id":"conv-terminate","agent_run_id":"run-1","interrupted_assistant_message_id":"asst-1"}`)
	spanCtx, _ := oteltrace.StartInternalSpan(c.Request.Context())
	c.Request = c.Request.WithContext(spanCtx)

	h.TerminateChat(c)
	oteltrace.EndSpan(spanCtx, nil)

	assert.Equal(t, http.StatusNoContent, recorderHTTP.Code)

	spans := recorder.Ended()
	require.NotEmpty(t, spans)
	assert.Equal(t, "conv-terminate", lastAttributeValue(spans, otelconst.AttrGenAIConversationID))
}

func lastAttributeValue(spans []sdktrace.ReadOnlySpan, key string) string {
	for i := len(spans) - 1; i >= 0; i-- {
		for _, attr := range spans[i].Attributes() {
			if string(attr.Key) == key {
				return attr.Value.AsString()
			}
		}
	}

	return ""
}
