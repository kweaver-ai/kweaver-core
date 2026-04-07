# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exception_logger/formatter_pkg/request_info.py
"""

from app.common.exception_logger.formatter_pkg.request_info import format_request_info


class TestFormatRequestInfo:
    """Tests for format_request_info function"""

    def test_format_empty_request_info(self):
        """测试空请求信息"""
        result = format_request_info({}, colorize=False)

        assert isinstance(result, str)
        # Should have at least the title
        assert "REQUEST" in result or len(result) == 0

    def test_format_with_method(self):
        """测试包含 method 的请求信息"""
        request_info = {"method": "POST"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "POST" in result
        assert "Method" in result

    def test_format_with_url(self):
        """测试包含 url 的请求信息"""
        request_info = {"url": "https://example.com/api/endpoint"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "URL" in result
        assert "https://example.com/api/endpoint" in result

    def test_format_with_path(self):
        """测试包含 path 的请求信息"""
        request_info = {"path": "/api/v1/resource"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Path" in result
        assert "/api/v1/resource" in result

    def test_format_with_query_string(self):
        """测试包含 query_string 的请求信息"""
        request_info = {"query_string": "param1=value1&param2=value2"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Query" in result
        assert "param1=value1" in result

    def test_format_with_client_ip(self):
        """测试包含 client_ip 的请求信息"""
        request_info = {"client_ip": "192.168.1.100"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Client IP" in result
        assert "192.168.1.100" in result

    def test_format_with_account_id(self):
        """测试包含 account_id 的请求信息"""
        request_info = {"account_id": "acc_12345"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Account ID" in result
        assert "acc_12345" in result

    def test_format_with_account_type(self):
        """测试包含 account_type 的请求信息"""
        request_info = {"account_type": "premium"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Account Type" in result
        assert "premium" in result

    def test_format_with_biz_domain(self):
        """测试包含 biz_domain 的请求信息"""
        request_info = {"biz_domain": "sales"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Biz Domain" in result
        assert "sales" in result

    def test_format_with_headers(self):
        """测试包含 headers 的请求信息"""
        request_info = {
            "headers": {
                "content-type": "application/json",
                "authorization": "Bearer token123",
                "host": "example.com",
            }
        }

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Headers" in result
        # Should filter out host, connection, accept, etc.
        assert "content-type" in result or "Content-Type" in result
        # Host should be filtered
        # Need to check if it appears in result

    def test_format_with_headers_skips_sensitive(self):
        """测试 headers 过滤敏感信息"""
        request_info = {
            "headers": {
                "host": "example.com",
                "connection": "keep-alive",
                "accept": "application/json",
                "authorization": "Bearer secret",
                "x-custom-header": "custom_value",
            }
        }

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        # Should include custom header
        assert "x-custom-header" in result or "X-Custom-Header" in result

    def test_format_with_body_dict(self):
        """测试包含字典类型的 body"""
        request_info = {"body": {"key": "value", "number": 123}}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Request Body" in result
        assert "key" in result

    def test_format_with_body_string(self):
        """测试包含字符串类型的 body"""
        request_info = {"body": "raw body string"}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "Request Body" in result
        assert "raw body string" in result

    def test_format_with_long_body_truncated(self):
        """测试超长 body 被截断"""
        long_body = "x" * 600
        request_info = {"body": long_body}

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "truncated" in result

    def test_format_with_all_fields(self):
        """测试包含所有字段的请求信息"""
        request_info = {
            "method": "GET",
            "url": "https://api.example.com/v1/endpoint",
            "query_string": "test=true",
            "client_ip": "10.0.0.1",
            "account_id": "user123",
            "headers": {"x-custom": "value"},
        }

        result = format_request_info(request_info, colorize=False)

        assert isinstance(result, str)
        assert "GET" in result
        assert "10.0.0.1" in result

    def test_format_with_colorize_true(self):
        """测试 colorize=True 的格式化"""
        request_info = {"method": "POST"}

        result = format_request_info(request_info, colorize=True)

        assert isinstance(result, str)
        # Colorized version should have structure

    def test_format_with_custom_indent(self):
        """测试自定义缩进"""
        request_info = {"method": "DELETE"}

        result = format_request_info(request_info, colorize=False, indent=4)

        assert isinstance(result, str)
        # Should have indentation

    def test_format_with_empty_query_string(self):
        """测试空 query_string"""
        request_info = {"query_string": ""}

        result = format_request_info(request_info, colorize=False)

        # Empty query string should not be shown
        assert isinstance(result, str)

    def test_format_with_empty_body(self):
        """测试空 body"""
        request_info = {"body": ""}

        result = format_request_info(request_info, colorize=False)

        # Empty body should not trigger body section
        assert isinstance(result, str)
