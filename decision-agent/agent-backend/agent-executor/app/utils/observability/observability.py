# -*- coding:utf-8 -*-

from app.utils.observability.observability_setting import (
    ServerInfo,
    ObservabilitySetting,
)
from app.utils.observability.observability_log import (
    init_log_provider,
    shutdown_log_provider,
)
from app.utils.observability.observability_trace import init_trace_provider
from app.common.stand_log import StandLogger


def init_observability(server_info: ServerInfo, setting: ObservabilitySetting):
    """初始化可观测性组件"""
    StandLogger.info_log(
        f"[Observability] Initializing observability: log_enabled={setting.log.log_enabled}, trace_enabled={setting.trace.trace_enabled}"
    )

    if setting.log.log_enabled:
        init_log_provider(server_info, setting.log)

    if setting.trace.trace_enabled:
        StandLogger.info_log(
            f"[Observability] Trace enabled, initializing trace provider: provider={setting.trace.trace_provider}, endpoint={setting.trace.otlp_endpoint}"
        )
        init_trace_provider(server_info, setting.trace)
    else:
        StandLogger.info_log(
            f"[Observability] Trace is disabled (trace_enabled={setting.trace.trace_enabled})"
        )

    # if setting.metric.metric_enabled:
    # pass
    # init_meter_provider(server_info, setting.metric)


def shutdown_observability() -> None:
    """关闭可观测性组件"""
    shutdown_log_provider()
