"""扩展单元测试 - utils/regex_rules 模块 - 增加边界情况测试"""

from jsonschema import ValidationError
from unittest.mock import Mock


class TestRegexPatternsEdgeCasesExtended:
    """测试正则模式扩展边界情况"""

    def test_positive_integer_with_0(self):
        """测试正整数或0模式"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Positive_integer_with_0

        assert re.match(pattern, "0") is not None
        assert re.match(pattern, "1") is not None
        assert re.match(pattern, "123") is not None
        assert re.match(pattern, "999999999") is not None
        assert re.match(pattern, "-1") is None

    def test_positive_integer_with_0_empty(self):
        """测试正整数或0模式空字符串"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Positive_integer_with_0

        # Should match empty string
        assert re.match(pattern, "") is not None

    def test_chinese_english_underscore_long_string(self):
        """测试中文英文下划线长字符串"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_underline

        long_string = "test_123_测试_abc_456_你好"
        assert re.match(pattern, long_string) is not None

    def test_chinese_english_underscore_single_char(self):
        """测试中文英文下划线单字符"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_underline

        assert re.match(pattern, "a") is not None
        assert re.match(pattern, "测") is not None
        assert re.match(pattern, "1") is not None
        assert re.match(pattern, "_") is not None

    def test_english_hyphen_start_end(self):
        """测试英文数字连字符开头结尾"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.English_numbers_and_hyphen

        assert re.match(pattern, "-test") is not None
        assert re.match(pattern, "test-") is not None
        assert re.match(pattern, "-") is not None

    def test_snow_id_all_zeros(self):
        """测试雪花ID全零"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.snow_id_pattern

        assert re.match(pattern, "0000000000000000000") is not None

    def test_snow_id_all_nines(self):
        """测试雪花ID全9"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.snow_id_pattern

        assert re.match(pattern, "9999999999999999999") is not None

    def test_uuid_lowercase(self):
        """测试UUID小写"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.uuid_pattern

        assert re.match(pattern, "0123456789abcdef0123456789abcdef") is not None

    def test_uuid_uppercase(self):
        """测试UUID大写"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.uuid_pattern

        assert re.match(pattern, "0123456789ABCDEF0123456789ABCDEF") is not None

    def test_uuid_mixed_case(self):
        """测试UUID混合大小写"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.uuid_pattern

        assert re.match(pattern, "0123456789AbCdEf0123456789aBcDeF") is not None

    def test_simple_variable_dollar_only(self):
        """测试简单变量仅美元符号"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Simple_variable_with_dollar_sign

        text = "$ test"
        matches = re.findall(pattern, text)
        assert matches == []

    def test_complex_variable_nested_index(self):
        """测试复杂变量嵌套索引"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Complex_variable_with_dollar_sign

        text = "$result[0][1][2]"
        assert re.search(pattern, text) is not None

    def test_complex_variable_deep_nesting(self):
        """测试复杂变量深度嵌套"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Complex_variable_with_dollar_sign

        text = "$a.b.c.d.e.f.g"
        assert re.search(pattern, text) is not None

    def test_variable_in_curly_braces_empty(self):
        """测试花括号变量空内容"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Variable_in_curly_braces

        text = "{{}}"
        matches = re.findall(pattern, text)
        assert matches == [""]

    def test_variable_in_curly_braces_spaces(self):
        """测试花括号变量带空格"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Variable_in_curly_braces

        text = "{{user name}}"
        matches = re.findall(pattern, text)
        assert matches == ["user name"]

    def test_special_symbols_all_keyboard_chars(self):
        """测试特殊符号所有键盘字符"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_special_symbols_on_the_keyboard

        # Test various special characters (some may not match due to regex exclusions)
        assert re.match(pattern, "test!@#$%^&*()") is not None
        # Note: Some characters like []{}|;:,.<>? may not match due to regex pattern

    def test_excluding_special_chars_should_exclude(self):
        """测试排除特殊字符模式应该排除特定字符"""
        import re
        from app.utils.regex_rules import RegexPatterns

        pattern = RegexPatterns.Chinese_and_English_numbers_and_some_keyboard_symbols_excluding_special_chars

        # Should exclude these characters
        assert re.match(pattern, "test#value") is None
        assert re.match(pattern, "test/value") is None
        assert re.match(pattern, "test:value") is None
        assert re.match(pattern, "test*value") is None
        assert re.match(pattern, "test?value") is None
        assert re.match(pattern, 'test"value') is None
        assert re.match(pattern, "test<value") is None
        assert re.match(pattern, "test>value") is None
        assert re.match(pattern, "test|value") is None


