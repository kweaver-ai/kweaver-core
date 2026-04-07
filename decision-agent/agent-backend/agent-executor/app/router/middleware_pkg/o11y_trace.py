# -*- coding:utf-8 -*-
"""
HTTP请求追踪中间件
负责OpenTelemetry分布式追踪
"""

from fastapi import Request, Response
from opentelemetry import trace
from opentelemetry.propagate import extract

from app.utils.observability.sdk_available import TELEMETRY_SDK_AVAILABLE
from app.utils.observability.observability_log import get_logger as o11y_logger


async def o11y_trace(request: Request, call_next) -> Response:
    """HTTP请求追踪中间件

    该中间件负责在HTTP请求处理过程中执行以下操作：
    1. 从请求头提取分布式追踪上下文
    2. 创建并管理OpenTelemetry span
    3. 记录请求处理的关键指标
    4. 处理请求过程中的异常追踪

    Args:
        request (Request): FastAPI请求对象，包含请求头、客户端信息等
        call_next: 请求处理链的下一个中间件或路由处理函数

    Returns:
        Response: 处理完成的响应对象

    Raises:
        Exception: 原始请求处理中发生的任何异常都会被重新抛出
    """
    # 从请求头中提取 trace 上下文
    headers = dict(request.headers)
    ctx = extract(headers)

    # 延迟导入 Config 避免循环依赖
    from app.common.config import Config

    # 如果 SDK 不可用或追踪未启用，直接调用下一个中间件
    if not TELEMETRY_SDK_AVAILABLE or not Config.o11y.trace_enabled:
        return await call_next(request)

    from exporter.ar_trace.trace_exporter import tracer

    # 创建 span 并设置为当前上下文
    with tracer.start_as_current_span(
        f"HTTP {request.method} {request.url.path}",
        context=ctx,
        kind=trace.SpanKind.SERVER,
        attributes={
            "http.method": request.method,
            "http.url": str(request.url),
            "http.client_ip": request.client.host,
        },
    ) as span:
        try:
            # 处理请求
            response = await call_next(request)

            # 添加响应状态码到 span
            span.set_attribute("http.status_code", response.status_code)
            logger = o11y_logger()
            if logger:
                logger.info(f"http status {response.status_code}")

            return response
        except Exception as e:
            # 错误处理
            span.set_attribute("http.status_code", 500)
            logger = o11y_logger()
            if logger:
                logger.error(f"Error: {e}")

            span.record_exception(e)
            raise
