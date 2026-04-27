"""
时间戳辅助工具

用于在 BIGINT 毫秒时间戳和 datetime 对象之间进行转换。
按照数据表命名规范，时间戳字段使用 BIGINT 存储毫秒级时间戳。
"""
import time
from datetime import datetime
from typing import Optional


def datetime_to_millis(dt: Optional[datetime]) -> int:
    """
    将 datetime 对象转换为毫秒时间戳

    Args:
        dt: datetime 对象，如果为 None 则返回当前时间的毫秒时间戳

    Returns:
        毫秒时间戳
    """
    if dt is None:
        return int(time.time() * 1000)
    return int(dt.timestamp() * 1000)


def millis_to_datetime(millis: Optional[int]) -> Optional[datetime]:
    """
    将毫秒时间戳转换为 datetime 对象

    Args:
        millis: 毫秒时间戳，如果为 None 或 0 则返回 None

    Returns:
        datetime 对象，如果输入无效则返回 None
    """
    if not millis or millis == 0:
        return None
    try:
        return datetime.fromtimestamp(millis / 1000)
    except (ValueError, OSError):
        return None


def current_millis() -> int:
    """
    获取当前时间的毫秒时间戳

    Returns:
        当前时间的毫秒时间戳
    """
    return int(time.time() * 1000)


def current_datetime() -> datetime:
    """
    获取当前时间的 datetime 对象

    Returns:
        当前时间的 datetime 对象
    """
    return datetime.now()
