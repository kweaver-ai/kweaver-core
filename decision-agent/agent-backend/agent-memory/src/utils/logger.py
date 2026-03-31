import logging
import sys
from pathlib import Path
from logging.handlers import RotatingFileHandler
from typing import Optional


class Logger:
    """日志工具类"""

    _instance: Optional["Logger"] = None
    _initialized = False

    def __new__(cls):
        if cls._instance is None:
            cls._instance = super().__new__(cls)
        return cls._instance

    def __init__(self):
        if not self._initialized:
            self._initialized = True
            self._setup_logger()

    def _setup_logger(self):
        """配置日志"""
        # 创建日志目录
        log_dir = Path("logs")
        log_dir.mkdir(exist_ok=True)

        # 创建日志记录器
        self.logger = logging.getLogger("agent_memory")
        self.logger.setLevel(logging.DEBUG)

        # 创建控制台处理器
        console_handler = logging.StreamHandler(sys.stdout)
        console_handler.setLevel(logging.INFO)
        console_format = logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )
        console_handler.setFormatter(console_format)

        # 创建文件处理器
        file_handler = RotatingFileHandler(
            log_dir / "agent_memory.log",
            maxBytes=10 * 1024 * 1024,  # 10MB
            backupCount=5,
            encoding="utf-8",
        )
        file_handler.setLevel(logging.DEBUG)
        file_format = logging.Formatter(
            "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        )
        file_handler.setFormatter(file_format)

        # 添加处理器
        self.logger.addHandler(console_handler)
        self.logger.addHandler(file_handler)

    def info(self, msg: str, *args, **kwargs):
        """记录信息日志"""
        self.logger.info(
            msg.format(*args),
            exc_info=kwargs.get("exc_info"),
            extra=kwargs.get("extra"),
        )

    def error(self, msg: str, *args, **kwargs):
        """记录信息日志"""
        self.logger.error(
            msg.format(*args),
            exc_info=kwargs.get("exc_info"),
            extra=kwargs.get("extra"),
        )

    def warning(self, msg: str, *args, **kwargs):
        """记录警告日志"""
        self.logger.warning(
            msg.format(*args),
            exc_info=kwargs.get("exc_info"),
            extra=kwargs.get("extra"),
        )

    def debug(self, msg: str, *args, **kwargs):
        """记录信息日志"""
        self.logger.debug(
            msg.format(*args),
            exc_info=kwargs.get("exc_info"),
            extra=kwargs.get("extra"),
        )

    def errorf(self, msg: str, *args, **kwargs):
        """格式化 error 日志"""
        self.logger.error(
            msg % args, exc_info=kwargs.get("exc_info"), extra=kwargs.get("extra")
        )

    def infof(self, msg: str, *args, **kwargs):
        """格式化 info 日志"""
        self.logger.info(msg % args, extra=kwargs.get("extra"))

    def debugf(self, msg: str, *args, **kwargs):
        """格式化 debug 日志"""
        self.logger.debug(msg % args, extra=kwargs.get("extra"))

    def warningf(self, msg: str, *args, **kwargs):
        """格式化 warning 日志"""
        self.logger.warning(msg % args, extra=kwargs.get("extra"))


# 创建全局日志实例
logger = Logger()
