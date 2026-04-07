"""单元测试 - config/config_v2/config_initializer 模块"""

from unittest.mock import patch


class TestConfigState:
    """测试 ConfigState 类"""

    def test_init_creates_none_attributes(self):
        """测试初始化创建所有属性为None"""
        from app.config.config_v2.config_initializer import ConfigState

        state = ConfigState()

        assert state.app is None
        assert state.rds is None
        assert state.redis is None
        assert state.graphdb is None
        assert state.services is None
        assert state.memory is None
        assert state.local_dev is None
        assert state.outer_llm is None
        assert state.features is None
        assert state.o11y is None
        assert state.dialog_logging is None


class TestConfigInitializer:
    """测试 ConfigInitializer 类"""

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    def test_initialize_with_empty_config(self, m_loader):
        """测试使用空配置初始化"""
        m_loader.load_config_file.return_value = {}

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            ConfigState,
        )

        state = ConfigState()
        ConfigInitializer.initialize(state)

        # All configs should be initialized (not None)
        assert state.app is not None
        assert state.rds is not None
        assert state.redis is not None
        assert state.graphdb is not None
        assert state.services is not None
        assert state.memory is not None
        assert state.local_dev is not None
        assert state.outer_llm is not None
        assert state.features is not None
        assert state.o11y is not None
        assert state.dialog_logging is not None

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    def test_initialize_with_config_data(self, m_loader):
        """测试使用配置数据初始化"""
        test_config = {
            "app": {"debug": True},
            "rds": {"port": "5432"},
            "redis": {"port": "6379"},
            "services": {"mf_model_api": {"host": "test"}},
        }
        m_loader.load_config_file.return_value = test_config

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            ConfigState,
        )

        state = ConfigState()
        ConfigInitializer.initialize(state)

        # Check that configs are initialized with the test data
        assert state.app.debug is True
        # Port conversion varies by config type
        assert str(state.rds.port) == "5432"  # RdsConfig converts to int
        assert state.redis.port == "6379"  # RedisConfig keeps as string
        assert state.services.mf_model_api.host == "test"

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    @patch("app.config.config_v2.config_initializer.sys")
    def test_post_process_app_config_in_development(self, m_sys, m_loader):
        """测试开发环境下的APP_ROOT后处理"""
        m_loader.load_config_file.return_value = {}
        m_sys.frozen = False
        m_sys._MEIPASS = "/tmp/pyinstaller"

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            AppConfig,
        )

        app_config = AppConfig()
        ConfigInitializer._post_process_app_config(app_config)

        # APP_ROOT should be set to parent directory
        assert app_config.app_root is not None
        assert "app" in app_config.app_root

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    @patch("app.config.config_v2.config_initializer.sys")
    def test_post_process_app_config_in_pyinstaller(self, m_sys, m_loader):
        """测试PyInstaller环境下的APP_ROOT后处理"""
        m_loader.load_config_file.return_value = {}
        m_sys.frozen = True
        m_sys._MEIPASS = "/tmp/pyinstaller"

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            AppConfig,
        )

        app_config = AppConfig()
        ConfigInitializer._post_process_app_config(app_config)

        # APP_ROOT should point to _MEIPASS/app
        assert app_config.app_root == "/tmp/pyinstaller/app"

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    def test_post_process_host_ip_ipv4(self, m_loader):
        """测试IPv4地址后处理"""
        m_loader.load_config_file.return_value = {}

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            AppConfig,
        )

        app_config = AppConfig(host_ip="192.168.1.1")
        ConfigInitializer._post_process_host_ip(app_config)

        assert app_config.host_ip == "0.0.0.0"

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    def test_post_process_host_ip_ipv6(self, m_loader):
        """测试IPv6地址后处理"""
        m_loader.load_config_file.return_value = {}

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            AppConfig,
        )

        app_config = AppConfig(host_ip="::1")
        ConfigInitializer._post_process_host_ip(app_config)

        assert app_config.host_ip == "::"

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    def test_post_process_host_ip_invalid(self, m_loader):
        """测试无效IP地址后处理"""
        m_loader.load_config_file.return_value = {}

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            AppConfig,
        )

        app_config = AppConfig(host_ip="invalid-ip")
        ConfigInitializer._post_process_host_ip(app_config)

        assert app_config.host_ip == "0.0.0.0"

    @patch("app.config.config_v2.config_initializer.ConfigLoader")
    def test_post_process_host_ip_localhost(self, m_loader):
        """测试localhost地址后处理"""
        m_loader.load_config_file.return_value = {}

        from app.config.config_v2.config_initializer import (
            ConfigInitializer,
            AppConfig,
        )

        app_config = AppConfig(host_ip="127.0.0.1")
        ConfigInitializer._post_process_host_ip(app_config)

        assert app_config.host_ip == "0.0.0.0"
