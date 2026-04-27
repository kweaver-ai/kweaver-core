"""JSON序列化器 - 优先使用

支持自定义类型的JSON序列化，包括：
- datetime -> ISO格式字符串
- Enum -> value
- Pydantic模型 -> dict
"""

import json
from datetime import datetime
from decimal import Decimal
from enum import Enum
from typing import Any

from pydantic import BaseModel


class JSONSerializer:
    """JSON序列化器 - 优先使用

    提供对常见Python类型的JSON序列化支持，包括：
    - datetime对象转ISO格式字符串
    - Enum枚举转value值
    - Pydantic模型转字典
    - Decimal数值类型
    """

    @staticmethod
    def serialize(data: Any) -> str:
        """序列化为JSON字符串

        Args:
            data: 要序列化的数据，支持dict、list、Pydantic模型等

        Returns:
            JSON格式的字符串

        Raises:
            TypeError: 当数据包含无法序列化的类型时

        Example:
            >>> serializer = JSONSerializer()
            >>> data = {"time": datetime.now(), "status": MyEnum.ACTIVE}
            >>> json_str = serializer.serialize(data)
        """
        return json.dumps(
            data, default=JSONSerializer._custom_encoder, ensure_ascii=False
        )

    @staticmethod
    def deserialize(data: str) -> Any:
        """从JSON字符串反序列化

        Args:
            data: JSON格式的字符串

        Returns:
            反序列化后的Python对象

        Raises:
            json.JSONDecodeError: 当字符串不是有效的JSON格式时

        Example:
            >>> serializer = JSONSerializer()
            >>> json_str = '{"name": "test", "count": 42}'
            >>> data = serializer.deserialize(json_str)
        """
        return json.loads(data)

    @staticmethod
    def _custom_encoder(obj: Any) -> Any:
        """自定义编码器，处理特殊类型

        Args:
            obj: 需要编码的对象

        Returns:
            可JSON序列化的值

        Raises:
            TypeError: 当对象类型无法处理时
        """
        # 处理 datetime
        if isinstance(obj, datetime):
            return obj.isoformat()

        # 处理 Enum
        if isinstance(obj, Enum):
            return obj.value

        # 处理 Pydantic 模型
        if isinstance(obj, BaseModel):
            return obj.model_dump()

        # 处理 Decimal
        if isinstance(obj, Decimal):
            return float(obj)

        # 处理 set
        if isinstance(obj, set):
            return list(obj)

        # 处理 bytes
        if isinstance(obj, bytes):
            return obj.decode("utf-8")

        # 无法处理的类型抛出异常
        raise TypeError(
            f"Object of type {type(obj).__name__} is not JSON serializable. "
            f"Consider using PickleSerializer for complex objects."
        )
