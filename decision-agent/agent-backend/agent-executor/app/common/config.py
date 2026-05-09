import sys
from app.config.builtin_ids_class import BuiltinIdsConfig
from app.config.config_v2 import ConfigClassV2
from app.utils.observability.observability_setting import (
    ServerInfo,
    ObservabilitySetting,
    LogSetting,
    TraceSetting,
)

## 1. 初始化Config配置
Config = ConfigClassV2()

## 2. 初始化 server info

server_info = ServerInfo(
    server_name=Config.o11y.service_name,
    server_version=Config.o11y.service_version,
    language="python",
    python_version=sys.version,
)

## 3. 初始化可观测性配置

observability_config = ObservabilitySetting(
    log=LogSetting(
        log_enabled=Config.o11y.log_enabled,
        log_level=Config.o11y.log_level,
    ),
    trace=TraceSetting(
        trace_enabled=Config.o11y.trace_enabled,
        trace_max_queue_size=512,
        otlp_endpoint=Config.o11y.trace_endpoint,
        environment=Config.o11y.environment,
        sampling_rate=Config.o11y.trace_sampling_rate,
    ),
)


# 4. 初始化BuiltinIds配置
# 创建内置ID配置实例
BuiltinIds = BuiltinIdsConfig()
