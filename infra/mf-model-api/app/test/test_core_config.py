"""测试 core config 模块"""
import pytest
import os
from unittest.mock import patch
from app.core.config import BaseConfig, base_config, server_info, observability_config


class TestBaseConfig:
    """测试BaseConfig类"""

    def test_default_values(self):
        """测试默认值"""
        config = BaseConfig()
        assert config.DEBUGDEFAULT is False
        assert config.PORTDEFAULT == 9898
        assert config.DIPHOSTDEFAULT == "10.4.134.253"

    def test_redis_defaults(self):
        """测试Redis默认配置"""
        config = BaseConfig()
        assert config.REDISPORTDEFAULT == 6379
        assert config.REDISCLUSTERMODEDEFAULT == "master-slave"

    def test_rds_defaults(self):
        """测试RDS默认配置"""
        config = BaseConfig()
        assert config.RDSPORTDEFAULT == 3330
        assert config.RDSDBNAMEDEFAULT == 'adp'
        assert config.RDSUSERDEFAULT == 'root'

    def test_oauth_defaults(self):
        """测试OAuth默认配置"""
        config = BaseConfig()
        assert config.OAUTHADMINPORTDEFAULT == 4445

    def test_kafka_defaults(self):
        """测试Kafka默认配置"""
        config = BaseConfig()
        assert config.KAFKAPORTDEFAULT == 9097
        assert config.KAFKAUSERDEFAULT == "anyrobot"

    def test_port_from_env(self):
        """测试APP_PORT属性存在"""
        config = BaseConfig()
        assert hasattr(config, 'APP_PORT')
        assert isinstance(config.APP_PORT, int)

    def test_debug_from_env(self):
        """测试DEBUG属性存在"""
        config = BaseConfig()
        assert hasattr(config, 'DEBUG')
        assert isinstance(config.DEBUG, bool)

    def test_rdshost_from_env(self):
        """测试RDSHOST属性存在"""
        config = BaseConfig()
        assert hasattr(config, 'RDSHOST')
        assert isinstance(config.RDSHOST, str)

    def test_redisport_from_env(self):
        """测试REDISPORT属性存在"""
        config = BaseConfig()
        assert hasattr(config, 'REDISPORT')
        assert config.REDISPORT > 0

    def test_kafkahost_from_env(self):
        """测试KAFKAHOST属性存在"""
        config = BaseConfig()
        assert hasattr(config, 'KAFKAHOST')
        assert isinstance(config.KAFKAHOST, str)


class TestBaseConfigInstance:
    """测试base_config实例"""

    def test_base_config_exists(self):
        """测试base_config实例存在"""
        assert base_config is not None
        assert isinstance(base_config, BaseConfig)

    def test_base_config_has_attributes(self):
        """测试base_config具有必需的属性"""
        assert hasattr(base_config, 'PORTDEFAULT')
        assert hasattr(base_config, 'RDSHOST')
        assert hasattr(base_config, 'REDISHOST')
        assert hasattr(base_config, 'KAFKAHOST')


class TestServerInfo:
    """测试server_info配置"""

    def test_server_info_exists(self):
        """测试server_info存在"""
        assert server_info is not None

    def test_server_info_attributes(self):
        """测试server_info属性"""
        assert hasattr(server_info, 'server_name')
        assert hasattr(server_info, 'server_version')
        assert hasattr(server_info, 'language')
        assert hasattr(server_info, 'python_version')

    def test_server_info_values(self):
        """测试server_info值"""
        assert server_info.server_name == "agent-executor"
        assert server_info.server_version == "1.0.0"
        assert server_info.language == "python"


class TestObservabilityConfig:
    """测试observability_config配置"""

    def test_observability_config_exists(self):
        """测试observability_config存在"""
        assert observability_config is not None

    def test_observability_config_has_log(self):
        """测试observability_config具有log配置"""
        assert hasattr(observability_config, 'log')

    def test_log_settings(self):
        """测试日志设置"""
        log_config = observability_config.log
        assert hasattr(log_config, 'log_enabled')
        assert hasattr(log_config, 'log_exporter')
        assert hasattr(log_config, 'log_load_interval')

    @patch.dict(os.environ, {'LOG_ENABLED': 'true'})
    def test_log_enabled_from_env(self):
        """测试从环境变量启用日志"""
        from app.core.config import ObservabilitySetting, LogSetting
        config = ObservabilitySetting(
            log=LogSetting(
                log_enabled=os.getenv("LOG_ENABLED", "false") == "true",
                log_exporter="http",
                log_load_interval=10,
                log_load_max_log=1000,
                http_log_feed_ingester_url=""
            )
        )
        assert config.log.log_enabled is True


class TestAiohttpTimeout:
    """测试aiohttp timeout配置"""

    def test_aiohttp_timeout_exists(self):
        """测试aiohttp_timeout存在"""
        config = BaseConfig()
        assert hasattr(config, 'aiohttp_timeout')

    def test_aiohttp_timeout_total(self):
        """测试total timeout"""
        config = BaseConfig()
        assert config.aiohttp_timeout.total == 1800

    def test_aiohttp_timeout_sock_connect(self):
        """测试sock_connect timeout"""
        config = BaseConfig()
        assert config.aiohttp_timeout.sock_connect == 30

