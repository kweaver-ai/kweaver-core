"""
配置类v2版本
参考agent-go-common-pkg/cconf/config.go的实现方式
通过yaml配置文件来管理配置，而不是环境变量
使用dataclass模型来组织配置项
"""

from .config_initializer import ConfigState, ConfigInitializer


class ConfigClassV2(ConfigState):
    """配置类v2版本 - 使用yaml配置文件和dataclass模型（单例模式）"""

    _instance = None  # 单例实例

    def __new__(cls):
        """单例模式：确保只有一个实例"""
        if cls._instance is None:
            cls._instance = super().__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        """初始化配置实例"""
        # 避免重复初始化
        if self._initialized:
            return

        # 先加载环境变量，确保配置初始化前环境变量已就绪
        from app.boot.load_env import load_env

        load_env()

        super().__init__()
        ConfigInitializer.initialize(self)
        self._initialized = True

    # ========== 公共方法 ==========

    def get_local_dev_model_config(self, model_name: str) -> dict:
        """根据模型名称获取模型配置"""
        return self.outer_llm.model_list.get(model_name, {})

    def is_o11y_log_enabled(self):
        """是否启用o11y日志"""
        return self.o11y.log_enabled

    def is_o11y_trace_enabled(self):
        """是否启用o11y追踪"""
        return self.o11y.trace_enabled

    def is_dolphin_trace_enabled(self):
        """是否启用dolphin trace"""
        return self.o11y.dolphin_trace_enabled

    def get_dolphin_trace_url(self):
        """获取dolphin trace上报URL"""
        return self.o11y.dolphin_trace_url

    def is_debug_mode(self):
        """是否启用调试模式"""
        return self.app.debug

    def to_dict(self) -> dict:
        """将配置对象转换为字典"""
        import dataclasses

        config_dict = {}
        for attr_name in dir(self):
            if not attr_name.startswith("_") and not callable(getattr(self, attr_name)):
                attr_value = getattr(self, attr_name)
                if hasattr(attr_value, "__dict__"):
                    try:
                        # 尝试使用dataclasses.asdict
                        config_dict[attr_name] = dataclasses.asdict(attr_value)
                    except (TypeError, ValueError):
                        # 如果不是dataclass，使用__dict__
                        config_dict[attr_name] = (
                            attr_value.__dict__
                            if hasattr(attr_value, "__dict__")
                            else str(attr_value)
                        )
                else:
                    config_dict[attr_name] = attr_value
        return config_dict
