#!/usr/bin/env python3
# coding=utf-8
import logging.handlers
import os
import sys
import logging
import traceback

class CustomFormatter(logging.Formatter):
    def format(self, record):
        call_stack = traceback.extract_stack()[:-1]
        current_path = os.getcwd()
        for item in call_stack[::-1]:
            filename, line, _, _ = item
            if not filename.startswith(current_path):
                continue
            record.lineno = line
            record.filename = filename
            break

        # 使用 Formatter 类的 format 方法进行格式化
        formatted_record = logging.Formatter.format(self, record)

        return formatted_record


_nameToLevel = {
    "CRITICAL": logging.CRITICAL,
    "FATAL": logging.FATAL,
    "ERROR": logging.ERROR,
    "WARN": logging.WARNING,
    "WARNING": logging.WARNING,
    "INFO": logging.INFO,
    "DEBUG": logging.DEBUG,
    "NOTSET": logging.NOTSET,
}

log_level = _nameToLevel.get(os.environ.get("LOG_LEVEL", "INFO").upper())
basic_log_level = _nameToLevel.get(os.environ.get("BASIC_LOG_LEVEL", os.environ.get("LOG_LEVEL", "INFO")).upper())

logging.basicConfig(level=basic_log_level)

logger = logging.Logger("CodeRunner")
logger.setLevel(level=log_level) # type: ignore

formatter = CustomFormatter("[%(asctime)s] %(levelname)s %(filename)s line:%(lineno)d message:%(message)s")
# 控制台输出
console_handler = logging.StreamHandler(sys.stdout)
console_handler.setFormatter(formatter)
logger.addHandler(console_handler)

class RequestPathExclusionFilter(logging.Filter):
    """
    请求路径排除过滤器
    
    用于过滤包含特定路径的日志记录，避免某些路径的日志被记录
    """

    def filter(self, record):
        message = record.getMessage()
        # 过滤包含特定路径的日志
        skip_patterns = ['/health']
        return not any(pattern in message for pattern in skip_patterns)

# 应用过滤器
access_log = logging.getLogger("tornado.access")
access_log.addFilter(RequestPathExclusionFilter())