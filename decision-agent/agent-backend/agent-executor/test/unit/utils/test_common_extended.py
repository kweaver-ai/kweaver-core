"""扩展单元测试 - utils/common.py 模块"""

import os
from unittest.mock import patch
from datetime import datetime
from decimal import Decimal
from enum import Enum
from pydantic import BaseModel


class TestGetCallerInfo:
    """测试get_caller_info函数"""

    def test_get_caller_info_returns_tuple(self):
        """测试返回元组"""
        from app.utils.common import get_caller_info

        filename, lineno = get_caller_info()
        assert isinstance(filename, str)
        assert isinstance(lineno, int)

    def test_get_caller_info_non_empty_filename(self):
        """测试文件名非空"""
        from app.utils.common import get_caller_info

        filename, _ = get_caller_info()
        assert len(filename) > 0

    def test_get_caller_info_positive_lineno(self):
        """测试行号为正数"""
        from app.utils.common import get_caller_info

        _, lineno = get_caller_info()
        assert lineno > 0


class TestIsInPod:
    """测试is_in_pod函数"""

    def test_is_in_pod_returns_bool(self):
        """测试返回布尔值"""
        from app.utils.common import is_in_pod

        result = is_in_pod()
        assert isinstance(result, bool)

    @patch.dict(os.environ, {}, clear=True)
    def test_is_in_pod_false_no_env(self):
        """测试无环境变量时返回False"""
        from app.utils.common import is_in_pod

        result = is_in_pod()
        assert result is False

    @patch.dict(
        os.environ,
        {"KUBERNETES_SERVICE_HOST": "10.0.0.1", "KUBERNETES_SERVICE_PORT": "443"},
    )
    def test_is_in_pod_true_with_env(self):
        """测试有Kubernetes环境变量时返回True"""
        from app.utils.common import is_in_pod

        result = is_in_pod()
        assert result is True

    @patch.dict(os.environ, {"KUBERNETES_SERVICE_HOST": "10.0.0.1"}, clear=True)
    def test_is_in_pod_false_partial_env(self):
        """测试部分环境变量时返回False"""
        from app.utils.common import is_in_pod

        result = is_in_pod()
        assert result is False


class TestFailureThreshold:
    """测试失败阈值相关函数"""

    def test_get_failure_threshold_default(self):
        """测试获取默认失败阈值"""
        from app.utils.common import get_failure_threshold

        assert get_failure_threshold() == 10

    def test_set_failure_threshold(self):
        """测试设置失败阈值"""
        from app.utils.common import set_failure_threshold, get_failure_threshold

        set_failure_threshold(20)
        assert get_failure_threshold() == 20

        # Reset to default
        set_failure_threshold(10)

    def test_set_failure_threshold_zero(self):
        """测试设置零阈值"""
        from app.utils.common import set_failure_threshold, get_failure_threshold

        set_failure_threshold(0)
        assert get_failure_threshold() == 0

        set_failure_threshold(10)

    def test_set_failure_threshold_large_value(self):
        """测试设置大阈值"""
        from app.utils.common import set_failure_threshold, get_failure_threshold

        set_failure_threshold(1000)
        assert get_failure_threshold() == 1000

        set_failure_threshold(10)


class TestRecoveryTimeout:
    """测试恢复超时相关函数"""

    def test_get_recovery_timeout_default(self):
        """测试获取默认恢复超时"""
        from app.utils.common import get_recovery_timeout

        assert get_recovery_timeout() == 5

    def test_set_recovery_timeout(self):
        """测试设置恢复超时"""
        from app.utils.common import set_recovery_timeout, get_recovery_timeout

        set_recovery_timeout(10)
        assert get_recovery_timeout() == 10

        # Reset
        set_recovery_timeout(5)

    def test_set_recovery_timeout_zero(self):
        """测试设置零超时"""
        from app.utils.common import set_recovery_timeout, get_recovery_timeout

        set_recovery_timeout(0)
        assert get_recovery_timeout() == 0

        set_recovery_timeout(5)


class TestLangFunctions:
    """测试语言相关函数"""

    def test_get_lang_returns_callable(self):
        """测试返回可调用对象"""
        from app.utils.common import get_lang

        lang_func = get_lang()
        assert callable(lang_func)

    def test_get_lang_returns_string(self):
        """测试返回字符串"""
        from app.utils.common import get_lang

        lang_func = get_lang()
        result = lang_func("test")
        assert isinstance(result, str)

    def test_set_lang(self):
        """测试设置语言函数"""
        from app.utils.common import set_lang, get_lang

        def custom_lang(s):
            return s.upper()

        set_lang(custom_lang)
        lang_func = get_lang()
        assert lang_func("test") == "TEST"

        # Reset
        import gettext

        set_lang(gettext.gettext)

    def test_get_request_lang_from_header_empty(self):
        """测试空header"""
        from app.utils.common import get_request_lang_from_header

        lang_func = get_request_lang_from_header({})
        assert callable(lang_func)

    def test_get_request_lang_from_header_english(self):
        """测试英语header"""
        from app.utils.common import get_request_lang_from_header

        lang_func = get_request_lang_from_header({"x-language": "en"})
        assert callable(lang_func)

    def test_get_request_lang_from_header_chinese(self):
        """测试中文header"""
        from app.utils.common import get_request_lang_from_header

        lang_func = get_request_lang_from_header({"x-language": "zh"})
        assert callable(lang_func)

    def test_get_request_lang_from_header_zh_cn(self):
        """测试zh-CN header"""
        from app.utils.common import get_request_lang_from_header

        lang_func = get_request_lang_from_header({"x-language": "zh-CN"})
        assert callable(lang_func)


