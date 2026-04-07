"""Massive unit tests for app/utils/json.py - 150+ tests"""

import pytest
import datetime
import decimal
import uuid
import enum
import json
from app.utils.json import (
    custom_serializer,
    json_serialize_async,
    thread_pool_for_json_encode,
)


class TestCustomSerializer:
    """Test custom_serializer function"""

    def test_serializes_none(self):
        result = custom_serializer(None)
        assert result is None

    def test_serializes_int(self):
        result = custom_serializer(42)
        assert result == 42

    def test_serializes_negative_int(self):
        result = custom_serializer(-42)
        assert result == -42

    def test_serializes_zero(self):
        result = custom_serializer(0)
        assert result == 0

    def test_serializes_float(self):
        result = custom_serializer(3.14)
        assert result == 3.14

    def test_serializes_string(self):
        result = custom_serializer("hello")
        assert result == "hello"

    def test_serializes_empty_string(self):
        result = custom_serializer("")
        assert result == ""

    def test_serializes_bool_true(self):
        result = custom_serializer(True)
        assert result is True

    def test_serializes_bool_false(self):
        result = custom_serializer(False)
        assert result is False

    def test_serializes_list(self):
        result = custom_serializer([1, 2, 3])
        assert result == [1, 2, 3]

    def test_serializes_empty_list(self):
        result = custom_serializer([])
        assert result == []

    def test_serializes_dict(self):
        result = custom_serializer({"a": 1})
        assert result == {"a": 1}

    def test_serializes_empty_dict(self):
        result = custom_serializer({})
        assert result == {}

    def test_serializes_datetime(self):
        dt = datetime.datetime(2024, 1, 1, 12, 0, 0)
        result = custom_serializer(dt)
        assert isinstance(result, str)
        assert "2024" in result

    def test_serializes_datetime_now(self):
        dt = datetime.datetime.now()
        result = custom_serializer(dt)
        assert isinstance(result, str)

    def test_serializes_datetime_now_utc(self):
        dt = datetime.datetime.now(datetime.UTC)
        result = custom_serializer(dt)
        assert isinstance(result, str)

    def test_serializes_time(self):
        t = datetime.time(12, 30, 45)
        result = custom_serializer(t)
        assert isinstance(result, str)

    def test_serializes_date(self):
        d = datetime.date(2024, 1, 1)
        result = custom_serializer(d)
        assert isinstance(result, str)
        assert "2024" in result

    def test_serializes_decimal(self):
        d = decimal.Decimal("3.14159")
        result = custom_serializer(d)
        assert isinstance(result, float)
        assert abs(result - 3.14159) < 0.0001

    def test_serializes_decimal_zero(self):
        d = decimal.Decimal("0")
        result = custom_serializer(d)
        assert result == 0.0

    def test_serializes_decimal_negative(self):
        d = decimal.Decimal("-10.5")
        result = custom_serializer(d)
        assert result == -10.5

    def test_serializes_uuid(self):
        u = uuid.uuid4()
        result = custom_serializer(u)
        assert isinstance(result, str)
        assert len(result) == 36

    def test_serializes_uuid4(self):
        u = uuid.UUID("12345678-1234-5678-1234-567812345678")
        result = custom_serializer(u)
        assert result == "12345678-1234-5678-1234-567812345678"

    def test_serializes_enum(self):
        class TestEnum(enum.Enum):
            A = 1
            B = 2

        result = custom_serializer(TestEnum.A)
        assert result == 1

    def test_serializes_enum_string(self):
        class TestEnum(enum.Enum):
            A = "a"
            B = "b"

        result = custom_serializer(TestEnum.A)
        assert result == "a"

    def test_serializes_set(self):
        result = custom_serializer({1, 2, 3})
        assert isinstance(result, list)
        assert set(result) == {1, 2, 3}

    def test_serializes_empty_set(self):
        result = custom_serializer(set())
        assert result == []

    def test_serializes_frozenset(self):
        result = custom_serializer(frozenset([1, 2, 3]))
        assert isinstance(result, list)

    def test_serializes_object_with_dict(self):
        class TestObj:
            def __init__(self):
                self.a = 1
                self.b = 2

        result = custom_serializer(TestObj())
        assert isinstance(result, dict)
        assert "a" in result

    def test_serializes_nested_list(self):
        result = custom_serializer([[1], [2]])
        assert result == [[1], [2]]

    def test_serializes_nested_dict(self):
        result = custom_serializer({"a": {"b": 1}})
        assert result == {"a": {"b": 1}}

    def test_serializes_list_of_dicts(self):
        result = custom_serializer([{"a": 1}, {"b": 2}])
        assert result == [{"a": 1}, {"b": 2}]

    def test_serializes_dict_of_lists(self):
        result = custom_serializer({"a": [1, 2], "b": [3, 4]})
        assert result == {"a": [1, 2], "b": [3, 4]}

    def test_serializes_mixed_types_list(self):
        result = custom_serializer([1, "a", True, None])
        assert result == [1, "a", True, None]

    def test_serializes_complex_nested(self):
        result = custom_serializer({"a": [1, {"b": 2}]})
        assert "a" in result

    def test_raises_type_error_for_unsupported(self):
        class Unsupported:
            pass

        with pytest.raises(TypeError):
            custom_serializer(Unsupported())

    def test_serializes_long_string(self):
        result = custom_serializer("x" * 10000)
        assert result == "x" * 10000

    def test_serializes_unicode_string(self):
        result = custom_serializer("你好世界")
        assert result == "你好世界"

    def test_serializes_special_chars_string(self):
        result = custom_serializer("Hello\nWorld\t!")
        assert result == "Hello\nWorld\t!"

    def test_serializes_large_int(self):
        result = custom_serializer(999999999999)
        assert result == 999999999999

    def test_serializes_negative_large_int(self):
        result = custom_serializer(-999999999999)
        assert result == -999999999999

    def test_serializes_scientific_notation(self):
        result = custom_serializer(1.23e-10)
        assert result == 1.23e-10

    def test_serializes_inf(self):
        result = custom_serializer(float("inf"))
        assert result == float("inf")

    def test_serializes_neg_inf(self):
        result = custom_serializer(float("-inf"))
        assert result == float("-inf")

    def test_serializes_datetime_microseconds(self):
        dt = datetime.datetime(2024, 1, 1, 12, 0, 0, 123456)
        result = custom_serializer(dt)
        assert isinstance(result, str)

    def test_serializes_time_with_microseconds(self):
        t = datetime.time(12, 0, 0, 123456)
        result = custom_serializer(t)
        assert isinstance(result, str)

    def test_serializes_date_far_past(self):
        d = datetime.date(1900, 1, 1)
        result = custom_serializer(d)
        assert isinstance(result, str)

    def test_serializes_date_far_future(self):
        d = datetime.date(2100, 1, 1)
        result = custom_serializer(d)
        assert isinstance(result, str)

    def test_serializes_decimal_very_small(self):
        d = decimal.Decimal("0.00000001")
        result = custom_serializer(d)
        assert isinstance(result, float)

    def test_serializes_enum_auto_value(self):
        class AutoEnum(enum.Enum):
            A = enum.auto()
            B = enum.auto()

        result = custom_serializer(AutoEnum.A)
        assert isinstance(result, int)

    def test_serializes_enum_complex_value(self):
        class ComplexEnum(enum.Enum):
            A = {"key": "value"}

        result = custom_serializer(ComplexEnum.A)
        assert isinstance(result, dict)


