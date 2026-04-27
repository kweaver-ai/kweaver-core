"""
应用配置单元测试

测试 Settings 类的功能。
"""
import os
import pytest
from unittest.mock import patch
from pydantic import ValidationError

from src.infrastructure.config.settings import Settings, get_settings


def create_settings_with_defaults(**kwargs):
    """创建使用默认值的 Settings 实例（不读取 .env 文件）"""
    # 使用 _env_file=None 来跳过 .env 文件加载
    return Settings(_env_file=None, **kwargs)


class TestSettings:
    """应用配置测试"""

    def test_default_values(self):
        """测试默认值"""
        settings = create_settings_with_defaults()

        assert settings.app_name == "Sandbox Control Plane"
        assert settings.app_version == "2.1.0"
        assert settings.environment == "development"
        assert settings.debug is False
        assert settings.host == "0.0.0.0"
        assert settings.port == 8000
        assert settings.workers == 4

    def test_default_database_config(self):
        """测试默认数据库配置"""
        settings = create_settings_with_defaults()

        assert settings.database_url == "mysql+aiomysql://sandbox:password@localhost:3308/kweaver"
        assert settings.db_pool_size == 20
        assert settings.db_max_overflow == 40
        assert settings.db_pool_recycle == 3600

    def test_default_s3_config(self):
        """测试默认 S3 配置"""
        settings = create_settings_with_defaults()

        assert settings.s3_bucket == "sandbox-workspace"
        assert settings.s3_region == "us-east-1"
        assert settings.s3_access_key_id == ""
        assert settings.s3_secret_access_key == ""

    def test_default_docker_config(self):
        """测试默认 Docker 配置"""
        settings = create_settings_with_defaults()

        assert settings.docker_host == "unix:///var/run/docker.sock"
        assert settings.docker_tls_verify is False

    def test_default_kubernetes_config(self):
        """测试默认 Kubernetes 配置"""
        settings = create_settings_with_defaults()

        assert settings.kubernetes_namespace == "sandbox-runtime"

    def test_default_execution_config(self):
        """测试默认执行配置"""
        settings = create_settings_with_defaults()

        assert settings.default_timeout == 300
        assert settings.max_timeout == 3600
        assert settings.default_cpu == "1"
        assert settings.default_memory == "512Mi"
        assert settings.default_disk == "1Gi"
        assert settings.disable_bwrap is False

    def test_default_cleanup_config(self):
        """测试默认清理配置"""
        settings = create_settings_with_defaults()

        assert settings.idle_threshold_minutes == -1
        assert settings.max_lifetime_hours == -1
        assert settings.cleanup_interval_seconds == 300
        assert settings.creating_timeout_seconds == 300

    def test_default_retry_config(self):
        """测试默认重试配置"""
        settings = create_settings_with_defaults()

        assert settings.max_retry_attempts == 3
        assert settings.retry_backoff_base == 1.0
        assert settings.retry_backoff_factor == 2.0

    def test_default_health_check_config(self):
        """测试默认健康检查配置"""
        settings = create_settings_with_defaults()

        assert settings.health_check_interval_seconds == 10
        assert settings.heartbeat_interval_seconds == 5
        assert settings.heartbeat_timeout_seconds == 15

    def test_default_log_config(self):
        """测试默认日志配置"""
        settings = create_settings_with_defaults()

        assert settings.log_level == "INFO"
        assert settings.log_format == "text"

    def test_validate_log_level_valid(self):
        """测试验证有效的日志级别"""
        for level in ["DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"]:
            settings = create_settings_with_defaults(log_level=level)
            assert settings.log_level == level

    def test_validate_log_level_lowercase(self):
        """测试验证小写日志级别（自动转大写）"""
        settings = create_settings_with_defaults(log_level="debug")
        assert settings.log_level == "DEBUG"

    def test_validate_log_level_invalid(self):
        """测试验证无效的日志级别"""
        with pytest.raises(ValidationError, match="log_level must be one of"):
            create_settings_with_defaults(log_level="INVALID")

    def test_validate_log_format_valid(self):
        """测试验证有效的日志格式"""
        for fmt in ["json", "text"]:
            settings = create_settings_with_defaults(log_format=fmt)
            assert settings.log_format == fmt

    def test_validate_log_format_invalid(self):
        """测试验证无效的日志格式"""
        with pytest.raises(ValidationError, match="log_format must be one of"):
            create_settings_with_defaults(log_format="xml")

    def test_validate_environment_valid(self):
        """测试验证有效的环境"""
        for env in ["development", "staging", "production"]:
            settings = create_settings_with_defaults(environment=env)
            assert settings.environment == env

    def test_validate_environment_invalid(self):
        """测试验证无效的环境"""
        with pytest.raises(ValidationError, match="environment must be one of"):
            create_settings_with_defaults(environment="testing")

    def test_effective_database_url_default(self):
        """测试有效的数据库 URL（默认值）"""
        settings = create_settings_with_defaults()

        # When RDS fields are not set, use default database_url
        assert settings.effective_database_url == settings.database_url

    def test_effective_database_url_with_rds_config(self):
        """测试有效的数据库 URL（使用 RDS 配置）"""
        settings = create_settings_with_defaults(
            db_type="MYSQL",
            db_host="localhost",
            db_port=3306,
            db_user="root",
            db_password="password",
            db_database="sandbox",
        )

        # When RDS fields are set, use RDS config
        assert "mysql+aiomysql://" in settings.effective_database_url
        assert "root" in settings.effective_database_url
        assert "localhost:3306" in settings.effective_database_url
        assert "sandbox" in settings.effective_database_url

    def test_effective_database_url_with_charset(self):
        """测试有效的数据库 URL（带字符集）"""
        settings = create_settings_with_defaults(
            db_type="MYSQL",
            db_host="localhost",
            db_port=3306,
            db_user="root",
            db_password="password",
            db_database="sandbox",
            db_charset="utf8mb4",
        )

        assert "charset=utf8mb4" in settings.effective_database_url

    def test_effective_database_url_partial_rds_config(self):
        """测试有效的数据库 URL（部分 RDS 配置）"""
        settings = create_settings_with_defaults(
            db_host="localhost",
            db_port=3306,
            # Missing other required fields
        )

        # When RDS fields are partially set, use default database_url
        assert settings.effective_database_url == settings.database_url

    def test_default_security_config(self):
        """测试默认安全配置"""
        settings = create_settings_with_defaults()

        assert settings.secret_key == "change-this-in-production"
        assert settings.allowed_hosts == ["*"]
        assert settings.cors_origins == ["http://localhost:3000"]

    def test_default_rate_limit_config(self):
        """测试默认限流配置"""
        settings = create_settings_with_defaults()

        assert settings.rate_limit_enabled is True
        assert settings.rate_limit_per_minute == 60

    def test_default_metrics_config(self):
        """测试默认监控配置"""
        settings = create_settings_with_defaults()

        assert settings.metrics_enabled is True
        assert settings.metrics_port == 9090

    def test_custom_values(self):
        """测试自定义值"""
        settings = create_settings_with_defaults(
            app_name="Custom App",
            environment="production",
            port=9000,
            debug=True,
        )

        assert settings.app_name == "Custom App"
        assert settings.environment == "production"
        assert settings.port == 9000
        assert settings.debug is True

    def test_get_settings_returns_singleton(self):
        """测试获取配置单例"""
        # Clear cache first
        get_settings.cache_clear()

        settings1 = get_settings()
        settings2 = get_settings()

        assert settings1 is settings2


