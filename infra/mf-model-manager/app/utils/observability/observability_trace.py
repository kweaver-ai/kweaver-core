# -*- coding:utf-8 -*-

"""
Python 实现的可观测性日志模块
提供带上下文追踪的日志记录功能，支持多种日志导出方式
"""
import os

from exporter.resource.resource import log_resource, set_service_info
from app.utils.observability.observability_setting import TraceSetting, ServerInfo
from exporter.public.client import HTTPClient
from exporter.public.public import WithAnyRobotURL
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor, ConsoleSpanExporter
from opentelemetry.trace import set_tracer_provider

from exporter.ar_trace.trace_exporter import ARTraceExporter
from exporter.public.client import HTTPClient
from exporter.public.public import WithAnyRobotURL
from exporter.resource.resource import set_service_info, trace_resource


# 初始化 Telemetry Log Provider

# 初始化 Trace Provider
def init_trace_provider(server_info :ServerInfo, setting: TraceSetting) -> None:
    """初始化日志导出器
    
    Args:
        server_info: 服务器信息
        setting: 日志配置设置
    """
    set_service_info(server_info.server_name, server_info.server_version, os.getenv("POD_NAME", "unknown"))
    
    trace_exporter = None
    
    

    if setting.trace_provider == "console":
        trace_exporter = ConsoleSpanExporter()
        
    elif setting.trace_provider == "http":
        trace_exporter = ARTraceExporter(HTTPClient(WithAnyRobotURL(setting.http_trace_feed_ingester_url))
        # StdoutClient("AnyRobotTrace.json")
    )

    trace_processor =  BatchSpanProcessor(
            span_exporter=trace_exporter,
            schedule_delay_millis=2000,
            max_queue_size=setting.trace_max_queue_size
            )
    trace_provider = TracerProvider(resource=trace_resource(), active_span_processor=trace_processor)
    
    set_tracer_provider(trace_provider)
