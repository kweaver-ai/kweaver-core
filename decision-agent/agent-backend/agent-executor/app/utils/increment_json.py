import asyncio
import json
from typing import Any, AsyncGenerator, Dict, List


async def incremental_async_generator(
    full_json_gen: AsyncGenerator[Dict[str, Any], None],
) -> AsyncGenerator[Dict[str, Any], None]:
    """处理JSON增量更新的异步生成器

    Args:
        full_json_gen: 完整的JSON生成器，每次yield一个完整的JSON对象

    Yields:
        Dict[str, Any]: 包含seq_id, key, content, action的增量更新字典
    """
    previous_json = None
    seq_id = 0

    async for current_json in full_json_gen:
        if previous_json is None:
            if not isinstance(current_json, dict):
                yield {
                    "seq_id": seq_id,
                    "key": [],
                    "content": current_json,
                    "action": "upsert",
                }
                seq_id += 1
            else:
                # 第一次处理，直接返回完整JSON
                for key, value in current_json.items():
                    yield {
                        "seq_id": seq_id,
                        "key": [key],
                        "content": value,
                        "action": "upsert",
                    }
                    seq_id += 1
            # 使用序列化/反序列化代替深拷贝
            previous_json = json.loads(json.dumps(current_json))
            continue

        # 比较当前JSON与之前JSON的差异
        changes = find_differences(previous_json, current_json, seq_id)
        for change in changes:
            yield change
            seq_id += 1

        # 更新previous_json
        previous_json = json.loads(json.dumps(current_json))

    # 标记结束
    yield {"seq_id": seq_id, "key": [], "content": None, "action": "end"}


def compare_values(
    prev: Any,
    curr: Any,
    seq_id: int,
    parent_keys: List[str],
) -> List[Dict[str, Any]]:
    """比较两个值并返回差异列表

    Args:
        prev: 前一个值
        curr: 当前值
        seq_id: 序列ID
        parent_keys: 父级键列表

    Returns:
        List[Dict[str, Any]]: 差异列表，每个差异包含seq_id, key, content, action
    """
    differences = []

    # 如果值相等，无需处理
    if curr == prev:
        return differences

    # 处理字典类型
    if isinstance(curr, dict) and isinstance(prev, dict):
        prev_keys = set(prev.keys())
        curr_keys = set(curr.keys())

        # 处理新增的键
        for key in curr_keys - prev_keys:
            differences.append(
                {
                    "seq_id": seq_id,
                    "key": parent_keys + [key],
                    "content": curr[key],
                    "action": "upsert",
                }
            )
            seq_id += 1

        # 处理删除的键
        for key in prev_keys - curr_keys:
            differences.append(
                {
                    "seq_id": seq_id,
                    "key": parent_keys + [key],
                    "content": None,
                    "action": "remove",
                }
            )
            seq_id += 1

        # 处理共有的键
        for key in prev_keys & curr_keys:
            if curr[key] == prev[key]:
                continue
            # 递归比较值
            differences.extend(
                compare_values(prev[key], curr[key], seq_id, parent_keys + [key])
            )

    # 处理列表类型
    elif isinstance(curr, list) and isinstance(prev, list):
        curr_len = len(curr)
        prev_len = len(prev)
        min_len = min(curr_len, prev_len)

        # 处理交集部分的变化
        for i in range(min_len):
            if curr[i] != prev[i]:
                differences.append(
                    {
                        "seq_id": seq_id,
                        "key": parent_keys + [i],
                        "content": curr[i],
                        "action": "upsert",
                    }
                )
                seq_id += 1

        # 处理新增的元素
        if curr_len > prev_len:
            for i in range(prev_len, curr_len):
                differences.append(
                    {
                        "seq_id": seq_id,
                        "key": parent_keys + [i],
                        "content": curr[i],
                        "action": "append",
                    }
                )
                seq_id += 1
        # 处理删除的元素
        elif curr_len < prev_len:
            for i in range(curr_len, prev_len):
                differences.append(
                    {
                        "seq_id": seq_id,
                        "key": parent_keys + [i],
                        "content": None,
                        "action": "remove",
                    }
                )
                seq_id += 1

    # 处理字符串类型
    elif isinstance(curr, str) and isinstance(prev, str):
        if curr.startswith(prev):
            differences.append(
                {
                    "seq_id": seq_id,
                    "key": parent_keys,
                    "content": curr[len(prev) :],
                    "action": "append",
                }
            )
            seq_id += 1
        else:
            differences.append(
                {
                    "seq_id": seq_id,
                    "key": parent_keys,
                    "content": curr,
                    "action": "upsert",
                }
            )
            seq_id += 1

    # 处理其他类型
    else:
        differences.append(
            {
                "seq_id": seq_id,
                "key": parent_keys,
                "content": curr,
                "action": "upsert",
            }
        )
        seq_id += 1

    return differences


