"""单元测试 - utils/regex_rules 模块"""

from jsonschema import ValidationError
from unittest.mock import Mock


class TestRegexPatterns:
    """测试 RegexPatterns 类"""

    def test_chinese_and_english_numbers_and_underline(self):
        """测试中文英文数字下划线模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_underline

        # Valid cases
        assert re.match(pattern, "test") is not None
        assert re.match(pattern, "test123") is not None
        assert re.match(pattern, "test_abc") is not None
        assert re.match(pattern, "测试") is not None
        assert re.match(pattern, "test测试123") is not None

        # Invalid cases
        assert re.match(pattern, "test-abc") is None
        assert re.match(pattern, "test abc") is None
        assert re.match(pattern, "") is None

    def test_chinese_and_english_numbers_and_special_symbols(self):
        """测试中文英文数字特殊键盘符号模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard

        # Valid cases
        assert re.match(pattern, "test测试") is not None
        assert re.match(pattern, "test-abc") is not None
        assert re.match(pattern, "test@abc") is not None
        assert re.match(pattern, "test#abc") is not None

        # Invalid - empty
        assert re.match(pattern, "") is None

    def test_chinese_and_english_numbers_and_special_symbols_allow_empty(self):
        """测试中文英文数字特殊键盘符号允许空模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard_allow_empty

        # Valid cases
        assert re.match(pattern, "") is not None
        assert re.match(pattern, "test测试") is not None

    def test_positive_integer(self):
        """测试正整数模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Positive_integer

        assert re.match(pattern, "1") is not None
        assert re.match(pattern, "123") is not None
        assert re.match(pattern, "999999") is not None

        assert re.match(pattern, "0") is None
        assert re.match(pattern, "-1") is None
        assert re.match(pattern, "1.5") is None

    def test_positive_integer_with_minus_1(self):
        """测试正整数或-1模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Positive_integer_with_minus_1

        assert re.match(pattern, "1") is not None
        assert re.match(pattern, "123") is not None
        assert re.match(pattern, "-1") is not None

        assert re.match(pattern, "0") is None
        assert re.match(pattern, "-2") is None

    def test_english_numbers_and_hyphen(self):
        """测试英文数字连字符模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.English_numbers_and_hyphen

        assert re.match(pattern, "test") is not None
        assert re.match(pattern, "test123") is not None
        assert re.match(pattern, "test-abc") is not None
        assert re.match(pattern, "test-abc-123") is not None

        assert re.match(pattern, "test_abc") is None
        assert re.match(pattern, "test abc") is None

    def test_snow_id_pattern(self):
        """测试雪花ID模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.snow_id_pattern

        assert re.match(pattern, "1234567890123456789") is not None  # 19 digits
        assert re.match(pattern, "0000000000000000001") is not None

        assert re.match(pattern, "12345678901234567890") is None  # 20 digits
        assert re.match(pattern, "123456789012345678") is None  # 18 digits

    def test_uuid_pattern(self):
        """测试UUID模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.uuid_pattern

        # 32 character hex string
        assert re.match(pattern, "0123456789abcdef0123456789abcdef") is not None
        assert re.match(pattern, "A" * 32) is not None

        assert re.match(pattern, "0123456789abcdef") is None  # 16 chars
        assert re.match(pattern, "g" * 32) is None  # Invalid hex

    def test_variable_in_curly_braces(self):
        """测试花括号变量模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Variable_in_curly_braces

        # Test extracting variables
        text = "Hello {{name}}, your age is {{age}}"
        matches = re.findall(pattern, text)

        assert matches == ["name", "age"]

    def test_simple_variable_with_dollar_sign(self):
        """测试简单美元符号变量模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Simple_variable_with_dollar_sign

        # Test extracting variables
        text = "Hello $name, your score is $score"
        matches = re.findall(pattern, text)

        assert matches == ["name", "score"]

    def test_complex_variable_with_dollar_sign(self):
        """测试复杂美元符号变量模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Complex_variable_with_dollar_sign

        # Test various variable formats
        assert re.search(pattern, "$x") is not None
        assert re.search(pattern, "$result[0]") is not None
        assert re.search(pattern, "$a.b.c") is not None


class TestGetErrorMessageByRegex:
    """测试 GetErrorMessageByRegex 函数"""

    def test_get_known_pattern_error(self):
        """测试获取已知模式的错误信息"""
        from app.utils.regex_rules import RegexPatterns, GetErrorMessageByRegex

        key = RegexPatterns.Positive_integer
        result = GetErrorMessageByRegex(key)

        assert "positive integer" in result.lower()

    def test_get_unknown_pattern_error(self):
        """测试获取未知模式的错误信息"""
        from app.utils.regex_rules import GetErrorMessageByRegex

        result = GetErrorMessageByRegex("unknown_pattern")

        assert "invalid" in result.lower()


class TestHandleJsonSchemaError:
    """测试 handleJsonSchemaError 函数"""

    def test_handle_required_field_error(self):
        """测试必填字段错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        # Create a mock ValidationError
        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = [""]
        exc.absolute_schema_path = ["required"]
        exc.validator_value = ["field1", "field2"]
        exc.instance = {}

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)

    def test_handle_type_error_string(self):
        """测试类型错误-字符串"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "string"

        result = handleJsonSchemaError(exc)

        assert "string" in result.lower() or "fieldName" in result

    def test_handle_type_error_integer(self):
        """测试类型错误-整数"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "integer"

        result = handleJsonSchemaError(exc)

        assert "int" in result.lower()

    def test_handle_min_length_error(self):
        """测试最小长度错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["minLength"]
        exc.validator_value = 5
        exc.instance = "abc"

        result = handleJsonSchemaError(exc)

        assert "5" in result

    def test_handle_max_length_error(self):
        """测试最大长度错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["maxLength"]
        exc.validator_value = 10
        exc.instance = "this is too long"

        result = handleJsonSchemaError(exc)

        assert "10" in result

    def test_handle_pattern_error(self):
        """测试正则表达式模式错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["pattern"]
        exc.validator_value = "^[a-z]+$"

        result = handleJsonSchemaError(exc)

        # Should include the regex error message
        assert "fieldName" in result

    def test_handle_unique_items_error(self):
        """测试唯一项错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["uniqueItems"]

        result = handleJsonSchemaError(exc)

        assert "duplicate" in result.lower()

    def test_handle_enum_error(self):
        """测试枚举错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["enum"]
        exc.validator_value = ["value1", "value2"]

        result = handleJsonSchemaError(exc)

        assert "value1" in result or "value2" in result


class TestRegexPatternsEdgeCases:
    """测试正则模式边界情况"""

    def test_oss_id_pattern_valid(self):
        """测试有效的OSS ID模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.oss_id_pattern_allow_empty

        # Valid OSS ID: AD-19digits-19digits
        assert (
            re.match(pattern, "AD-1234567890123456789-1234567890123456789") is not None
        )
        assert re.match(pattern, "") is not None  # Allow empty

    def test_oss_id_pattern_invalid(self):
        """测试无效的OSS ID模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.oss_id_pattern_allow_empty

        assert (
            re.match(pattern, "AD-123456789012345678-123456789012345678") is None
        )  # 18 digits

    def test_snow_id_pattern_allow_empty(self):
        """测试允许空的雪花ID模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.snow_id_pattern_allow_empty

        assert re.match(pattern, "") is not None
        assert re.match(pattern, "1234567890123456789") is not None

    def test_chinese_and_english_excluding_special_chars(self):
        """测试排除特殊字符的模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_some_keyboard_symbols_excluding_special_chars

        # Should allow letters, numbers, some symbols
        assert re.match(pattern, "test测试") is not None
        assert re.match(pattern, "test-abc") is not None
        assert re.match(pattern, "test_abc") is not None

        # Should exclude special chars like # / : * ? " < > |
        assert re.match(pattern, "test#abc") is None


class TestErrorMessageMapping:
    """测试错误消息映射"""

    def test_error_message_dict_keys(self):
        """测试错误消息字典包含所有模式"""
        from app.utils.regex_rules import RegexPatterns, regexErrorMessage

        # Check that all regex patterns have error messages
        patterns_with_errors = [
            RegexPatterns.Chinese_and_English_numbers_and_underline,
            RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard,
            RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard_allow_empty,
            RegexPatterns.Chinese_and_English_numbers_and_some_keyboard_symbols_excluding_special_chars,
            RegexPatterns.Positive_integer,
            RegexPatterns.Positive_integer_with_minus_1,
            RegexPatterns.Positive_integer_with_0,
            RegexPatterns.English_numbers_and_hyphen,
            RegexPatterns.oss_id_pattern_allow_empty,
            RegexPatterns.snow_id_pattern,
            RegexPatterns.snow_id_pattern_allow_empty,
            RegexPatterns.uuid_pattern,
        ]

        for pattern in patterns_with_errors:
            assert pattern in regexErrorMessage


class TestHandleJsonSchemaEdgeCases:
    """测试 handleJsonSchemaError 边界情况"""

    def test_handle_error_with_no_path(self):
        """测试无路径的错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = []
        exc.absolute_schema_path = []

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)
        assert len(result) > 0

    def test_handle_error_with_array_index_path(self):
        """测试带数组索引路径的错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["items", 0, "fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "string"

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)
        assert len(result) > 0
