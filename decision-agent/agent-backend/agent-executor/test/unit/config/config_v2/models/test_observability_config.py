"""单元测试 - config/config_v2/models/observability_config 模块"""


class TestO11yConfig:
    """测试 O11yConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig()

        assert config.log_enabled is False
        assert config.trace_enabled is False

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig(log_enabled=True, trace_enabled=True)

        assert config.log_enabled is True
        assert config.trace_enabled is True

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.observability_config import O11yConfig

        data = {"log_enabled": True, "trace_enabled": True}

        config = O11yConfig.from_dict(data)

        assert config.log_enabled is True
        assert config.trace_enabled is True

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict({})

        assert config.log_enabled is False
        assert config.trace_enabled is False

    def test_from_dict_log_only(self):
        """测试只启用日志"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict({"log_enabled": True})

        assert config.log_enabled is True
        assert config.trace_enabled is False

    def test_from_dict_trace_only(self):
        """测试只启用追踪"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict({"trace_enabled": True})

        assert config.log_enabled is False
        assert config.trace_enabled is True


class TestDialogLoggingConfig:
    """测试 DialogLoggingConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig()

        assert config.enable_dialog_logging is True
        assert config.use_single_log_file is False
        assert config.single_profile_file_path == "./data/debug_logs/profile.log"
        assert config.single_trajectory_file_path == "./data/debug_logs/trajectory.log"

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig(
            enable_dialog_logging=False,
            use_single_log_file=True,
            single_profile_file_path="/tmp/profile.log",
            single_trajectory_file_path="/tmp/trajectory.log",
        )

        assert config.enable_dialog_logging is False
        assert config.use_single_log_file is True
        assert config.single_profile_file_path == "/tmp/profile.log"
        assert config.single_trajectory_file_path == "/tmp/trajectory.log"

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        data = {
            "enable_dialog_logging": False,
            "use_single_log_file": True,
            "single_profile_file_path": "/tmp/profile.log",
            "single_trajectory_file_path": "/tmp/trajectory.log",
        }

        config = DialogLoggingConfig.from_dict(data)

        assert config.enable_dialog_logging is False
        assert config.use_single_log_file is True
        assert config.single_profile_file_path == "/tmp/profile.log"
        assert config.single_trajectory_file_path == "/tmp/trajectory.log"

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig.from_dict({})

        assert config.enable_dialog_logging is True
        assert config.use_single_log_file is False
        assert config.single_profile_file_path == "./data/debug_logs/profile.log"
        assert config.single_trajectory_file_path == "./data/debug_logs/trajectory.log"

    def test_from_dict_enable_only(self):
        """测试只设置enable_dialog_logging"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig.from_dict({"enable_dialog_logging": False})

        assert config.enable_dialog_logging is False
        assert config.use_single_log_file is False
        assert config.single_profile_file_path == "./data/debug_logs/profile.log"

    def test_from_dict_use_single_file_only(self):
        """测试只设置use_single_log_file"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig.from_dict({"use_single_log_file": True})

        assert config.enable_dialog_logging is True
        assert config.use_single_log_file is True

    def test_from_dict_custom_profile_path(self):
        """测试自定义profile路径"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig.from_dict(
            {"single_profile_file_path": "/var/log/profile.log"}
        )

        assert config.single_profile_file_path == "/var/log/profile.log"

    def test_from_dict_custom_trajectory_path(self):
        """测试自定义trajectory路径"""
        from app.config.config_v2.models.observability_config import DialogLoggingConfig

        config = DialogLoggingConfig.from_dict(
            {"single_trajectory_file_path": "/var/log/trajectory.log"}
        )

        assert config.single_trajectory_file_path == "/var/log/trajectory.log"
