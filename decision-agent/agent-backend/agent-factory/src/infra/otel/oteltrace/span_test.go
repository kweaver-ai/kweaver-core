package oteltrace

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
)

func TestSetConversationID(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	oldTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())

		otel.SetTracerProvider(oldTP)
	})

	ctx, _ := StartInternalSpan(context.Background())
	SetConversationID(ctx, "conv-123")
	EndSpan(ctx, nil)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Equal(t, "conv-123", attributeValue(spans[0].Attributes(), otelconst.AttrGenAIConversationID))
}

func TestSetConversationID_EmptySkipsAttribute(t *testing.T) {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	oldTP := otel.GetTracerProvider()
	otel.SetTracerProvider(tp)
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())

		otel.SetTracerProvider(oldTP)
	})

	ctx, _ := StartInternalSpan(context.Background())
	SetConversationID(ctx, "")
	EndSpan(ctx, nil)

	spans := recorder.Ended()
	require.Len(t, spans, 1)
	assert.Empty(t, attributeValue(spans[0].Attributes(), otelconst.AttrGenAIConversationID))
}

func attributeValue(attrs []attribute.KeyValue, key string) string {
	for _, attr := range attrs {
		if string(attr.Key) == key {
			return attr.Value.AsString()
		}
	}

	return ""
}
