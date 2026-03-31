import datetime
import decimal
import uuid
import enum

from concurrent.futures import ThreadPoolExecutor
import asyncio
import functools
import json


def custom_serializer(obj):
    """
    通用 JSON 序列化器，支持多种 Python 特殊类型和自定义对象。

    支持的类型包括：
    - 基本类型 (None, int, float, str, bool, list, dict)
    - datetime/datetime.time/datetime.date
    - decimal.Decimal
    - uuid.UUID
    - enum.Enum
    - 自定义类实例（转换为字典）
    - 集合（set/frozenset 转换为列表）
    - 生成器和迭代器（转换为列表）
    - 其他具有 __dict__ 属性的对象
    """
    # 基本类型直接返回（JSON原生支持）
    if obj is None or isinstance(obj, (int, float, str, bool, list, dict)):
        return obj

    # 日期时间类型
    if isinstance(obj, (datetime.datetime, datetime.time, datetime.date)):
        return obj.isoformat()

    # 数值类型
    if isinstance(obj, decimal.Decimal):
        return float(obj)

    # UUID 类型
    if isinstance(obj, uuid.UUID):
        return str(obj)

    # 枚举类型
    if isinstance(obj, enum.Enum):
        return obj.value

    # 集合类型
    if isinstance(obj, (set, frozenset)):
        return list(obj)

    # 尝试使用 __dict__ 属性序列化对象
    if hasattr(obj, "__dict__"):
        # 如果 __dict__ 为空，检查是否有其他可序列化的属性
        obj_dict = obj.__dict__
        if obj_dict or hasattr(obj, "__slots__"):
            return obj_dict

    # 其他类型抛出错误
    raise TypeError(f"Object of type {obj.__class__.__name__} is not JSON serializable")


thread_pool_for_json_encode = ThreadPoolExecutor()


async def json_serialize_async(data: dict) -> str:
    """异步JSON序列化（使用线程池）"""
    loop = asyncio.get_event_loop()
    # 使用 functools.partial 包装 json.dumps 及参数
    dumps_partial = functools.partial(
        json.dumps,
        data,
        ensure_ascii=False,
        default=custom_serializer,  # 假设 custom_serializer 是自定义的序列化函数
    )
    # 将同步的json.dumps放到线程池执行
    return await loop.run_in_executor(thread_pool_for_json_encode, dumps_partial)
