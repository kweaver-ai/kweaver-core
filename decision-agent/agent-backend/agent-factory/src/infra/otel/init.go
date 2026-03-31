package otel

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
)

// InitOTel 初始化 traces 和 logs 两类 provider。
// 示例：服务启动时调用一次，退出时再调用 providers.Shutdown(ctx)。
func InitOTel(ctx context.Context, cfg *OtelV2Config) (*Providers, error) {
	cfg.SetDefaults()
	otellog.SetServiceName(cfg.ServiceName)

	res, err := buildResource(ctx, cfg)
	if err != nil {
		return nil, err
	}

	providers := &Providers{}

	// 初始化 Trace 提供者
	if cfg.Trace.Enabled {
		tracerProvider, err := newTracerProvider(ctx, cfg.OTLPEndpoint, cfg.Trace.SamplingRate, res)
		if err != nil {
			return nil, err
		}

		otel.SetTracerProvider(tracerProvider)
		providers.TracerProvider = tracerProvider
	}

	// 设置全局传播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 初始化 Log 提供者
	if cfg.Log.Enabled {
		loggerProvider, err := newLoggerProvider(ctx, cfg.OTLPEndpoint, res)
		if err != nil {
			return nil, err
		}

		global.SetLoggerProvider(loggerProvider)
		providers.LoggerProvider = loggerProvider
	}

	log.Printf("[OTel] Initialized for service=%s, endpoint=%s (HTTP), trace=%v, log=%v",
		cfg.ServiceName, cfg.OTLPEndpoint, cfg.Trace.Enabled, cfg.Log.Enabled)

	return providers, nil
}
