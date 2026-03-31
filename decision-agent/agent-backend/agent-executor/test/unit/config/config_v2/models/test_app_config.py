"""单元测试 - config/config_v2/models/app_config 模块"""

import logging


class TestAppConfig:
    """测试 AppConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig()

        assert config.debug is False
        assert config.host_ip == "0.0.0.0"
        assert config.port == 30778
        assert config.host_prefix == "/api/agent-executor/v1"
        assert config.host_prefix_v2 == "/api/agent-executor/v2"
        assert config.rps_limit == 100
        assert config.enable_system_log == "false"
        assert config.log_level == "info"
        assert config.app_root == ""
        assert config.enable_dolphin_agent_verbose is False
        assert config.log_conversation_session_init is False
        assert config.is_write_exception_log_to_file is False

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig(
            debug=True,
            host_ip="127.0.0.1",
            port=8080,
            host_prefix="/api/v1",
            rps_limit=200,
        )

        assert config.debug is True
        assert config.host_ip == "127.0.0.1"
        assert config.port == 8080
        assert config.host_prefix == "/api/v1"
        assert config.rps_limit == 200

    def test_get_stdlib_log_level_info(self):
        """测试获取INFO日志级别"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig(log_level="info")

        level = config.get_stdlib_log_level()

        assert level == logging.INFO

    def test_get_stdlib_log_level_debug(self):
        """测试获取DEBUG日志级别"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig(log_level="debug")

        level = config.get_stdlib_log_level()

        assert level == logging.DEBUG

    def test_get_stdlib_log_level_warning(self):
        """测试获取WARNING日志级别"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig(log_level="warning")

        level = config.get_stdlib_log_level()

        assert level == logging.WARNING

    def test_get_stdlib_log_level_invalid(self):
        """测试无效日志级别默认为INFO"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig(log_level="invalid")

        level = config.get_stdlib_log_level()

        # Should default to INFO for invalid level
        assert level == logging.INFO

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.app_config import AppConfig

        data = {
            "debug": True,
            "host_ip": "127.0.0.1",
            "port": 8080,
            "host_prefix": "/api/v1",
            "host_prefix_v2": "/api/v2",
            "rps_limit": 200,
            "enable_system_log": "true",
            "log_level": "debug",
            "app_root": "/app",
            "enable_dolphin_agent_verbose": True,
            "log_conversation_session_init": True,
            "is_write_exception_log_to_file": True,
        }

        config = AppConfig.from_dict(data)

        assert config.debug is True
        assert config.host_ip == "127.0.0.1"
        assert config.port == 8080

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.app_config import AppConfig

        config = AppConfig.from_dict({})

        assert config.debug is False
        assert config.port == 30778

    def test_from_dict_port_string_conversion(self):
        """测试从字典创建时端口字符串转换"""
        from app.config.config_v2.models.app_config import AppConfig

        data = {"port": "8080"}
        config = AppConfig.from_dict(data)

        assert config.port == 8080

    def test_from_dict_log_level_case_insensitive(self):
        """测试日志级别不区分大小写"""
        from app.config.config_v2.models.app_config import AppConfig

        # Should convert to lowercase
        config = AppConfig.from_dict({"log_level": "DEBUG"})

        assert config.log_level == "debug"

    def test_enable_system_log_string_conversion(self):
        """测试enable_system_log字符串转换"""
        from app.config.config_v2.models.app_config import AppConfig

        # Should convert to lowercase
        config = AppConfig.from_dict({"enable_system_log": "TRUE"})

        assert config.enable_system_log == "true"


class TestLocalDevConfig:
    """测试 LocalDevConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.local_dev_config import LocalDevConfig

        config = LocalDevConfig()

        assert config.is_skip_pms_check is False
        assert config.is_use_outer_llm is False
        assert config.is_mock_get_llm_config_resp is False
        assert config.do_not_init_built_in_agent_and_tool is False
        assert config.dolphin_agent_output_variables is None
        assert config.enable_streaming_response_rate_limit is False
        assert config.is_show_config_on_boot is False

    def test_with_all_fields(self):
        """测试所有字段"""
        from app.config.config_v2.models.local_dev_config import LocalDevConfig

        config = LocalDevConfig(
            is_skip_pms_check=True,
            is_use_outer_llm=True,
            is_mock_get_llm_config_resp=True,
            do_not_init_built_in_agent_and_tool=True,
            dolphin_agent_output_variables=["var1", "var2"],
            enable_streaming_response_rate_limit=True,
            is_show_config_on_boot=True,
        )

        assert config.is_skip_pms_check is True
        assert config.is_use_outer_llm is True
        assert config.dolphin_agent_output_variables == ["var1", "var2"]

    def test_from_dict_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.local_dev_config import LocalDevConfig

        config = LocalDevConfig.from_dict({})

        assert config.is_skip_pms_check is False

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.local_dev_config import LocalDevConfig

        data = {
            "is_skip_pms_check": True,
            "is_use_outer_llm": True,
            "is_mock_get_llm_config_resp": True,
            "do_not_init_built_in_agent_and_tool": True,
            "dolphin_agent_output_variables": ["var1", "var2"],
            "enable_streaming_response_rate_limit": True,
            "is_show_config_on_boot": True,
        }

        config = LocalDevConfig.from_dict(data)

        assert config.is_skip_pms_check is True
        assert config.dolphin_agent_output_variables == ["var1", "var2"]


class TestFeaturesConfig:
    """测试 FeaturesConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.feature_config import FeaturesConfig

        config = FeaturesConfig()

        assert config.use_explore_block_v2 is True
        assert config.disable_dolphin_sdk_llm_cache is False
        assert config.enable_dolphin_agent_output_variables_ctrl is True
        assert config.is_skill_agent_need_progress is False

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.feature_config import FeaturesConfig

        config = FeaturesConfig(
            use_explore_block_v2=False,
            disable_dolphin_sdk_llm_cache=True,
            enable_dolphin_agent_output_variables_ctrl=False,
            is_skill_agent_need_progress=True,
        )

        assert config.use_explore_block_v2 is False
        assert config.disable_dolphin_sdk_llm_cache is True
        assert config.enable_dolphin_agent_output_variables_ctrl is False
        assert config.is_skill_agent_need_progress is True

    def test_from_dict_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.feature_config import FeaturesConfig

        config = FeaturesConfig.from_dict({})

        assert config.use_explore_block_v2 is True
        assert config.disable_dolphin_sdk_llm_cache is False

    def test_from_dict_with_custom_values(self):
        """测试从字典创建（自定义值）"""
        from app.config.config_v2.models.feature_config import FeaturesConfig

        data = {
            "use_explore_block_v2": False,
            "disable_dolphin_sdk_llm_cache": True,
            "enable_dolphin_agent_output_variables_ctrl": False,
            "is_skill_agent_need_progress": True,
        }

        config = FeaturesConfig.from_dict(data)

        assert config.use_explore_block_v2 is False
        assert config.disable_dolphin_sdk_llm_cache is True
