"""测试 str_util 模块"""
import pytest
from app.utils.str_util import generate_random_string, get_md5_key, has_common_substring


class TestGenerateRandomString:
    """测试generate_random_string函数"""

    def test_default_length(self):
        """测试默认长度"""
        result = generate_random_string()
        assert len(result) == 32
        assert result.isalnum()

    def test_custom_length(self):
        """测试自定义长度"""
        for length in [10, 20, 50, 100]:
            result = generate_random_string(length)
            assert len(result) == length
            assert result.isalnum()

    def test_zero_length(self):
        """测试零长度"""
        result = generate_random_string(0)
        assert len(result) == 0
        assert result == ""

    def test_randomness(self):
        """测试随机性"""
        results = set()
        for _ in range(100):
            results.add(generate_random_string(32))
        # 应该生成100个不同的字符串
        assert len(results) == 100

    def test_valid_characters(self):
        """测试字符有效性"""
        result = generate_random_string(1000)
        valid_chars = set("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
        assert all(c in valid_chars for c in result)


class TestGetMd5Key:
    """测试get_md5_key函数"""

    def test_simple_string(self):
        """测试简单字符串"""
        result = get_md5_key("hello")
        assert isinstance(result, str)
        assert len(result) == 32
        # MD5哈希应该是固定的
        assert result == get_md5_key("hello")

    def test_empty_string(self):
        """测试空字符串"""
        result = get_md5_key("")
        assert isinstance(result, str)
        assert len(result) == 32

    def test_chinese_string(self):
        """测试中文字符串"""
        result = get_md5_key("你好世界")
        assert isinstance(result, str)
        assert len(result) == 32

    def test_special_characters(self):
        """测试特殊字符"""
        result = get_md5_key("!@#$%^&*()")
        assert isinstance(result, str)
        assert len(result) == 32

    def test_long_string(self):
        """测试长字符串"""
        long_str = "a" * 10000
        result = get_md5_key(long_str)
        assert isinstance(result, str)
        assert len(result) == 32

    def test_consistency(self):
        """测试一致性"""
        input_str = "test_consistency"
        result1 = get_md5_key(input_str)
        result2 = get_md5_key(input_str)
        assert result1 == result2

    def test_different_inputs(self):
        """测试不同输入产生不同哈希"""
        result1 = get_md5_key("test1")
        result2 = get_md5_key("test2")
        assert result1 != result2


class TestHasCommonSubstring:
    """测试has_common_substring函数"""

    def test_has_common_suffix_prefix(self):
        """测试有公共后缀和前缀"""
        assert has_common_substring("hello", "hello world") is True
        assert has_common_substring("world", "hello") is False

    def test_exact_match(self):
        """测试完全匹配"""
        assert has_common_substring("test", "test") is True

    def test_no_common_substring(self):
        """测试无公共子串"""
        assert has_common_substring("abc", "def") is False

    def test_partial_overlap(self):
        """测试部分重叠"""
        assert has_common_substring("abc", "bcd") is True
        assert has_common_substring("hello", "orld") is True

    def test_empty_strings(self):
        """测试空字符串"""
        assert has_common_substring("", "") is False
        assert has_common_substring("test", "") is False
        assert has_common_substring("", "test") is False

    def test_single_character(self):
        """测试单字符"""
        assert has_common_substring("a", "a") is True
        assert has_common_substring("a", "b") is False

    def test_chinese_characters(self):
        """测试中文字符"""
        assert has_common_substring("你好", "好世界") is True
        assert has_common_substring("你好", "世界") is False

    def test_case_sensitive(self):
        """测试大小写敏感"""
        assert has_common_substring("Hello", "hello") is False
        assert has_common_substring("TEST", "EST") is True

    def test_longer_first_string(self):
        """测试第一个字符串更长时的后缀匹配"""
        # "hello world"的后缀"world"与"world"的前缀匹配
        assert has_common_substring("hello world", "world") is True
        # "testing"的后缀"ing"与"ing"的前缀匹配
        assert has_common_substring("testing", "ing") is True

    def test_multiple_character_overlap(self):
        """测试多字符重叠"""
        assert has_common_substring("abcde", "cdefg") is True
        assert has_common_substring("12345", "45678") is True

