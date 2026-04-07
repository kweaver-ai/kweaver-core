# -*- coding: utf-8 -*-
"""单元测试 - enhanced_unknown_handler 模块"""

from unittest.mock import MagicMock, patch


class TestCacheRequestBody:
    """测试 cache_request_body 函数"""

    def test_cache_request_body(self):
        """测试缓存请求体"""
        request = MagicMock()
        request.state = MagicMock()

        from app.router.exception_handler.enhanced_unknown_handler import (
            cache_request_body,
        )

        cache_request_body(request, {"key": "value"})

        assert request.state.cached_body == {"key": "value"}


class TestExtractRequestInfo:
    """测试 _extract_request_info 函数"""

    def test_extract_request_info_basic(self):
        """测试基本请求信息提取"""
        request = MagicMock()
        request.method = "POST"
        request.url.path = "/api/test"
        request.url.query = ""
        request.url = MagicMock()
        request.url.path = "/api/test"
        request.url.query = ""
        request.url.__str__ = lambda self: "http://localhost/api/test"
        request.client = None
        request.headers = MagicMock()
        request.headers.items.return_value = []
        request.state = MagicMock()

        from app.router.exception_handler.enhanced_unknown_handler import (
            _extract_request_info,
        )

        with patch(
            "app.router.exception_handler.enhanced_unknown_handler.get_user_account_id",
            return_value=None,
        ):
            with patch(
                "app.router.exception_handler.enhanced_unknown_handler.get_user_account_type",
                return_value=None,
            ):
                with patch(
                    "app.router.exception_handler.enhanced_unknown_handler.get_biz_domain_id",
                    return_value=None,
                ):
                    result = _extract_request_info(request)

                    assert result["method"] == "POST"
                    assert result["path"] == "/api/test"

    def test_extract_request_info_with_client(self):
        """测试带客户端信息的请求"""
        request = MagicMock()
        request.method = "GET"
        request.url.path = "/api/test"
        request.url.query = "id=123"
        request.url.__str__ = lambda self: "http://localhost/api/test?id=123"
        request.client = MagicMock()
        request.client.host = "127.0.0.1"
        request.headers = MagicMock()
        request.headers.items.return_value = []
        request.state = MagicMock()

        from app.router.exception_handler.enhanced_unknown_handler import (
            _extract_request_info,
        )

        with patch(
            "app.router.exception_handler.enhanced_unknown_handler.get_user_account_id",
            return_value=None,
        ):
            with patch(
                "app.router.exception_handler.enhanced_unknown_handler.get_user_account_type",
                return_value=None,
            ):
                with patch(
                    "app.router.exception_handler.enhanced_unknown_handler.get_biz_domain_id",
                    return_value=None,
                ):
                    result = _extract_request_info(request)

                    assert result["client_ip"] == "127.0.0.1"
                    assert result["query_string"] == "id=123"

    def test_extract_request_info_with_headers(self):
        """测试带请求头的请求"""
        request = MagicMock()
        request.method = "POST"
        request.url.path = "/api/test"
        request.url.query = ""
        request.url.__str__ = lambda self: "http://localhost/api/test"
        request.client = None
        request.headers = MagicMock()
        request.headers.items.return_value = [
            ("Content-Type", "application/json"),
            ("Authorization", "Bearer token"),
        ]
        request.state = MagicMock()

        from app.router.exception_handler.enhanced_unknown_handler import (
            _extract_request_info,
        )

        with patch(
            "app.router.exception_handler.enhanced_unknown_handler.get_user_account_id",
            return_value=None,
        ):
            with patch(
                "app.router.exception_handler.enhanced_unknown_handler.get_user_account_type",
                return_value=None,
            ):
                with patch(
                    "app.router.exception_handler.enhanced_unknown_handler.get_biz_domain_id",
                    return_value=None,
                ):
                    result = _extract_request_info(request)

                    # Authorization 应该被过滤
                    assert result["headers"]["Authorization"] == "***REDACTED***"

    def test_extract_request_info_with_cached_body(self):
        """测试带缓存请求体的请求"""
        request = MagicMock()
        request.method = "POST"
        request.url.path = "/api/test"
        request.url.query = ""
        request.url.__str__ = lambda self: "http://localhost/api/test"
        request.client = None
        request.headers = MagicMock()
        request.headers.items.return_value = []
        request.state = MagicMock()
        request.state.cached_body = {"data": "test"}

        from app.router.exception_handler.enhanced_unknown_handler import (
            _extract_request_info,
        )

        with patch(
            "app.router.exception_handler.enhanced_unknown_handler.get_user_account_id",
            return_value=None,
        ):
            with patch(
                "app.router.exception_handler.enhanced_unknown_handler.get_user_account_type",
                return_value=None,
            ):
                with patch(
                    "app.router.exception_handler.enhanced_unknown_handler.get_biz_domain_id",
                    return_value=None,
                ):
                    result = _extract_request_info(request)

                    assert result["body"] == {"data": "test"}


class TestGetActualException:
    """测试 _get_actual_exception 函数"""

    def test_get_actual_exception_normal(self):
        """测试普通异常"""
        exc = ValueError("test error")

        from app.router.exception_handler.enhanced_unknown_handler import (
            _get_actual_exception,
        )

        result = _get_actual_exception(exc)

        assert result == exc

    def test_get_actual_exception_group(self):
        """测试异常组"""
        inner_exc = ValueError("inner error")
        exc = MagicMock()
        exc.__class__.__name__ = "ExceptionGroup"
        exc.exceptions = [inner_exc]

        from app.router.exception_handler.enhanced_unknown_handler import (
            _get_actual_exception,
        )

        result = _get_actual_exception(exc)

        assert result == inner_exc


class TestHandleEnhancedUnknownException:
    """测试 handle_enhanced_unknown_exception 函数"""

    def test_handle_enhanced_unknown_exception(self):
        """测试处理未知异常"""
        request = MagicMock()
        request.method = "POST"
        request.url.path = "/api/test"
        request.url.query = ""
        request.url.__str__ = lambda self: "http://localhost/api/test"
        request.client = None
        request.headers = MagicMock()
        request.headers.items.return_value = []
        request.state = MagicMock()

        exc = ValueError("test error")

        with patch(
            "app.router.exception_handler.enhanced_unknown_handler._extract_request_info",
            return_value={},
        ):
            with patch(
                "app.router.exception_handler.enhanced_unknown_handler.exception_logger"
            ):
                with patch(
                    "app.router.exception_handler.enhanced_unknown_handler._get_actual_exception",
                    return_value=exc,
                ):
                    with patch(
                        "app.router.exception_handler.enhanced_unknown_handler.GetRequestLangFunc",
                        return_value=lambda x: x,
                    ):
                        with patch(
                            "app.router.exception_handler.enhanced_unknown_handler.GetUnknowError",
                            return_value={"error": "unknown"},
                        ):
                            with patch("sys.exc_info", return_value=(None, None, None)):
                                from app.router.exception_handler.enhanced_unknown_handler import (
                                    handle_enhanced_unknown_exception,
                                )

                                result = handle_enhanced_unknown_exception(request, exc)

                                assert result.status_code == 500
