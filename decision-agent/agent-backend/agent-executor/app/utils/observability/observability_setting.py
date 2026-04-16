# -*- coding:utf-8 -*-

"""可观测性配置模块。"""

from dataclasses import dataclass, field
from opentelemetry import trace
from opentelemetry.propagate import get_global_textmap


@dataclass
class LogSetting:
    """日志配置类"""
    log_enabled: bool = False
    log_level: str = "info"


@dataclass
class TraceSetting:
    """追踪配置类"""
    trace_enabled: bool = False
    trace_max_queue_size: int = 512
    otlp_endpoint: str = ""
    environment: str = ""
    sampling_rate: float = 1.0


@dataclass
class ObservabilitySetting:
    """可观测性配置组合类"""
    log: LogSetting = field(default_factory=LogSetting)
    trace: TraceSetting = field(default_factory=TraceSetting)


@dataclass
class ServerInfo:
    """服务器信息类"""
    server_name: str = ""
    server_version: str = ""
    language: str = ""
    python_version: str = ""


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
    trace.set_span_in_context(
        trace.NonRecordingSpan(context.get(trace.SPAN_KEY).get_span_context())
    )
