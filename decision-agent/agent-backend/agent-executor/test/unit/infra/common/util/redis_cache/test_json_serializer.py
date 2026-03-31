"""单元测试 - infra/common/util/redis_cache/json_serializer 模块"""

import pytest
import sys
from datetime import datetime
from decimal import Decimal
from enum import Enum

# Add app directory to path
sys.path.insert(
    0, "/Users/guochenguang/project/decision-agent/agent-backend/agent-executor/app"
)

# Direct import to avoid __init__.py issues
import importlib.util

spec = importlib.util.spec_from_file_location(
    "json_serializer",
    "/Users/guochenguang/project/decision-agent/agent-backend/agent-executor/app/infra/common/util/redis_cache/json_serializer.py",
)
json_serializer_module = importlib.util.module_from_spec(spec)
sys.modules["json_serializer"] = json_serializer_module
spec.loader.exec_module(json_serializer_module)

JSONSerializer = json_serializer_module.JSONSerializer


class TestJSONSerializer:
    """测试 JSONSerializer 类"""

    def test_serialize_simple_dict(self):
        """测试序列化简单字典"""
        data = {"key": "value", "number": 42}
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)
        assert "key" in serialized
        assert "value" in serialized

    def test_serialize_list(self):
        """测试序列化列表"""
        data = [1, 2, 3, "four"]
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)

    def test_serialize_datetime(self):
        """测试序列化datetime对象"""
        dt = datetime(2025, 1, 1, 12, 0, 0)
        data = {"timestamp": dt}
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)
        assert "2025-01-01T12:00:00" in serialized

    def test_serialize_enum(self):
        """测试序列化Enum"""

        class TestEnum(Enum):
            VALUE1 = "value1"
            VALUE2 = "value2"

        data = {"status": TestEnum.VALUE1}
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)
        assert "value1" in serialized

    def test_serialize_pydantic_model(self):
        """测试序列化Pydantic模型"""
        from pydantic import BaseModel

        class TestModel(BaseModel):
            name: str
            count: int

        model = TestModel(name="test", count=42)
        serialized = JSONSerializer.serialize(model)

        assert isinstance(serialized, str)
        assert "test" in serialized
        assert "42" in serialized

    def test_serialize_decimal(self):
        """测试序列化Decimal"""
        data = {"value": Decimal("3.14159")}
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)
        assert "3.14159" in serialized

    def test_serialize_set(self):
        """测试序列化set"""
        data = {"items": {1, 2, 3}}
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)

    def test_serialize_bytes(self):
        """测试序列化bytes"""
        data = {"data": b"test bytes"}
        serialized = JSONSerializer.serialize(data)

        assert isinstance(serialized, str)
        assert "test bytes" in serialized

    def test_serialize_unsupported_type_raises_error(self):
        """测试序列化不支持类型抛出异常"""

        class CustomClass:
            pass

        data = {"custom": CustomClass()}

        with pytest.raises(TypeError):
            JSONSerializer.serialize(data)

    def test_deserialize_simple_dict(self):
        """测试反序列化简单字典"""
        json_str = '{"key": "value", "number": 42}'
        deserialized = JSONSerializer.deserialize(json_str)

        assert deserialized == {"key": "value", "number": 42}

    def test_deserialize_list(self):
        """测试反序列化列表"""
        json_str = '[1, 2, 3, "four"]'
        deserialized = JSONSerializer.deserialize(json_str)

        assert deserialized == [1, 2, 3, "four"]

    def test_deserialize_nested_structure(self):
        """测试反序列化嵌套结构"""
        json_str = '{"level1": {"level2": "value"}}'
        deserialized = JSONSerializer.deserialize(json_str)

        assert deserialized == {"level1": {"level2": "value"}}

    def test_serialize_deserialize_roundtrip(self):
        """测试序列化和反序列化往返"""
        original = {
            "string": "test",
            "int": 42,
            "float": 3.14,
            "bool": True,
            "none": None,
            "list": [1, 2, 3],
            "dict": {"nested": "value"},
        }

        serialized = JSONSerializer.serialize(original)
        deserialized = JSONSerializer.deserialize(serialized)

        assert deserialized == original

    def test_serialize_datetime_roundtrip(self):
        """测试datetime序列化反序列化往返"""
        dt = datetime(2025, 1, 1, 12, 0, 0)
        original = {"timestamp": dt}

        serialized = JSONSerializer.serialize(original)
        deserialized = JSONSerializer.deserialize(serialized)

        assert isinstance(deserialized["timestamp"], str)
        assert "2025-01-01T12:00:00" in deserialized["timestamp"]

    def test_serialize_enum_roundtrip(self):
        """测试Enum序列化反序列化往返"""

        class TestEnum(Enum):
            VALUE1 = "value1"

        original = {"status": TestEnum.VALUE1}

        serialized = JSONSerializer.serialize(original)
        deserialized = JSONSerializer.deserialize(serialized)

        assert deserialized["status"] == "value1"

    def test_deserialize_invalid_json_raises_error(self):
        """测试反序列化无效JSON抛出异常"""
        import json

        invalid_json = "{invalid json}"

        with pytest.raises(json.JSONDecodeError):
            JSONSerializer.deserialize(invalid_json)

    def test_ensure_ascii_false(self):
        """测试ensure_ascii为False"""
        data = {"chinese": "中文测试"}
        serialized = JSONSerializer.serialize(data)

        assert "中文测试" in serialized
        assert "\\u" not in serialized
