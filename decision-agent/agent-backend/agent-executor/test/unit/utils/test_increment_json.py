"""单元测试 - utils/increment_json 模块"""

import pytest

from app.utils.increment_json import (
    compare_values,
    find_differences,
    incremental_async_generator,
    restore_full_json,
)


class TestCompareValues:
    """测试 compare_values 函数"""

    def test_equal_values(self):
        """测试相等的值"""
        result = compare_values("test", "test", 0, [])
        assert result == []

    def test_different_strings(self):
        """测试不同的字符串"""
        result = compare_values("old", "new", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "upsert"
        assert result[0]["content"] == "new"

    def test_string_append(self):
        """测试字符串追加"""
        result = compare_values("hello", "hello world", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "append"
        assert result[0]["content"] == " world"

    def test_dict_new_key(self):
        """测试字典新增键"""
        prev = {"a": 1}
        curr = {"a": 1, "b": 2}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == ["b"]
        assert result[0]["content"] == 2
        assert result[0]["action"] == "upsert"

    def test_dict_removed_key(self):
        """测试字典删除键"""
        prev = {"a": 1, "b": 2}
        curr = {"a": 1}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == ["b"]
        assert result[0]["action"] == "remove"

    def test_dict_changed_value(self):
        """测试字典值变化"""
        prev = {"a": 1}
        curr = {"a": 2}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == ["a"]
        assert result[0]["content"] == 2

    def test_list_append(self):
        """测试列表追加"""
        prev = [1, 2]
        curr = [1, 2, 3]
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == [2]
        assert result[0]["content"] == 3
        assert result[0]["action"] == "append"

    def test_list_remove(self):
        """测试列表删除"""
        prev = [1, 2, 3]
        curr = [1, 2]
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == [2]
        assert result[0]["action"] == "remove"

    def test_list_change(self):
        """测试列表元素变化"""
        prev = [1, 2, 3]
        curr = [1, 5, 3]
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == [1]
        assert result[0]["content"] == 5

    def test_nested_dict(self):
        """测试嵌套字典"""
        prev = {"outer": {"inner": "old"}}
        curr = {"outer": {"inner": "new"}}
        result = compare_values(prev, curr, 0, [])
        assert len(result) == 1
        assert result[0]["key"] == ["outer", "inner"]
        assert result[0]["content"] == "new"

    def test_integer_change(self):
        """测试整数变化"""
        result = compare_values(1, 2, 0, [])
        assert len(result) == 1
        assert result[0]["content"] == 2
        assert result[0]["action"] == "upsert"


class TestFindDifferences:
    """测试 find_differences 函数"""

    def test_no_parent_keys(self):
        """测试无父级键"""
        result = find_differences("old", "new", 0)
        assert len(result) == 1
        assert result[0]["key"] == []

    def test_with_parent_keys(self):
        """测试带父级键"""
        result = find_differences("old", "new", 0, ["parent"])
        assert len(result) == 1
        assert result[0]["key"] == ["parent"]

    def test_equal_values(self):
        """测试相等的值"""
        result = find_differences("same", "same", 0)
        assert result == []

    def test_nested_difference(self):
        """测试嵌套差异"""
        prev = {"a": {"b": 1}}
        curr = {"a": {"b": 2}}
        result = find_differences(prev, curr, 0)
        assert len(result) == 1
        assert result[0]["key"] == ["a", "b"]


class TestIncrementalAsyncGenerator:
    """测试 incremental_async_generator 函数"""

    @pytest.mark.asyncio
    async def test_single_dict(self):
        """测试单个字典"""

        async def gen():
            yield {"a": 1, "b": 2}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        assert len(result) == 3  # 2 keys + 1 end marker
        assert result[0]["key"] == ["a"]
        assert result[0]["content"] == 1
        assert result[1]["key"] == ["b"]
        assert result[1]["content"] == 2
        assert result[2]["action"] == "end"

    @pytest.mark.asyncio
    async def test_dict_changes(self):
        """测试字典变化"""

        async def gen():
            yield {"a": 1}
            yield {"a": 1, "b": 2}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        # First yield: a=1
        assert result[0]["key"] == ["a"]
        assert result[0]["content"] == 1
        # Second yield: new key b=2
        assert result[1]["key"] == ["b"]
        assert result[1]["action"] == "upsert"
        # End marker
        assert result[-1]["action"] == "end"

    @pytest.mark.asyncio
    async def test_non_dict_first_yield(self):
        """测试首次yield非字典"""

        async def gen():
            yield "string value"
            yield {"a": 1}

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        # First yield non-dict: returns the value as-is
        assert result[0]["content"] == "string value"
        assert result[0]["key"] == []
        # Second yield dict: since previous was not a dict,
        # it processes as a new full dict (not incremental)
        # The function treats it as a completely new object
        assert result[1]["content"] == {"a": 1}
        assert result[1]["key"] == []
        # End marker
        assert result[-1]["action"] == "end"

    @pytest.mark.asyncio
    async def test_string_append(self):
        """测试字符串追加"""

        async def gen():
            yield "hello"
            yield "hello world"
            yield "hello world!"

        result = []
        async for item in incremental_async_generator(gen()):
            result.append(item)

        # First yield: full string
        assert result[0]["content"] == "hello"
        # Second yield: append " world"
        assert result[1]["action"] == "append"
        assert result[1]["content"] == " world"
        # Third yield: append "!"
        assert result[2]["action"] == "append"
        assert result[2]["content"] == "!"
        # End marker
        assert result[-1]["action"] == "end"


class TestRestoreFullJson:
    """测试 restore_full_json 函数"""

    @pytest.mark.asyncio
    async def test_simple_restore(self):
        """测试简单恢复"""

        async def gen():
            yield {"seq_id": 0, "key": ["a"], "content": 1, "action": "upsert"}
            yield {"seq_id": 1, "key": ["b"], "content": 2, "action": "upsert"}
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"a": 1, "b": 2}

    @pytest.mark.asyncio
    async def test_nested_restore(self):
        """测试嵌套恢复"""

        async def gen():
            yield {"seq_id": 0, "key": ["outer"], "content": {}, "action": "upsert"}
            yield {
                "seq_id": 1,
                "key": ["outer", "inner"],
                "content": "value",
                "action": "upsert",
            }
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"outer": {"inner": "value"}}

    @pytest.mark.asyncio
    async def test_list_append_restore(self):
        """测试列表追加恢复"""

        async def gen():
            yield {"seq_id": 0, "key": ["text"], "content": "hello", "action": "upsert"}
            yield {
                "seq_id": 1,
                "key": ["text"],
                "content": " world",
                "action": "append",
            }
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"text": "hello world"}

    @pytest.mark.asyncio
    async def test_list_append_array(self):
        """测试数组元素更新恢复"""

        async def gen():
            # Create array with 3 elements
            yield {
                "seq_id": 0,
                "key": ["arr"],
                "content": [1, 2, 3],
                "action": "upsert",
            }
            # Update element at index 1 (change 2 to 5)
            yield {"seq_id": 1, "key": ["arr", 1], "content": 5, "action": "upsert"}
            yield {"seq_id": 2, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"arr": [1, 5, 3]}

    @pytest.mark.asyncio
    async def test_remove_restore(self):
        """测试删除恢复"""

        async def gen():
            yield {"seq_id": 0, "key": ["a"], "content": 1, "action": "upsert"}
            yield {"seq_id": 1, "key": ["b"], "content": 2, "action": "upsert"}
            yield {"seq_id": 2, "key": ["b"], "content": None, "action": "remove"}
            yield {"seq_id": 3, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {"a": 1}
        assert "b" not in result

    @pytest.mark.asyncio
    async def test_empty_restore(self):
        """测试空恢复"""

        async def gen():
            yield {"seq_id": 0, "key": [], "content": None, "action": "end"}

        result = await restore_full_json(gen())
        assert result == {}

    @pytest.mark.asyncio
    async def test_round_trip(self):
        """测试往返转换"""
        original = {"a": 1, "b": {"c": 2}, "d": [3, 4]}

        async def full_gen():
            yield original

        # Convert to incremental
        incremental_gen = incremental_async_generator(full_gen())

        # Collect incremental updates
        updates = []
        async for update in incremental_gen:
            if update["action"] != "end":
                updates.append(update)

        # Restore from incremental
        async def update_gen():
            for update in updates:
                yield update
            yield {"seq_id": 999, "key": [], "content": None, "action": "end"}

        restored = await restore_full_json(update_gen())
        assert restored == original
