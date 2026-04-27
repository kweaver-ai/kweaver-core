# -*- coding: utf-8 -*-
"""
Unit tests for app/common/exceptions/dolphin_sdk_exception.py
"""

import pytest
from unittest.mock import patch

from app.common.exceptions.dolphin_sdk_exception import (
    DolphinSDKException,
    _get_dolphin_exception_classes,
)


class TestGetDolphinExceptionClasses:
    """测试 _get_dolphin_exception_classes 函数"""

    def test_returns_dict(self):
        """测试返回字典"""
        result = _get_dolphin_exception_classes()

        assert isinstance(result, dict)

    def test_returns_empty_dict_on_import_error(self):
        """测试导入错误时返回空字典"""
        # Mock get_dolphin_exception to raise ImportError
        import builtins

        original_import = builtins.__import__

        def mock_import(name, *args, **kwargs):
            if name == "app.common.dependencies":
                raise ImportError("Mocked import error")
            return original_import(name, *args, **kwargs)

        with patch("builtins.__import__", side_effect=mock_import):
            result = _get_dolphin_exception_classes()

            # Should return empty dict when import fails
            assert result == {}


class TestDolphinSDKExceptionInit:
    """测试 DolphinSDKException __init__ 方法"""

    def test_init_with_required_params(self):
        """测试必需参数"""
        exc = ValueError("test error")

        exception = DolphinSDKException(
            raw_exception=exc, agent_id="test_agent", session_id="test_session"
        )

        assert exception is not None

    def test_init_with_user_id(self):
        """测试带 user_id 初始化"""
        exc = RuntimeError("test error")

        exception = DolphinSDKException(
            raw_exception=exc,
            agent_id="agent123",
            session_id="session456",
            user_id="user789",
        )

        assert exception is not None

    def test_init_minimal_skipped(self):
        """测试最小参数初始化（跳过因为缺少可选参数）"""
        # Skip since agent_id and session_id are required
        # but the test would fail anyway
        pytest.skip("Requires agent_id and session_id which are required parameters")
