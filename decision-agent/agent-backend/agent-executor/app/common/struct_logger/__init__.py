"""
结构化日志模块 - 基于 structlog
提供更好的日志格式化和可读性

使用示例:
    from app.common.struct_logger import struct_logger

    struct_logger.info("用户登录", user_id="123", action="login")
    struct_logger.error("处理失败", error="timeout", retry_count=3)
"""

from .logger import StructLogger
from .constants import SYSTEM_LOG, BUSINESS_LOG
from .utils import safe_json_serialize, _safe_json_serialize

# 创建全局实例
struct_logger = StructLogger()

# 导出独立的logger实例，方便直接使用
file_logger = struct_logger.file_logger  # 仅写入文件
console_logger = struct_logger.console_logger  # 仅输出到控制台

# 导出所有公共接口
__all__ = [
    "struct_logger",
    "file_logger",
    "console_logger",
    "StructLogger",
    "SYSTEM_LOG",
    "BUSINESS_LOG",
    "safe_json_serialize",
    "_safe_json_serialize",  # 向后兼容
]
