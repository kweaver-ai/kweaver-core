"""单元测试 - config/config_v2/models/observability_config 模块"""

import pytest


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

        data = {
            "log_enabled": True,
            "trace_enabled": True,
            "trace_endpoint": "http://otelcol-contrib:4318",
        }

        config = O11yConfig.from_dict(data)

        assert config.log_enabled is True
        assert config.trace_enabled is True
        assert config.trace_endpoint == "http://otelcol-contrib:4318"

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict({})

        assert config.log_enabled is False
        assert config.trace_enabled is False

    def test_from_dict_log_only(self):
        """测试只启用日志时要求带 trace_endpoint"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict(
            {"log_enabled": True, "trace_endpoint": "http://otelcol-contrib:4318"}
        )

        assert config.log_enabled is True
        assert config.trace_enabled is False
        assert config.trace_endpoint == "http://otelcol-contrib:4318"

    def test_from_dict_trace_only(self):
        """测试只启用追踪时要求带 trace_endpoint"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict(
            {"trace_enabled": True, "trace_endpoint": "http://otelcol-contrib:4318"}
        )

        assert config.log_enabled is False
        assert config.trace_enabled is True
        assert config.trace_endpoint == "http://otelcol-contrib:4318"

    def test_from_dict_raises_when_log_enabled_without_trace_endpoint(self):
        """测试启用日志但未配置 trace_endpoint 时抛异常"""
        from app.config.config_v2.models.observability_config import O11yConfig

        with pytest.raises(
            ValueError,
            match="o11y.trace_endpoint is required when o11y.log_enabled or o11y.trace_enabled is true",
        ):
            O11yConfig.from_dict({"log_enabled": True, "trace_endpoint": "   "})

    def test_from_dict_raises_when_trace_enabled_without_trace_endpoint(self):
        """测试启用追踪但未配置 trace_endpoint 时抛异常"""
        from app.config.config_v2.models.observability_config import O11yConfig

        with pytest.raises(
            ValueError,
            match="o11y.trace_endpoint is required when o11y.log_enabled or o11y.trace_enabled is true",
        ):
            O11yConfig.from_dict({"trace_enabled": True, "trace_endpoint": ""})

    def test_from_dict_reads_otel_fields_from_yaml(self):
        """测试从 YAML 读取统一 OTel 字段"""
        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict(
            {
                "service_name": "agent-executor",
                "service_version": "1.0.1",
                "environment": "staging",
                "log_enabled": True,
                "log_level": "warning",
                "trace_enabled": True,
                "trace_endpoint": "http://otelcol-contrib:4318",
                "trace_sampling_rate": 0.25,
            }
        )

        assert config.service_name == "agent-executor"
        assert config.service_version == "1.0.1"
        assert config.environment == "staging"
        assert config.log_enabled is True
        assert config.log_level == "warning"
        assert config.trace_enabled is True
        assert config.dolphin_trace_enabled is True
        assert config.trace_endpoint == "http://otelcol-contrib:4318"
        assert config.dolphin_trace_url == "http://otelcol-contrib:4318"
        assert config.trace_sampling_rate == pytest.approx(0.25)

    def test_from_dict_does_not_read_trace_env_vars(self, monkeypatch):
        """测试不再从 TRACE 环境变量读取配置"""
        monkeypatch.setenv("TRACE_ENABLE", "true")
        monkeypatch.setenv("TRACE_URL", "http://env-only:4318")

        from app.config.config_v2.models.observability_config import O11yConfig

        config = O11yConfig.from_dict(
            {
                "trace_enabled": False,
                "trace_endpoint": "http://yaml-only:4318",
            }
        )

        assert config.trace_enabled is False
        assert config.dolphin_trace_enabled is False
        assert config.trace_endpoint == "http://yaml-only:4318"
        assert config.dolphin_trace_url == "http://yaml-only:4318"


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
