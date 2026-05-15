"""
应用配置

使用 Pydantic Settings 管理应用配置。
"""

from functools import lru_cache
from urllib.parse import quote_plus

from pydantic import Field, computed_field, field_validator
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    """应用配置类"""

    model_config = SettingsConfigDict(
        env_file=".env",
        env_file_encoding="utf-8",
        case_sensitive=False,
        extra="ignore",
    )

    # ============== 应用配置 ==============
    app_name: str = Field(default="Sandbox Control Plane")
    app_version: str = Field(default="2.1.0")
    environment: str = Field(default="development")
    debug: bool = Field(default=False)

    # ============== 服务器配置 ==============
    host: str = Field(default="0.0.0.0")
    port: int = Field(default=8000)
    workers: int = Field(default=4)

    # ============== 数据库配置 ==============
    database_url: str = Field(default="mysql+aiomysql://sandbox:password@localhost:3308/kweaver")
    db_pool_size: int = Field(default=20)
    db_max_overflow: int = Field(default=40)
    db_pool_recycle: int = Field(default=3600)

    # ============== RDS 数据库配置（从 depServices.rds 注入） ==============
    # 这些字段优先于 database_url，如果设置了则使用这些值构建数据库连接
    db_type: str | None = Field(default=None)  # 数据库类型，如 MYSQL, POSTGRESQL
    db_host: str | None = Field(default=None)  # 主库主机
    db_port: int | None = Field(default=None)  # 主库端口
    db_host_read: str | None = Field(default=None)  # 从库主机（读写分离）
    db_port_read: int | None = Field(default=None)  # 从库端口
    db_user: str | None = Field(default=None)  # 数据库用户
    db_password: str | None = Field(default=None)  # 数据库密码
    db_database: str | None = Field(default=None)  # 数据库名
    db_max_connections: int | None = Field(default=None)  # 最大连接数
    db_max_read_connections: int | None = Field(default=None)  # 最大读连接数
    db_charset: str | None = Field(default=None)  # 字符集
    db_timeout: int | None = Field(default=None)  # 连接超时
    db_read_timeout: int | None = Field(default=None)  # 读超时
    db_write_timeout: int | None = Field(default=None)  # 写超时

    @computed_field
    @property
    def effective_database_url(self) -> str:
        """
        计算有效的数据库 URL

        优先使用 RDS 环境变量构建数据库连接，如果未设置则使用 database_url
        """
        # 如果所有必需的 RDS 字段都已设置，则使用 RDS 配置
        if all(
            [
                self.db_type,
                self.db_host,
                self.db_port is not None,
                self.db_user,
                self.db_password is not None,
                self.db_database,
            ]
        ):
            # 构建 aiomysql 连接 URL
            # 格式: mysql+aiomysql://user:password@host:port/database
            user = quote_plus(self.db_user)
            password = quote_plus(self.db_password)
            host = self.db_host
            port = self.db_port
            database = self.db_database

            # 添加连接参数
            params = []
            if self.db_charset:
                params.append(f"charset={self.db_charset}")

            url = f"mysql+aiomysql://{user}:{password}@{host}:{port}/{database}"
            if params:
                url += "?" + "&".join(params)

            return url

        # 否则使用默认的 database_url
        return self.database_url

    # ============== S3 配置 ==============
    s3_bucket: str = Field(default="sandbox-workspace")
    s3_region: str = Field(default="us-east-1")
    s3_access_key_id: str = Field(default="")
    s3_secret_access_key: str = Field(default="")
    s3_endpoint_url: str = Field(default="")

    # ============== Docker 配置 ==============
    docker_host: str = Field(default="unix:///var/run/docker.sock")
    docker_tls_verify: bool = Field(default=False)
    docker_cert_path: str = Field(default="")

    # ============== Kubernetes 配置 ==============
    kubernetes_namespace: str = Field(default="sandbox-runtime")
    executor_image_pull_policy: str = Field(default="IfNotPresent")
    executor_image_pull_secrets: str = Field(default="")

    # ============== 执行配置 ==============
    default_timeout: int = Field(default=300)
    max_timeout: int = Field(default=3600)
    default_cpu: str = Field(default="1")
    default_memory: str = Field(default="512Mi")
    default_disk: str = Field(default="1Gi")
    default_template_id: str = Field(default="python-basic")
    default_multi_language_template_image: str = Field(default="")
    max_upload_file_size_mb: int = Field(default=100, ge=1)
    max_extracted_file_count: int = Field(default=10000, ge=1)
    max_extracted_total_size_mb: int = Field(default=512, ge=1)
    disable_bwrap: bool = Field(default=False)  # 禁用 Bubblewrap（本地开发环境）
    control_plane_url: str | None = Field(
        default=None
    )  # Control Plane URL for executor callback (None = auto-generate from namespace)

    # ============== 清理配置 ==============
    idle_threshold_minutes: int = Field(
        default=-1, ge=-1, description="空闲超时时间（分钟），-1 表示无限期（不清理空闲会话）"
    )
    max_lifetime_hours: int = Field(
        default=-1, ge=-1, description="最大生命周期（小时），-1 表示无限期"
    )
    cleanup_interval_seconds: int = Field(default=300, ge=1)
    creating_timeout_seconds: int = Field(
        default=300,
        ge=30,
        description="会话创建超时时间（秒），超过此时间的 creating 状态会话将被标记为 failed",
    )

    # ============== 重试配置 ==============
    max_retry_attempts: int = Field(default=3)
    retry_backoff_base: float = Field(default=1.0)
    retry_backoff_factor: float = Field(default=2.0)
    max_retry_backoff: float = Field(default=10.0)

    # ============== 预热池配置 ==============
    warm_pool_enabled: bool = Field(default=True)
    warm_pool_default_size: int = Field(default=10)
    warm_pool_min_size: int = Field(default=5)
    warm_pool_max_idle_time: int = Field(default=300)

    # ============== 健康检查配置 ==============
    health_check_interval_seconds: int = Field(default=10)
    heartbeat_interval_seconds: int = Field(default=5)
    heartbeat_timeout_seconds: int = Field(default=15)

    # ============== 日志配置 ==============
    log_level: str = Field(default="INFO")
    log_format: str = Field(default="text")  # json, text (default: text for human-readable)

    @field_validator("log_level")
    @classmethod
    def validate_log_level(cls, v: str) -> str:
        allowed = {"DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"}
        if v.upper() not in allowed:
            raise ValueError(f"log_level must be one of {allowed}")
        return v.upper()

    @field_validator("log_format")
    @classmethod
    def validate_log_format(cls, v: str) -> str:
        allowed = {"json", "text"}
        if v not in allowed:
            raise ValueError(f"log_format must be one of {allowed}")
        return v

    # ============== 监控配置 ==============
    metrics_enabled: bool = Field(default=True)
    metrics_port: int = Field(default=9090)

    # ============== 安全配置 ==============
    secret_key: str = Field(default="change-this-in-production")
    allowed_hosts: list[str] = Field(default=["*"])
    cors_origins: list[str] = Field(default=["http://localhost:3000"])

    # ============== 限流配置 ==============
    rate_limit_enabled: bool = Field(default=True)
    rate_limit_per_minute: int = Field(default=60)

    @field_validator("environment")
    @classmethod
    def validate_environment(cls, v: str) -> str:
        allowed = {"development", "staging", "production"}
        if v not in allowed:
            raise ValueError(f"environment must be one of {allowed}")
        return v


@lru_cache
def get_settings() -> Settings:
    """
    获取配置单例

    使用 lru_cache 确保配置只加载一次。
    """
    return Settings()
