"""
文件日志设置模块
"""

import logging
from logging.handlers import TimedRotatingFileHandler
import structlog

from app.common.config import Config
from .processors import add_caller_info
from .file_formatter import format_file_log


def setup_file_logging():
    """
    配置文件日志

    Returns:
        structlog.BoundLogger: 配置好的文件日志记录器
    """
    # 配置标准 logging
    log_level = Config.app.get_stdlib_log_level()

    # 创建文件处理器
    file_handler = TimedRotatingFileHandler(
        "log/agent-executor.log",
        when="midnight",
        interval=1,
        backupCount=30,
    )
    file_handler.setLevel(logging.NOTSET)  # 不在 handler 层过滤，完全由 Logger 控制
    file_handler.setFormatter(logging.Formatter("%(message)s"))

    # 创建文件专用的 stdlib logger
    file_stdlib_logger = logging.getLogger("agent-executor-file")
    file_stdlib_logger.setLevel(log_level)
    file_stdlib_logger.addHandler(file_handler)
    file_stdlib_logger.propagate = False

    # 配置文件 logger（格式: "时间 - 级别 - {JSON内容}"）
    file_logger = structlog.wrap_logger(
        file_stdlib_logger,
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
            format_file_log,  # 自定义格式化: "时间 - 级别 - {JSON}" (直接返回字符串)
        ],
        wrapper_class=structlog.stdlib.BoundLogger,
        context_class=dict,
        cache_logger_on_first_use=True,
    )

    return file_logger
