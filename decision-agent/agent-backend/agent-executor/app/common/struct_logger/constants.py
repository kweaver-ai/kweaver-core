"""
结构化日志常量定义
"""

# 日志目录
LOG_DIR = "log"

# 日志类型
SYSTEM_LOG = "SystemLog"
BUSINESS_LOG = "BusinessLog"

# ANSI 颜色代码
RESET = "\033[0m"
BOLD = "\033[1m"
DIM = "\033[2m"

# 颜色定义
COLORS = {
    "timestamp": "\033[90m",  # 灰色 - 时间戳
    "debug": "\033[36m",  # 青色 - DEBUG级别
    "info": "\033[32m",  # 绿色 - INFO级别
    "warning": "\033[33m",  # 黄色 - WARNING级别
    "error": "\033[31m",  # 红色 - ERROR级别
    "critical": "\033[35m",  # 紫色 - CRITICAL级别
    "caller": "\033[94m",  # 亮蓝色 - 调用位置
    "key": "\033[96m",  # 亮青色 - 字段名
    "value": "\033[37m",  # 白色 - 字段值 (从亮白色改为白色,更明显)
    "error_value": "\033[31m",  # 红色 - 错误相关字段值
    "border": "\033[90m",  # 灰色 - 边界线
    "exception_type": "\033[91m",  # 亮红色 - 异常类型
    "exception_msg": "\033[93m",  # 亮黄色 - 异常消息
    "traceback": "\033[90m",  # 灰色 - 堆栈信息
}

# 级别表情符号
LEVEL_EMOJI = {
    "DEBUG": "🔍",
    "INFO": "ℹ️",
    "WARNING": "⚠️",
    "ERROR": "❌",
    "CRITICAL": "🔥",
}
