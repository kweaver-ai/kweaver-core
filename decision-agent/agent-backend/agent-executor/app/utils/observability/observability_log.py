# -*- coding:utf-8 -*-

"""
Python 实现的可观测性日志模块
提供带上下文追踪的日志记录功能，支持多种日志导出方式
"""

import inspect
import os
import time
from typing import Optional, Any
from opentelemetry import context
from opentelemetry import trace as otel_trace
from opentelemetry.trace import format_span_id, format_trace_id

from app.utils.observability.sdk_available import (
    TELEMETRY_SDK_AVAILABLE,
    SamplerLogger,
    log_resource,
    set_service_info,
)
from app.utils.observability.observability_setting import LogSetting, ServerInfo


_LEVEL_PRIORITY = {
    "trace": 10,
    "debug": 20,
    "info": 30,
    "warn": 40,
    "warning": 40,
    "error": 50,
    "fatal": 60,
    "off": 100,
}


class NullLogger:
    """空操作日志器，当 TelemetrySDK 不可用时使用

    实现与 SamplerLogger 相同的接口，但所有方法都是空操作（no-op）。
    这样可以避免在 SDK 不可用时调用 o11y_logger().info() 等方法报错。
    """

    def info(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        pass

    def error(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        pass

    def warn(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        pass

    def debug(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        pass

    def fatal(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        pass

    def trace(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        pass

    def set_level(self, level: str) -> None:
        pass

    def get_level(self) -> int:
        return 0

    def set_exporters(self, *exporters) -> None:
        pass

    def shutdown(self) -> None:
        pass


class StandardOtelLogger:
    """标准 OpenTelemetry 日志适配器。"""

    def __init__(self, otel_logger, provider, level: str = "info"):
        self._otel_logger = otel_logger
        self._provider = provider
        self._level = "info"
        self.set_level(level)

    def info(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        self._emit("info", message, attributes, etype, ctx)

    def error(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        self._emit("error", message, attributes, etype, ctx)

    def warn(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        self._emit("warn", message, attributes, etype, ctx)

    def debug(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        self._emit("debug", message, attributes, etype, ctx)

    def fatal(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        self._emit("fatal", message, attributes, etype, ctx)

    def trace(
        self,
        message: Any = "",
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        self._emit("trace", message, attributes, etype, ctx)

    def set_level(self, level: str) -> None:
        normalized = str(level or "info").lower()
        if normalized not in _LEVEL_PRIORITY:
            normalized = "info"
        self._level = normalized

    def get_level(self) -> int:
        return _LEVEL_PRIORITY[self._level]

    def set_exporters(self, *exporters) -> None:
        return None

    def shutdown(self) -> None:
        if self._provider is not None:
            self._provider.shutdown()

    def _emit(
        self,
        severity_text: str,
        message: Any,
        attributes: Any = None,
        etype: Any = None,
        ctx: Any = None,
    ) -> None:
        if _LEVEL_PRIORITY.get(severity_text, 0) < self.get_level():
            return

        active_context = ctx or context.get_current()
        payload = dict(attributes or {})
        if etype is not None:
            payload["etype"] = etype

        span = otel_trace.get_current_span(active_context)
        span_context = span.get_span_context() if span is not None else None
        if span_context is not None and span_context.is_valid:
            payload.setdefault("trace_id", format_trace_id(span_context.trace_id))
            payload.setdefault("span_id", format_span_id(span_context.span_id))

        self._otel_logger.emit(
            timestamp=time.time_ns(),
            context=active_context,
            severity_number=_severity_number(severity_text),
            severity_text=severity_text.upper(),
            body=str(message),
            attributes=payload or None,
        )


def _severity_number(severity_text: str):
    from opentelemetry._logs import SeverityNumber

    severity_map = {
        "trace": SeverityNumber.TRACE,
        "debug": SeverityNumber.DEBUG,
        "info": SeverityNumber.INFO,
        "warn": SeverityNumber.WARN,
        "warning": SeverityNumber.WARN,
        "error": SeverityNumber.ERROR,
        "fatal": SeverityNumber.FATAL,
    }
    return severity_map.get(severity_text.lower(), SeverityNumber.INFO)


def _normalize_otlp_http_endpoint(otlp_endpoint: str, signal_path: str) -> str:
    endpoint = otlp_endpoint.strip()
    if not endpoint:
        return ""
    if not endpoint.startswith("http://") and not endpoint.startswith("https://"):
        endpoint = f"http://{endpoint}"
    endpoint = endpoint.rstrip("/")
    if not endpoint.endswith(signal_path):
        endpoint = f"{endpoint}{signal_path}"
    return endpoint


def _build_standard_logger(server_info: ServerInfo, setting: LogSetting):
    from opentelemetry._logs import set_logger_provider
    from opentelemetry.exporter.otlp.proto.http._log_exporter import OTLPLogExporter
    from opentelemetry.sdk._logs import LoggerProvider
    from opentelemetry.sdk._logs.export import (
        BatchLogRecordProcessor,
        ConsoleLogExporter,
        SimpleLogRecordProcessor,
    )
    from opentelemetry.sdk.resources import Resource

    from app.common.config import Config

    resource_attributes = {
        "service.name": server_info.server_name,
        "service.version": server_info.server_version,
    }

    if Config.o11y.environment:
        resource_attributes["deployment.environment"] = Config.o11y.environment

    pod_name = os.getenv("POD_NAME")
    if pod_name:
        resource_attributes["pod.name"] = pod_name

    provider = LoggerProvider(resource=Resource.create(resource_attributes))

    if setting.log_exporter == "console":
        processor = SimpleLogRecordProcessor(ConsoleLogExporter())
    else:
        endpoint = _normalize_otlp_http_endpoint(Config.o11y.trace_endpoint, "/v1/logs")
        if not endpoint:
            return NullLogger()
        processor = BatchLogRecordProcessor(OTLPLogExporter(endpoint=endpoint))

    provider.add_log_record_processor(processor)
    set_logger_provider(provider)
    otel_logger = provider.get_logger(
        server_info.server_name, version=server_info.server_version
    )
    return StandardOtelLogger(otel_logger, provider, setting.log_level)


# 定义 全局 Telemetry Logger
logger = None

if TELEMETRY_SDK_AVAILABLE:
    logger = SamplerLogger(log_resource())
    # 默认级别为off，不打印日志
    logger.set_level("off")
else:
    logger = NullLogger()


def get_caller_info() -> str:
    """获取调用者信息（文件名、行号、函数名）"""
    frame = inspect.stack()[2]
    filename = frame.filename
    line_number = frame.lineno
    function_name = frame.function
    return f"{filename}:{line_number}:{function_name}"


def info(msg: str, ctx: Optional[context.Context] = None) -> None:
    """记录INFO级别日志

    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    global logger
    if logger is None:
        return
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.info(message=full_msg, ctx=ctx)


def error(msg: str, ctx: Optional[context.Context] = None) -> None:
    """记录ERROR级别日志

    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    global logger
    if logger is None:
        return
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.error(message=full_msg, ctx=ctx)


def warn(msg: str, ctx: Optional[context.Context] = None) -> None:
    """记录WARNING级别日志

    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    global logger
    if logger is None:
        return
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.warn(message=full_msg, ctx=ctx)


def debug(msg: str, ctx: Optional[context.Context] = None) -> None:
    """记录DEBUG级别日志

    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    global logger
    if logger is None:
        return
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.debug(message=full_msg, ctx=ctx)


def fatal(msg: str, ctx: Optional[context.Context] = None) -> None:
    """记录FATAL级别日志并退出程序

    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    global logger
    if logger is None:
        exit(1)
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.fatal(message=full_msg, ctx=ctx)
    exit(1)


def init_log_provider(server_info: ServerInfo, setting: LogSetting) -> None:
    """初始化日志导出器

    Args:
        server_info: 服务器信息
        setting: 日志配置设置
    """
    global logger

    if not setting.log_enabled:
        return

    if not TELEMETRY_SDK_AVAILABLE:
        logger = _build_standard_logger(server_info, setting)
        return

    set_service_info(
        server_info.server_name,
        server_info.server_version,
        os.getenv("POD_NAME", "unknown"),
    )

    logger = SamplerLogger(log_resource())
    logger.set_level(setting.log_level)
    if setting.log_exporter == "console":
        logger.set_exporters()
    elif setting.log_exporter == "http":
        from exporter.ar_log.log_exporter import ARLogExporter
        from exporter.public.client import HTTPClient
        from exporter.public.public import WithAnyRobotURL

        http_exporter = ARLogExporter(
            HTTPClient(
                WithAnyRobotURL(setting.http_log_feed_ingester_url),
            )
        )
        logger.set_exporters(http_exporter)


def get_logger():
    global logger
    if logger is None:
        logger = NullLogger()
    return logger


def shutdown_log_provider(*args, **kwargs):
    global logger
    if logger is None:
        return
    logger.shutdown()
