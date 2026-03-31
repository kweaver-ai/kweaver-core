import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, MagicMock, mock_open
import yaml


@pytest.fixture(autouse=True)
def reset_env():
    """Reset environment before each test"""
    os.environ.clear()
    yield


class TestConfig:
    def test_config_initialization(self):
        """Test config initialization"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        config = Config()
        assert config.config is not None
        assert isinstance(config.config, dict)

    def test_get_with_valid_key(self):
        """Test getting config value with valid key"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        with patch.object(Config, "__init__", lambda self: None):
            config = Config()
            config.config = {"app": {"name": "Test App"}}

        result = config.get("app.name")
        assert result == "Test App"

    def test_get_with_invalid_key(self):
        """Test getting config value with invalid key"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        with patch.object(Config, "__init__", lambda self: None):
            config = Config()
            config.config = {"app": {"name": "Test App"}}

        result = config.get("invalid.key")
        assert result is None

    def test_get_with_default_value(self):
        """Test getting config value with default"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        with patch.object(Config, "__init__", lambda self: None):
            config = Config()
            config.config = {"app": {"name": "Test App"}}

        result = config.get("invalid.key", default="default")
        assert result == "default"

    def test_get_nested_key(self):
        """Test getting nested config value"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        with patch.object(Config, "__init__", lambda self: None):
            config = Config()
            config.config = {"app": {"name": "Test", "version": "1.0.0"}}

        result = config.get("app.version")
        assert result == "1.0.0"

    def test_get_db_config(self):
        """Test getting database configuration"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        with patch.object(Config, "__init__", lambda self: None):
            config = Config()
            config.config = {
                "db": {
                    "host": "localhost",
                    "port": 3306,
                    "user": "root",
                    "password": "password",
                    "database": "memory_db",
                }
            }

        db_config = config.get_db_config()

        assert db_config["host"] == "localhost"
        assert db_config["port"] == 3306

    def test_get_rerank_config(self):
        """Test getting rerank configuration"""
        from src.config.config import Config

        Config._instance = None
        Config._initialized = False

        with patch.object(Config, "__init__", lambda self: None):
            config = Config()
            config.config = {
                "rerank": {
                    "rerank_url": "http://test.com",
                    "rerank_model": "reranker",
                }
            }

        rerank_config = config.get_rerank_config()

        assert rerank_config.rerank_url == "http://test.com"
        assert rerank_config.rerank_model == "reranker"
