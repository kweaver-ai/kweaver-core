# -*- coding: utf-8 -*-
"""
Unit tests for app/common/struct_logger/utils module
"""

import pytest

from app.common.struct_logger.utils import safe_json_serialize, _safe_json_serialize


class TestSafeJsonSerialize:
    """Tests for safe_json_serialize function"""

    @pytest.mark.asyncio
    async def test_serialize_primitives(self):
        """Test serializing primitive types"""
        assert safe_json_serialize("string") == "string"
        assert safe_json_serialize(123) == 123
        assert safe_json_serialize(3.14) == 3.14
        assert safe_json_serialize(True) is True
        assert safe_json_serialize(False) is False
        assert safe_json_serialize(None) is None

    @pytest.mark.asyncio
    async def test_serialize_list(self):
        """Test serializing lists"""
        result = safe_json_serialize([1, "two", 3.0, True, None])
        assert result == [1, "two", 3.0, True, None]

    @pytest.mark.asyncio
    async def test_serialize_tuple(self):
        """Test serializing tuples"""
        result = safe_json_serialize((1, 2, 3))
        assert result == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_serialize_dict(self):
        """Test serializing dictionaries"""
        result = safe_json_serialize({"key": "value", "number": 42})
        assert result == {"key": "value", "number": 42}

    @pytest.mark.asyncio
    async def test_serialize_nested_structures(self):
        """Test serializing nested structures"""
        result = safe_json_serialize(
            {
                "list": [1, 2, 3],
                "dict": {"nested": "value"},
                "mixed": [{"a": 1}, {"b": 2}],
            }
        )
        assert "list" in result
        assert "dict" in result
        assert "mixed" in result

    @pytest.mark.asyncio
    async def test_serialize_exception(self):
        """Test serializing exception objects"""
        exc = ValueError("Test error")
        result = safe_json_serialize(exc)

        assert isinstance(result, dict)
        assert result["type"] == "ValueError"
        assert result["message"] == "Test error"
        assert "args" in result

    @pytest.mark.asyncio
    async def test_serialize_exception_with_args(self):
        """Test serializing exception with arguments"""
        exc = ValueError("arg1", "arg2")
        result = safe_json_serialize(exc)

        assert isinstance(result, dict)
        assert result["type"] == "ValueError"
        assert len(result["args"]) == 2

    @pytest.mark.asyncio
    async def test_serialize_exception_without_args(self):
        """Test serializing exception without arguments"""
        exc = Exception()
        result = safe_json_serialize(exc)

        assert isinstance(result, dict)
        assert result["type"] == "Exception"
        assert result["args"] == []

    @pytest.mark.asyncio
    async def test_serialize_custom_object(self):
        """Test serializing custom objects (converts to string)"""

        class CustomObject:
            def __str__(self):
                return "CustomObject()"

        obj = CustomObject()
        result = safe_json_serialize(obj)

        assert result == "CustomObject()"

    @pytest.mark.asyncio
    async def test_serialize_object_with_str_error(self):
        """Test serializing object that raises exception in __str__"""

        class BadObject:
            def __str__(self):
                raise RuntimeError("Cannot convert to string")

        obj = BadObject()
        result = safe_json_serialize(obj)

        # Should return a placeholder string
        assert "BadObject" in result

    @pytest.mark.asyncio
    async def test_serialize_complex_nested(self):
        """Test serializing complex nested structures with mixed types"""
        data = {
            "users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}],
            "metadata": {"count": 2, "active": True},
            "errors": None,
        }
        result = safe_json_serialize(data)

        assert result["users"][0]["name"] == "Alice"
        assert result["metadata"]["count"] == 2
        assert result["errors"] is None

    @pytest.mark.asyncio
    async def test_backward_compatibility_alias(self):
        """Test that _safe_json_serialize is an alias for safe_json_serialize"""
        assert _safe_json_serialize is safe_json_serialize

        # Test both functions produce same result
        data = {"key": "value"}
        result1 = safe_json_serialize(data)
        result2 = _safe_json_serialize(data)

        assert result1 == result2


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that utils module can be imported"""
        from app.common.struct_logger import utils

        assert utils is not None
        assert hasattr(utils, "safe_json_serialize")
        assert hasattr(utils, "_safe_json_serialize")
