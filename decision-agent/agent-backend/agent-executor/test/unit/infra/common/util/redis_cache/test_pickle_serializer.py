"""单元测试 - infra/common/util/redis_cache/pickle_serializer 模块"""

import sys
import pickle
from datetime import datetime

# Add app directory to path
sys.path.insert(
    0, "/Users/guochenguang/project/decision-agent/agent-backend/agent-executor/app"
)

# Direct import to avoid __init__.py issues
import importlib.util

spec = importlib.util.spec_from_file_location(
    "pickle_serializer",
    "/Users/guochenguang/project/decision-agent/agent-backend/agent-executor/app/infra/common/util/redis_cache/pickle_serializer.py",
)
pickle_serializer_module = importlib.util.module_from_spec(spec)
sys.modules["pickle_serializer"] = pickle_serializer_module
spec.loader.exec_module(pickle_serializer_module)

PickleSerializer = pickle_serializer_module.PickleSerializer


# Define test class at module level for pickle compatibility
class CustomObject:
    def __init__(self, value):
        self.value = value

    def __eq__(self, other):
        return isinstance(other, CustomObject) and self.value == other.value


class TestPickleSerializer:
    """测试 PickleSerializer 类"""

    def test_serialize_simple_dict(self):
        """测试序列化简单字典"""
        data = {"key": "value", "number": 42}
        serialized = PickleSerializer.serialize(data)

        assert isinstance(serialized, bytes)
        assert len(serialized) > 0

    def test_serialize_list(self):
        """测试序列化列表"""
        data = [1, 2, 3, "four", {"five": 5}]
        serialized = PickleSerializer.serialize(data)

        assert isinstance(serialized, bytes)

    def test_serialize_complex_object(self):
        """测试序列化复杂对象"""
        data = CustomObject(42)
        serialized = PickleSerializer.serialize(data)

        assert isinstance(serialized, bytes)

    def test_serialize_datetime(self):
        """测试序列化datetime对象"""
        data = {"timestamp": datetime.now(), "value": 42}
        serialized = PickleSerializer.serialize(data)

        assert isinstance(serialized, bytes)

    def test_serialize_nested_structures(self):
        """测试序列化嵌套结构"""
        data = {"level1": {"level2": {"level3": "deep_value"}, "list": [1, 2, 3]}}
        serialized = PickleSerializer.serialize(data)

        assert isinstance(serialized, bytes)

    def test_deserialize_simple_dict(self):
        """测试反序列化简单字典"""
        original = {"key": "value", "number": 42}
        serialized = PickleSerializer.serialize(original)
        deserialized = PickleSerializer.deserialize(serialized)

        assert deserialized == original

    def test_deserialize_list(self):
        """测试反序列化列表"""
        original = [1, 2, 3, "four"]
        serialized = PickleSerializer.serialize(original)
        deserialized = PickleSerializer.deserialize(serialized)

        assert deserialized == original

    def test_deserialize_complex_object(self):
        """测试反序列化复杂对象"""
        original = CustomObject(42)
        serialized = PickleSerializer.serialize(original)
        deserialized = PickleSerializer.deserialize(serialized)

        assert deserialized.value == original.value

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

        serialized = PickleSerializer.serialize(original)
        deserialized = PickleSerializer.deserialize(serialized)

        assert deserialized == original

    def test_serialize_returns_bytes(self):
        """测试序列化返回bytes类型"""
        serialized = PickleSerializer.serialize({"test": "data"})

        assert isinstance(serialized, bytes)

    def test_deserialize_accepts_bytes(self):
        """测试反序列化接受bytes类型"""
        data = {"test": "data"}
        serialized = pickle.dumps(data)

        deserialized = PickleSerializer.deserialize(serialized)
        assert deserialized == data
