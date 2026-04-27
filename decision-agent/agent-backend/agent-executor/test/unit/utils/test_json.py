"""单元测试 - utils/json 模块"""

import pytest
from datetime import datetime
from decimal import Decimal
from enum import Enum


class TestCustomSerializer:
    """测试 custom_serializer 函数"""

    def test_serialize_datetime(self):
        """测试序列化datetime"""
        from app.utils.json import custom_serializer

        dt = datetime(2025, 1, 1, 12, 0, 0)
        result = custom_serializer(dt)

        assert result == "2025-01-01T12:00:00"

    def test_serialize_decimal(self):
        """测试序列化Decimal"""
        from app.utils.json import custom_serializer

        dec = Decimal("3.14159")
        result = custom_serializer(dec)

        assert result == 3.14159

    def test_serialize_enum(self):
        """测试序列化Enum"""
        from app.utils.json import custom_serializer

        class TestEnum(Enum):
            VALUE1 = "value1"

        result = custom_serializer(TestEnum.VALUE1)

        assert result == "value1"

    def test_serialize_set(self):
        """测试序列化set"""
        from app.utils.json import custom_serializer

        result = custom_serializer({1, 2, 3})

        assert result == [1, 2, 3]

    def test_serialize_frozenset(self):
        """测试序列化frozenset"""
        from app.utils.json import custom_serializer

        result = custom_serializer(frozenset([1, 2, 3]))

        assert result == [1, 2, 3]

    def test_serialize_object_with_dict(self):
        """测试序列化具有__dict__的对象"""
        from app.utils.json import custom_serializer

        class CustomObject:
            def __init__(self):
                self.value = "test"

        obj = CustomObject()
        result = custom_serializer(obj)

        assert result == {"value": "test"}


class TestJsonSerializeAsync:
    """测试 json_serialize_async 函数"""

    @pytest.mark.asyncio
    async def test_serialize_simple_dict(self):
        """测试异步序列化简单字典"""
        from app.utils.json import json_serialize_async

        data = {"key": "value", "number": 42}
        result = await json_serialize_async(data)

        assert isinstance(result, str)
        assert "key" in result
        assert "value" in result

    @pytest.mark.asyncio
    async def test_serialize_with_datetime(self):
        """测试异步序列化包含datetime"""
        from app.utils.json import json_serialize_async

        data = {"timestamp": datetime(2025, 1, 1, 12, 0, 0)}
        result = await json_serialize_async(data)

        assert isinstance(result, str)
        assert "2025-01-01T12:00:00" in result

    @pytest.mark.asyncio
    async def test_ensure_ascii_false(self):
        """测试ensure_ascii为False"""
        from app.utils.json import json_serialize_async

        data = {"chinese": "中文测试"}
        result = await json_serialize_async(data)

        assert "中文测试" in result
