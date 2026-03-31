"""单元测试 - logic/agent_core_logic_v2/dialog_log 模块"""

import pytest
from unittest.mock import Mock, patch, mock_open


class TestDialogLogHandler:
    """测试 DialogLogHandler 类"""

    def setup_method(self):
        """每个测试方法前的设置"""
        # Mock agent
        self.mock_agent = Mock()

        # Mock config
        self.mock_config = Mock()
        self.mock_config.agent_id = "test_agent_123"
        self.mock_config.conversation_id = "conv_456"
        self.mock_config.agent_run_id = "run_789"

        # Mock headers
        self.mock_headers = {"x-account-id": "user_123", "x-account-type": "user"}

        # Import after setup to ensure mocks are in place
        from app.logic.agent_core_logic_v2.dialog_log import DialogLogHandler

        self.DialogLogHandler = DialogLogHandler

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    def test_init(self, mock_config_class):
        """测试 DialogLogHandler 初始化"""
        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        assert handler.agent == self.mock_agent
        assert handler.config == self.mock_config
        assert handler.headers == self.mock_headers
        assert handler.user_id == "user_123"

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    def test_init_unknown_user(self, mock_config_class):
        """测试初始化时用户ID未知"""
        # Provide empty strings for both old and new user ID keys to trigger "unknown" fallback
        headers_without_user = {"x-account-id": "", "x-user": ""}
        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=headers_without_user
        )

        assert handler.user_id == "unknown"

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    def test_save_dialog_logs_disabled(self, mock_config_class):
        """测试对话日志功能未启用时不保存"""
        mock_config_class.dialog_logging.enable_dialog_logging = False

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Should return early without calling any save methods
        handler.save_dialog_logs()

        # Verify no save methods were called
        assert not self.mock_agent.save_trajectory.called
        assert not self.mock_agent.get_profile.called

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    def test_save_to_single_file(self, mock_makedirs, mock_config_class):
        """测试保存到单一文件模式"""
        mock_config_class.dialog_logging.enable_dialog_logging = True
        mock_config_class.dialog_logging.use_single_log_file = True
        mock_config_class.dialog_logging.single_profile_file_path = (
            "./data/debug_logs/profile.log"
        )
        mock_config_class.dialog_logging.single_trajectory_file_path = (
            "./data/debug_logs/trajectory.log"
        )

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Mock the save methods
        handler._save_trajectory_to_single_file = Mock()
        handler._save_profile_to_single_file = Mock()

        handler.save_dialog_logs()

        # Verify both save methods were called
        handler._save_trajectory_to_single_file.assert_called_once()
        handler._save_profile_to_single_file.assert_called_once()

        # Verify directories were created
        assert mock_makedirs.call_count == 2

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    def test_save_to_multi_directories(self, mock_makedirs, mock_config_class):
        """测试保存到多目录模式"""
        mock_config_class.dialog_logging.enable_dialog_logging = True
        mock_config_class.dialog_logging.use_single_log_file = False

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Mock the save methods
        handler._save_trajectory_to_multi_directories = Mock(
            return_value="trajectory.jsonl"
        )
        handler._save_profile_to_multi_directories = Mock(return_value="profile.txt")

        handler.save_dialog_logs()

        # Verify both save methods were called
        handler._save_trajectory_to_multi_directories.assert_called_once()
        handler._save_profile_to_multi_directories.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    @patch("builtins.open", new_callable=mock_open, read_data="trajectory content")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.path.exists")
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_save_trajectory_to_single_file(
        self, mock_datetime, mock_exists, mock_file, mock_makedirs, mock_config_class
    ):
        """测试保存trajectory到单一文件"""
        mock_config_class.dialog_logging.single_trajectory_file_path = (
            "./data/debug_logs/trajectory.log"
        )
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        mock_exists.return_value = True

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        handler._save_trajectory_to_single_file()

        # Verify save_trajectory was called
        self.mock_agent.save_trajectory.assert_called_once()
        call_args = self.mock_agent.save_trajectory.call_args
        assert call_args[1]["agent_name"] == "test_agent_123"
        assert "dialog_" in call_args[1]["trajectory_path"]
        assert call_args[1]["force_save"] is True

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("builtins.open", new_callable=mock_open)
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_save_profile_to_single_file(
        self, mock_datetime, mock_file, mock_config_class
    ):
        """测试保存profile到单一文件"""
        mock_config_class.dialog_logging.single_profile_file_path = (
            "./data/debug_logs/profile.log"
        )
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        self.mock_agent.get_profile.return_value = "Profile content"

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        handler._save_profile_to_single_file()

        # Verify get_profile was called
        self.mock_agent.get_profile.assert_called_once_with(
            "Dolphin Runtime Profile - test_agent_123"
        )

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_save_trajectory_to_multi_directories(
        self, mock_datetime, mock_makedirs, mock_config_class
    ):
        """测试保存trajectory到多目录"""
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        result = handler._save_trajectory_to_multi_directories()

        # Verify save_trajectory was called
        self.mock_agent.save_trajectory.assert_called_once()
        call_args = self.mock_agent.save_trajectory.call_args
        assert call_args[1]["agent_name"] == "test_agent_123"
        assert result.endswith(".jsonl")

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    @patch("builtins.open", new_callable=mock_open)
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_save_profile_to_multi_directories(
        self, mock_datetime, mock_file, mock_makedirs, mock_config_class
    ):
        """测试保存profile到多目录"""
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        self.mock_agent.get_profile.return_value = "Profile content"

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        result = handler._save_profile_to_multi_directories()

        # Verify get_profile was called
        self.mock_agent.get_profile.assert_called_once()
        assert result.endswith(".txt")

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    @patch("builtins.open", new_callable=mock_open, read_data="trajectory content")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.path.exists")
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_trajectory_file_not_exists(
        self, mock_datetime, mock_exists, mock_file, mock_makedirs, mock_config_class
    ):
        """测试trajectory文件不存在时的处理"""
        mock_config_class.dialog_logging.single_trajectory_file_path = (
            "./data/debug_logs/trajectory.log"
        )
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        mock_exists.return_value = False

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Should not raise exception
        handler._save_trajectory_to_single_file()

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    def test_empty_config_values(self, mock_config_class):
        """测试空配置值"""
        empty_config = Mock()
        empty_config.agent_id = None
        empty_config.conversation_id = None
        empty_config.agent_run_id = None

        # Provide empty strings for both user ID keys to trigger "unknown" fallback
        handler = self.DialogLogHandler(
            agent=self.mock_agent,
            config=empty_config,
            headers={"x-account-id": "", "x-user": ""},
        )

        # Should handle None values gracefully
        assert handler.user_id == "unknown"

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_directory_creation_error(
        self, mock_datetime, mock_makedirs, mock_config_class
    ):
        """测试目录创建错误"""
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        mock_makedirs.side_effect = OSError("Permission denied")

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Should raise OSError
        with pytest.raises(OSError, match="Permission denied"):
            handler._save_trajectory_to_multi_directories()

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("builtins.open", new_callable=mock_open)
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_profile_get_error(self, mock_datetime, mock_file, mock_config_class):
        """测试获取profile时出错"""
        mock_config_class.dialog_logging.single_profile_file_path = (
            "./data/debug_logs/profile.log"
        )
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        self.mock_agent.get_profile.side_effect = Exception("Get profile failed")

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Should raise exception
        with pytest.raises(Exception, match="Get profile failed"):
            handler._save_profile_to_single_file()

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    @patch("builtins.open", new_callable=mock_open)
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.path.exists")
    @patch("app.logic.agent_core_logic_v2.dialog_log.datetime")
    def test_trajectory_write_error(
        self, mock_datetime, mock_exists, mock_file, mock_makedirs, mock_config_class
    ):
        """测试写入trajectory文件时出错"""
        mock_config_class.dialog_logging.single_trajectory_file_path = (
            "./data/debug_logs/trajectory.log"
        )
        mock_datetime.datetime.now.return_value.strftime.return_value = (
            "20250101_120000_000000"
        )
        mock_exists.return_value = True
        mock_file.side_effect = IOError("Write error")

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        # Should raise IOError
        with pytest.raises(IOError, match="Write error"):
            handler._save_trajectory_to_single_file()

    @patch("app.logic.agent_core_logic_v2.dialog_log.Config")
    @patch("app.logic.agent_core_logic_v2.dialog_log.os.makedirs")
    def test_multiple_save_calls(self, mock_makedirs, mock_config_class):
        """测试多次调用保存方法"""
        mock_config_class.dialog_logging.enable_dialog_logging = True
        mock_config_class.dialog_logging.use_single_log_file = False

        handler = self.DialogLogHandler(
            agent=self.mock_agent, config=self.mock_config, headers=self.mock_headers
        )

        handler._save_trajectory_to_multi_directories = Mock(
            return_value="trajectory.jsonl"
        )
        handler._save_profile_to_multi_directories = Mock(return_value="profile.txt")

        # Call save_dialog_logs multiple times
        handler.save_dialog_logs()
        handler.save_dialog_logs()

        # Verify methods were called twice
        assert handler._save_trajectory_to_multi_directories.call_count == 2
        assert handler._save_profile_to_multi_directories.call_count == 2
