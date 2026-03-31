"""
控制台日志设置模块
"""

import logging
import sys
import structlog

from app.common.config import Config
from .processors import add_caller_info
from .console_formatter import format_console_log


def _create_stdout_handler():
    """
    创建 stdout 处理器，用于输出 INFO 和 DEBUG 级别的日志

    Returns:
        logging.StreamHandler: 配置好的 stdout 处理器
    """

    class StdoutFilter(logging.Filter):
        def filter(self, record):
            return record.levelno <= logging.INFO

    stdout_handler = logging.StreamHandler(sys.stdout)
    stdout_handler.setLevel(logging.NOTSET)  # 不在 handler 层过滤，完全由 Filter 控制
    stdout_handler.addFilter(StdoutFilter())
    stdout_handler.setFormatter(logging.Formatter("%(message)s"))

    return stdout_handler


def _create_stderr_handler():
    """
    创建 stderr 处理器，用于输出 WARNING、ERROR 和 CRITICAL 级别的日志

    Returns:
        logging.StreamHandler: 配置好的 stderr 处理器
    """

    class StderrFilter(logging.Filter):
        def filter(self, record):
            return record.levelno >= logging.WARNING

    stderr_handler = logging.StreamHandler(sys.stderr)
    stderr_handler.setLevel(logging.NOTSET)  # 不在 handler 层过滤，完全由 Filter 控制
    stderr_handler.addFilter(StderrFilter())
    stderr_handler.setFormatter(logging.Formatter("%(message)s"))

    return stderr_handler


def setup_console_logging():
    """
    配置控制台日志

    Returns:
        structlog.BoundLogger: 配置好的控制台日志记录器
    """
    # 配置标准 logging
    log_level = Config.app.get_stdlib_log_level()

    # 创建两个控制台处理器：INFO/DEBUG 输出到 stdout，WARNING/ERROR/CRITICAL 输出到 stderr
    # 这样 K8s 日志收集器可以正确识别日志级别
    stdout_handler = _create_stdout_handler()
    stderr_handler = _create_stderr_handler()

    # 创建控制台专用的 stdlib logger
    console_stdlib_logger = logging.getLogger("agent-executor-console")
    console_stdlib_logger.setLevel(log_level)
    console_stdlib_logger.addHandler(stdout_handler)  # 添加 stdout 处理器
    console_stdlib_logger.addHandler(stderr_handler)  # 添加 stderr 处理器
    console_stdlib_logger.propagate = False

    # 配置控制台 logger（使用增强的格式化器）
    console_logger = structlog.wrap_logger(
        console_stdlib_logger,
        processors=[
            structlog.stdlib.filter_by_level,
            structlog.stdlib.add_logger_name,
            structlog.stdlib.add_log_level,
            structlog.stdlib.PositionalArgumentsFormatter(),
            structlog.processors.TimeStamper(fmt="%Y-%m-%d %H:%M:%S", utc=False),
            structlog.processors.StackInfoRenderer(),
            add_caller_info,
            structlog.processors.format_exc_info,
            structlog.processors.UnicodeDecoder(),
            format_console_log,  # 使用自定义的增强控制台格式化器
        ],
        wrapper_class=structlog.stdlib.BoundLogger,
        context_class=dict,
        cache_logger_on_first_use=True,
    )

    return console_logger
