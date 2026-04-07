# -*- coding:utf-8 -*-
"""
TelemetrySDK 可用性检测模块
在模块加载时检测 SDK 是否已安装
"""

from app.infra.common.helper.env_helper import is_aaron_local_dev

TELEMETRY_SDK_AVAILABLE = False

# SDK 组件占位符
tlogging_module = None
SamplerLogger = None
log_resource = None
trace_resource = None
set_service_info = None
sdk_tracer = None

# 如果是 Aaron 本地开发环境，跳过 SDK 导入
if is_aaron_local_dev():
    print("Info: Aaron local dev mode, skipping TelemetrySDK import.")
else:
    try:
        import tlogging as tlogging_module  # noqa: F401
        from tlogging import SamplerLogger  # noqa: F401
        from exporter.resource.resource import (  # noqa: F401
            log_resource,  # noqa: F401
            trace_resource,  # noqa: F401
            set_service_info,  # noqa: F401
        )
        from exporter.ar_trace.trace_exporter import tracer as sdk_tracer  # noqa: F401

        TELEMETRY_SDK_AVAILABLE = True
    except ImportError:
        print(
            "Warning: TelemetrySDK is not installed. Observability features will be disabled."
        )