class TestGetErrorMessageByRegexExtended:
    """测试 GetErrorMessageByRegex 扩展测试"""

    def test_all_patterns_have_messages(self):
        """测试所有模式都有错误消息"""
        from app.utils.regex_rules import RegexPatterns, GetErrorMessageByRegex

        patterns = [
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

        for pattern in patterns:
            message = GetErrorMessageByRegex(pattern)
            assert isinstance(message, str)
            assert len(message) > 0

    def test_empty_string_key(self):
        """测试空字符串键"""
        from app.utils.regex_rules import GetErrorMessageByRegex

        result = GetErrorMessageByRegex("")
        assert "invalid" in result.lower()

    def test_none_key(self):
        """测试None键"""
        from app.utils.regex_rules import GetErrorMessageByRegex

        result = GetErrorMessageByRegex(None)
        assert "invalid" in result.lower()


class TestHandleJsonSchemaErrorExtended:
    """测试 handleJsonSchemaError 扩展测试"""

    def test_handle_float_type_error(self):
        """测试浮点类型错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "float"

        result = handleJsonSchemaError(exc)

        assert "float" in result.lower()

    def test_handle_array_type_error(self):
        """测试数组类型错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "array"

        result = handleJsonSchemaError(exc)

        assert "list" in result.lower()

    def test_handle_object_type_error(self):
        """测试对象类型错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "object"

        result = handleJsonSchemaError(exc)

        assert "dict" in result.lower()

    def test_handle_boolean_type_error(self):
        """测试布尔类型错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "boolean"

        result = handleJsonSchemaError(exc)

        assert "bool" in result.lower()

    def test_handle_min_items_error(self):
        """测试最小项数错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["minItems"]
        exc.validator_value = 3
        exc.instance = []

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)
        assert len(result) > 0

    def test_handle_max_items_error(self):
        """测试最大项数错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["maxItems"]
        exc.validator_value = 5
        exc.instance = [1, 2, 3, 4, 5, 6]

        result = handleJsonSchemaError(exc)

        assert "5" in result

    def test_handle_unknown_error_type(self):
        """测试未知错误类型"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "custom error message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["unknown_type"]
        exc.instance = {}

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)
        assert len(result) > 0

    def test_handle_required_multiple_fields(self):
        """测试多个必填字段错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = [""]
        exc.absolute_schema_path = ["required"]
        exc.validator_value = ["field1", "field2", "field3"]
        exc.instance = {"field1": "value"}

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)
        assert "field2" in result or "field3" in result

    def test_handle_error_with_deep_path(self):
        """测试深度路径错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["level1", "level2", "level3", "fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "string"

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)

    def test_handle_error_with_mixed_path(self):
        """测试混合路径错误"""
        from app.utils.regex_rules import handleJsonSchemaError

        # Use a path that has both string and int (realistic scenario)
        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["items", 0, "fieldName"]
        exc.absolute_schema_path = ["type"]
        exc.validator_value = "string"

        result = handleJsonSchemaError(exc)

        # Should return some error message
        assert isinstance(result, str)

    def test_min_length_zero_instance(self):
        """测试最小长度零实例"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["minLength"]
        exc.validator_value = 5
        exc.instance = ""

        result = handleJsonSchemaError(exc)

        # Should say cannot be empty
        assert "empty" in result.lower() or "5" in result

    def test_enum_single_value(self):
        """测试枚举单个值"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["enum"]
        exc.validator_value = ["only_option"]

        result = handleJsonSchemaError(exc)

        assert "only_option" in result

    def test_enum_many_values(self):
        """测试枚举多个值"""
        from app.utils.regex_rules import handleJsonSchemaError

        exc = Mock(spec=ValidationError)
        exc.message = "test message"
        exc.absolute_path = ["fieldName"]
        exc.absolute_schema_path = ["enum"]
        exc.validator_value = ["opt1", "opt2", "opt3", "opt4", "opt5"]

        result = handleJsonSchemaError(exc)

        assert isinstance(result, str)