class TestConvertToCamelCase:
    """测试convert_to_camel_case函数"""

    def test_convert_simple_string(self):
        """测试简单字符串"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("hello")
        assert result == "Hello"

    def test_convert_with_underscores(self):
        """测试带下划线字符串"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("hello_world")
        assert result == "HelloWorld"

    def test_convert_multiple_underscores(self):
        """测试多个下划线"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("hello_world_test")
        assert result == "HelloWorldTest"

    def test_convert_single_char(self):
        """测试单字符"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("a")
        assert result == "A"

    def test_convert_empty_string(self):
        """测试空字符串"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("")
        assert result == ""

    def test_convert_non_string_returns_none(self):
        """测试非字符串返回None"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case(123)
        assert result is None

    def test_convert_none_returns_none(self):
        """测试None返回None"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case(None)
        assert result is None

    def test_convert_already_camel_case(self):
        """测试已经是驼峰命名"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("HelloWorld")
        assert result == "HelloWorld"

    def test_convert_with_numbers(self):
        """测试包含数字"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("test_123_value")
        assert "Test" in result
        assert "123" in result
        assert "Value" in result


class TestConvertToValidClassName:
    """测试convert_to_valid_class_name函数"""

    def test_convert_simple_string(self):
        """测试简单字符串"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("TestClass")
        assert result == "TestClass"

    def test_convert_with_special_chars(self):
        """测试带特殊字符"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("Test-Class")
        # Function replaces special chars with underscore
        assert result == "Test_Class"

    def test_convert_with_spaces(self):
        """测试带空格"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("Test Class")
        assert "Test" in result
        assert "Class" in result

    def test_convert_empty_string(self):
        """测试空字符串"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("")
        assert result == ""

    def test_convert_starts_with_digit(self):
        """测试以数字开头"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("123test")
        assert result.startswith("_")

    def test_convert_all_special_chars(self):
        """测试全特殊字符"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("!@#$%")
        assert isinstance(result, str)


class TestTruncateByByteLen:
    """测试truncate_by_byte_len函数"""

    def test_truncate_short_string(self):
        """测试短字符串"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("hello", 100)
        assert result == "hello"

    def test_truncate_ascii_string(self):
        """测试ASCII字符串"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("a" * 100, 50)
        assert len(result) == 50

    def test_truncate_unicode_string(self):
        """测试Unicode字符串"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("你好" * 100, 50)
        assert len(result.encode("utf-8")) <= 50

    def test_truncate_empty_string(self):
        """测试空字符串"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("", 100)
        assert result == ""

    def test_truncate_zero_length(self):
        """测试零长度"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("hello", 0)
        assert result == ""

    def test_truncate_default_length(self):
        """测试默认长度"""
        from app.utils.common import truncate_by_byte_len

        long_string = "a" * 70000
        result = truncate_by_byte_len(long_string)
        assert len(result.encode("utf-8")) <= 65535


class TestIsValidUrl:
    """测试is_valid_url函数"""

    def test_valid_http_url(self):
        """测试有效HTTP URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://example.com") is True

    def test_valid_https_url(self):
        """测试有效HTTPS URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("https://example.com") is True

    def test_valid_url_with_path(self):
        """测试带路径的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("https://example.com/path/to/resource") is True

    def test_valid_url_with_query(self):
        """测试带查询参数的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("https://example.com?query=value") is True

    def test_valid_url_with_port(self):
        """测试带端口的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("https://example.com:8080") is True

    def test_invalid_url_no_scheme(self):
        """测试无协议的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("example.com") is False

    def test_invalid_url_no_netloc(self):
        """测试无网络位置的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("https://") is False

    def test_invalid_url_empty_string(self):
        """测试空字符串"""
        from app.utils.common import is_valid_url

        assert is_valid_url("") is False

    def test_valid_ftp_url(self):
        """测试FTP URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("ftp://example.com") is True


