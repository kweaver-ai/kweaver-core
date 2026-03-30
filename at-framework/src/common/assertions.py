import re
import traceback
from typing import List, Dict, Any, Optional, Union

from _pytest.assertion.util import _compare_eq_any
from genson import SchemaBuilder
from jsonschema.exceptions import ValidationError, SchemaError
from jsonschema.validators import validate

from src.common.db_connector import MySQLConnector
from src.common.logger import logger


class Assertions:
    """
    断言工具类，提供丰富的断言方法和完善的日志记录
    每个断言都会记录执行结果（通过/失败）和详细信息
    """

    @staticmethod
    def _log_assertion_result(
            assertion_name: str, success: bool, details: str, error_msg: str = None
    ):
        """
        记录断言执行结果

        :param assertion_name: 断言方法名称
        :param success: 断言是否成功
        :param details: 断言详细信息
        :param error_msg: 错误信息（失败时）
        """
        status = "✅ 通过" if success else "❌ 失败"
        log_msg = f"[断言] {assertion_name} - {status}: {details}"

        if success:
            logger.info(log_msg)
        else:
            logger.error(f"{log_msg} | 错误: {error_msg}")

    @staticmethod
    def equal(actual: Any, expected: Any, message: str = None) -> None:
        """断言两个值相等"""
        assertion_name = "相等断言"
        details = f"预期值: {expected}, 实际值: {actual}"

        try:
            assert actual == expected
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"值不相等: 预期 {expected}, 实际 {actual}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def not_equal(actual: Any, expected: Any, message: str = None) -> None:
        """断言两个值不相等"""
        assertion_name = "不相等断言"
        details = f"预期值不等于: {expected}, 实际值: {actual}"

        try:
            assert actual != expected
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"值相等: 预期不等于 {expected}, 但实际为 {actual}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_true(condition: bool, message: str = None) -> None:
        """断言条件为True"""
        assertion_name = "True断言"
        details = f"条件值: {condition}"

        try:
            assert condition
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or "预期条件为True，但实际为False"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_false(condition: bool, message: str = None) -> None:
        """断言条件为False"""
        assertion_name = "False断言"
        details = f"条件值: {condition}"

        try:
            assert not condition
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or "预期条件为False，但实际为True"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_none(value: Any, message: str = None) -> None:
        """断言值为None"""
        assertion_name = "None断言"
        details = f"值: {value}"

        try:
            assert value is None
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"预期值为None，但实际为: {value}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_not_none(value: Any, message: str = None) -> None:
        """断言值不为None"""
        assertion_name = "非None断言"
        details = f"值: {value}"

        try:
            assert value is not None
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or "预期值不为None，但实际为None"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_instance(obj: Any, expected_type: type, message: str = None) -> None:
        """断言对象是指定类型的实例"""
        assertion_name = "类型断言"
        details = f"对象类型: {type(obj).__name__}, 预期类型: {expected_type.__name__}"

        try:
            assert isinstance(obj, expected_type)
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = (
                    message
                    or f"对象类型不匹配: 预期 {expected_type.__name__}, 实际 {type(obj).__name__}"
            )
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def contains(
            container: Union[List, str, Dict], item: Any, message: str = None
    ) -> None:
        """断言容器包含指定元素"""
        assertion_name = "包含断言"
        container_type = type(container).__name__
        details = f"容器类型: {container_type}, 元素: {item}"

        try:
            assert item in container
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"容器不包含元素: {item}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def not_contains(
            container: Union[List, str, Dict], item: Any, message: str = None
    ) -> None:
        """断言容器不包含指定元素"""
        assertion_name = "不包含断言"
        container_type = type(container).__name__
        details = f"容器类型: {container_type}, 元素: {item}"

        try:
            assert item not in container
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"容器不应包含元素: {item}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def dict_contains_keys(
            dictionary: Dict, keys: List[str], message: str = None
    ) -> None:
        """断言字典包含指定的所有键"""
        assertion_name = "字典键包含断言"
        missing_keys = [key for key in keys if key not in dictionary]
        details = f"要求的键: {keys}, 缺少的键: {missing_keys}"

        try:
            assert not missing_keys
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"字典缺少键: {missing_keys}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def dict_contains(
            dictionary: Dict, expected_subset: Dict, message: str = None
    ) -> None:
        """断言字典包含指定的子集"""
        assertion_name = "字典子集断言"
        details = f"预期子集: {expected_subset}"

        try:
            for key, value in expected_subset.items():
                assert key in dictionary, f"缺少键: {key}"
                assert (
                        dictionary[key] == value
                ), f"键 {key} 的值不匹配，预期: {value}, 实际: {dictionary[key]}"
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"字典不包含预期子集: {expected_subset}"
            Assertions._log_assertion_result(assertion_name, False, details, str(e))
            raise AssertionError(error_msg) from e

    @staticmethod
    def list_contains_all_items(
            actual_list: List, expected_items: List, message: str = None
    ) -> None:
        """断言列表包含所有预期元素（不考虑顺序）"""
        assertion_name = "列表包含所有元素断言"
        missing_items = [item for item in expected_items if item not in actual_list]
        details = f"预期元素: {expected_items}, 缺少元素: {missing_items}"

        try:
            assert not missing_items
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"列表缺少元素: {missing_items}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def list_equals(
            actual_list: List, expected_list: List, message: str = None
    ) -> None:
        """断言列表完全相等（包括顺序）"""
        assertion_name = "列表相等断言"
        details = (
            f"预期列表长度: {len(expected_list)}, 实际列表长度: {len(actual_list)}"
        )

        try:
            assert actual_list == expected_list
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = (
                    message or f"列表不相等，预期: {expected_list}, 实际: {actual_list}"
            )
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def string_matches(pattern: str, text: str, message: str = None) -> None:
        """断言字符串匹配正则表达式"""
        assertion_name = "字符串正则匹配断言"
        details = f"正则表达式: {pattern}, 文本长度: {len(text)}"

        try:
            assert re.search(pattern, text)
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"字符串: {text} 不匹配正则表达式: {pattern}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def greater_than(
            actual: Union[int, float], expected: Union[int, float], message: str = None
    ) -> None:
        """断言实际值大于预期值"""
        assertion_name = "大于断言"
        details = f"实际值: {actual}, 比较值: {expected}"

        try:
            assert actual > expected
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"预期 {actual} 大于 {expected}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def less_than(
            actual: Union[int, float], expected: Union[int, float], message: str = None
    ) -> None:
        """断言实际值小于预期值"""
        assertion_name = "小于断言"
        details = f"实际值: {actual}, 比较值: {expected}"

        try:
            assert actual < expected
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"预期 {actual} 小于 {expected}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def greater_or_equal(
            actual: Union[int, float], expected: Union[int, float], message: str = None
    ) -> None:
        """断言实际值大于等于预期值"""
        assertion_name = "大于等于断言"
        details = f"实际值: {actual}, 比较值: {expected}"

        try:
            assert actual >= expected
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"预期 {actual} 大于等于 {expected}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def less_or_equal(
            actual: Union[int, float], expected: Union[int, float], message: str = None
    ) -> None:
        """断言实际值小于等于预期值"""
        assertion_name = "小于等于断言"
        details = f"实际值: {actual}, 比较值: {expected}"

        try:
            assert actual <= expected
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"预期 {actual} 小于等于 {expected}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def approx_equal(
            actual: Union[int, float],
            expected: Union[int, float],
            tolerance: float = 0.001,
            message: str = None,
    ) -> None:
        """断言浮点数在指定 tolerance 范围内相等"""
        assertion_name = "近似相等断言"
        diff = abs(actual - expected)
        details = (
            f"实际值: {actual}, 预期值: {expected}, 差值: {diff}, 容差: {tolerance}"
        )

        try:
            assert diff <= tolerance
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = (
                    message
                    or f"预期 {actual} 在 {tolerance} 范围内接近 {expected}，实际差值: {diff}"
            )
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def has_length(
            container: Union[List, str, Dict], expected_length: int, message: str = None
    ) -> None:
        """断言容器长度等于预期值"""
        assertion_name = "长度断言"
        actual_length = len(container)
        container_type = type(container).__name__
        details = f"容器类型: {container_type}, 实际长度: {actual_length}, 预期长度: {expected_length}"

        try:
            assert actual_length == expected_length
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = (
                    message
                    or f"容器长度不匹配: 预期 {expected_length}, 实际 {actual_length}"
            )
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_empty(container: Union[List, str, Dict], message: str = None) -> None:
        """断言容器为空"""
        assertion_name = "空容器断言"
        actual_length = len(container)
        container_type = type(container).__name__
        details = f"容器类型: {container_type}, 长度: {actual_length}"

        try:
            assert actual_length == 0
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"容器不为空: 长度为 {actual_length}"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def is_not_empty(container: Union[List, str, Dict], message: str = None) -> None:
        """断言容器不为空"""
        assertion_name = "非空容器断言"
        actual_length = len(container)
        container_type = type(container).__name__
        details = f"容器类型: {container_type}, 长度: {actual_length}"

        try:
            assert actual_length > 0
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or "容器为空"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def in_range(
            value: Union[int, float],
            min_val: Union[int, float],
            max_val: Union[int, float],
            message: str = None,
    ) -> None:
        """断言数值在指定范围内（包含边界）"""
        assertion_name = "范围断言"
        details = f"值: {value}, 范围: [{min_val}, {max_val}]"

        try:
            assert min_val <= value <= max_val
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"值 {value} 不在范围 [{min_val}, {max_val}] 内"
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def response_status_code(
            response: Any, expected_code: int, message: str = None
    ) -> None:
        """断言HTTP响应状态码为预期值"""
        assertion_name = "HTTP状态码断言"
        details = f"预期状态码: {expected_code}, 实际状态码: {response.status_code}"

        try:
            assert response.status_code == expected_code
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = (
                    message
                    or f"预期响应状态码为 {expected_code}，实际为 {response.status_code}"
            )
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e

    @staticmethod
    def json_contains(
            json_data: Dict, expected_subset: Dict, message: str = None
    ) -> None:
        """断言JSON数据包含预期的子集"""
        assertion_name = "JSON包含断言"
        details = f"预期子集键数: {len(expected_subset) if expected_subset else 0}"

        def _compare(actual, expected, path=""):
            if isinstance(expected, dict):
                if not isinstance(actual, dict):
                    raise AssertionError(
                        f"{path}: 预期为字典，实际为 {type(actual).__name__}"
                    )
                for key, value in expected.items():
                    current_path = f"{path}.{key}" if path else key
                    if key not in actual:
                        raise AssertionError(f"{current_path}: 缺少键")
                    _compare(actual[key], value, current_path)
            elif isinstance(expected, list):
                if not isinstance(actual, list):
                    raise AssertionError(
                        f"{path}: 预期为列表，实际为 {type(actual).__name__}"
                    )
                if len(actual) < len(expected):
                    raise AssertionError(
                        f"{path}: 实际列表长度 {len(actual)} 小于预期 {len(expected)}"
                    )
                for i, item in enumerate(expected):
                    current_path = f"{path}[{i}]"
                    _compare(actual[i], item, current_path)
            else:
                if not _compare_eq_any(actual, expected):
                    raise AssertionError(f"{path}: 预期值 {expected}，实际值 {actual}")

        try:
            _compare(json_data, expected_subset)
            Assertions._log_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = message or f"JSON数据不包含预期子集: {str(e)}"
            Assertions._log_assertion_result(assertion_name, False, details, str(e))
            raise AssertionError(error_msg) from e

    @staticmethod
    def contains_exact_items(
            key: str, actual_items: list, expected_ids: set, message: str = None
    ) -> None:
        """断言集合包含且仅包含指定ID，同时长度匹配（优化日志：精准标注失败项）"""
        assertion_name = "元素精确匹配断言"
        # 提取实际ID集合
        actual_ids = {item.get(key) for item in actual_items} if actual_items else set()
        # 长度校验
        actual_len = len(actual_items)
        expected_len = len(expected_ids)

        # ========== 核心优化：拆分校验项，明确失败类型 ==========
        # 1. 校验元素是否匹配
        missing_ids = expected_ids - actual_ids  # 预期有但实际没有的ID
        extra_ids = actual_ids - expected_ids  # 实际有但预期没有的ID
        is_ids_match = missing_ids == set() and extra_ids == set()
        # 2. 校验长度是否匹配
        is_len_match = actual_len == expected_len

        # 构建精准的日志详情
        details = []
        details.append(f"【基础信息】")
        details.append(f"  - 实际{key}列表: {sorted(actual_ids)}")
        details.append(f"  - 预期{key}列表: {sorted(expected_ids)}")
        details.append(f"  - 实际长度: {actual_len} | 预期长度: {expected_len}")

        details.append(f"\n【校验结果】")
        if is_ids_match:
            details.append(f"  - 元素匹配: ✅ 所有预期ID都存在，无多余ID")
        else:
            details.append(f"  - 元素匹配: ❌ 存在不匹配")
            if missing_ids:
                details.append(f"    → 缺失预期ID: {sorted(missing_ids)}")
            if extra_ids:
                details.append(f"    → 多出非预期ID: {sorted(extra_ids)}")

        if is_len_match:
            details.append(f"  - 长度匹配: ✅ 长度一致")
        else:
            details.append(
                f"  - 长度匹配: ❌ 长度不符（差异: {actual_len - expected_len}）"
            )

        # 拼接最终详情字符串
        details_str = "\n".join(details)
        is_all_match = is_ids_match and is_len_match

        try:
            assert is_all_match
            Assertions._log_assertion_result(
                assertion_name, success=True, details=details_str
            )
        except AssertionError as e:
            # 生成精准的错误提示
            error_reasons = []
            if not is_ids_match:
                error_reasons.append(
                    f"元素不匹配（缺失: {missing_ids}, 多余: {extra_ids}）"
                )
            if not is_len_match:
                error_reasons.append(
                    f"长度不符（实际{actual_len}, 预期{expected_len}）"
                )
            error_msg = message or f"断言失败: {'; '.join(error_reasons)}"

            Assertions._log_assertion_result(
                assertion_name, success=False, details=details_str, error_msg=error_msg
            )
            raise AssertionError(error_msg) from e

    @staticmethod
    def assert_json_schema(example: dict, instance: dict) -> None:
        """
        json格式断言，基于样例判断返回结果是否符合格式
        """
        assertion_name = "响应格式断言"
        details = "预期格式: %s, 实际响应: %s" % (example, instance)

        builder = SchemaBuilder()
        builder.add_object(example)
        json_schema = builder.to_schema()
        json_schema["$schema"] = "https://json-schema.org/draft/2020-12/schema"

        try:
            validate(instance=instance, schema=json_schema)
            Assertions._log_assertion_result(assertion_name, True, details)
        except (ValidationError, SchemaError) as e:
            error_msg = "响应格式错误: %s" % traceback.format_exc()
            Assertions._log_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(error_msg) from e


