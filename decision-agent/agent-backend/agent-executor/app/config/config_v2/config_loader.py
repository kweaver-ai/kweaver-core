"""
配置文件加载模块
负责处理yaml配置文件的路径查找和加载
"""

import os
from typing import Optional, Dict, Any
import yaml


class ConfigLoader:
    """配置文件加载器"""

    # 配置文件路径缓存
    _config_path: Optional[str] = None
    _config_data: Optional[Dict[str, Any]] = None

    @classmethod
    def get_config_path(cls) -> str:
        """
        获取配置文件路径
        优先级：环境变量 > 挂载路径 > 本地路径
        """
        if cls._config_path:
            return cls._config_path

        # 1. 从环境变量获取
        env_path = os.getenv("AGENT_EXECUTOR_CONFIG_PATH", "")
        if env_path and os.path.exists(env_path):
            cls._config_path = env_path
            return cls._config_path

        # 2. 默认挂载路径
        default_mount_path = "/sysvol/conf/"
        if os.path.exists(default_mount_path):
            cls._config_path = default_mount_path
            return cls._config_path

        # 3. 本地开发路径
        local_path = "./conf/"
        cls._config_path = local_path
        return cls._config_path

    @classmethod
    def load_config_file(cls) -> Dict[str, Any]:
        """加载yaml配置文件"""
        if cls._config_data is not None:
            return cls._config_data

        config_path = cls.get_config_path()
        config_file = os.path.join(config_path, "agent-executor.yaml")
        if not os.path.exists(config_file):
            print(
                f"Warning: Config file not found at {config_file}, using default values"
            )
            cls._config_data = {}
            return cls._config_data

        try:
            with open(config_file, "r", encoding="utf-8") as f:
                cls._config_data = yaml.safe_load(f) or {}
            print(f"[CONFIG] Loaded config from {config_file}")
        except Exception as e:
            print(f"[CONFIG] Error loading config file {config_file}: {e}")
            cls._config_data = {}

        return cls._config_data

    @classmethod
    def reset(cls):
        """重置配置缓存(主要用于测试)"""
        cls._config_path = None
        cls._config_data = None
