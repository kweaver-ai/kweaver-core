// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package telemetry

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type ctxSpanKey struct{}

// RDSHook rds hook
type RDSHook struct {
	System string
}

func generateSQL(sqlstr string, args ...interface{}) string {
	var str string
	for _, arg := range args {
		switch v := arg.(type) {
		case bool:
			str = fmt.Sprint(v)
		case string:
			str = fmt.Sprintf("'%s'", v)
		case time.Time:
			str = v.Format("2006-01-02 15:04:05")
		default:
			f, err := strconv.ParseFloat(fmt.Sprint(v), 64)
			if err == nil {
				str = fmt.Sprint(f)
			} else {
				str = fmt.Sprintf("'%s'", fmt.Sprint(v))
			}
		}
		sqlstr = strings.Replace(sqlstr, "?", str, 1)
	}
	return sqlstr
}

// Before 开始
func (h *RDSHook) Before(ctx context.Context, sqltmp string, args ...interface{}) (context.Context, error) {
	tracer := otel.GetTracerProvider()
	if tracer != nil {
		opname := BuildUpOperateName("MySQL")
		sqlstr := generateSQL(sqltmp, args...)
		nctx, span := o11y.GlobalTracer().Start(ctx, opname, trace.WithSpanKind(trace.SpanKindInternal))
		span.SetAttributes(attribute.Key("db.system").String(h.System))
		span.SetAttributes(attribute.Key("db.statement").String(sqlstr))
		nctx = context.WithValue(nctx, (*ctxSpanKey)(nil), span)
		return nctx, nil
	}
	return ctx, nil
}

// After 结束
func (h *RDSHook) After(ctx context.Context, _ string, _ ...interface{}) (context.Context, error) {
	if span, ok := ctx.Value((*ctxSpanKey)(nil)).(trace.Span); ok {
		o11y.TelemetrySpanEnd(span, nil)
	}
	return ctx, nil
}

// OnError 错误
func (h *RDSHook) OnError(ctx context.Context, err error, _ string, _ ...interface{}) error {
	if err != nil && !errors.Is(err, driver.ErrSkip) {
		if span, ok := ctx.Value((*ctxSpanKey)(nil)).(trace.Span); ok {
			span.SetAttributes(attribute.Key("db.error").String(err.Error()))
			o11y.TelemetrySpanEnd(span, err)
		}
	}
	return nil
}
