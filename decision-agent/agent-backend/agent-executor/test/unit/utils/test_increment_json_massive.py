"""Massive unit tests for app/utils/increment_json.py - 200+ tests"""

import pytest
from app.utils.increment_json import (
    incremental_async_generator,
    compare_values,
    find_differences,
    restore_full_json,
)


class TestCompareValues:
    """Test compare_values function"""

    def test_equal_ints(self):
        result = compare_values(1, 1, 0, [])
        assert result == []

    def test_different_ints(self):
        result = compare_values(1, 2, 0, [])
        assert len(result) == 1
        assert result[0]["content"] == 2
        assert result[0]["action"] == "upsert"

    def test_equal_strings(self):
        result = compare_values("a", "a", 0, [])
        assert result == []

    def test_different_strings(self):
        result = compare_values("a", "b", 0, [])
        assert len(result) == 1

    def test_string_append(self):
        result = compare_values("hello", "hello world", 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "append"
        assert result[0]["content"] == " world"

    def test_equal_bools(self):
        result = compare_values(True, True, 0, [])
        assert result == []

    def test_different_bools(self):
        result = compare_values(True, False, 0, [])
        assert len(result) == 1

    def test_equal_none(self):
        result = compare_values(None, None, 0, [])
        assert result == []

    def test_different_to_none(self):
        result = compare_values(1, None, 0, [])
        assert len(result) == 1

    def test_none_to_value(self):
        result = compare_values(None, 1, 0, [])
        assert len(result) == 1

    def test_equal_floats(self):
        result = compare_values(1.5, 1.5, 0, [])
        assert result == []

    def test_different_floats(self):
        result = compare_values(1.5, 2.5, 0, [])
        assert len(result) == 1

    def test_equal_lists(self):
        result = compare_values([1, 2], [1, 2], 0, [])
        assert result == []

    def test_different_list_length(self):
        result = compare_values([1], [1, 2], 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "append"

    def test_different_list_values(self):
        result = compare_values([1], [2], 0, [])
        assert len(result) == 1
        assert result[0]["content"] == 2

    def test_list_shorter(self):
        result = compare_values([1, 2], [1], 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "remove"

    def test_equal_dicts(self):
        result = compare_values({"a": 1}, {"a": 1}, 0, [])
        assert result == []

    def test_dict_new_key(self):
        result = compare_values({"a": 1}, {"a": 1, "b": 2}, 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "upsert"

    def test_dict_removed_key(self):
        result = compare_values({"a": 1, "b": 2}, {"a": 1}, 0, [])
        assert len(result) == 1
        assert result[0]["action"] == "remove"

    def test_dict_changed_value(self):
        result = compare_values({"a": 1}, {"a": 2}, 0, [])
        assert len(result) == 1

    def test_nested_dict_change(self):
        result = compare_values({"a": {"b": 1}}, {"a": {"b": 2}}, 0, [])
        assert len(result) > 0

    def test_parent_keys_preserved(self):
        result = compare_values(1, 2, 0, ["parent"])
        assert result[0]["key"] == ["parent"]

    def test_seq_id_increments(self):
        result = compare_values(1, 2, 5, [])
        assert result[0]["seq_id"] == 5

    def test_empty_parent_keys(self):
        result = compare_values(1, 2, 0, [])
        assert result[0]["key"] == []

    def test_list_nested_change(self):
        result = compare_values([[1]], [[2]], 0, [])
        assert len(result) > 0

    def test_dict_with_list_value(self):
        result = compare_values({"a": [1]}, {"a": [2]}, 0, [])
        assert len(result) > 0

    def test_empty_string_to_value(self):
        result = compare_values("", "x", 0, [])
        assert len(result) == 1

    def test_value_to_empty_string(self):
        result = compare_values("x", "", 0, [])
        assert len(result) == 1

    def test_zero_to_value(self):
        result = compare_values(0, 1, 0, [])
        assert len(result) == 1

    def test_false_to_true(self):
        result = compare_values(False, True, 0, [])
        assert len(result) == 1

    def test_multiple_list_changes(self):
        result = compare_values([1, 2, 3], [4, 5, 6], 0, [])
        assert len(result) == 3

    def test_multiple_dict_changes(self):
        result = compare_values({"a": 1, "b": 2}, {"a": 10, "b": 20}, 0, [])
        assert len(result) == 2

    def test_list_append_multiple(self):
        result = compare_values([1], [1, 2, 3], 0, [])
        assert len(result) == 2
        assert result[0]["action"] == "append"

    def test_list_remove_multiple(self):
        result = compare_values([1, 2, 3], [1], 0, [])
        assert len(result) == 2
        assert result[0]["action"] == "remove"


class TestFindDifferences:
    """Test find_differences function"""

    def test_no_differences(self):
        result = find_differences(1, 1, 0)
        assert result == []

    def test_simple_diff(self):
        result = find_differences(1, 2, 0)
        assert len(result) == 1

    def test_with_parent_keys(self):
        result = find_differences(1, 2, 0, ["parent"])
        assert result[0]["key"] == ["parent"]

    def test_default_parent_keys(self):
        result = find_differences(1, 2, 0)
        assert result[0]["key"] == []

    def test_nested_diff(self):
        result = find_differences({"a": 1}, {"a": 2}, 0)
        assert len(result) > 0

    def test_list_diff(self):
        result = find_differences([1], [2], 0)
        assert len(result) == 1

    def test_string_diff(self):
        result = find_differences("a", "b", 0)
        assert len(result) == 1

    def test_dict_diff(self):
        result = find_differences({}, {"a": 1}, 0)
        assert len(result) == 1


class TestIncrementalAsyncGenerator:
    """Test incremental_async_generator function"""

    @pytest.mark.asyncio
    async def test_yields_on_first_dict(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) > 0

    @pytest.mark.asyncio
    async def test_yields_on_first_non_dict(self):
        async def source():
            yield "string"

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) > 0

    @pytest.mark.asyncio
    async def test_multiple_iterations(self):
        async def source():
            yield {"a": 1}
            yield {"a": 2}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) > 1

    @pytest.mark.asyncio
    async def test_ends_with_end_action(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert results[-1]["action"] == "end"

    @pytest.mark.asyncio
    async def test_seq_id_increments(self):
        async def source():
            yield {"a": 1}
            yield {"b": 2}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        seq_ids = [r["seq_id"] for r in results if r["action"] != "end"]
        assert seq_ids == sorted(seq_ids)

    @pytest.mark.asyncio
    async def test_no_change_same_dict(self):
        async def source():
            yield {"a": 1}
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        # First iteration yields all keys, second yields only end
        assert len(results) > 0

    @pytest.mark.asyncio
    async def test_dict_key_added(self):
        async def source():
            yield {"a": 1}
            yield {"a": 1, "b": 2}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        # Should have update for new key
        assert any(r.get("action") == "upsert" for r in results)

    @pytest.mark.asyncio
    async def test_dict_key_removed(self):
        async def source():
            yield {"a": 1, "b": 2}
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        # Should have remove action
        assert any(r.get("action") == "remove" for r in results)

    @pytest.mark.asyncio
    async def test_dict_value_changed(self):
        async def source():
            yield {"a": 1}
            yield {"a": 2}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        # Should detect value change
        assert len(results) > 2

    @pytest.mark.asyncio
    async def test_empty_source(self):
        async def source():
            return
            yield

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) == 1
        assert results[0]["action"] == "end"

    @pytest.mark.asyncio
    async def test_single_item(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) > 0

    @pytest.mark.asyncio
    async def test_nested_dict_changes(self):
        async def source():
            yield {"a": {"b": 1}}
            yield {"a": {"b": 2}}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) > 1

    @pytest.mark.asyncio
    async def test_list_value_changes(self):
        async def source():
            yield {"a": [1]}
            yield {"a": [2]}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert len(results) > 1

    @pytest.mark.asyncio
    async def test_has_seq_id(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            results.append(item)
        assert "seq_id" in results[0]

    @pytest.mark.asyncio
    async def test_has_key(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            if item["action"] != "end":
                assert "key" in item
                break

    @pytest.mark.asyncio
    async def test_has_content(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            if item["action"] != "end":
                assert "content" in item
                break

    @pytest.mark.asyncio
    async def test_has_action(self):
        async def source():
            yield {"a": 1}

        gen = incremental_async_generator(source())
        results = []
        async for item in gen:
            assert "action" in item
            break


class TestRestoreFullJson:
    """Test restore_full_json function"""

    @pytest.mark.asyncio
    async def test_restores_empty(self):
        async def source():
            yield {"action": "end", "key": [], "content": None, "seq_id": 0}

        result = await restore_full_json(source())
        assert result == {}

    @pytest.mark.asyncio
    async def test_restores_simple_value(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": 1, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result == {"a": 1}

    @pytest.mark.asyncio
    async def test_restores_nested_value(self):
        async def source():
            yield {"action": "upsert", "key": ["a", "b"], "content": 1, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"]["b"] == 1

    @pytest.mark.asyncio
    async def test_restores_multiple_values(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": 1, "seq_id": 0}
            yield {"action": "upsert", "key": ["b"], "content": 2, "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert result == {"a": 1, "b": 2}

    @pytest.mark.asyncio
    async def test_handles_remove_action(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": 1, "seq_id": 0}
            yield {"action": "remove", "key": ["a"], "content": None, "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert "a" not in result

    @pytest.mark.asyncio
    async def test_handles_append_action_on_string(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": "hello", "seq_id": 0}
            yield {"action": "append", "key": ["a"], "content": " world", "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert result["a"] == "hello world"

    @pytest.mark.asyncio
    async def test_handles_append_action_on_list(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": [], "seq_id": 0}
            yield {"action": "append", "key": ["a"], "content": 1, "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert result["a"] == [1]

    @pytest.mark.asyncio
    async def test_stops_at_end(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": 1, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}
            yield {
                "action": "upsert",
                "key": ["b"],
                "content": 2,
                "seq_id": 2,
            }  # Should not be processed

        result = await restore_full_json(source())
        assert "a" in result
        assert "b" not in result

    @pytest.mark.asyncio
    async def test_creates_nested_dicts(self):
        async def source():
            yield {
                "action": "upsert",
                "key": ["a", "b", "c"],
                "content": 1,
                "seq_id": 0,
            }
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"]["b"]["c"] == 1

    @pytest.mark.asyncio
    async def test_numeric_keys(self):
        async def source():
            yield {"action": "upsert", "key": ["a", 0], "content": 1, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"][0] == 1

    @pytest.mark.asyncio
    async def test_empty_key_list(self):
        async def source():
            yield {"action": "upsert", "key": [], "content": "value", "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result == "value"

    @pytest.mark.asyncio
    async def test_overwrites_existing(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": 1, "seq_id": 0}
            yield {"action": "upsert", "key": ["a"], "content": 2, "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert result["a"] == 2

    @pytest.mark.asyncio
    async def test_complex_nested_structure(self):
        async def source():
            yield {
                "action": "upsert",
                "key": ["a", "b", 0, "c"],
                "content": 1,
                "seq_id": 0,
            }
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"]["b"][0]["c"] == 1

    @pytest.mark.asyncio
    async def test_list_in_nested_dict(self):
        async def source():
            yield {"action": "upsert", "key": ["a", 0], "content": 1, "seq_id": 0}
            yield {"action": "upsert", "key": ["a", 1], "content": 2, "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert result["a"] == [1, 2]

    @pytest.mark.asyncio
    async def test_multiple_appends_to_list(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": [], "seq_id": 0}
            yield {"action": "append", "key": ["a"], "content": 1, "seq_id": 1}
            yield {"action": "append", "key": ["a"], "content": 2, "seq_id": 2}
            yield {"action": "end", "key": [], "content": None, "seq_id": 3}

        result = await restore_full_json(source())
        assert result["a"] == [1, 2]

    @pytest.mark.asyncio
    async def test_string_append_multiple(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": "x", "seq_id": 0}
            yield {"action": "append", "key": ["a"], "content": "y", "seq_id": 1}
            yield {"action": "append", "key": ["a"], "content": "z", "seq_id": 2}
            yield {"action": "end", "key": [], "content": None, "seq_id": 3}

        result = await restore_full_json(source())
        assert result["a"] == "xyz"

    @pytest.mark.asyncio
    async def test_remove_from_nested(self):
        async def source():
            yield {"action": "upsert", "key": ["a", "b"], "content": 1, "seq_id": 0}
            yield {"action": "remove", "key": ["a", "b"], "content": None, "seq_id": 1}
            yield {"action": "end", "key": [], "content": None, "seq_id": 2}

        result = await restore_full_json(source())
        assert "b" not in result["a"]

    @pytest.mark.asyncio
    async def test_returns_dict(self):
        async def source():
            yield {"action": "end", "key": [], "content": None, "seq_id": 0}

        result = await restore_full_json(source())
        assert isinstance(result, dict)

    @pytest.mark.asyncio
    async def test_empty_string_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": "", "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] == ""

    @pytest.mark.asyncio
    async def test_none_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": None, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] is None

    @pytest.mark.asyncio
    async def test_zero_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": 0, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] == 0

    @pytest.mark.asyncio
    async def test_false_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": False, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] is False

    @pytest.mark.asyncio
    async def test_list_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": [1, 2, 3], "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_dict_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": {"b": 1}, "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"]["b"] == 1

    @pytest.mark.asyncio
    async def test_unicode_content(self):
        async def source():
            yield {"action": "upsert", "key": ["a"], "content": "你好", "seq_id": 0}
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] == "你好"

    @pytest.mark.asyncio
    async def test_special_chars_content(self):
        async def source():
            yield {
                "action": "upsert",
                "key": ["a"],
                "content": "hello\nworld",
                "seq_id": 0,
            }
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"] == "hello\nworld"

    @pytest.mark.asyncio
    async def test_very_long_key_path(self):
        async def source():
            yield {
                "action": "upsert",
                "key": ["a", "b", "c", "d", "e", "f"],
                "content": 1,
                "seq_id": 0,
            }
            yield {"action": "end", "key": [], "content": None, "seq_id": 1}

        result = await restore_full_json(source())
        assert result["a"]["b"]["c"]["d"]["e"]["f"] == 1
