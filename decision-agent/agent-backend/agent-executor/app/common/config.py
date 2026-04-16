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
        log_exporter="http",
        log_load_interval=10,
        log_load_max_log=1000,
        http_log_feed_ingester_url="http://feed-ingester-service:13031/api/feed_ingester/v1/jobs/dip-o11y-log/events",
        log_level=Config.o11y.log_level,
    ),
    trace=TraceSetting(
        trace_enabled=Config.o11y.trace_enabled,
        trace_provider="otlp",
        trace_max_queue_size=512,
        max_export_batch_size=512,
        http_trace_feed_ingester_url="http://feed-ingester-service:13031/api/feed_ingester/v1/jobs/dip-o11y-trace/events",
        otlp_endpoint=Config.o11y.trace_endpoint,
        environment=Config.o11y.environment,
        sampling_rate=Config.o11y.trace_sampling_rate,
    ),
)


# 4. 初始化BuiltinIds配置
# 创建内置ID配置实例
BuiltinIds = BuiltinIdsConfig()