class DatabaseAssertion:
    """数据库查询断言工具类，基于MySQLConnector实现数据库验证功能，包含完善的日志记录"""

    def __init__(self, db_connector: MySQLConnector):
        """
        初始化断言工具类

        :param db_connector: 已实例化的MySQLConnector对象
        """
        self.db = db_connector

    @staticmethod
    def _log_db_assertion_result(
            assertion_name: str, success: bool, details: str, error_msg: str = None
    ):
        """
        记录数据库断言执行结果

        :param assertion_name: 断言方法名称
        :param success: 断言是否成功
        :param details: 断言详细信息
        :param error_msg: 错误信息（失败时）
        """
        status = "✅ 通过" if success else "❌ 失败"
        log_msg = f"[数据库断言] {assertion_name} - {status}: {details}"

        if success:
            logger.info(log_msg)
        else:
            logger.error(f"{log_msg} | 错误: {error_msg}")

    def assert_record_exists(
            self, table: str, condition: str, params: Optional[Union[tuple, dict]] = None
    ) -> None:
        """
        断言符合条件的记录存在

        :param table: 表名
        :param condition: 查询条件，如"id = %s AND status = %s"
        :param params: 条件参数，与查询条件中的占位符对应
        :raises AssertionError: 当记录不存在时触发
        """
        assertion_name = "记录存在断言"
        sql = f"SELECT 1 FROM {table} WHERE {condition} LIMIT 1"
        details = f"表: {table}, 条件: {condition}, 参数: {params}"

        try:
            result = self.db.fetch_one(sql, params)
            assert result is not None
            self._log_db_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = f"表 [{table}] 中不存在符合条件 [{condition}] 的记录"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(f"断言失败: {error_msg}，参数: {params}") from e
        except Exception as e:
            error_msg = f"数据库查询异常: {str(e)}"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise

    def assert_record_not_exists(
            self, table: str, condition: str, params: Optional[Union[tuple, dict]] = None
    ) -> None:
        """
        断言符合条件的记录不存在

        :param table: 表名
        :param condition: 查询条件
        :param params: 条件参数
        :raises AssertionError: 当记录存在时触发
        """
        assertion_name = "记录不存在断言"
        sql = f"SELECT 1 FROM {table} WHERE {condition} LIMIT 1"
        details = f"表: {table}, 条件: {condition}, 参数: {params}"

        try:
            result = self.db.fetch_one(sql, params)
            assert result is None
            self._log_db_assertion_result(assertion_name, True, details)
        except AssertionError as e:
            error_msg = f"表 [{table}] 中存在符合条件 [{condition}] 的记录"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(f"断言失败: {error_msg}，参数: {params}") from e
        except Exception as e:
            error_msg = f"数据库查询异常: {str(e)}"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise

    def assert_record_count(
            self,
            table: str,
            condition: str,
            expected_count: int,
            params: Optional[Union[tuple, dict]] = None,
    ) -> None:
        """
        断言符合条件的记录数量与预期一致

        :param table: 表名
        :param condition: 查询条件，无条件时用"1=1"
        :param expected_count: 预期记录数量
        :param params: 条件参数
        :raises AssertionError: 当实际数量与预期不符时触发
        """
        assertion_name = "记录数量断言"
        sql = f"SELECT COUNT(*) AS count FROM {table} WHERE {condition}"
        details = f"表: {table}, 条件: {condition}, 预期数量: {expected_count}, 参数: {params}"

        try:
            result = self.db.fetch_one(sql, params)
            actual_count = result["count"] if isinstance(result, dict) else result[0]
            assert actual_count == expected_count
            self._log_db_assertion_result(
                assertion_name, True, f"{details}, 实际数量: {actual_count}"
            )
        except AssertionError as e:
            error_msg = f"记录数量不匹配: 实际 {actual_count}, 预期 {expected_count}"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(
                f"断言失败: 表 [{table}] 中符合条件 [{condition}] 的记录数为 {actual_count}，预期为 {expected_count}，参数: {params}"
            ) from e
        except Exception as e:
            error_msg = f"数据库查询异常: {str(e)}"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise

    def assert_field_equals(
            self,
            table: str,
            condition: str,
            field: str,
            expected_value: Any,
            params: Optional[Union[tuple, dict]] = None,
    ) -> None:
        """
        断言符合条件的记录的指定字段值等于预期值

        :param table: 表名
        :param condition: 查询条件（应确保唯一匹配一条记录）
        :param field: 要验证的字段名
        :param expected_value: 预期字段值
        :param params: 条件参数
        :raises AssertionError: 当字段值与预期不符或记录不存在时触发
        """
        assertion_name = "字段值断言"
        sql = f"SELECT {field} FROM {table} WHERE {condition} LIMIT 1"
        details = f"表: {table}, 条件: {condition}, 字段: {field}, 预期值: {expected_value}, 参数: {params}"

        try:
            result = self.db.fetch_one(sql, params)
            assert result is not None, f"记录不存在"

            actual_value = result[field] if isinstance(result, dict) else result[0]
            assert actual_value == expected_value
            self._log_db_assertion_result(
                assertion_name, True, f"{details}, 实际值: {actual_value}"
            )
        except AssertionError as e:
            if "记录不存在" in str(e):
                error_msg = f"表 [{table}] 中不存在符合条件 [{condition}] 的记录，无法验证字段 [{field}]"
            else:
                error_msg = f"字段值不匹配: 实际 {actual_value}, 预期 {expected_value}"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise AssertionError(f"断言失败: {error_msg}，参数: {params}") from e
        except Exception as e:
            error_msg = f"数据库查询异常: {str(e)}"
            self._log_db_assertion_result(assertion_name, False, details, error_msg)
            raise

    def assert_field_in(
            self,
            table: str,
            condition: str,
            field: str,
            expected_values: List[Any],
            params: Optional[Union[tuple, dict]] = None,
    ) -> None:
        """
        断言符合条件的记录的指定字段值在预期列表中

        :param table: 表名
        :param condition: 查询条件
        :param field: 要验证的字段名
        :param expected_values: 预期值列表
        :param params: 条件参数
        :raises AssertionError: 当字段值不在预期列表中时触发
        """
        sql = f"SELECT {field} FROM {table} WHERE {condition}"
        results = self.db.fetch_all(sql, params)

        assert (
                len(results) > 0
        ), f"断言失败: 表 [{table}] 中不存在符合条件 [{condition}] 的记录，无法验证字段 [{field}]，参数: {params}"

        for idx, result in enumerate(results):
            actual_value = result[field] if isinstance(result, dict) else result[0]
            assert (
                    actual_value in expected_values
            ), f"断言失败: 表 [{table}] 中第 {idx + 1} 条符合条件 [{condition}] 的记录的字段 [{field}] 值为 {actual_value}，不在预期列表 {expected_values} 中，参数: {params}"

        logger.info(
            f"断言成功: 表 [{table}] 中所有符合条件 [{condition}] 的记录的字段 [{field}] 值均在预期列表中"
        )

    def assert_multiple_fields(
            self,
            table: str,
            condition: str,
            expected_fields: Dict[str, Any],
            params: Optional[Union[tuple, dict]] = None,
    ) -> None:
        """
        断言符合条件的记录的多个字段值与预期一致

        :param table: 表名
        :param condition: 查询条件（应确保唯一匹配一条记录）
        :param expected_fields: 预期字段值字典，如{"name": "test", "status": 1}
        :param params: 条件参数
        :raises AssertionError: 当任何字段值与预期不符时触发
        """
        fields = ", ".join(expected_fields.keys())
        sql = f"SELECT {fields} FROM {table} WHERE {condition} LIMIT 1"
        result = self.db.fetch_one(sql, params)

        assert (
                result is not None
        ), f"断言失败: 表 [{table}] 中不存在符合条件 [{condition}] 的记录，无法验证字段，参数: {params}"

        for field, expected_value in expected_fields.items():
            actual_value = (
                result[field]
                if isinstance(result, dict)
                else result[list(expected_fields.keys()).index(field)]
            )
            assert (
                    actual_value == expected_value
            ), f"断言失败: 表 [{table}] 中符合条件 [{condition}] 的记录的字段 [{field}] 值为 {actual_value}，预期为 {expected_value}，参数: {params}"

        logger.info(
            f"断言成功: 表 [{table}] 中符合条件 [{condition}] 的记录的所有字段值均符合预期"
        )


# 全局断言实例
asserts = Assertions()


def get_db_asserts(
        database: str = None, env_config: dict = None, db_connector: MySQLConnector = None
):
    """
    获取数据库断言实例，支持多种配置方式

    :param database: 指定数据库名称，其他参数使用默认配置
    :param env_config: 环境配置字典
    :param db_connector: 直接传入数据库连接器实例
    :return: DatabaseAssertion实例
    """
    from src.common.db_connector import MySQLConnector

    if db_connector:
        # 直接使用传入的连接器
        return DatabaseAssertion(db_connector)
    elif env_config:
        # 使用环境配置
        connector = MySQLConnector.from_env_config(env_config)
        return DatabaseAssertion(connector)
    elif database:
        # 指定数据库名称
        connector = MySQLConnector.for_database(database)
        return DatabaseAssertion(connector)
    else:
        # 使用默认配置
        connector = MySQLConnector()
        return DatabaseAssertion(connector)
