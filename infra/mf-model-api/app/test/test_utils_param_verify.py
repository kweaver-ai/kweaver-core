"""测试 param_verify_utils 模块"""
import pytest
from app.utils.param_verify_utils import (
    verify_icon_color_config,
    verify_icon_color_config_metric,
    verify_text_field,
    verify_id,
    include_dataset_id,
    llm_source_verify
)


class TestVerifyIconColorConfig:
    """测试verify_icon_color_config函数"""

    def test_valid_colors(self):
        """测试有效颜色"""
        valid_colors = [
            "icon-color-pz-019688",
            "icon-color-pz-F759AB",
            "icon-color-pz-FADB14",
            "icon-color-pz-FF8501",
            "icon-color-pz-F75959",
            "icon-color-pz-8C8C8C",
            "icon-color-pz-126EE3",
            "icon-color-pz-13C2C2",
            "icon-color-pz-52C41A",
            "icon-color-pz-9254DE"
        ]
        for color in valid_colors:
            assert verify_icon_color_config(color) is True

    def test_invalid_colors(self):
        """测试无效颜色"""
        invalid_colors = [
            "icon-color-pz-FFFFFF",
            "invalid-color",
            "",
            "icon-color-zbk-FF8501",  # 错误的前缀
            "icon-color-pz-"
        ]
        for color in invalid_colors:
            assert verify_icon_color_config(color) is False


class TestVerifyIconColorConfigMetric:
    """测试verify_icon_color_config_metric函数"""

    def test_valid_metric_colors(self):
        """测试有效的指标颜色"""
        valid_colors = [
            "icon-color-zbk-FF8501",
            "icon-color-zbk-13C2C2",
            "icon-color-zbk-FADB14",
            "icon-color-zbk-019688",
            "icon-color-zbk-9254DE",
            "icon-color-zbk-8C8C8C",
            "icon-color-zbk-126EE3",
            "icon-color-zbk-52C41A",
            "icon-color-zbk-F759AB",
            "icon-color-zbk-F75959"
        ]
        for color in valid_colors:
            assert verify_icon_color_config_metric(color) is True

    def test_invalid_metric_colors(self):
        """测试无效的指标颜色"""
        invalid_colors = [
            "icon-color-zbk-FFFFFF",
            "invalid-color",
            "",
            "icon-color-pz-FF8501",  # 错误的前缀
            "icon-color-zbk-"
        ]
        for color in invalid_colors:
            assert verify_icon_color_config_metric(color) is False


class TestVerifyTextField:
    """测试verify_text_field函数"""

    def test_valid_text_within_limit(self):
        """测试有效文本在长度限制内"""
        assert verify_text_field("测试文本", 10) is True
        assert verify_text_field("test", 100) is True
        assert verify_text_field("", 10) is True

    def test_valid_text_exact_limit(self):
        """测试有效文本正好在长度限制"""
        assert verify_text_field("a" * 50, 50) is True

    def test_text_exceeds_limit(self):
        """测试文本超过长度限制"""
        assert verify_text_field("a" * 101, 100) is False

    def test_valid_characters(self):
        """测试有效字符"""
        valid_strings = [
            "hello123",
            "你好世界",
            "test@test.com",
            "!@#$%^&*()",
            "中英文混合123"
        ]
        for s in valid_strings:
            assert verify_text_field(s, 100) is True

    def test_invalid_type(self):
        """测试无效类型"""
        assert verify_text_field(123, 10) is False
        assert verify_text_field(None, 10) is False
        assert verify_text_field([], 10) is False

    def test_special_characters(self):
        """测试特殊字符"""
        assert verify_text_field("测试-文本_123", 50) is True
        assert verify_text_field("test/path", 50) is True