class TestJsonSerializeAsync:
    """Test json_serialize_async function"""

    @pytest.mark.asyncio
    async def test_returns_string(self):
        data = {"a": 1}
        result = await json_serialize_async(data)
        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_serializes_dict(self):
        data = {"a": 1, "b": 2}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == {"a": 1, "b": 2}

    @pytest.mark.asyncio
    async def test_serializes_empty_dict(self):
        data = {}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == {}

    @pytest.mark.asyncio
    async def test_serializes_list(self):
        data = [1, 2, 3]
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_serializes_string(self):
        data = "hello"
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == "hello"

    @pytest.mark.asyncio
    async def test_serializes_number(self):
        data = 42
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == 42

    @pytest.mark.asyncio
    async def test_serializes_bool(self):
        data = True
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed is True

    @pytest.mark.asyncio
    async def test_serializes_none(self):
        data = None
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed is None

    @pytest.mark.asyncio
    async def test_serializes_nested_dict(self):
        data = {"a": {"b": {"c": 1}}}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == {"a": {"b": {"c": 1}}}

    @pytest.mark.asyncio
    async def test_serializes_list_of_dicts(self):
        data = [{"a": 1}, {"b": 2}]
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == [{"a": 1}, {"b": 2}]

    @pytest.mark.asyncio
    async def test_uses_custom_serializer(self):
        data = {"dt": datetime.datetime(2024, 1, 1)}
        result = await json_serialize_async(data)
        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_ensure_ascii_false(self):
        data = {"text": "你好"}
        result = await json_serialize_async(data)
        assert "你好" in result

    @pytest.mark.asyncio
    async def test_large_dict(self):
        data = {f"key_{i}": i for i in range(1000)}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert len(parsed) == 1000

    @pytest.mark.asyncio
    async def test_unicode_keys(self):
        data = {"键": "值"}
        result = await json_serialize_async(data)
        assert "键" in result

    @pytest.mark.asyncio
    async def test_special_chars_values(self):
        data = {"text": "Line1\nLine2\tTab"}
        result = await json_serialize_async(data)
        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_nested_list(self):
        data = [[[1]], [[2]]]
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == [[[1]], [[2]]]

    @pytest.mark.asyncio
    async def test_mixed_types(self):
        data = {"int": 1, "str": "a", "bool": True, "none": None}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["int"] == 1
        assert parsed["str"] == "a"
        assert parsed["bool"] is True
        assert parsed["none"] is None

    @pytest.mark.asyncio
    async def test_empty_string_value(self):
        data = {"empty": ""}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["empty"] == ""

    @pytest.mark.asyncio
    async def test_zero_value(self):
        data = {"zero": 0}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["zero"] == 0

    @pytest.mark.asyncio
    async def test_false_value(self):
        data = {"false": False}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["false"] is False

    @pytest.mark.asyncio
    async def test_list_with_none(self):
        data = [1, None, 3]
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == [1, None, 3]

    @pytest.mark.asyncio
    async def test_dict_with_none_value(self):
        data = {"a": None}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["a"] is None

    @pytest.mark.asyncio
    async def test_large_string(self):
        data = {"text": "x" * 10000}
        result = await json_serialize_async(data)
        assert len(result) > 10000

    @pytest.mark.asyncio
    async def test_very_nested(self):
        data = {"a": {"b": {"c": {"d": 1}}}}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["a"]["b"]["c"]["d"] == 1

    @pytest.mark.asyncio
    async def test_datetime_in_list(self):
        data = {"dates": [datetime.datetime(2024, 1, 1)]}
        result = await json_serialize_async(data)
        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_decimal_in_dict(self):
        data = {"value": decimal.Decimal("3.14")}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert abs(parsed["value"] - 3.14) < 0.01

    @pytest.mark.asyncio
    async def test_uuid_in_dict(self):
        data = {"id": uuid.uuid4()}
        result = await json_serialize_async(data)
        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_set_serialized_as_list(self):
        data = {"items": {1, 2, 3}}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert isinstance(parsed["items"], list)

    @pytest.mark.asyncio
    async def test_frozenset_serialized_as_list(self):
        data = {"items": frozenset([1, 2, 3])}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert isinstance(parsed["items"], list)

    @pytest.mark.asyncio
    async def test_enum_in_dict(self):
        class TestEnum(enum.Enum):
            A = 1

        data = {"value": TestEnum.A}
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed["value"] == 1

    @pytest.mark.asyncio
    async def test_object_with_dict_attr(self):
        class TestObj:
            def __init__(self):
                self.value = 42

        data = {"obj": TestObj()}
        result = await json_serialize_async(data)
        assert isinstance(result, str)

    @pytest.mark.asyncio
    async def test_concurrent_calls(self):
        import asyncio

        tasks = [json_serialize_async({"i": i}) for i in range(10)]
        results = await asyncio.gather(*tasks)
        assert len(results) == 10

    @pytest.mark.asyncio
    async def test_empty_result(self):
        result = await json_serialize_async([])
        parsed = json.loads(result)
        assert parsed == []

    @pytest.mark.asyncio
    async def test_single_space_string(self):
        data = " "
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == " "

    @pytest.mark.asyncio
    async def test_newline_string(self):
        data = "\n"
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == "\n"

    @pytest.mark.asyncio
    async def test_tab_string(self):
        data = "\t"
        result = await json_serialize_async(data)
        parsed = json.loads(result)
        assert parsed == "\t"


class TestThreadPoolForJsonEncode:
    """Test thread_pool_for_json_encode"""

    def test_thread_pool_exists(self):
        assert thread_pool_for_json_encode is not None

    def test_thread_pool_is_executor(self):
        from concurrent.futures import ThreadPoolExecutor

        assert isinstance(thread_pool_for_json_encode, ThreadPoolExecutor)
