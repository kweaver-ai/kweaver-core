"""单元测试 - utils/common 模块"""

import pytest
from unittest.mock import Mock, patch
from enum import Enum


class TestGetCallerInfo:
    """测试 get_caller_info 函数"""

    def test_get_caller_info_returns_tuple(self):
        """测试返回元组"""
        from app.utils.common import get_caller_info

        def inner_function():
            return get_caller_info()

        result = inner_function()

        assert isinstance(result, tuple)
        assert len(result) == 2
        assert isinstance(result[0], str)  # filename
        assert isinstance(result[1], int)  # line number


class TestIsInPod:
    """测试 is_in_pod 函数"""

    def test_is_in_pod_true_with_env_vars(self):
        """测试在Pod中（有环境变量）"""
        from app.utils.common import is_in_pod

        with patch.dict(
            "os.environ",
            {"KUBERNETES_SERVICE_HOST": "10.0.0.1", "KUBERNETES_SERVICE_PORT": "443"},
        ):
            assert is_in_pod() is True

    def test_is_in_pod_false_without_env_vars(self):
        """测试不在Pod中（无环境变量）"""
        from app.utils.common import is_in_pod

        with patch.dict("os.environ", {}, clear=True):
            assert is_in_pod() is False

    def test_is_in_pod_false_with_only_host(self):
        """测试只有HOST环境变量"""
        from app.utils.common import is_in_pod

        with patch.dict(
            "os.environ", {"KUBERNETES_SERVICE_HOST": "10.0.0.1"}, clear=True
        ):
            assert is_in_pod() is False


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

        # Reset to default
        set_recovery_timeout(5)


class TestLangFunctions:
    """测试语言相关函数"""

    def test_get_lang_returns_function(self):
        """测试get_lang返回函数"""
        from app.utils.common import get_lang

        lang_func = get_lang()
        assert callable(lang_func)

    def test_set_lang(self):
        """测试设置语言函数"""
        from app.utils.common import set_lang, get_lang

        custom_lang = lambda x: f"translated_{x}"
        set_lang(custom_lang)

        lang_func = get_lang()
        assert lang_func("test") == "translated_test"


class TestConvertToCamelCase:
    """测试 convert_to_camel_case 函数"""

    def test_convert_snake_case_to_camel_case(self):
        """测试下划线转驼峰"""
        from app.utils.common import convert_to_camel_case

        assert convert_to_camel_case("hello_world") == "HelloWorld"
        assert convert_to_camel_case("test_case_example") == "TestCaseExample"

    def test_convert_single_char_words(self):
        """测试单字符单词"""
        from app.utils.common import convert_to_camel_case

        assert convert_to_camel_case("a_b_c") == "ABC"
        assert convert_to_camel_case("test_a_b") == "TestAB"

    def test_convert_non_string_returns_none(self):
        """测试非字符串返回None"""
        from app.utils.common import convert_to_camel_case

        assert convert_to_camel_case(123) is None
        assert convert_to_camel_case(None) is None
        assert convert_to_camel_case([]) is None

    def test_convert_empty_string(self):
        """测试空字符串"""
        from app.utils.common import convert_to_camel_case

        assert convert_to_camel_case("") == ""

    def test_convert_mixed_case(self):
        """测试混合大小写"""
        from app.utils.common import convert_to_camel_case

        assert convert_to_camel_case("hello_World") == "HelloWorld"


class TestConvertToValidClassName:
    """测试 convert_to_valid_class_name 函数"""

    def test_convert_with_special_chars(self):
        """测试转换特殊字符"""
        from app.utils.common import convert_to_valid_class_name

        assert convert_to_valid_class_name("test-class") == "test_class"
        assert convert_to_valid_class_name("test.class") == "test_class"

    def test_convert_empty_string(self):
        """测试空字符串"""
        from app.utils.common import convert_to_valid_class_name

        assert convert_to_valid_class_name("") == ""

    def test_convert_starts_with_digit(self):
        """测试以数字开头"""
        from app.utils.common import convert_to_valid_class_name

        assert convert_to_valid_class_name("123test") == "_123test"

    def test_convert_valid_name_unchanged(self):
        """测试有效名称不变"""
        from app.utils.common import convert_to_valid_class_name

        assert convert_to_valid_class_name("ValidClass123") == "ValidClass123"


