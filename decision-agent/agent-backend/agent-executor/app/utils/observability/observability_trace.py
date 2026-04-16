# -*- coding:utf-8 -*-

"""Python 实现的可观测性追踪模块。"""

import os

from app.utils.observability.observability_setting import TraceSetting, ServerInfo
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
from opentelemetry.trace import set_tracer_provider
from opentelemetry.sdk.resources import Resource
from app.common.stand_log import StandLogger


def init_trace_provider(server_info: ServerInfo, setting: TraceSetting) -> None:
    """初始化追踪导出器

    Args:
        server_info: 服务器信息
        setting: 追踪配置设置
    """
    try:
        StandLogger.info_log("[OTel] init_trace_provider called")
        StandLogger.info_log(f"[OTel]   server_name={server_info.server_name}")
        StandLogger.info_log(f"[OTel]   server_version={server_info.server_version}")
        StandLogger.info_log(f"[OTel]   trace_provider={setting.trace_provider}")
        StandLogger.info_log(f"[OTel]   otlp_endpoint={setting.otlp_endpoint}")

        # 创建 trace exporter
        trace_exporter = None

        if setting.trace_provider == "console":
            trace_exporter = ConsoleSpanExporter()
            StandLogger.info_log("[OTel] Created ConsoleSpanExporter")

        elif setting.trace_provider == "otlp":
            # 使用标准 OTLP HTTP exporter（参考 agent-executor-tmp/otel_test_setup.py）
            from opentelemetry.exporter.otlp.proto.http.trace_exporter import (
                OTLPSpanExporter,
            )

            otlp_endpoint = setting.otlp_endpoint
            if not otlp_endpoint:
                StandLogger.info_log(
                    "[OTel] ❌ OTLP endpoint is empty, trace will not be exported"
                )
                return

            StandLogger.info_log(f"[OTel] Raw OTLP endpoint: {otlp_endpoint}")

            # 构造完整 URL
            if not otlp_endpoint.startswith("http://") and not otlp_endpoint.startswith(
                "https://"
            ):
                otlp_endpoint = f"http://{otlp_endpoint}"
            if not otlp_endpoint.endswith("/v1/traces"):
                otlp_endpoint = f"{otlp_endpoint}/v1/traces"

            StandLogger.info_log(f"[OTel] Final OTLP endpoint: {otlp_endpoint}")
            trace_exporter = OTLPSpanExporter(endpoint=otlp_endpoint)
            StandLogger.info_log("[OTel] ✅ OTLPSpanExporter created successfully")

        else:
            StandLogger.info_log(
                f"[OTel] ❌ Unsupported trace provider: {setting.trace_provider}"
            )

        # 如果没有配置任何 exporter，直接返回
        if trace_exporter is None:
            StandLogger.info_log(
                f"[OTel] ❌ No trace exporter configured for provider: {setting.trace_provider}"
            )
            return

        StandLogger.info_log("[OTel] Creating BatchSpanProcessor...")
        trace_processor = BatchSpanProcessor(
            span_exporter=trace_exporter,
            schedule_delay_millis=2000,
            max_queue_size=setting.trace_max_queue_size,
        )

        # 构建 Resource（使用标准 OpenTelemetry SDK）
        StandLogger.info_log("[OTel] Building resource attributes...")
        resource_attributes = {
            "service.name": server_info.server_name,
            "service.version": server_info.server_version,
        }

        # 添加 deployment.environment
        if setting.environment:
            resource_attributes["deployment.environment"] = setting.environment
            StandLogger.info_log(
                f"[OTel] Adding deployment.environment={setting.environment}"
            )

        # 添加 pod.name
        pod_name = os.getenv("POD_NAME")
        if pod_name:
            resource_attributes["pod.name"] = pod_name
            StandLogger.info_log(f"[OTel] Adding pod.name={pod_name}")

        resource = Resource.create(resource_attributes)
        StandLogger.info_log(
            f"[OTel] Resource created with attributes: {resource_attributes}"
        )

        # 设置采样率
        from opentelemetry.sdk.trace.sampling import ParentBasedTraceIdRatio

        sampling_rate = setting.sampling_rate
        if sampling_rate <= 0 or sampling_rate > 1:
            sampling_rate = 1.0
        sampler = ParentBasedTraceIdRatio(sampling_rate)
        StandLogger.info_log(f"[OTel] Sampling rate: {sampling_rate}")

        # 创建 TracerProvider
        StandLogger.info_log("[OTel] Creating TracerProvider...")
        trace_provider = TracerProvider(
            resource=resource, active_span_processor=trace_processor, sampler=sampler
        )

        StandLogger.info_log("[OTel] Setting global tracer provider...")
        set_tracer_provider(trace_provider)
        StandLogger.info_log("[OTel] ✅ Trace provider initialized successfully!")
        StandLogger.info_log(
            f"[OTel] Summary: service={server_info.server_name}, version={server_info.server_version}, endpoint={setting.otlp_endpoint}, provider={setting.trace_provider}"
        )

    except Exception as e:
        import traceback

        trace_details = traceback.format_exc()
        StandLogger.info_log(
            f"[OTel] ❌ Error initializing trace provider: {type(e).__name__}: {e}\n{trace_details}"
        )
