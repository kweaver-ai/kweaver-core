"""单元测试 - boot/built_in 模块"""

import pytest
from unittest.mock import patch, MagicMock
import sys


class TestHandleBuiltIn:
    """测试 handle_built_in 函数"""

    def setup_method(self):
        """在每次测试前设置 mock"""
        # Mock the problematic import before it gets loaded
        sys.modules["data_migrations"] = MagicMock()
        sys.modules["data_migrations.init"] = MagicMock()
        sys.modules["data_migrations.init.manage_built_in_agent_and_tool"] = MagicMock()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_initializes_when_enabled(self, mock_config, mock_init):
        """测试当 do_not_init_built_in_agent_and_tool 为 False 时进行初始化"""
        # Setup
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = False

        # Execute
        from app.boot.built_in import handle_built_in

        handle_built_in()

        # Verify
        mock_init.assert_called_once()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_skips_when_disabled(self, mock_config, mock_init):
        """测试当 do_not_init_built_in_agent_and_tool 为 True 时跳过初始化"""
        # Setup
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = True

        # Execute
        from app.boot.built_in import handle_built_in

        handle_built_in()

        # Verify
        mock_init.assert_not_called()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_multiple_calls_with_skip(self, mock_config, mock_init):
        """测试多次调用时的跳过行为"""
        # Setup
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = True

        # Execute multiple times
        from app.boot.built_in import handle_built_in

        handle_built_in()
        handle_built_in()
        handle_built_in()

        # Verify never called
        mock_init.assert_not_called()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_multiple_calls_with_init(self, mock_config, mock_init):
        """测试多次调用时的初始化行为"""
        # Setup
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = False

        # Execute multiple times
        from app.boot.built_in import handle_built_in

        handle_built_in()
        handle_built_in()

        # Verify called twice (once per call)
        assert mock_init.call_count == 2

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_returns_none(self, mock_config, mock_init):
        """测试函数返回 None"""
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = False

        from app.boot.built_in import handle_built_in

        result = handle_built_in()

        # Verify returns None
        assert result is None

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_init_exception_propagates(self, mock_config, mock_init):
        """测试 init_built_in_agent_and_tool 抛出异常时的行为"""
        # Setup
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = False
        mock_init.side_effect = RuntimeError("Initialization failed")

        # Execute and verify exception propagates
        from app.boot.built_in import handle_built_in

        with pytest.raises(RuntimeError, match="Initialization failed"):
            handle_built_in()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_with_config_access(self, mock_config, mock_init):
        """测试正确访问 Config.local_dev"""
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = False

        from app.boot.built_in import handle_built_in

        handle_built_in()

        # Verify Config.local_dev was accessed
        assert hasattr(mock_config, "local_dev")
        mock_init.assert_called_once()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_boolean_false_value(self, mock_config, mock_init):
        """测试布尔值为 False 时的处理"""
        # Python 中 False == 0, so test with actual False
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = False

        from app.boot.built_in import handle_built_in

        handle_built_in()

        mock_init.assert_called_once()

    @patch("app.boot.built_in.init_built_in_agent_and_tool")
    @patch("app.boot.built_in.Config")
    def test_handle_built_in_boolean_true_value(self, mock_config, mock_init):
        """测试布尔值为 True 时的处理"""
        # Python 中 True == 1, so test with actual True
        mock_config.local_dev.do_not_init_built_in_agent_and_tool = True

        from app.boot.built_in import handle_built_in

        handle_built_in()

        mock_init.assert_not_called()
