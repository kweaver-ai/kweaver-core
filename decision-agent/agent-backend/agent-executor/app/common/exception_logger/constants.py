# -*- coding:utf-8 -*-
"""
异常日志常量定义
"""

import os

# 项目根目录标识（用于过滤第三方库代码）
PROJECT_ROOT = os.path.dirname(
    os.path.dirname(os.path.dirname(os.path.dirname(os.path.dirname(__file__))))
)

# 日志目录
EXCEPTION_LOG_DIR = "log/exceptions"

# 日志文件名
EXCEPTION_LOG_SIMPLE = "exception_simple.log"  # 简单日志（仅项目代码）
EXCEPTION_LOG_DETAILED = "exception_detailed.log"  # 详细日志（包含第三方库）

# ANSI 颜色代码
RESET = "\033[0m"
BOLD = "\033[1m"
DIM = "\033[2m"
UNDERLINE = "\033[4m"

# 颜色定义
COLORS = {
    "timestamp": "\033[90m",  # 灰色 - 时间戳
    "error": "\033[31m",  # 红色 - ERROR级别
    "critical": "\033[35m",  # 紫色 - CRITICAL级别
    "warning": "\033[33m",  # 黄色 - WARNING级别
    "caller": "\033[94m",  # 亮蓝色 - 调用位置（方便点击跳转）
    "key": "\033[96m",  # 亮青色 - 字段名
    "value": "\033[37m",  # 白色 - 字段值
    "error_value": "\033[31m",  # 红色 - 错误相关字段值
    "border": "\033[90m",  # 灰色 - 边界线
    "exception_type": "\033[91m",  # 亮红色 - 异常类型
    "exception_msg": "\033[93m",  # 亮黄色 - 异常消息
    "traceback": "\033[90m",  # 灰色 - 堆栈信息
    "project_code": "\033[92m",  # 亮绿色 - 项目代码（更醒目）
    "separator": "\033[95m",  # 亮紫色 - 分隔符
}

# 级别表情符号
LEVEL_EMOJI = {
    "ERROR": "❌",
    "CRITICAL": "🔥",
}

# 边界字符
BORDER_DOUBLE = "═"
BORDER_SINGLE = "─"
BORDER_DOT = "┄"

# 边界宽度
BORDER_WIDTH = 100