class TestSettingsValidation:
    """配置验证测试"""

    def test_cleanup_interval_validation(self):
        """测试清理间隔验证"""
        # Valid value
        settings = create_settings_with_defaults(cleanup_interval_seconds=60)
        assert settings.cleanup_interval_seconds == 60

        # Invalid value (below minimum)
        with pytest.raises(ValidationError):
            create_settings_with_defaults(cleanup_interval_seconds=0)

    def test_creating_timeout_validation(self):
        """测试创建超时验证"""
        # Valid value
        settings = create_settings_with_defaults(creating_timeout_seconds=60)
        assert settings.creating_timeout_seconds == 60

        # Invalid value (below minimum)
        with pytest.raises(ValidationError):
            create_settings_with_defaults(creating_timeout_seconds=10)

    def test_idle_threshold_validation(self):
        """测试空闲阈值验证"""
        # Valid value
        settings = create_settings_with_defaults(idle_threshold_minutes=30)
        assert settings.idle_threshold_minutes == 30

        # -1 is valid (disabled)
        settings = create_settings_with_defaults(idle_threshold_minutes=-1)
        assert settings.idle_threshold_minutes == -1

    def test_max_lifetime_validation(self):
        """测试最大生命周期验证"""
        # Valid value
        settings = create_settings_with_defaults(max_lifetime_hours=6)
        assert settings.max_lifetime_hours == 6

        # -1 is valid (disabled)
        settings = create_settings_with_defaults(max_lifetime_hours=-1)
        assert settings.max_lifetime_hours == -1
