"""
字典路径解析器 - 非扁平化版本
保留原始数据结构和层级关系
"""

from typing import Any, List, Union
import copy


class DictPathParser:
    """
    字典路径解析器类 - 非扁平化版本

    关键区别：
    - a[*].b      : 返回每个a中b字段的值，保持数组结构 [[b1, b2], [b3]]
    - a[*].b[*]   : 在上面基础上展开最后一层 [b1, b2, b3]

    支持路径形式:
    - 简单路径: 'a.b.c'
    - 数组路径: 'a.b[*].c'  (保持结构)
    - 数组展开: 'a.b[*].c[*]'  (展开最后一层)
    - 数组索引: 'a.b[0].c'
    """

    def __init__(self, data: Union[dict, list] = None):
        """
        初始化解析器

        :param data: 要解析的数据结构
        """
        self.data = data if data is not None else {}

    def get(self, path: str, flatten_final: bool = False) -> Any:
        """
        根据路径获取数据

        :param path: 路径字符串，如 'a.b[*].c'
        :param flatten_final: 是否扁平化最终结果
        :return: 解析结果
        """
        if not path:
            return self.data

        keys = self._parse_path(path)
        return self._get_recursive(self.data, keys, flatten_final)

    def get_flat(self, path: str) -> Any:
        """
        获取扁平化结果（等价于原版本的行为）

        :param path: 路径字符串
        :return: 扁平化的解析结果
        """
        return self.get(path, flatten_final=True)

    def set(self, path: str, value: Any) -> None:
        """
        根据路径设置数据

        :param path: 路径字符串
        :param value: 要设置的值
        """
        if not path:
            self.data = value
            return

        keys = self._parse_path(path)
        self._set_recursive(self.data, keys, value)

    def has(self, path: str) -> bool:
        """
        检查路径是否存在

        :param path: 路径字符串
        :return: 是否存在
        """
        try:
            self.get(path)
            return True
        except (ValueError, KeyError, IndexError, TypeError):
            return False

    def delete(self, path: str) -> bool:
        """
        删除指定路径的数据

        :param path: 路径字符串
        :return: 是否成功删除
        """
        if not path:
            return False

        keys = self._parse_path(path)
        return self._delete_recursive(self.data, keys)

    def get_all_paths(self, prefix: str = "") -> List[str]:
        """
        获取所有可能的路径

        :param prefix: 路径前缀
        :return: 路径列表
        """
        paths = []
        self._collect_paths(self.data, prefix, paths)
        return paths

    def _parse_path(self, path: str) -> List[Union[str, int, None]]:
        """
        解析路径字符串为键列表

        :param path: 路径字符串
        :return: 键列表，None表示数组遍历，'[*]'表示最终展开
        """
        if not path:
            return []

        keys = []
        current_part = ""
        i = 0

        while i < len(path):
            if path[i] == ".":
                if current_part:
                    keys.append(current_part)
                    current_part = ""
            elif path[i] == "[":
                # 找到匹配的右括号
                if current_part:
                    keys.append(current_part)
                    current_part = ""

                bracket_start = i
                bracket_count = 1
                i += 1

                while i < len(path) and bracket_count > 0:
                    if path[i] == "[":
                        bracket_count += 1
                    elif path[i] == "]":
                        bracket_count -= 1
                    i += 1

                if bracket_count > 0:
                    raise ValueError(f"不匹配的方括号: {path}")

                bracket_content = path[bracket_start:i]

                if bracket_content == "[*]":
                    keys.append(None)  # None表示遍历数组（保持结构）
                elif bracket_content.startswith("[") and bracket_content.endswith("]"):
                    # 处理数组索引，如 [0], [1] 等
                    index_str = bracket_content[1:-1]
                    if index_str.isdigit():
                        keys.append(int(index_str))
                    else:
                        raise ValueError(f"无效的数组索引: {bracket_content}")
                else:
                    raise ValueError(f"无效的方括号表达式: {bracket_content}")

                i -= 1  # 回退一位，因为外层循环会增加
            else:
                current_part += path[i]

            i += 1

        # 处理最后一部分
        if current_part:
            keys.append(current_part)

        return keys

    def _get_recursive(
        self,
        current: Any,
        keys: List[Union[str, int, None]],
        flatten_final: bool = False,
    ) -> Any:
        """
        递归获取数据 - 非扁平化版本

        :param current: 当前数据
        :param keys: 剩余键列表
        :param flatten_final: 是否在最终结果中扁平化
        :return: 解析结果
        """
        if not keys:
            return current

        key = keys[0]
        remaining_keys = keys[1:]

        if key is None:  # [*] 遍历数组
            if not isinstance(current, list):
                raise ValueError(f"尝试遍历数组但当前数据不是列表: {type(current)}")

            results = []
            for item in current:
                try:
                    value = self._get_recursive(item, remaining_keys, flatten_final)
                    results.append(value)
                except (ValueError, KeyError, IndexError, TypeError):
                    # 忽略无法访问的项
                    continue

            # 扁平化逻辑：只有当 flatten_final=True 时才递归扁平化
            if flatten_final:
                return self._flatten_deeply(results)

            return results

        elif isinstance(key, int):  # 数组索引
            if not isinstance(current, list):
                raise ValueError(f"尝试使用索引但当前数据不是列表: {type(current)}")

            if key < 0 or key >= len(current):
                raise IndexError(f"数组索引超出范围: {key}")

            return self._get_recursive(current[key], remaining_keys, flatten_final)

        else:  # 字典键
            if not isinstance(current, dict):
                raise ValueError(
                    f"尝试获取键'{key}'但当前数据不是字典: {type(current)}"
                )

            if key not in current:
                raise KeyError(f"键'{key}'不存在")

            return self._get_recursive(current[key], remaining_keys, flatten_final)

    def _flatten_deeply(self, data: Any) -> Any:
        """
        深度扁平化列表

        :param data: 要扁平化的数据
        :return: 扁平化后的结果
        """
        if not isinstance(data, list):
            return data

        result = []
        for item in data:
            if isinstance(item, list):
                result.extend(self._flatten_deeply(item))
            else:
                result.append(item)
        return result

    def _set_recursive(
        self, current: Any, keys: List[Union[str, int, None]], value: Any
    ) -> Any:
        """
        递归设置数据

        :param current: 当前数据
        :param keys: 剩余键列表
        :param value: 要设置的值
        :return: 修改后的数据
        """
        if not keys:
            return value

        key = keys[0]
        remaining_keys = keys[1:]

        if key is None:  # [*] 遍历数组
            if not isinstance(current, list):
                raise ValueError(f"尝试遍历数组但当前数据不是列表: {type(current)}")

            for i, item in enumerate(current):
                current[i] = self._set_recursive(item, remaining_keys, value)
            return current

        elif isinstance(key, int):  # 数组索引
            if not isinstance(current, list):
                current = []

            # 扩展数组以适应索引
            while len(current) <= key:
                current.append(None)

            current[key] = self._set_recursive(current[key], remaining_keys, value)
            return current

        else:  # 字典键
            if not isinstance(current, dict):
                current = {}

            if remaining_keys:
                if key not in current:
                    # 根据下一个键的类型决定创建字典还是列表
                    next_key = remaining_keys[0]
                    current[key] = (
                        [] if (isinstance(next_key, int) or next_key is None) else {}
                    )

                current[key] = self._set_recursive(current[key], remaining_keys, value)
            else:
                current[key] = value

            return current

    def _delete_recursive(
        self, current: Any, keys: List[Union[str, int, None]]
    ) -> bool:
        """
        递归删除数据

        :param current: 当前数据
        :param keys: 剩余键列表
        :return: 是否成功删除
        """
        if not keys:
            return False

        if len(keys) == 1:
            key = keys[0]

            if isinstance(key, str) and isinstance(current, dict):
                if key in current:
                    del current[key]
                    return True
            elif isinstance(key, int) and isinstance(current, list):
                if 0 <= key < len(current):
                    current.pop(key)
                    return True

            return False

        key = keys[0]
        remaining_keys = keys[1:]

        if key is None:  # [*] 遍历数组
            if not isinstance(current, list):
                return False

            success = False
            for item in current:
                if self._delete_recursive(item, remaining_keys):
                    success = True
            return success

        elif isinstance(key, int):  # 数组索引
            if not isinstance(current, list) or key >= len(current):
                return False

            return self._delete_recursive(current[key], remaining_keys)

        else:  # 字典键
            if not isinstance(current, dict) or key not in current:
                return False

            return self._delete_recursive(current[key], remaining_keys)

    def _collect_paths(self, current: Any, prefix: str, paths: List[str]) -> None:
        """
        收集所有路径

        :param current: 当前数据
        :param prefix: 当前路径前缀
        :param paths: 路径收集列表
        """
        if isinstance(current, dict):
            for key, value in current.items():
                new_prefix = f"{prefix}.{key}" if prefix else key
                paths.append(new_prefix)
                self._collect_paths(value, new_prefix, paths)
        elif isinstance(current, list):
            for i, item in enumerate(current):
                new_prefix = f"{prefix}[{i}]" if prefix else f"[{i}]"
                paths.append(new_prefix)
                self._collect_paths(item, new_prefix, paths)

    def copy(self) -> "DictPathParser":
        """
        创建解析器的深拷贝

        :return: 新的解析器实例
        """
        return DictPathParser(copy.deepcopy(self.data))

    def to_dict(self) -> Union[dict, list]:
        """
        获取数据的副本

        :return: 数据副本
        """
        return copy.deepcopy(self.data)

    def __str__(self) -> str:
        return str(self.data)

    def __repr__(self) -> str:
        return f"DictPathParser({self.data})"


