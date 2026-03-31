"""单元测试 - app/utils/common.py 中的 is_dolphin_var 函数

验证修复后的 is_dolphin_var 函数与原始 VarOutput.is_serialized_dict 实现保持一致。
"""

from app.utils.common import is_dolphin_var


class TestIsDolphinVar:
    """测试 is_dolphin_var 函数"""

    def test_returns_true_for_valid_var_output_dict(self):
        """测试：有效的 VarOutput 字典应该返回 True"""
        var_output_dict = {
            "__type__": "VarOutput",
            "name": "test_var",
            "value": 123,
            "source_type": "OTHER",
        }
        assert is_dolphin_var(var_output_dict) is True

    def test_returns_false_when_type_is_not_var_output(self):
        """测试：__type__ 不是 'VarOutput' 时应该返回 False"""
        other_dict = {"__type__": "OtherType", "name": "test", "value": 123}
        assert is_dolphin_var(other_dict) is False

    def test_returns_false_when_type_field_missing(self):
        """测试：缺少 __type__ 字段时应该返回 False"""
        dict_without_type = {"name": "test", "value": 123}
        assert is_dolphin_var(dict_without_type) is False

    def test_returns_false_for_non_dict_values(self):
        """测试：非字典类型应该返回 False"""
        assert is_dolphin_var(123) is False
        assert is_dolphin_var("string") is False
        assert is_dolphin_var([1, 2, 3]) is False
        assert is_dolphin_var(None) is False

    def test_returns_false_for_empty_dict(self):
        """测试：空字典应该返回 False"""
        assert is_dolphin_var({}) is False

    def test_returns_false_for_type_field_set_to_none(self):
        """测试：__type__ 字段为 None 时应该返回 False"""
        dict_with_none_type = {"__type__": None, "value": 123}
        assert is_dolphin_var(dict_with_none_type) is False

    def test_consistency_with_var_output_is_serialized_dict(self):
        """测试：与原始 VarOutput.is_serialized_dict 实现保持一致

        根据 https://github.com/kweaver-ai/dolphin 的实现：
        @staticmethod
        def is_serialized_dict(data) -> bool:
            return isinstance(data, dict) and data.get("__type__") == "VarOutput"
        """
        # 这是一个有效的 VarOutput 序列化字典
        valid_var_output = {
            "__type__": "VarOutput",
            "name": "agent_result",
            "value": {"answer": "test answer"},
            "source_type": "LLM",
        }
        assert is_dolphin_var(valid_var_output) is True

        # 这不是一个 VarOutput 字典
        invalid_dict = {"__type__": "OtherType", "value": "something"}
        assert is_dolphin_var(invalid_dict) is False

        # 没有 __type__ 字段
        no_type_field = {"value": "something", "name": "test"}
        assert is_dolphin_var(no_type_field) is False
