"""单元测试 - config/config_v2/config_loader 模块"""

import os
import yaml
from unittest.mock import patch
import tempfile


class TestConfigLoader:
    """测试 ConfigLoader 类"""

    def setup_method(self):
        """每个测试前重置ConfigLoader"""
        from app.config.config_v2.config_loader import ConfigLoader

        ConfigLoader.reset()

    def test_get_config_path_from_env(self):
        """测试从环境变量获取配置路径"""
        from app.config.config_v2.config_loader import ConfigLoader

        # Mock环境变量和文件存在
        with patch.dict(os.environ, {"AGENT_EXECUTOR_CONFIG_PATH": "/test/path"}):
            with patch("os.path.exists") as m_exists:
                m_exists.side_effect = lambda x: x == "/test/path"

                result = ConfigLoader.get_config_path()

                assert result == "/test/path"

    def test_get_config_path_from_mount(self):
        """测试从默认挂载路径获取配置"""
        from app.config.config_v2.config_loader import ConfigLoader

        with patch("os.path.exists") as m_exists:
            m_exists.side_effect = lambda x: x == "/sysvol/conf/"

            result = ConfigLoader.get_config_path()

            assert result == "/sysvol/conf/"

    def test_get_config_path_local_fallback(self):
        """测试回退到本地开发路径"""
        from app.config.config_v2.config_loader import ConfigLoader

        with patch("os.path.exists") as m_exists:
            m_exists.return_value = False

            result = ConfigLoader.get_config_path()

            assert result == "./conf/"

    def test_get_config_path_caches_result(self):
        """测试配置路径被缓存"""
        from app.config.config_v2.config_loader import ConfigLoader

        with patch("os.path.exists") as m_exists:
            m_exists.return_value = False

            first_call = ConfigLoader.get_config_path()
            second_call = ConfigLoader.get_config_path()

            assert first_call == second_call
            assert first_call == "./conf/"

    def test_load_config_file_not_found(self, capsys):
        """测试配置文件不存在"""
        from app.config.config_v2.config_loader import ConfigLoader

        with patch("os.path.exists") as m_exists:
            m_exists.return_value = False

            ConfigLoader._config_path = "/test/path"
            result = ConfigLoader.load_config_file()

            assert result == {}
            captured = capsys.readouterr()
            assert "Warning: Config file not found" in captured.out

    def test_load_config_file_success(self):
        """测试成功加载配置文件"""
        from app.config.config_v2.config_loader import ConfigLoader

        # Create a temporary config file
        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".yaml", delete=False, dir="/tmp"
        ) as f:
            config_data = {"test": "value", "number": 123}
            yaml.dump(config_data, f)
            temp_file = f.name

        try:
            # Set config path to temp directory
            ConfigLoader._config_path = os.path.dirname(temp_file)
            config_filename = os.path.basename(temp_file)

            # Mock to return the temp file
            def mock_join(path, filename):
                if path == os.path.dirname(temp_file):
                    return temp_file
                return os.path.join(path, filename)

            with patch("os.path.join", side_effect=mock_join):
                with patch("os.path.exists") as m_exists:
                    m_exists.return_value = True

                    result = ConfigLoader.load_config_file()

                    assert result == config_data
        finally:
            os.unlink(temp_file)

    def test_load_config_file_caches_result(self):
        """测试配置数据被缓存"""
        from app.config.config_v2.config_loader import ConfigLoader

        with patch("os.path.exists") as m_exists:
            m_exists.return_value = False

            ConfigLoader._config_path = "/test/path"
            first_call = ConfigLoader.load_config_file()
            second_call = ConfigLoader.load_config_file()

            assert first_call == second_call
            assert first_call == {}

    def test_load_config_file_error_handling(self, capsys):
        """测试配置文件加载错误处理"""
        from app.config.config_v2.config_loader import ConfigLoader

        ConfigLoader._config_path = "/test/path"

        with patch("os.path.exists") as m_exists:
            m_exists.return_value = True

            # Mock open to raise an exception
            with patch("builtins.open", side_effect=IOError("Test error")):
                result = ConfigLoader.load_config_file()

                assert result == {}
                captured = capsys.readouterr()
                assert "Error loading config file" in captured.out

    def test_reset_clears_cache(self):
        """测试reset清除缓存"""
        from app.config.config_v2.config_loader import ConfigLoader

        # Set some values
        ConfigLoader._config_path = "/test/path"
        ConfigLoader._config_data = {"test": "value"}

        # Reset
        ConfigLoader.reset()

        # Check cache is cleared
        assert ConfigLoader._config_path is None
        assert ConfigLoader._config_data is None
