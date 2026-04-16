# -*- coding:utf-8 -*-
"""
HTTP请求追踪中间件
负责OpenTelemetry分布式追踪
"""

from fastapi import Request, Response
from opentelemetry import trace
from opentelemetry.propagate import extract

from app.common.stand_log import StandLogger
from app.utils.observability.observability_log import get_logger as o11y_logger


def _get_request_tracer():
    """返回标准 OTel HTTP tracer。"""
    return trace.get_tracer("agent-executor.http")


def _should_skip_trace(path: str, config) -> bool:
    """健康检查接口不进入 OTel 上报，避免污染链路列表。"""
    if not path:
        return False

    configured_paths = {
        f"{config.app.host_prefix}/health/alive",
        f"{config.app.host_prefix}/health/ready",
    }
    return (
        path in configured_paths
        or path.endswith("/health/alive")
        or path.endswith("/health/ready")
    )


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

    request_path = request.url.path

    # 如果追踪未启用，或者当前请求显式排除上报，直接调用下一个中间件
    if not Config.o11y.trace_enabled or _should_skip_trace(request_path, Config):
        return await call_next(request)

    tracer = _get_request_tracer()

    # 创建 span 并设置为当前上下文
    with tracer.start_as_current_span(
        f"{request.method} {request_path}",
        context=ctx,
        kind=trace.SpanKind.SERVER,
        attributes={
            "http.method": request.method,
            "http.route": request_path,
            "http.url": str(request.url),
            "http.client_ip": request.client.host,
        },
    ) as span:
        try:
            # 处理请求
            response = await call_next(request)

            # 添加响应状态码到 span
            span.set_attribute("http.status_code", response.status_code)
            StandLogger.info(f"http status {response.status_code}")

            return response
        except Exception as e:
            # 错误处理
            span.set_attribute("http.status_code", 500)
            StandLogger.error(f"Error: {e}")
            logger = o11y_logger()
            if logger:
                logger.error(f"Error: {e}")

            span.record_exception(e)
            raise
