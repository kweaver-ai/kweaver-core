"""
配置初始化模块
负责从配置数据创建和初始化各个dataclass模型实例
"""

import ipaddress
import os
import sys
from typing import Optional

from .config_loader import ConfigLoader
from .models import (
    AppConfig,
    RdsConfig,
    RedisConfig,
    GraphDBConfig,
    ServicesConfig,
    MemoryConfig,
    LocalDevConfig,
    OuterLLMConfig,
    FeaturesConfig,
    O11yConfig,
    DialogLoggingConfig,
    LLMMessageLoggingConfig,
)


class ConfigState:
    """配置状态管理器 - 持有所有配置模型实例"""

    def __init__(self):
        # 配置模型实例
        self.app: Optional[AppConfig] = None
        self.rds: Optional[RdsConfig] = None
        self.redis: Optional[RedisConfig] = None
        self.graphdb: Optional[GraphDBConfig] = None
        self.services: Optional[ServicesConfig] = None
        self.memory: Optional[MemoryConfig] = None
        self.local_dev: Optional[LocalDevConfig] = None
        self.outer_llm: Optional[OuterLLMConfig] = None
        self.features: Optional[FeaturesConfig] = None
        self.o11y: Optional[O11yConfig] = None
        self.dialog_logging: Optional[DialogLoggingConfig] = None
        self.llm_message_logging: Optional[LLMMessageLoggingConfig] = None


class ConfigInitializer:
    """配置初始化器 - 负责创建和初始化所有配置模型"""

    @staticmethod
    def initialize(state: ConfigState):
        """初始化配置，使用dataclass模型"""
        config = ConfigLoader.load_config_file()

        # 初始化各个配置模型
        state.app = AppConfig.from_dict(config.get("app", {}))
        state.rds = RdsConfig.from_dict(config.get("rds", {}))
        state.redis = RedisConfig.from_dict(config.get("redis", {}))
        state.graphdb = GraphDBConfig.from_dict(config.get("graphdb", {}))
        state.services = ServicesConfig.from_dict(config.get("services", {}))
        state.memory = MemoryConfig.from_dict(config.get("memory", {}))
        state.local_dev = LocalDevConfig.from_dict(config.get("local_dev", {}))
        state.outer_llm = OuterLLMConfig.from_dict(config.get("outer_llm", {}))
        state.features = FeaturesConfig.from_dict(config.get("features", {}))
        state.o11y = O11yConfig.from_dict(config.get("o11y", {}))
        state.dialog_logging = DialogLoggingConfig.from_dict(
            config.get("dialog_logging", {})
        )
        state.llm_message_logging = LLMMessageLoggingConfig.from_dict(
            config.get("llm_message_logging", {})
        )

        # 后处理：设置APP_ROOT
        ConfigInitializer._post_process_app_config(state.app)

        # 后处理：设置HOST_IP
        ConfigInitializer._post_process_host_ip(state.app)

    @staticmethod
    def _post_process_app_config(app_config: AppConfig):
        """后处理应用配置 - 设置APP_ROOT"""
        # 设置APP_ROOT为app目录的绝对路径
        # 支持 PyInstaller 打包后的环境
        if getattr(sys, "frozen", False):
            # PyInstaller 打包后的环境，使用 _MEIPASS
            # _MEIPASS 指向临时解压目录的根，数据文件在 app/ 子目录下
            app_config.app_root = os.path.join(sys._MEIPASS, "app")
        else:
            # 开发环境
            # 当前文件在 app/config/config_v2/config_initializer.py
            # 向上三级得到 app 目录
            app_config.app_root = os.path.dirname(
                os.path.dirname(os.path.dirname(__file__))
            )

        # 打印app_root信息（使用print避免循环依赖）
        print(f"[CONFIG] APP_ROOT initialized: {app_config.app_root}", flush=True)
        print(
            f"[CONFIG] PyInstaller frozen: {getattr(sys, 'frozen', False)}", flush=True
        )
        if getattr(sys, "frozen", False):
            print(f"[CONFIG] _MEIPASS: {sys._MEIPASS}", flush=True)

    @staticmethod
    def _post_process_host_ip(app_config: AppConfig):
        """后处理HOST_IP - 根据IP类型设置绑定地址"""
        try:
            if isinstance(
                ipaddress.ip_address(app_config.host_ip), ipaddress.IPv6Address
            ):
                app_config.host_ip = "::"
            else:
                app_config.host_ip = "0.0.0.0"
        except ValueError:
            app_config.host_ip = "0.0.0.0"