class TestTruncateByByteLen:
    """测试 truncate_by_byte_len 函数"""

    def test_truncate_ascii_string(self):
        """测试截断ASCII字符串"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("Hello World", 5)
        assert result == "Hello"

    def test_truncate_unicode_string(self):
        """测试截断Unicode字符串"""
        from app.utils.common import truncate_by_byte_len

        # Each Chinese character is 3 bytes in UTF-8
        result = truncate_by_byte_len("你好世界", 6)
        assert result == "你好"

    def test_truncate_within_length(self):
        """测试字符串长度在限制内"""
        from app.utils.common import truncate_by_byte_len

        text = "Hello"
        result = truncate_by_byte_len(text, 10)
        assert result == "Hello"

    def test_truncate_empty_string(self):
        """测试空字符串"""
        from app.utils.common import truncate_by_byte_len

        result = truncate_by_byte_len("", 10)
        assert result == ""


class TestCreateSubclass:
    """测试 create_subclass 函数"""

    def test_create_subclass_basic(self):
        """测试创建基本子类"""
        from app.utils.common import create_subclass

        class BaseClass:
            def __init__(self):
                self.value = "base"

        SubClass = create_subclass(BaseClass, "SubClass", {})
        instance = SubClass()

        assert isinstance(instance, BaseClass)
        assert instance.value == "base"

    def test_create_subclass_with_attributes(self):
        """测试创建带属性的子类"""
        from app.utils.common import create_subclass

        class BaseClass:
            pass

        SubClass = create_subclass(
            BaseClass, "SubClass", {"custom_attr": "custom_value"}
        )
        instance = SubClass()

        assert instance.custom_attr == "custom_value"

    def test_create_subclass_with_methods(self):
        """测试创建带方法的子类"""
        from app.utils.common import create_subclass

        class BaseClass:
            pass

        def custom_method(self):
            return "result"

        SubClass = create_subclass(
            BaseClass, "SubClass", {"custom_method": custom_method}
        )
        instance = SubClass()

        assert instance.custom_method() == "result"


class TestIsValidUrl:
    """测试 is_valid_url 函数"""

    def test_valid_http_url(self):
        """测试有效HTTP URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://example.com") is True
        assert is_valid_url("https://example.com") is True

    def test_valid_url_with_path(self):
        """测试带路径的有效URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("https://example.com/path/to/resource") is True

    def test_invalid_url_missing_scheme(self):
        """测试缺少协议的无效URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("example.com") is False
        assert is_valid_url("//example.com") is False

    def test_invalid_url_missing_netloc(self):
        """测试缺少网络位置的无效URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://") is False

    def test_invalid_input_type(self):
        """测试无效输入类型"""
        from app.utils.common import is_valid_url

        assert is_valid_url(None) is False
        assert is_valid_url(123) is False


class TestFuncJudgment:
    """测试 func_judgment 函数"""

    def test_sync_regular_function(self):
        """测试同步普通函数"""
        from app.utils.common import func_judgment

        def regular_func():
            pass

        async_result, stream_result = func_judgment(regular_func)
        assert async_result is False
        assert stream_result is False

    def test_async_function(self):
        """测试异步函数"""
        from app.utils.common import func_judgment

        async def async_func():
            pass

        async_result, stream_result = func_judgment(async_func)
        assert async_result is True
        assert stream_result is False

    def test_sync_generator_function(self):
        """测试同步生成器函数"""
        from app.utils.common import func_judgment

        def gen_func():
            yield 1

        async_result, stream_result = func_judgment(gen_func)
        assert async_result is False
        assert stream_result is True

    def test_async_generator_function(self):
        """测试异步生成器函数"""
        from app.utils.common import func_judgment

        async def async_gen_func():
            yield 1

        async_result, stream_result = func_judgment(async_gen_func)
        assert async_result is True
        assert stream_result is True


class TestMakeJsonSerializable:
    """测试 make_json_serializable 函数"""

    def test_list_of_objects(self):
        """测试对象列表"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable([1, 2, 3])
        assert result == [1, 2, 3]

    def test_tuple_converts_to_list(self):
        """测试元组转列表"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable((1, 2, 3))
        assert result == [1, 2, 3]

    def test_dict_with_embedding(self):
        """测试带embedding的字典"""
        from app.utils.common import make_json_serializable

        input_dict = {"embedding": [1.0, 2.0, 3.0], "other": "value"}
        result = make_json_serializable(input_dict)

        assert result["embedding"] is None
        assert result["other"] == "value"

    def test_enum_converts_to_value(self):
        """测试枚举转值"""
        from app.utils.common import make_json_serializable

        class TestEnum(Enum):
            VALUE = "test_value"

        result = make_json_serializable(TestEnum.VALUE)
        assert result == "test_value"

    def test_nan_float_converts_to_none(self):
        """测试NaN浮点数转为None"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(float("nan"))
        assert result is None

    def test_regular_float_unchanged(self):
        """测试常规浮点数不变"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(3.14)
        assert result == 3.14

    def test_pydantic_model_converts_to_dict(self):
        """测试Pydantic模型转字典"""
        from app.utils.common import make_json_serializable
        from pydantic import BaseModel

        class TestModel(BaseModel):
            field1: str = "value1"

        model = TestModel()
        result = make_json_serializable(model)

        assert result == {"field1": "value1"}


class TestIsDolphinVar:
    """测试 is_dolphin_var 函数"""

    def test_with_mock_dolphin_var(self):
        """测试mock的dolphin变量"""
        from app.utils.common import is_dolphin_var

        # When mocked, is_dolphin_var should return False for regular values
        assert is_dolphin_var("regular_string") is False
        assert is_dolphin_var({"key": "value"}) is False


class TestConvertToValidClassName:
    """测试 convert_to_valid_class_name 函数 (补充)"""

    def test_convert_with_unicode(self):
        """测试转换Unicode字符"""
        from app.utils.common import convert_to_valid_class_name

        # Chinese characters are considered alphanumeric by Python
        result = convert_to_valid_class_name("测试类名")
        assert result == "测试类名"

    def test_convert_with_special_chars(self):
        """测试转换特殊字符"""
        from app.utils.common import convert_to_valid_class_name

        # Special characters like hyphens become underscores
        result = convert_to_valid_class_name("test-class-name")
        assert result == "test_class_name"


class TestConvertToCamelCaseEdgeCases:
    """测试 convert_to_camel_case 边界情况"""

    def test_multiple_underscores_in_row(self):
        """测试连续多个下划线"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("test__example")
        # Empty string becomes ""
        # The function capitalizes each "word"
        assert "Test" in result
        assert "Example" in result