class TestFuncJudgment:
    """测试func_judgment函数"""

    def test_sync_function(self):
        """测试同步函数"""

        def sync_func():
            pass

        from app.utils.common import func_judgment

        async_flag, stream_flag = func_judgment(sync_func)
        assert async_flag is False
        assert stream_flag is False

    def test_async_function(self):
        """测试异步函数"""

        async def async_func():
            pass

        from app.utils.common import func_judgment

        async_flag, stream_flag = func_judgment(async_func)
        assert async_flag is True
        assert stream_flag is False

    def test_sync_generator_function(self):
        """测试同步生成器函数"""

        def sync_gen():
            yield 1

        from app.utils.common import func_judgment

        async_flag, stream_flag = func_judgment(sync_gen)
        assert async_flag is False
        assert stream_flag is True

    def test_async_generator_function(self):
        """测试异步生成器函数"""

        async def async_gen():
            yield 1

        from app.utils.common import func_judgment

        async_flag, stream_flag = func_judgment(async_gen)
        assert async_flag is True
        assert stream_flag is True


class TestMakeJsonSerializable:
    """测试make_json_serializable函数"""

    def test_serialize_simple_types(self):
        """测试简单类型"""
        from app.utils.common import make_json_serializable

        assert make_json_serializable("string") == "string"
        assert make_json_serializable(42) == 42
        assert make_json_serializable(3.14) == 3.14
        assert make_json_serializable(True) is True
        assert make_json_serializable(None) is None

    def test_serialize_list(self):
        """测试列表"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable([1, 2, 3])
        assert result == [1, 2, 3]

    def test_serialize_tuple(self):
        """测试元组"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable((1, 2, 3))
        assert result == [1, 2, 3]

    def test_serialize_dict(self):
        """测试字典"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable({"key": "value"})
        assert result == {"key": "value"}

    def test_serialize_nested_dict(self):
        """测试嵌套字典"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable({"outer": {"inner": "value"}})
        assert result == {"outer": {"inner": "value"}}

    def test_serialize_dict_with_embedding(self):
        """测试带embedding的字典"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable({"embedding": [1, 2, 3]})
        assert result["embedding"] is None

    def test_serialize_pydantic_model(self):
        """测试Pydantic模型"""
        from app.utils.common import make_json_serializable

        class TestModel(BaseModel):
            name: str
            value: int

        model = TestModel(name="test", value=42)
        result = make_json_serializable(model)
        assert result == {"name": "test", "value": 42}

    def test_serialize_enum(self):
        """测试枚举"""
        from app.utils.common import make_json_serializable

        class TestEnum(Enum):
            VALUE1 = "value1"

        result = make_json_serializable(TestEnum.VALUE1)
        assert result == "value1"

    def test_serialize_float_nan(self):
        """测试NaN浮点数"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(float("nan"))
        assert result is None

    def test_serialize_float_inf(self):
        """测试无穷大"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(float("inf"))
        # Should remain as inf (not NaN)
        assert result == float("inf")

    def test_serialize_datetime(self):
        """测试datetime"""
        from app.utils.common import make_json_serializable

        dt = datetime(2025, 1, 1, 12, 0, 0)
        result = make_json_serializable(dt)
        assert isinstance(result, datetime)

    def test_serialize_decimal(self):
        """测试Decimal"""
        from app.utils.common import make_json_serializable

        dec = Decimal("3.14")
        result = make_json_serializable(dec)
        assert result == dec

    def test_serialize_list_of_models(self):
        """测试模型列表"""
        from app.utils.common import make_json_serializable

        class TestModel(BaseModel):
            value: int

        models = [TestModel(value=1), TestModel(value=2)]
        result = make_json_serializable(models)
        assert result == [{"value": 1}, {"value": 2}]


class TestCreateSubclass:
    """测试create_subclass函数"""

    def test_create_simple_subclass(self):
        """测试创建简单子类"""
        from app.utils.common import create_subclass

        class Base:
            pass

        Subclass = create_subclass(Base, "Subclass", {})
        assert issubclass(Subclass, Base)

    def test_create_subclass_with_attributes(self):
        """测试带属性创建子类"""
        from app.utils.common import create_subclass

        class Base:
            pass

        attributes = {"custom_attr": "value"}
        Subclass = create_subclass(Base, "Subclass", attributes)

        instance = Subclass()
        assert instance.custom_attr == "value"

    def test_create_subclass_with_methods(self):
        """测试带方法创建子类"""
        from app.utils.common import create_subclass

        class Base:
            pass

        def custom_method(self):
            return "custom"

        attributes = {"custom_method": custom_method}
        Subclass = create_subclass(Base, "Subclass", attributes)

        instance = Subclass()
        assert instance.custom_method() == "custom"


class TestBackwardCompatibilityAliases:
    """测试向后兼容别名"""

    def test_get_caller_info_alias(self):
        """测试GetCallerInfo别名"""
        from app.utils.common import GetCallerInfo

        filename, lineno = GetCallerInfo()
        assert isinstance(filename, str)
        assert isinstance(lineno, int)

    def test_is_in_pod_alias(self):
        """测试IsInPod别名"""
        from app.utils.common import IsInPod

        result = IsInPod()
        assert isinstance(result, bool)

    def test_convert_to_camel_case_alias(self):
        """测试ConvertToCamelCase别名"""
        from app.utils.common import ConvertToCamelCase

        result = ConvertToCamelCase("hello_world")
        assert result == "HelloWorld"
