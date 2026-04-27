# -*- coding:utf-8 -*-

"""Python 实现的可观测性追踪模块。"""

import os

from app.utils.observability.observability_setting import TraceSetting, ServerInfo
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
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
        otlp_endpoint = setting.otlp_endpoint
        if not otlp_endpoint:
            StandLogger.warn(
                "[OTel] ❌ OTLP endpoint is empty, trace will not be exported"
            )
            return

        from opentelemetry.exporter.otlp.proto.http.trace_exporter import (
            OTLPSpanExporter,
        )

        if not otlp_endpoint.startswith("http://") and not otlp_endpoint.startswith(
            "https://"
        ):
            otlp_endpoint = f"http://{otlp_endpoint}"
        if not otlp_endpoint.endswith("/v1/traces"):
            otlp_endpoint = f"{otlp_endpoint}/v1/traces"

        trace_exporter = OTLPSpanExporter(endpoint=otlp_endpoint)
        trace_processor = BatchSpanProcessor(
            span_exporter=trace_exporter,
            schedule_delay_millis=2000,
            max_queue_size=setting.trace_max_queue_size,
        )

        # 构建 Resource（使用标准 OpenTelemetry SDK）
        resource_attributes = {
            "service.name": server_info.server_name,
            "service.version": server_info.server_version,
        }

        # 添加 deployment.environment
        if setting.environment:
            resource_attributes["deployment.environment"] = setting.environment

        # 添加 pod.name
        pod_name = os.getenv("POD_NAME")
        if pod_name:
            resource_attributes["pod.name"] = pod_name

        resource = Resource.create(resource_attributes)

        # 设置采样率
        from opentelemetry.sdk.trace.sampling import ParentBasedTraceIdRatio

        sampling_rate = setting.sampling_rate
        if sampling_rate <= 0 or sampling_rate > 1:
            sampling_rate = 1.0
        sampler = ParentBasedTraceIdRatio(sampling_rate)

        # 创建 TracerProvider
        trace_provider = TracerProvider(
            resource=resource, active_span_processor=trace_processor, sampler=sampler
        )

        set_tracer_provider(trace_provider)

    except Exception as e:
        import traceback

        trace_details = traceback.format_exc()
        StandLogger.error(
            f"[OTel] ❌ Error initializing trace provider: {type(e).__name__}: {e}\n{trace_details}"
        )
