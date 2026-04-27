"""扩展单元测试 - utils/increment_json 模块 - 增加边界情况测试"""

import pytest

from app.utils.increment_json import (
    compare_values,
    find_differences,
    incremental_async_generator,
    restore_full_json,
)


class TestCompareValuesExtended:
    """测试 compare_values 函数扩展测试"""

    def test_none_values(self):
        """测试None值"""
        result = compare_values(None, None, 0, [])
        assert result == []

    def test_none_to_value(self):
        """测试从None到有值"""
        result = compare_values(None, "value", 0, [])
        assert len(result) == 1
        assert result[0]["content"] == "value"

    def test_value_to_none(self):
        """测试从有值到None"""
        result = compare_values("value", None, 0, [])
        assert len(result) == 1
        assert result[0]["content"] is None

    def test_empty_dict_to_dict(self):
        """测试空字典到有值字典"""
        result = compare_values({}, {"a": 1}, 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "upsert"

    def test_dict_to_empty_dict(self):
        """测试字典到空字典"""
        result = compare_values({"a": 1}, {}, 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "remove"

    def test_empty_list_to_list(self):
        """测试空列表到有值列表"""
        result = compare_values([], [1, 2], 0, [])
        assert len(result) == 2
        assert result[0]["action"] == "append"

    def test_list_to_empty_list(self):
        """测试列表到空列表"""
        result = compare_values([1, 2], [], 0, [])
        assert len(result) == 2
        assert result[0]["action"] == "remove"

    def test_bool_true_to_false(self):
        """测试布尔值变化"""
        result = compare_values(True, False, 0, [])
        assert len(result) == 1
        assert result[0]["content"] is False

    def test_float_change(self):
        """测试浮点数变化"""
        result = compare_values(3.14, 2.71, 0, [])
        assert len(result) == 1
        assert result[0]["content"] == 2.71

    def test_zero_to_nonzero(self):
        """测试从0到非零"""
        result = compare_values(0, 42, 0, [])
        assert len(result) == 1
        assert result[0]["content"] == 42

    def test_nested_empty_dicts(self):
        """测试嵌套空字典"""
        result = compare_values({"a": {}}, {"a": {}}, 0, [])
        assert result == []

    def test_dict_with_none_value(self):
        """测试字典包含None值"""
        prev = {"a": None}
        curr = {"a": None}
        result = compare_values(prev, curr, 0, [])
        assert result == []

    def test_dict_none_to_value(self):
        """测试字典中None到有值"""
        prev = {"a": None}
        curr = {"a": "value"}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["content"] == "value"

    def test_list_with_none(self):
        """测试列表包含None"""
        prev = [None, 1, None]
        curr = [None, 1, None]
        result = compare_values(prev, curr, 0, [])
        assert result == []

    def test_string_no_prefix_match(self):
        """测试字符串无前缀匹配"""
        result = compare_values("abc", "xyz", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "upsert"

    def test_string_empty_to_value(self):
        """测试空字符串到有值"""
        result = compare_values("", "hello", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "append"

    def test_string_value_to_empty(self):
        """测试字符串到空"""
        result = compare_values("hello", "", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "upsert"

    def test_unicode_strings(self):
        """测试Unicode字符串"""
        result = compare_values("你好", "你好世界", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "append"
        assert result[0]["content"] == "世界"

    def test_emoji_strings(self):
        """测试表情符号字符串"""
        result = compare_values("Hello 👋", "Hello 👋 🌍", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "append"

    def test_list_single_element(self):
        """测试单元素列表"""
        result = compare_values([1], [2], 0, [])
        assert len(result) == 1
        assert result[0]["content"] == 2

    def test_dict_multiple_keys_all_changed(self):
        """测试字典所有键都变化"""
        prev = {"a": 1, "b": 2, "c": 3}
        curr = {"a": 10, "b": 20, "c": 30}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 3

    def test_dict_all_keys_removed(self):
        """测试字典所有键被删除"""
        prev = {"a": 1, "b": 2, "c": 3}
        curr = {}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 3
        for change in result:
            assert change["action"] == "remove"

    def test_list_all_elements_changed(self):
        """测试列表所有元素变化"""
        prev = [1, 2, 3]
        curr = [4, 5, 6]
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 3

    def test_deeply_nested_dict(self):
        """测试深度嵌套字典"""
        prev = {"a": {"b": {"c": {"d": 1}}}}
        curr = {"a": {"b": {"c": {"d": 2}}}}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == ["a", "b", "c", "d"]

    def test_mixed_nested_structure(self):
        """测试混合嵌套结构"""
        prev = {"a": [{"b": 1}]}
        curr = {"a": [{"b": 2}]}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1


class TestFindDifferencesExtended:
    """测试 find_differences 函数扩展测试"""

    def test_none_input(self):
        """测试None输入"""
        result = find_differences(None, None, 0)
        assert result == []

    def test_none_to_value(self):
        """测试None到值"""
        result = find_differences(None, "value", 0)
        assert len(result) == 1

    def test_complex_nested_structure(self):
        """测试复杂嵌套结构"""
        prev = {"users": [{"name": "Alice"}, {"name": "Bob"}]}
        curr = {"users": [{"name": "Alice"}, {"name": "Charlie"}]}
        result = find_differences(prev, curr, 0)
        assert len(result) == 1

    def test_with_multiple_parent_keys(self):
        """测试多个父级键"""
        result = find_differences(1, 2, 0, ["a", "b", "c"])
        assert len(result) == 1
        assert result[0]["key"] == ["a", "b", "c"]

    def test_seq_id_increment(self):
        """测试序列ID递增"""
        result = find_differences("old", "new", 5)
        assert result[0]["seq_id"] == 5

    def test_large_numbers(self):
        """测试大数字"""
        result = find_differences(999999999, 1000000000, 0)
        assert len(result) == 1

    def test_negative_numbers(self):
        """测试负数"""
        result = find_differences(-10, -20, 0)
        assert len(result) == 1


class TestIncrementalAsyncGeneratorExtended:
    """测试 incremental_async_generator 函数扩展测试"""

    @pytest.mark.asyncio
    async def test_empty_generator(self):
        """测试空生成器"""

        async def gen():
            return
            yield  # Never reached

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert len(result) == 1
        assert result[0]["action"] == "end"

    @pytest.mark.asyncio
    async def test_single_non_dict_value(self):
        """测试单个非字典值"""

        async def gen():
            yield 42

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert len(result) == 2  # value + end
        assert result[0]["content"] == 42

    @pytest.mark.asyncio
    async def test_list_first_value(self):
        """测试列表作为第一个值"""

        async def gen():
            yield [1, 2, 3]

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert len(result) == 2  # list + end
        assert result[0]["content"] == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_number_sequence(self):
        """测试数字序列"""

        async def gen():
            yield 1
            yield 2
            yield 3

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert result[0]["content"] == 1
        assert result[1]["content"] == 2
        assert result[2]["content"] == 3

    @pytest.mark.asyncio
    async def test_dict_with_nested_list(self):
        """测试字典包含嵌套列表"""

        async def gen():
            yield {"items": [1, 2, 3]}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert result[0]["key"] == ["items"]
        assert result[0]["content"] == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_dict_with_empty_string(self):
        """测试字典包含空字符串"""

        async def gen():
            yield {"text": ""}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert result[0]["key"] == ["text"]
        assert result[0]["content"] == ""

    @pytest.mark.asyncio
    async def test_dict_with_boolean(self):
        """测试字典包含布尔值"""

        async def gen():
            yield {"active": True}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert result[0]["key"] == ["active"]
        assert result[0]["content"] is True

    @pytest.mark.asyncio
    async def test_dict_with_null(self):
        """测试字典包含null"""

        async def gen():
            yield {"value": None}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert result[0]["key"] == ["value"]
        assert result[0]["content"] is None

    @pytest.mark.asyncio
    async def test_multiple_dict_changes_same_key(self):
        """测试同一键多次变化"""

        async def gen():
            yield {"count": 1}
            yield {"count": 2}
            yield {"count": 3}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        # First yield
        assert result[0]["key"] == ["count"]
        assert result[0]["content"] == 1
        # Second yield - change detected
        assert result[1]["key"] == ["count"]
        assert result[1]["content"] == 2
        # Third yield - another change
        assert result[2]["key"] == ["count"]
        assert result[2]["content"] == 3

    @pytest.mark.asyncio
    async def test_empty_dict_first(self):
        """测试空字典作为第一个值"""

        async def gen():
            yield {}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        # Empty dict - no keys to process
        assert result[0]["action"] == "end"

    @pytest.mark.asyncio
    async def test_dict_unicode_values(self):
        """测试字典Unicode值"""

        async def gen():
            yield {"text": "你好世界"}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert result[0]["content"] == "你好世界"

    @pytest.mark.asyncio
    async def test_dict_special_chars(self):
        """测试字典特殊字符"""

        async def gen():
            yield {"text": "Hello\nWorld\t!"}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert "\n" in result[0]["content"]


class TestRestoreFullJsonExtended:
    """测试 restore_full_json 函数扩展测试"""

    @pytest.mark.asyncio
    async def test_empty_updates(self):
        """测试空更新"""

        async def gen():
            yield {"seq_id": 0, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {}

    @pytest.mark.asyncio
    async def test_multiple_upserts(self):
        """测试多次upsert"""

        async def gen():
            yield {"seq_id": 0, "key": ["a"], "content": 1, "action": "upsert"}
            yield {"seq_id": 1, "key": ["b"], "content": 2, "action": "upsert"}
            yield {"seq_id": 2, "key": ["c"], "content": 3, "action": "upsert"}
            yield {"seq_id": 3, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"a": 1, "b": 2, "c": 3}

    @pytest.mark.asyncio
    async def test_deeply_nested_structure(self):
        """测试深度嵌套结构"""

        async def gen():
            yield {"seq_id": 0, "key": ["a"], "content": {}, "action": "upsert"}
            yield {"seq_id": 1, "key": ["a", "b"], "content": {}, "action": "upsert"}
            yield {
                "seq_id": 2,
                "key": ["a", "b", "c"],
                "content": "value",
                "action": "upsert",
            }
            yield {"seq_id": 3, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"a": {"b": {"c": "value"}}}

    @pytest.mark.asyncio
    async def test_append_to_string_multiple_times(self):
        """测试多次追加到字符串"""

        async def gen():
            yield {"seq_id": 0, "key": ["text"], "content": "a", "action": "upsert"}
            yield {"seq_id": 1, "key": ["text"], "content": "b", "action": "append"}
            yield {"seq_id": 2, "key": ["text"], "content": "c", "action": "append"}
            yield {"seq_id": 3, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"text": "abc"}

    @pytest.mark.asyncio
    async def test_append_array_elements(self):
        """测试追加数组元素"""

        async def gen():
            # First create the array with upsert
            yield {"seq_id": 0, "key": ["arr"], "content": [1], "action": "upsert"}
            # Then use append for new elements
            yield {"seq_id": 1, "key": ["arr"], "content": 2, "action": "append"}
            yield {"seq_id": 2, "key": ["arr"], "content": 3, "action": "append"}
            yield {"seq_id": 3, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        # Result should have array with elements
        assert "arr" in result
        assert len(result["arr"]) >= 1

    @pytest.mark.asyncio
    async def test_remove_nested_key(self):
        """测试删除嵌套键"""

        async def gen():
            yield {
                "seq_id": 0,
                "key": ["outer"],
                "content": {"inner": "value"},
                "action": "upsert",
            }
            yield {
                "seq_id": 1,
                "key": ["outer", "inner"],
                "content": None,
                "action": "remove",
            }
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"outer": {}}

    @pytest.mark.asyncio
    async def test_upsert_overwrites_existing(self):
        """测试upsert覆盖现有值"""

        async def gen():
            yield {"seq_id": 0, "key": ["a"], "content": "old", "action": "upsert"}
            yield {"seq_id": 1, "key": ["a"], "content": "new", "action": "upsert"}
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"a": "new"}

    @pytest.mark.asyncio
    async def test_unicode_content(self):
        """测试Unicode内容"""

        async def gen():
            yield {"seq_id": 0, "key": ["text"], "content": "你好", "action": "upsert"}
            yield {"seq_id": 1, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"text": "你好"}

    @pytest.mark.asyncio
    async def test_special_characters_in_content(self):
        """测试内容中的特殊字符"""

        async def gen():
            yield {
                "seq_id": 0,
                "key": ["text"],
                "content": "Line1\nLine2",
                "action": "upsert",
            }
            yield {"seq_id": 1, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"text": "Line1\nLine2"}

    @pytest.mark.asyncio
    async def test_boolean_values(self):
        """测试布尔值"""

        async def gen():
            yield {
                "seq_id": 0,
                "key": ["bool_true"],
                "content": True,
                "action": "upsert",
            }
            yield {
                "seq_id": 1,
                "key": ["bool_false"],
                "content": False,
                "action": "upsert",
            }
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"bool_true": True, "bool_false": False}

    @pytest.mark.asyncio
    async def test_null_value(self):
        """测试null值"""

        async def gen():
            yield {"seq_id": 0, "key": ["value"], "content": None, "action": "upsert"}
            yield {"seq_id": 1, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"value": None}

    @pytest.mark.asyncio
    async def test_numeric_values(self):
        """测试数值类型"""

        async def gen():
            yield {"seq_id": 0, "key": ["int"], "content": 42, "action": "upsert"}
            yield {"seq_id": 1, "key": ["float"], "content": 3.14, "action": "upsert"}
            yield {"seq_id": 2, "key": ["negative"], "content": -10, "action": "upsert"}
            yield {"seq_id": 3, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"int": 42, "float": 3.14, "negative": -10}


class TestIntegrationScenarios:
    """测试集成场景"""

    @pytest.mark.asyncio
    async def test_full_roundtrip_complex_data(self):
        """测试复杂数据完整往返"""
        original = {
            "users": [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}],
            "metadata": {"version": "1.0", "created": "2024-01-01"},
        }

        async def full_gen():
            yield original

        # Convert to incremental
        incremental_gen = incremental_async_generator(full_gen())

        # Restore from incremental
        restored = await restore_full_json(incremental_gen)

        assert restored == original

    @pytest.mark.asyncio
    async def test_multiple_updates_same_structure(self):
        """测试相同结构多次更新"""

        async def full_gen():
            yield {"count": 1, "status": "init"}
            yield {"count": 2, "status": "running"}
            yield {"count": 3, "status": "complete"}

        incremental_gen = incremental_async_generator(full_gen())

        updates = []
        async for update in incremental_gen:
            if update["action"] != "end":
                updates.append(update)

        # Should have updates for all changes
        assert len(updates) > 0