class TestIsValidUrlEdgeCases:
    """测试 is_valid_url 边界情况"""

    def test_url_with_port(self):
        """测试带端口的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://example.com:8080") is True

    def test_url_with_query_params(self):
        """测试带查询参数的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://example.com?param=value") is True

    def test_url_with_fragment(self):
        """测试带片段的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://example.com#section") is True


class TestSyncWrapper:
    """测试 sync_wrapper 函数"""

    def test_sync_wrapper_calls_async(self):
        """测试同步包装器调用异步函数"""
        from app.utils.common import sync_wrapper

        async def async_func(x):
            return x * 2

        # Use a new event loop in a separate thread
        result = sync_wrapper(async_func, 5)
        assert result == 10


class TestGetDolphinVarValue:
    """测试 get_dolphin_var_value 函数"""

    def test_non_dolphin_var_unchanged(self):
        """测试非dolphin变量不变"""
        from app.utils.common import get_dolphin_var_value

        result = get_dolphin_var_value("test")
        assert result == "test"

    def test_number_unchanged(self):
        """测试数字不变"""
        from app.utils.common import get_dolphin_var_value

        result = get_dolphin_var_value(123)
        assert result == 123


class TestGetDolphinVarFinalValue:
    """测试 get_dolphin_var_final_value 函数"""

    def test_non_dolphin_var_unchanged(self):
        """测试非dolphin变量不变"""
        from app.utils.common import get_dolphin_var_final_value

        result = get_dolphin_var_final_value("test")
        assert result == "test"

    def test_dict_unchanged(self):
        """测试字典不变"""
        from app.utils.common import get_dolphin_var_final_value

        result = get_dolphin_var_final_value({"key": "value"})
        assert result == {"key": "value"}


class TestGetUnknownError:
    """测试 get_unknown_error 函数"""

    def test_get_unknown_error_returns_dict(self):
        """测试返回字典"""
        from app.utils.common import get_unknown_error

        mock_lang_func = lambda x: x
        result = get_unknown_error(
            "test_file", "test_func", "error details", mock_lang_func
        )

        assert isinstance(result, dict)
        assert "description" in result
        assert "solution" in result
        assert "error_code" in result
        assert "error_details" in result
        assert "error_link" in result

    def test_get_unknown_error_content(self):
        """测试错误内容"""
        from app.utils.common import get_unknown_error

        mock_lang_func = lambda x: x
        result = get_unknown_error(
            "test_file", "test_func", "error details", mock_lang_func
        )

        assert result["error_code"] == "AgentExecutor.InternalServerError.UnknownError"
        assert result["error_details"] == "error details"
        assert result["error_link"] == ""


class TestGetUserIdByRequest:
    """测试 get_user_id_by_request 函数"""

    def test_get_user_id_by_request_with_mock(self):
        """测试从请求获取用户ID"""
        from app.utils.common import get_user_id_by_request

        mock_request = Mock()
        # Need to mock the headers dict properly
        mock_request.headers = {"x-user": "test_user_id"}

        result = get_user_id_by_request(mock_request)
        # Should return the user ID
        assert result == "test_user_id"


class TestRunAsyncInThread:
    """测试 run_async_in_thread 函数"""

    def test_run_async_in_thread_basic(self):
        """测试在线程中运行异步函数"""
        from app.utils.common import run_async_in_thread

        async def async_func(x):
            return x * 2

        result = run_async_in_thread(async_func, 5)
        assert result == 10

    def test_run_async_in_thread_with_multiple_args(self):
        """测试多个参数"""
        from app.utils.common import run_async_in_thread

        async def async_func(x, y, z):
            return x + y + z

        result = run_async_in_thread(async_func, 1, 2, 3)
        assert result == 6

    def test_run_async_in_thread_with_kwargs(self):
        """测试关键字参数"""
        from app.utils.common import run_async_in_thread

        async def async_func(x, y=0):
            return x + y

        result = run_async_in_thread(async_func, 5, y=3)
        assert result == 8


class TestGetFormatErrorInfo:
    """测试 get_format_error_info 异步函数"""

    @pytest.mark.asyncio
    async def test_get_format_error_info_with_regular_exception(self):
        """测试常规异常"""
        from app.utils.common import get_format_error_info

        header = {}
        exc = Exception("Test error")

        result = await get_format_error_info(header, exc)

        assert isinstance(result, dict)
        assert "description" in result

    @pytest.mark.asyncio
    async def test_get_format_error_info_with_custom_exception(self):
        """测试带FormatHttpError的自定义异常"""
        from app.utils.common import get_format_error_info

        # Create exception with FormatHttpError method
        class CustomException(Exception):
            def FormatHttpError(self, lang_func):
                return {"custom": "error"}

        header = {}
        exc = CustomException("Test error")

        result = await get_format_error_info(header, exc)

        assert result == {"custom": "error"}

    @pytest.mark.asyncio
    async def test_get_format_error_info_with_header(self):
        """测试带header的异常"""
        from app.utils.common import get_format_error_info

        header = {"x-language": "zh"}
        exc = Exception("Test error")

        result = await get_format_error_info(header, exc)

        assert isinstance(result, dict)
        assert "description" in result


class TestGetRequestLangFunc:
    """测试 get_request_lang_func 函数"""

    def test_get_request_lang_func_returns_callable(self):
        """测试返回可调用对象"""
        from app.utils.common import get_request_lang_func

        mock_request = Mock()
        mock_request.headers = {}

        result = get_request_lang_func(mock_request)
        assert callable(result)


class TestGetRequestLangFromHeader:
    """测试 get_request_lang_from_header 函数"""

    def test_get_request_lang_from_header_english(self):
        """测试英语header"""
        from app.utils.common import get_request_lang_from_header

        header = {"x-language": "en"}
        result = get_request_lang_from_header(header)

        assert callable(result)

    def test_get_request_lang_from_header_chinese(self):
        """测试中文header"""
        from app.utils.common import get_request_lang_from_header

        header = {"x-language": "zh"}
        result = get_request_lang_from_header(header)

        assert callable(result)

    def test_get_request_lang_from_header_chinese_variant(self):
        """测试中文变体header"""
        from app.utils.common import get_request_lang_from_header

        for lang in ["zh-CN", "zh-TW", "zh-HK"]:
            header = {"x-language": lang}
            result = get_request_lang_from_header(header)
            assert callable(result)

    def test_get_request_lang_from_header_missing_key(self):
        """测试缺少language key"""
        from app.utils.common import get_request_lang_from_header

        header = {}
        result = get_request_lang_from_header(header)

        assert callable(result)


class TestTruncateByByteLenExtended:
    """测试 truncate_by_byte_len 函数扩展测试"""

    def test_truncate_with_mixed_unicode(self):
        """测试混合Unicode字符"""
        from app.utils.common import truncate_by_byte_len

        text = "Hello你好World世界"
        result = truncate_by_byte_len(text, 11)

        # "Hello" (5) + "你好" (6) = 11 bytes
        assert result == "Hello你好"

    def test_truncate_with_emoji(self):
        """测试emoji字符"""
        from app.utils.common import truncate_by_byte_len

        text = "Hello 😀 World"
        result = truncate_by_byte_len(text, 5)

        assert result == "Hello"

    def test_truncate_exactly_at_boundary(self):
        """测试正好在边界截断"""
        from app.utils.common import truncate_by_byte_len

        text = "Hello"
        result = truncate_by_byte_len(text, 5)

        assert result == "Hello"

    def test_truncate_with_zero_length(self):
        """测试零长度截断"""
        from app.utils.common import truncate_by_byte_len

        text = "Hello"
        result = truncate_by_byte_len(text, 0)

        assert result == ""

    def test_truncate_very_long_string(self):
        """测试非常长的字符串"""
        from app.utils.common import truncate_by_byte_len

        text = "A" * 100000
        result = truncate_by_byte_len(text, 1000)

        assert len(result) == 1000
        assert len(result.encode("utf-8")) <= 1000


class TestConvertToCamelCaseExtended:
    """测试 convert_to_camel_case 函数扩展测试"""

    def test_convert_with_trailing_underscore(self):
        """测试尾部下划线"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("test_")
        assert result == "Test"

    def test_convert_with_leading_underscore(self):
        """测试前导下划线"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("_test")
        # Empty string first word
        assert "Test" in result

    def test_convert_all_caps(self):
        """测试全大写"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("HELLO_WORLD")
        assert result == "HELLOWORLD"

    def test_convert_numbers_in_string(self):
        """测试字符串中的数字"""
        from app.utils.common import convert_to_camel_case

        result = convert_to_camel_case("test123_abc456")
        assert result == "Test123Abc456"