class TestVerifyId:
    """测试verify_id函数"""

    def test_valid_18_digit_id(self):
        """测试有效的18位ID"""
        assert verify_id("123456789012345678") is True

    def test_valid_19_digit_id(self):
        """测试有效的19位ID"""
        assert verify_id("1234567890123456789") is True

    def test_invalid_length(self):
        """测试无效长度"""
        assert verify_id("12345") is False
        assert verify_id("12345678901234567") is False  # 17位
        assert verify_id("12345678901234567890") is False  # 20位

    def test_invalid_characters(self):
        """测试无效字符"""
        assert verify_id("12345678901234567a") is False
        assert verify_id("123456789012345-78") is False
        assert verify_id("123456789012345 78") is False

    def test_invalid_type(self):
        """测试无效类型"""
        assert verify_id(123456789012345678) is False
        assert verify_id(None) is False
        assert verify_id([]) is False


class TestIncludeDatasetId:
    """测试include_dataset_id函数"""

    def test_dataset_id_found(self):
        """测试找到数据集ID"""
        dataset_version_id_list = ["123/v1", "456/v2", "789/v3"]
        assert include_dataset_id(dataset_version_id_list, "123") is True
        assert include_dataset_id(dataset_version_id_list, "456") is True

    def test_dataset_id_not_found(self):
        """测试未找到数据集ID"""
        dataset_version_id_list = ["123/v1", "456/v2"]
        assert include_dataset_id(dataset_version_id_list, "999") is False

    def test_empty_list(self):
        """测试空列表"""
        assert include_dataset_id([], "123") is False

    def test_invalid_format(self):
        """测试无效格式"""
        dataset_version_id_list = ["invalid", "no-slash"]
        result = include_dataset_id(dataset_version_id_list, "123")
        assert result is False

    def test_exception_handling(self):
        """测试异常处理"""
        # 传入非列表
        assert include_dataset_id("not a list", "123") is False


class TestLlmSourceVerify:
    """测试llm_source_verify函数"""

    def test_valid_parameters(self):
        """测试有效参数"""
        result = llm_source_verify(
            order="desc",
            page="1",
            size="10",
            rule="update_time",
            series="openai",
            name="test_model",
            model_type="llm"
        )
        assert result is False  # 验证通过返回False

    def test_invalid_page(self):
        """测试无效的page"""
        result = llm_source_verify(
            order="desc",
            page="0",  # 无效
            size="10",
            rule="update_time",
            series="openai",
            name="",
            model_type=""
        )
        assert result is not False  # 应该返回错误

    def test_invalid_size(self):
        """测试无效的size"""
        result = llm_source_verify(
            order="desc",
            page="1",
            size="abc",  # 无效
            rule="update_time",
            series="openai",
            name="",
            model_type=""
        )
        assert result is not False

    def test_invalid_order(self):
        """测试无效的order"""
        result = llm_source_verify(
            order="invalid",  # 无效
            page="1",
            size="10",
            rule="update_time",
            series="openai",
            name="",
            model_type=""
        )
        assert result is not False

    def test_invalid_rule(self):
        """测试无效的rule"""
        result = llm_source_verify(
            order="desc",
            page="1",
            size="10",
            rule="invalid",  # 无效
            series="openai",
            name="",
            model_type=""
        )
        assert result is not False

    def test_empty_series(self):
        """测试空series"""
        result = llm_source_verify(
            order="desc",
            page="1",
            size="10",
            rule="update_time",
            series="",  # 无效
            name="",
            model_type=""
        )
        assert result is not False

    def test_invalid_model_type(self):
        """测试无效的model_type"""
        result = llm_source_verify(
            order="desc",
            page="1",
            size="10",
            rule="update_time",
            series="openai",
            name="",
            model_type="invalid"  # 无效
        )
        assert result is not False

    def test_valid_model_types(self):
        """测试有效的model_type"""
        for model_type in ["llm", "rlm", "vu", ""]:
            result = llm_source_verify(
                order="desc",
                page="1",
                size="10",
                rule="update_time",
                series="openai",
                name="",
                model_type=model_type
            )
            # 空字符串是有效的
            if model_type == "":
                assert result is False

    def test_name_with_special_characters(self):
        """测试带特殊字符的名称"""
        result = llm_source_verify(
            order="desc",
            page="1",
            size="10",
            rule="update_time",
            series="openai",
            name="测试模型123!@#",
            model_type=""
        )
        # 应该验证通过
        assert result is False

