import os
import sys
from app.config.builtin_ids_class import BuiltinIdsConfig
from app.config.config_v2 import ConfigClassV2
from app.utils.observability.observability_setting import (
    ServerInfo,
    ObservabilitySetting,
    LogSetting,
    TraceSetting,
)

## 1. 初始化 server info

server_info = ServerInfo(
    server_name=os.getenv("OTEL_SERVICE_NAME", "agent-executor"),
    server_version=os.getenv("OTEL_SERVICE_VERSION", "1.0.0"),
    language="python",
    python_version=sys.version,
)

## 2. 初始化配置
# 统一使用 otel 配置（只从 TRACE_ENABLE 和 TRACE_URL 读取）
_trace_enable_env = os.getenv("TRACE_ENABLE", "false").lower()
_trace_enabled = _trace_enable_env == "true"

observability_config = ObservabilitySetting(
    log=LogSetting(
        log_enabled=os.getenv("O11Y_LOG_ENABLED", "false") == "true",
        log_exporter=os.getenv("O11Y_LOG_EXPORTER", "http"),
        log_load_interval=int(os.getenv("O11Y_LOG_LOAD_INTERVAL", "10")),
        log_load_max_log=int(os.getenv("O11Y_LOG_LOAD_MAX_LOG", "1000")),
        http_log_feed_ingester_url=os.getenv(
            "O11Y_HTTP_LOG_FEED_INGESTER_URL",
            "http://feed-ingester-service:13031/api/feed_ingester/v1/jobs/dip-o11y-log/events",
        ),
    ),
    trace=TraceSetting(
        trace_enabled=_trace_enabled,
        trace_provider="otlp",
        trace_max_queue_size=int(os.getenv("O11Y_TRACE_MAX_QUEUE_SIZE", "512")),
        max_export_batch_size=int(os.getenv("O11Y_TRACE_MAX_EXPORT_BATCH_SIZE", "512")),
        http_trace_feed_ingester_url=os.getenv(
            "O11Y_HTTP_TRACE_FEED_INGESTER_URL",
            "http://feed-ingester-service:13031/api/feed_ingester/v1/jobs/dip-o11y-trace/events",
        ),
        otlp_endpoint=os.getenv("TRACE_URL", ""),
    ),
)


# 3. 初始化Config配置
Config = ConfigClassV2()


# 4. 初始化BuiltinIds配置
# 创建内置ID配置实例
BuiltinIds = BuiltinIdsConfig()