class DictPathParserFlat:
    """
    专门用于扁平化操作的解析器版本
    等价于原始版本的行为
    """

    def __init__(self, data: Union[dict, list] = None):
        self._parser = DictPathParser(data)

    def get(self, path: str) -> Any:
        """扁平化获取"""
        return self._parser.get_flat(path)

    def set(self, path: str, value: Any) -> None:
        """设置数据"""
        return self._parser.set(path, value)

    def has(self, path: str) -> bool:
        """检查路径"""
        return self._parser.has(path)

    def delete(self, path: str) -> bool:
        """删除数据"""
        return self._parser.delete(path)

    @property
    def data(self):
        """获取数据"""
        return self._parser.data


def get_dict_val_by_path(
    data: Union[dict, list], path: str, preserve_structure: bool = True
) -> Any:
    """
    便捷函数：根据路径获取数据

    :param data: 数据结构
    :param path: 路径字符串
    :param preserve_structure: 是否保持结构（False为扁平化）
    :return: 解析结果
    """
    parser = DictPathParser(data)
    return parser.get(path, flatten_final=not preserve_structure)


def get_dic_val_by_path_flat(data: Union[dict, list], path: str) -> Any:
    """
    便捷函数：根据路径获取数据（扁平化版本）

    :param data: 数据结构
    :param path: 路径字符串
    :return: 扁平化的解析结果
    """
    return get_dict_val_by_path(data, path, preserve_structure=False)


def set_dict_val_by_path(
    data: Union[dict, list], path: str, value: Any
) -> Union[dict, list]:
    """
    便捷函数：根据路径设置数据

    :param data: 数据结构
    :param path: 路径字符串
    :param value: 要设置的值
    :return: 修改后的数据结构
    """
    parser = DictPathParser(copy.deepcopy(data))
    parser.set(path, value)
    return parser.to_dict()
