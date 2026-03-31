// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

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
