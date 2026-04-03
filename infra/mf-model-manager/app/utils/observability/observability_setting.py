# -*- coding:utf-8 -*-

"""
可观测性配置模块
提供日志、追踪、指标的配置结构及初始化功能
"""

from dataclasses import dataclass
from typing import Optional
from opentelemetry import trace
from opentelemetry.propagate import get_global_textmap


@dataclass
class LogSetting:
    """日志配置类"""
    def __init__(self, log_enabled: bool = False, log_exporter: str = "", 
                 log_load_interval: int = 0, log_load_max_log: int = 0,
                 http_log_feed_ingester_url: str = ""):
        self.log_enabled = log_enabled
        self.log_exporter = log_exporter
        self.log_load_interval = log_load_interval
        self.log_load_max_log = log_load_max_log
        self.http_log_feed_ingester_url = http_log_feed_ingester_url

@dataclass
class TraceSetting:
    """追踪配置类"""
    def __init__(self, trace_enabled: bool = False, trace_provider: str = "",
                 trace_max_queue_size: int = 512, 
                 max_export_batch_size: int =512,
                 http_trace_feed_ingester_url: str = "",
                 grpc_trace_feed_ingester_url: str = "",
                 grpc_trace_job_id: str = ""):
        self.trace_enabled = trace_enabled
        self.trace_provider = trace_provider
        self.trace_max_queue_size = trace_max_queue_size
        self.max_export_batch_size = max_export_batch_size
        self.http_trace_feed_ingester_url = http_trace_feed_ingester_url
        self.grpc_trace_feed_ingester_url = grpc_trace_feed_ingester_url
        self.grpc_trace_job_id = grpc_trace_job_id

@dataclass
class MetricSetting:
    """指标配置类"""
    def __init__(self, metric_enabled: bool = False, metric_provider: str = "",
                 http_metric_feed_ingester_url: str = "",
                 metric_interval_second: int = 0):
        self.metric_enabled = metric_enabled
        self.metric_provider = metric_provider
        self.http_metric_feed_ingester_url = http_metric_feed_ingester_url
        self.metric_interval_second = metric_interval_second

@dataclass
class ObservabilitySetting:
    """可观测性配置组合类"""
    def __init__(self, log: Optional[LogSetting] = None,
                 trace: Optional[TraceSetting] = None,
                 metric: Optional[MetricSetting] = None):
        self.log = log or LogSetting()
        self.trace = trace or TraceSetting()
        self.metric = metric or MetricSetting()


@dataclass
class ServerInfo:
    """服务器信息类"""
    def __init__(self, server_name: str = "", server_version: str = "",
                 language: str = "", python_version: str = ""):
        self.server_name = server_name
        self.server_version = server_version
        self.language = language
        self.python_version = python_version

def inject_trace_context(headers: dict) -> dict:
    """
    将当前的 trace context 注入到 HTTP 请求头中
    
    Args:
        headers: 原始 HTTP 请求头字典
    
    Returns:
        包含 trace context 的 HTTP 请求头字典
    """
    # 获取当前活动的 span 上下文
    current_span = trace.get_current_span()
    if not current_span.is_recording():
        return headers
    
    # 使用全局文本映射器将上下文注入到 headers 中
    propagator = get_global_textmap()
    propagator.inject(headers)
    
    return headers

def extract_trace_context(headers: dict) -> None:
    """
    从 HTTP 请求头中提取 trace 信息并设置到当前上下文中
    
    Args:
        headers: 包含 trace context 的 HTTP 请求头字典
    """
    # 使用全局文本映射器从 headers 中提取上下文
    propagator = get_global_textmap()
    context = propagator.extract(headers)
    
    # 设置提取的上下文为当前上下文
    trace.set_span_in_context(trace.NonRecordingSpan(
        context.get(trace.SPAN_KEY).get_span_context()
    ))