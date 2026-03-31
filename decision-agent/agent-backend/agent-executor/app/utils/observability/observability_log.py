# -*- coding:utf-8 -*-

"""
Python 实现的可观测性日志模块
提供带上下文追踪的日志记录功能，支持多种日志导出方式
"""

import inspect
import os
from typing import Optional, Any
from opentelemetry import context

from app.utils.observability.sdk_available import (
    TELEMETRY_SDK_AVAILABLE,
    SamplerLogger,
    log_resource,
    set_service_info,
)
from app.utils.observability.observability_setting import LogSetting, ServerInfo


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

    # 如果 SDK 不可用，直接返回
    if not TELEMETRY_SDK_AVAILABLE:
        return

    # 延迟导入 Config 避免循环依赖
    from app.common.config import Config

    set_service_info(
        server_info.server_name,
        server_info.server_version,
        os.getenv("POD_NAME", "unknown"),
    )

    # 如果没有启用 o11y 日志，直接返回
    if not Config.is_o11y_log_enabled():
        return

    logger = SamplerLogger(log_resource())
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
