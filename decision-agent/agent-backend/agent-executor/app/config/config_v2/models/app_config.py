"""
应用相关配置
"""

import logging
from dataclasses import dataclass


@dataclass
class AppConfig:
    """应用相关配置"""

    # 调试模式
    debug: bool = False

    # 服务监听配置
    host_ip: str = "0.0.0.0"
    port: int = 30778

    # API路径前缀
    host_prefix: str = "/api/agent-executor/v1"
    host_prefix_v2: str = "/api/agent-executor/v2"

    # 限流配置
    rps_limit: int = 100

    # 日志配置
    enable_system_log: str = "false"
    log_level: str = "info"

    # 应用根目录
    app_root: str = ""

    # dolphin agent 是否启用详细输出模式
    enable_dolphin_agent_verbose: bool = False

    # 是否记录 conversation-session/init 请求日志
    log_conversation_session_init: bool = False

    # 是否将异常日志写入文件
    is_write_exception_log_to_file: bool = False

    def get_stdlib_log_level(self) -> int:
        """获取标准库 logging 模块的日志级别"""
        return logging._nameToLevel.get(self.log_level.upper(), logging.INFO)

    @classmethod
    def from_dict(cls, data: dict) -> "AppConfig":
        """从字典创建配置对象"""
        return cls(
            debug=data.get("debug", False),
            host_ip=data.get("host_ip", "0.0.0.0"),
            port=int(data.get("port", 30778)),
            host_prefix=data.get("host_prefix", "/api/agent-executor/v1"),
            host_prefix_v2=data.get("host_prefix_v2", "/api/agent-executor/v2"),
            rps_limit=int(data.get("rps_limit", 100)),
            enable_system_log=str(data.get("enable_system_log", "false")).lower(),
            log_level=data.get("log_level", "info").lower(),
            enable_dolphin_agent_verbose=data.get(
                "enable_dolphin_agent_verbose", False
            ),
            log_conversation_session_init=data.get(
                "log_conversation_session_init", False
            ),
            is_write_exception_log_to_file=data.get(
                "is_write_exception_log_to_file", False
            ),
        )
