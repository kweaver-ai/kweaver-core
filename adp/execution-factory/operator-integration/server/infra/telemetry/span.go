package telemetry

import (
	"context"
	"fmt"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel/attribute"
)

// ExporterType Export类型
type ExporterType string

const (
	ExporterTypeOTLP   ExporterType = "otlp"   // otlp导出
	ExporterTypeJaeger ExporterType = "jaeger" // jaeger导出
)

// SetSpanAttributes 设置Span属性
func SetSpanAttributes(ctx context.Context, attrs map[string]interface{}) {
	if attrs == nil || ctx == nil {
		return
	}
	attrsList := make([]attribute.KeyValue, 0)
	for k, v := range attrs {
		attrsList = append(attrsList, attribute.String(k, fmt.Sprintf("%v", v)))
	}
	o11y.SetAttributes(ctx, attrsList...)
}
