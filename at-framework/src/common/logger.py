import os
import logging
from datetime import datetime
from src.config.setting import LOG_DIR


class Logger:
    def __init__(self, name=__name__):
        # 创建日志器
        self.logger = logging.getLogger(name)
        
        # 从环境变量获取日志级别，默认INFO
        log_level_str = os.getenv("LOCUST_LOG_LEVEL", "INFO").upper()
        log_level = getattr(logging, log_level_str, logging.INFO)
        self.logger.setLevel(log_level)

        # 避免日志重复输出
        if self.logger.handlers:
            return

        # 日志格式
        fmt = "%(asctime)s - %(name)s - %(levelname)s - %(message)s"
        formatter = logging.Formatter(fmt)

        # 控制台处理器
        console_handler = logging.StreamHandler()
        console_handler.setLevel(log_level)
        console_handler.setFormatter(formatter)

        # 日志文件名称
        now = datetime.now().strftime("%Y-%m-%d")
        log_file = os.path.join(LOG_DIR, f"{now}.log")

        # 文件处理器 - 始终记录DEBUG级别以上的日志到文件
        file_handler = logging.FileHandler(log_file, encoding="utf-8")
        file_handler.setLevel(logging.DEBUG)
        file_handler.setFormatter(formatter)

        # 添加处理器
        self.logger.addHandler(console_handler)
        self.logger.addHandler(file_handler)

    def get_logger(self):
        return self.logger


# 实例化日志对象
logger = Logger().get_logger()