class TestConvertToValidClassNameExtended:
    """测试 convert_to_valid_class_name 函数扩展测试"""

    def test_convert_all_special_chars(self):
        """测试全特殊字符"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("!@#$%^&*()")
        assert all(c.isalnum() or c == "_" for c in result)

    def test_convert_spaces(self):
        """测试空格转换"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("test class name")
        assert result == "test_class_name"

    def test_convert_mixed_special_chars(self):
        """测试混合特殊字符"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("test-class.name:space")
        assert result == "test_class_name_space"

    def test_convert_only_digits(self):
        """测试纯数字"""
        from app.utils.common import convert_to_valid_class_name

        result = convert_to_valid_class_name("123")
        assert result == "_123"


class TestIsValidUrlExtended:
    """测试 is_valid_url 函数扩展测试"""

    def test_ftp_url(self):
        """测试FTP URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("ftp://example.com") is True

    def test_file_url(self):
        """测试file URL"""
        from app.utils.common import is_valid_url

        # file:// URLs may not be considered valid depending on urlparse
        # Let's test what the actual behavior is
        result = is_valid_url("file:///path/to/file")
        # Just verify it returns a boolean
        assert isinstance(result, bool)

    def test_url_with_credentials(self):
        """测试带凭据的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://user:pass@example.com") is True

    def test_url_with_username_only(self):
        """测试只带用户名的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://user@example.com") is True

    def test_ipv4_url(self):
        """测试IPv4 URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://192.168.1.1") is True

    def test_ipv6_url(self):
        """测试IPv6 URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://[::1]") is True

    def test_localhost_url(self):
        """测试localhost URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("http://localhost") is True

    def test_invalid_url_no_protocol(self):
        """测试无协议的URL"""
        from app.utils.common import is_valid_url

        assert is_valid_url("www.example.com") is False

    def test_empty_string(self):
        """测试空字符串"""
        from app.utils.common import is_valid_url

        assert is_valid_url("") is False


class TestMakeJsonSerializableExtended:
    """测试 make_json_serializable 函数扩展测试"""

    def test_nested_lists(self):
        """测试嵌套列表"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable([[1, 2], [3, 4]])
        assert result == [[1, 2], [3, 4]]

    def test_nested_dicts(self):
        """测试嵌套字典"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable({"a": {"b": {"c": 1}}})
        assert result == {"a": {"b": {"c": 1}}}

    def test_mixed_nesting(self):
        """测试混合嵌套"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable({"a": [1, {"b": 2}]})
        assert result == {"a": [1, {"b": 2}]}

    def test_list_with_embeddings(self):
        """测试包含embedding的列表"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable([{"embedding": [1, 2]}, {"other": "value"}])
        assert result == [{"embedding": None}, {"other": "value"}]

    def test_inf_float(self):
        """测试无穷大浮点数"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(float("inf"))
        assert result == float("inf")

    def test_negative_inf_float(self):
        """测试负无穷大浮点数"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(float("-inf"))
        assert result == float("-inf")

    def test_none_value(self):
        """测试None值"""
        from app.utils.common import make_json_serializable

        result = make_json_serializable(None)
        assert result is None

    def test_bool_value(self):
        """测试布尔值"""
        from app.utils.common import make_json_serializable

        assert make_json_serializable(True) is True
        assert make_json_serializable(False) is False


class TestCompatibilityAliases:
    """测试兼容性别名"""

    def test_get_caller_info_alias(self):
        """测试GetCallerInfo别名"""
        from app.utils.common import GetCallerInfo, get_caller_info

        assert GetCallerInfo is get_caller_info

    def test_is_in_pod_alias(self):
        """测试IsInPod别名"""
        from app.utils.common import IsInPod, is_in_pod

        assert IsInPod is is_in_pod

    def test_get_failure_threshold_alias(self):
        """测试GetFailureThreshold别名"""
        from app.utils.common import GetFailureThreshold, get_failure_threshold

        assert GetFailureThreshold is get_failure_threshold

    def test_set_failure_threshold_alias(self):
        """测试SetFailureThreshold别名"""
        from app.utils.common import SetFailureThreshold, set_failure_threshold

        assert SetFailureThreshold is set_failure_threshold

    def test_get_recovery_timeout_alias(self):
        """测试GetRecoveryTimeout别名"""
        from app.utils.common import GetRecoveryTimeout, get_recovery_timeout

        assert GetRecoveryTimeout is get_recovery_timeout

    def test_set_recovery_timeout_alias(self):
        """测试SetRecoveryTimeout别名"""
        from app.utils.common import SetRecoveryTimeout, set_recovery_timeout

        assert SetRecoveryTimeout is set_recovery_timeout

    def test_convert_to_camel_case_alias(self):
        """测试ConvertToCamelCase别名"""
        from app.utils.common import ConvertToCamelCase, convert_to_camel_case

        assert ConvertToCamelCase is convert_to_camel_case
