# -*- coding:utf-8 -*-

"""
Python 实现的可观测性日志模块
提供带上下文追踪的日志记录功能，支持多种日志导出方式
"""
import inspect, os
from typing import Optional
from opentelemetry import context
from tlogging import SamplerLogger
from exporter.resource.resource import log_resource, set_service_info
from app.utils.observability.observability_setting import LogSetting, ServerInfo
from exporter.ar_log.log_exporter import ARLogExporter
from exporter.public.client import HTTPClient
from exporter.public.public import WithAnyRobotURL


# 定义 全局 Telemetry Logger
logger = None

def get_caller_info() -> str:
    """获取调用者信息（文件名、行号、函数名）"""
    frame = inspect.stack()[2]
    filename = frame.filename
    line_number = frame.lineno
    function_name = frame.function
    return f"{filename}:{line_number}:{function_name}"

def info(msg: str, ctx: Optional[context.Context] = None) -> None:
    global logger
    """记录INFO级别日志
    
    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.info(message=full_msg, ctx=ctx)

def error(msg: str, ctx: Optional[context.Context] = None) -> None:
    global logger

    """记录ERROR级别日志
    
    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.error(message=full_msg, ctx=ctx)

def warn(msg: str, ctx: Optional[context.Context] = None) -> None:
    global logger

    """记录WARNING级别日志
    
    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.warning(message=full_msg, ctx=ctx)

def debug(msg: str, ctx: Optional[context.Context] = None) -> None:
    global logger

    """记录DEBUG级别日志
    
    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.debug(message=full_msg, ctx=ctx)

def fatal(msg: str, ctx: Optional[context.Context] = None) -> None:
    global logger

    """记录FATAL级别日志并退出程序
    
    Args:
        msg: 日志消息内容
        ctx: OpenTelemetry上下文
    """
    caller_info = get_caller_info()
    full_msg = f"{caller_info}: {msg}"
    logger.fatal(message=full_msg, ctx=ctx)
    exit(1)

# 初始化 Log Exporters，设置 Logger 上报 Exporters
def init_log_provider(server_info :ServerInfo, setting: LogSetting) -> None:
    global logger
    """初始化日志导出器
    
    Args:
        server_info: 服务器信息
        setting: 日志配置设置
    """
    set_service_info(server_info.server_name, server_info.server_version, os.getenv("POD_NAME", "unknown"))

    
    logger = SamplerLogger(log_resource())
    if setting.log_exporter == "console":
        logger.set_exporters()
    elif setting.log_exporter == "http":
       http_exporter = ARLogExporter(
        HTTPClient(WithAnyRobotURL(setting.http_log_feed_ingester_url),
                #    WithSyncMode()
                ))
       logger.set_exporters(http_exporter)

def get_logger():
    return logger 
    

def shutdown_log_provider(*args, **kwargs):
    global logger
    if logger is None:
        return
    logger.shutdown()