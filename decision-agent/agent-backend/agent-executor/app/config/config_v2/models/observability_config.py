"""
可观测性相关配置
"""

import os
from dataclasses import dataclass


@dataclass
class O11yConfig:
    """可观测性配置"""

    # 日志开关
    log_enabled: bool = False

    # 追踪开关
    trace_enabled: bool = False

    # Dolphin SDK trace开关
    dolphin_trace_enabled: bool = False

    # Dolphin SDK trace上报URL
    dolphin_trace_url: str = ""

    @classmethod
    def from_dict(cls, data: dict) -> "O11yConfig":
        """从字典创建配置对象

        统一使用 otel.trace.enabled 配置:
        - TRACE_ENABLE: 统一的 trace 开关（来自 otel.trace.enabled）
        - TRACE_URL: OTLP endpoint（来自 otel.otlp_endpoint）

        不再使用其它开关变量（O11Y_TRACE_ENABLED 等）
        """
        # 统一从 TRACE_ENABLE 环境变量读取（对应 otel.trace.enabled）
        trace_enable_env = os.getenv("TRACE_ENABLE", "").lower()
        trace_enabled = trace_enable_env == "true"

        trace_url = os.getenv("TRACE_URL", "")

        return cls(
            log_enabled=data.get("log_enabled", False),
            trace_enabled=trace_enabled,
            dolphin_trace_enabled=trace_enabled,  # dolphin trace 与 o11y trace 使用相同开关
            dolphin_trace_url=trace_url,
        )


@dataclass
class DialogLoggingConfig:
    """对话日志配置"""

    # 是否启用对话日志
    enable_dialog_logging: bool = True

    # 是否使用单一日志文件
    use_single_log_file: bool = False

    # profile日志文件路径
    single_profile_file_path: str = "./data/debug_logs/profile.log"

    # trajectory日志文件路径
    single_trajectory_file_path: str = "./data/debug_logs/trajectory.log"

    @classmethod
    def from_dict(cls, data: dict) -> "DialogLoggingConfig":
        """从字典创建配置对象"""
        return cls(
            enable_dialog_logging=data.get("enable_dialog_logging", True),
            use_single_log_file=data.get("use_single_log_file", False),
            single_profile_file_path=data.get(
                "single_profile_file_path", "./data/debug_logs/profile.log"
            ),
            single_trajectory_file_path=data.get(
                "single_trajectory_file_path", "./data/debug_logs/trajectory.log"
            ),
        )
