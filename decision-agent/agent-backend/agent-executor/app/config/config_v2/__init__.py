"""
配置管理v2版本
使用yaml配置文件替代环境变量
使用dataclass模型组织配置项
"""

from .config_class_v2 import ConfigClassV2
from .models import (
    AppConfig,
    RdsConfig,
    RedisConfig,
    GraphDBConfig,
    OpenSearchConfig,
    ServicesConfig,
    MemoryConfig,
    LocalDevConfig,
    OuterLLMConfig,
    FeaturesConfig,
    O11yConfig,
    DialogLoggingConfig,
)

__all__ = [
    "ConfigClassV2",
    # 配置模型
    "AppConfig",
    "RdsConfig",
    "RedisConfig",
    "GraphDBConfig",
    "OpenSearchConfig",
    "ServicesConfig",
    "MemoryConfig",
    "LocalDevConfig",
    "OuterLLMConfig",
    "FeaturesConfig",
    "O11yConfig",
    "DialogLoggingConfig",
]
