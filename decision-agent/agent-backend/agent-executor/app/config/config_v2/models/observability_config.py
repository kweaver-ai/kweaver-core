"""可观测性相关配置"""

from dataclasses import dataclass


O11Y_TRACE_ENDPOINT_REQUIRED_ERROR = (
    "o11y.trace_endpoint is required when o11y.log_enabled or o11y.trace_enabled is true"
)


@dataclass
class O11yConfig:
    """可观测性配置"""

    # 服务名称
    service_name: str = "agent-executor"

    # 服务版本
    service_version: str = "1.0.0"

    # 部署环境
    environment: str = "production"

    # 日志开关
    log_enabled: bool = False

    # 日志级别
    log_level: str = "info"

    # 追踪开关
    trace_enabled: bool = False

    # 统一 OTLP endpoint
    trace_endpoint: str = ""

    # 追踪采样率
    trace_sampling_rate: float = 1.0

    # Dolphin SDK trace开关
    dolphin_trace_enabled: bool = False

    # Dolphin SDK trace上报URL
    dolphin_trace_url: str = ""

    @classmethod
    def from_dict(cls, data: dict) -> "O11yConfig":
        """从字典创建配置对象。

        OTel 配置统一由 YAML 驱动，不再读取 TRACE_* / OTEL_* 环境变量。
        """
        trace_enabled = data.get("trace_enabled", False)
        log_enabled = data.get("log_enabled", False)
        trace_endpoint = str(data.get("trace_endpoint") or "").strip()

        if (log_enabled or trace_enabled) and not trace_endpoint:
            raise ValueError(O11Y_TRACE_ENDPOINT_REQUIRED_ERROR)

        return cls(
            service_name=data.get("service_name", "agent-executor"),
            service_version=data.get("service_version", "1.0.0"),
            environment=data.get("environment", "production"),
            log_enabled=log_enabled,
            log_level=data.get("log_level", "info").lower(),
            trace_enabled=trace_enabled,
            trace_endpoint=trace_endpoint,
            trace_sampling_rate=data.get("trace_sampling_rate", 1.0),
            dolphin_trace_enabled=trace_enabled,  # dolphin trace 与 o11y trace 使用相同开关
            dolphin_trace_url=trace_endpoint,
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


@dataclass
class LLMMessageLoggingConfig:
    """LLM message 日志配置"""

    enabled: bool = False
    log_dir: str = ".local/dolphin_llm_log"

    @classmethod
    def from_dict(cls, data: dict) -> "LLMMessageLoggingConfig":
        """从字典创建配置对象"""
        return cls(
            enabled=data.get("enabled", False),
            log_dir=data.get("log_dir", ".local/dolphin_llm_log"),
        )
