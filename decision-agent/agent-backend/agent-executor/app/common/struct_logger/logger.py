"""
结构化日志记录器主类
"""

import os
import structlog


from .constants import LOG_DIR
from .file_logging_setup import setup_file_logging
from .console_logging_setup import setup_console_logging


class StructLogger:
    """结构化日志记录器"""

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
        self._file_logger = None  # 文件日志
        self._console_logger = None  # 控制台日志
        self._setup_logging()

    def _setup_logging(self):
        """设置日志系统"""
        # 创建日志目录
        if not os.path.exists(LOG_DIR):
            os.makedirs(LOG_DIR)

        # 配置文件日志
        self._file_logger = setup_file_logging()

        # 配置控制台日志
        self._console_logger = setup_console_logging()

    def debug(self, event: str, **kwargs):
        """记录 DEBUG 级别日志"""
        self._file_logger.debug(event, **kwargs)
        self._console_logger.debug(event, **kwargs)

    def info(self, event: str, **kwargs):
        """记录 INFO 级别日志"""
        self._file_logger.info(event, **kwargs)
        self._console_logger.info(event, **kwargs)

    def warning(self, event: str, **kwargs):
        """记录 WARNING 级别日志"""
        self._file_logger.warning(event, **kwargs)
        self._console_logger.warning(event, **kwargs)

    def warn(self, event: str, **kwargs):
        """warning 的别名"""
        self.warning(event, **kwargs)

    def error(self, event: str, **kwargs):
        """记录 ERROR 级别日志"""

        self._file_logger.error(event, **kwargs)
        self._console_logger.error(event, **kwargs)

    def fatal(self, event: str, **kwargs):
        """记录 FATAL 级别日志"""
        self._file_logger.critical(event, **kwargs)
        self._console_logger.critical(event, **kwargs)

    def bind(self, **kwargs) -> structlog.BoundLogger:
        """绑定上下文信息到日志（同时绑定到文件和控制台）"""
        # 返回文件logger的绑定，控制台logger会自动跟随
        return self._file_logger.bind(**kwargs)

    @property
    def file_logger(self) -> structlog.BoundLogger:
        """获取文件日志记录器"""
        return self._file_logger

    @property
    def console_logger(self) -> structlog.BoundLogger:
        """获取控制台日志记录器"""
        return self._console_logger