def find_differences(
    prev: Any,
    curr: Any,
    seq_id: int,
    parent_keys: List[str] = None,
) -> List[Dict[str, Any]]:
    """查找两个对象之间的差异，支持任何类型对象的比较

    Args:
        prev: 前一个对象
        curr: 当前对象
        seq_id: 序列ID
        parent_keys: 父级键列表

    Returns:
        List[Dict[str, Any]]: 差异列表，每个差异包含seq_id, key, content, action
    """
    if parent_keys is None:
        parent_keys = []

    # 直接使用compare_values函数处理任意类型的比较
    return compare_values(prev, curr, seq_id, parent_keys)


async def restore_full_json(
    incremental_gen: AsyncGenerator[Dict[str, Any], None],
) -> Dict[str, Any]:
    """从增量更新恢复完整的JSON对象

    Args:
        incremental_gen: 增量更新生成器，每个更新包含seq_id, key, content, action

    Returns:
        Dict[str, Any]: 恢复后的完整JSON对象

    Example:
        ```python
        incremental_gen = incremental_async_generator(full_json_gen)
        full_json = await restore_full_json(incremental_gen)
        ```
    """
    full_json = {}

    async for update in incremental_gen:
        if update["action"] == "end":
            break

        keys = update["key"]
        content = update["content"]
        action = update["action"]

        # Handle empty key list - return content directly
        if not keys:
            full_json = content
            continue

        # 遍历到目标键
        current_level = full_json
        for i, key in enumerate(keys[:-1]):
            next_key = keys[i + 1]

            if isinstance(key, int):
                # Current level is a list, key is an index
                # Ensure the list is large enough
                while len(current_level) <= key:
                    current_level.append({})
                # Check if we need to convert dict to list at this index
                if isinstance(next_key, int) and not isinstance(
                    current_level[key], list
                ):
                    current_level[key] = []
                elif not isinstance(next_key, int) and isinstance(
                    current_level[key], list
                ):
                    # Convert list to dict if needed
                    current_level[key] = {
                        i: v for i, v in enumerate(current_level[key])
                    }
                current_level = current_level[key]
            else:
                # Current level is a dict, key is a string key
                if key not in current_level:
                    # Determine if the next level should be a list or dict
                    if isinstance(next_key, int):
                        # Create a list if the next key is an integer
                        current_level[key] = []
                    else:
                        current_level[key] = {}
                elif isinstance(next_key, int) and not isinstance(
                    current_level[key], list
                ):
                    # Convert dict to list if needed
                    current_level[key] = []
                elif not isinstance(next_key, int) and isinstance(
                    current_level[key], list
                ):
                    # Convert list to dict if needed
                    current_level[key] = {
                        i: v for i, v in enumerate(current_level[key])
                    }
                current_level = current_level[key]

        last_key = keys[-1]

        if action == "upsert":
            if isinstance(last_key, int) and isinstance(current_level, list):
                # Ensure the list is large enough for the index
                while len(current_level) <= last_key:
                    current_level.append(None)
            current_level[last_key] = content
        elif action == "append":
            if isinstance(last_key, int) and isinstance(current_level, list):
                # Ensure the list is large enough for the index
                while len(current_level) <= last_key:
                    current_level.append(None)
            if isinstance(current_level[last_key], str):
                current_level[last_key] += content
            else:
                current_level[last_key].append(content)
        elif action == "remove":
            if isinstance(last_key, int) and isinstance(current_level, list):
                # Ensure the list is large enough for the index
                while len(current_level) <= last_key:
                    current_level.append(None)
            if last_key in current_level:
                del current_level[last_key]

    return full_json


if __name__ == "__main__":
    # Example Usage
    async def main():
        origin_json_list = ["a", "ab", "", None]

        # origin_json_list = []
        # example_dir = os.path.join(os.path.dirname(__file__), "真实输出样例")
        # example_file = os.path.join(example_dir, "doc_qa.txt")
        # with open(example_file, "r", encoding="utf-8") as f:
        #     while line := f.readline():
        #         if not line.startswith("data: "):
        #             continue
        #         line = line.split("data: ")[1]
        #         origin_json_list.append(json.loads(line))
        async def full_json_async_generator():
            for json_obj in origin_json_list:
                yield json_obj

        full_json_gen = full_json_async_generator()
        incremental_gen = incremental_async_generator(full_json_gen)

        increment_json_list = []
        async for inc in incremental_gen:
            print(json.dumps(inc, indent=2))
            increment_json_list.append(inc)
            pass

        restored_json = await restore_full_json(
            incremental_async_generator(full_json_async_generator())
        )
        print("Restored JSON:", restored_json)
        assert restored_json == origin_json_list[-1]

    asyncio.run(main())
