package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// StopTrace 停止 trace
func StopTrace() {
	tp := otel.GetTracerProvider()
	if tp == nil {
		return
	}
	p, ok := tp.(*sdktrace.TracerProvider)
	if ok {
		_ = p.Shutdown(context.Background())
	}
}
