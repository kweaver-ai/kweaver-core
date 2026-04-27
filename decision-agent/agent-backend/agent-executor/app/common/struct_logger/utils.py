"""
结构化日志工具函数
"""


def safe_json_serialize(obj):
    """
    安全的 JSON 序列化函数,处理不可序列化的对象

    Args:
        obj: 需要序列化的对象

    Returns:
        可 JSON 序列化的对象
    """
    if isinstance(obj, (str, int, float, bool, type(None))):
        return obj
    elif isinstance(obj, (list, tuple)):
        return [safe_json_serialize(item) for item in obj]
    elif isinstance(obj, dict):
        return {k: safe_json_serialize(v) for k, v in obj.items()}
    elif isinstance(obj, Exception):
        # 将异常对象转换为字典
        return {
            "type": type(obj).__name__,
            "message": str(obj),
            "args": [str(arg) for arg in obj.args] if obj.args else [],
        }
    else:
        # 其他不可序列化的对象转换为字符串
        try:
            return str(obj)
        except Exception:
            return f"<{type(obj).__name__} object>"


# 向后兼容的别名
_safe_json_serialize = safe_json_serialize
