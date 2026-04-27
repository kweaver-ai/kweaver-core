# -*- coding:utf-8 -*-
"""
异常日志记录器
专门用于记录异常日志，支持同时输出到控制台和文件
"""

import os
import sys
import logging
from datetime import datetime
from logging.handlers import TimedRotatingFileHandler
from typing import Optional, Dict, Any, List

from .constants import EXCEPTION_LOG_DIR, EXCEPTION_LOG_SIMPLE, EXCEPTION_LOG_DETAILED
from app.common.config import Config
from .formatter_pkg import (
    format_error_console,
    format_error_file_simple,
    format_error_file_detailed,
    format_multiple_errors_separator,
)


class ExceptionLogger:
    """
    异常日志记录器

    功能：
    1. 同时输出到控制台和文件
    2. 控制台日志：仅项目代码的 Traceback，带颜色
    3. 简单文件日志：仅项目代码的 Traceback
    4. 详细文件日志：完整的 Traceback（包含第三方库）
    """

    _instance = None
    _initialized = False

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if self._initialized:
            return

        self._initialized = True
        self._console_handler = None
        self._simple_file_handler = None
        self._detailed_file_handler = None
        self._setup_logging()

    def _setup_logging(self):
        """设置日志系统"""
        # 创建日志目录
        if not os.path.exists(EXCEPTION_LOG_DIR):
            os.makedirs(EXCEPTION_LOG_DIR)

        # 控制台处理器
        self._console_handler = logging.StreamHandler(sys.stderr)
        self._console_handler.setLevel(logging.ERROR)
        self._console_handler.setFormatter(logging.Formatter("%(message)s"))

        # 简单文件处理器
        simple_log_path = os.path.join(EXCEPTION_LOG_DIR, EXCEPTION_LOG_SIMPLE)
        self._simple_file_handler = TimedRotatingFileHandler(
            simple_log_path,
            when="midnight",
            interval=1,
            backupCount=30,
            encoding="utf-8",
        )
        self._simple_file_handler.setLevel(logging.ERROR)
        self._simple_file_handler.setFormatter(logging.Formatter("%(message)s"))

        # 详细文件处理器
        detailed_log_path = os.path.join(EXCEPTION_LOG_DIR, EXCEPTION_LOG_DETAILED)
        self._detailed_file_handler = TimedRotatingFileHandler(
            detailed_log_path,
            when="midnight",
            interval=1,
            backupCount=30,
            encoding="utf-8",
        )
        self._detailed_file_handler.setLevel(logging.ERROR)
        self._detailed_file_handler.setFormatter(logging.Formatter("%(message)s"))

        # 创建 logger
        self._logger = logging.getLogger("exception_logger")
        self._logger.setLevel(logging.ERROR)
        self._logger.addHandler(self._console_handler)
        self._logger.propagate = False

        # 文件 logger（简单）
        self._simple_logger = logging.getLogger("exception_logger_simple")
        self._simple_logger.setLevel(logging.ERROR)
        self._simple_logger.addHandler(self._simple_file_handler)
        self._simple_logger.propagate = False

        # 文件 logger（详细）
        self._detailed_logger = logging.getLogger("exception_logger_detailed")
        self._detailed_logger.setLevel(logging.ERROR)
        self._detailed_logger.addHandler(self._detailed_file_handler)
        self._detailed_logger.propagate = False

    def log_error(
        self,
        exc: BaseException,
        request_info: Optional[Dict[str, Any]] = None,
        timestamp: Optional[datetime] = None,
    ):
        """
        记录单个错误

        Args:
            exc: 异常对象
            request_info: 请求信息（可选）
            timestamp: 时间戳（可选，默认当前时间）
        """
        if timestamp is None:
            timestamp = datetime.now()

        # 格式化并输出到控制台
        console_msg = format_error_console(exc, timestamp, request_info)
        self._logger.error(console_msg)

        # 根据配置决定是否写入文件
        if Config.app.is_write_exception_log_to_file:
            # 格式化并输出到简单文件
            simple_msg = format_error_file_simple(exc, timestamp, request_info)
            self._simple_logger.error(simple_msg)

            # 格式化并输出到详细文件
            detailed_msg = format_error_file_detailed(exc, timestamp, request_info)
            self._detailed_logger.error(detailed_msg)

    def log_exception_group(
        self,
        exceptions: List[BaseException],
        request_info: Optional[Dict[str, Any]] = None,
        timestamp: Optional[datetime] = None,
    ):
        """
        记录多个错误（如 ExceptionGroup）

        Args:
            exceptions: 异常列表
            request_info: 请求信息（可选）
            timestamp: 时间戳（可选，默认当前时间）
        """
        if timestamp is None:
            timestamp = datetime.now()

        total = len(exceptions)

        for i, exc in enumerate(exceptions, 1):
            # 添加分隔符
            if i > 1:
                separator_console = format_multiple_errors_separator(
                    i, total, colorize=True
                )
                separator_file = format_multiple_errors_separator(
                    i, total, colorize=False
                )

                self._logger.error(separator_console)
                if Config.app.is_write_exception_log_to_file:
                    self._simple_logger.error(separator_file)
                    self._detailed_logger.error(separator_file)

            # 记录每个错误
            self.log_error(exc, request_info, timestamp)

    def log_exception(
        self,
        exc: BaseException,
        request_info: Optional[Dict[str, Any]] = None,
        timestamp: Optional[datetime] = None,
    ):
        """
        智能记录异常（自动处理 ExceptionGroup）

        Args:
            exc: 异常对象（可能是 ExceptionGroup）
            request_info: 请求信息（可选）
            timestamp: 时间戳（可选，默认当前时间）
        """
        # 处理 ExceptionGroup
        if hasattr(exc, "exceptions") and exc.__class__.__name__ == "ExceptionGroup":
            self.log_exception_group(
                list(exc.exceptions),
                request_info,
                timestamp,
            )
        else:
            self.log_error(exc, request_info, timestamp)


# 全局实例
exception_logger = ExceptionLogger()
