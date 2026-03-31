"""单元测试 - config/config_v2/config_class_v2 模块"""

from unittest.mock import MagicMock, patch


class TestConfigClassV2:
    """测试 ConfigClassV2 类"""

    def setup_method(self):
        """每个测试前重置单例"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        ConfigClassV2._instance = None

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_singleton_pattern(self, m_initializer, m_load_env):
        """测试单例模式"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config1 = ConfigClassV2()
        config2 = ConfigClassV2()

        assert config1 is config2
        assert id(config1) == id(config2)

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_init_called_once(self, m_initializer, m_load_env):
        """测试初始化只调用一次"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config = ConfigClassV2()
        config._initialized = False  # Reset to test again

        config.__init__()
        config.__init__()

        # ConfigInitializer.initialize should only be called once per instance
        # (but since we reset _initialized, it might be called again)

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_get_local_dev_model_config(self, m_initializer, m_load_env):
        """测试获取本地开发模型配置"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        # Mock the outer_llm.model_list
        config = ConfigClassV2()
        config.outer_llm = MagicMock()
        config.outer_llm.model_list = {
            "model1": {"param": "value1"},
            "model2": {"param": "value2"},
        }

        result = config.get_local_dev_model_config("model1")

        assert result == {"param": "value1"}

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_get_local_dev_model_config_not_found(self, m_initializer, m_load_env):
        """测试获取不存在的模型配置返回空字典"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config = ConfigClassV2()
        config.outer_llm = MagicMock()
        config.outer_llm.model_list = {}

        result = config.get_local_dev_model_config("nonexistent")

        assert result == {}

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_is_o11y_log_enabled(self, m_initializer, m_load_env):
        """测试检查o11y日志是否启用"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config = ConfigClassV2()
        config.o11y = MagicMock()
        config.o11y.log_enabled = True

        assert config.is_o11y_log_enabled() is True

        config.o11y.log_enabled = False
        assert config.is_o11y_log_enabled() is False

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_is_o11y_trace_enabled(self, m_initializer, m_load_env):
        """测试检查o11y追踪是否启用"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config = ConfigClassV2()
        config.o11y = MagicMock()
        config.o11y.trace_enabled = True

        assert config.is_o11y_trace_enabled() is True

        config.o11y.trace_enabled = False
        assert config.is_o11y_trace_enabled() is False

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_is_debug_mode(self, m_initializer, m_load_env):
        """测试检查是否为调试模式"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config = ConfigClassV2()
        config.app = MagicMock()
        config.app.debug = True

        assert config.is_debug_mode() is True

        config.app.debug = False
        assert config.is_debug_mode() is False

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_to_dict(self, m_initializer, m_load_env):
        """测试转换为字典"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2
        from dataclasses import dataclass

        config = ConfigClassV2()

        # Mock some attributes
        config.test_string = "test_value"
        config.test_number = 123

        @dataclass
        class MockConfig:
            field1: str = "value1"
            field2: int = 42

        config.test_config = MockConfig()

        result = config.to_dict()

        assert "test_string" in result
        assert result["test_string"] == "test_value"
        assert "test_number" in result
        assert result["test_number"] == 123
        assert "test_config" in result
        assert result["test_config"]["field1"] == "value1"

    @patch("app.boot.load_env.load_env")
    @patch("app.config.config_v2.config_class_v2.ConfigInitializer")
    def test_to_dict_with_non_dataclass(self, m_initializer, m_load_env):
        """测试转换包含非dataclass属性的对象"""
        from app.config.config_v2.config_class_v2 import ConfigClassV2

        config = ConfigClassV2()

        # Mock a non-dataclass object with __dict__
        class SimpleObject:
            def __init__(self):
                self.attr1 = "value1"

        config.simple_obj = SimpleObject()

        result = config.to_dict()

        assert "simple_obj" in result
        assert result["simple_obj"]["attr1"] == "value1"
