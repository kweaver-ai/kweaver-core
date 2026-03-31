# -*- coding:utf-8 -*-
"""
请求日志中间件
记录HTTP请求和响应的详细信息，支持流式响应统计
"""

import time
import uuid
from fastapi import Request, Response

from app.common.config import Config
from app.common.stand_log import StandLogger
from app.common.struct_logger import struct_logger
from .streaming_response_handler import handle_streaming_response
from app.router.exception_handler.enhanced_unknown_handler import cache_request_body


request_logger = StandLogger.get_request_logger()


async def log_requests(request: Request, call_next) -> Response:
    """请求日志中间件

    记录请求和响应的详细信息，包括：
    - 请求基本信息（方法、路径、参数等）
    - 请求体内容
    - 响应状态码
    - 处理时间
    - 流式响应的块数和大小统计

    Args:
        request (Request): FastAPI请求对象
        call_next: 下一个中间件或路由处理函数

    Returns:
        Response: 处理完成的响应对象
    """
    # 1. 排除健康检查端点的日志记录
    if request.url.path in [
        Config.app.host_prefix + "/health/alive",
        Config.app.host_prefix + "/health/ready",
    ]:
        response = await call_next(request)
        return response

    # 1.1 根据配置决定是否排除 conversation-session/init 请求日志
    if not Config.app.log_conversation_session_init:
        if (
            request.url.path
            == Config.app.host_prefix + "/agent/conversation-session/init"
        ):
            response = await call_next(request)
            return response

    # 2. 记录请求信息
    start_time = time.time()

    request_id = request.headers.get("X-Request-ID") or str(uuid.uuid4())

    # 2.1. 构建请求信息
    request_info = {
        "request_id": request_id,
        "method": request.method,
        "path": request.url.path,
        "query_params": dict(request.query_params),
        "client": request.client.host if request.client else None,
        "headers": dict(request.headers),
    }

    # 2.2. 读取请求体（Starlette 会自动缓存，后续 handler 仍可读取）
    body_bytes = await request.body()
    request_body = body_bytes.decode("utf-8") if body_bytes else ""

    request_info["body"] = request_body

    # 2.2.1. 缓存请求体供异常处理器使用
    try:
        import json

        body_dict = json.loads(request_body) if request_body else None
        if body_dict:
            cache_request_body(request, body_dict)
    except (json.JSONDecodeError, Exception):
        # 非 JSON 格式的请求体，直接缓存字符串
        if request_body:
            cache_request_body(request, request_body)

    # 2.3. 记录请求信息
    StandLogger.console_request_log(request_info)

    # 2.4. 处理请求
    response = await call_next(request)

    # 3. 计算处理时间
    process_time = (time.time() - start_time) * 1000

    # 4. 检查是否为流式响应
    # 注意：在 BaseHTTPMiddleware 中，所有响应都会被包装成 _StreamingResponse（ai生成，不确定是否正确）
    # 需要通过响应头来判断原始响应是否为真正的流式响应
    is_streaming = False
    try:
        content_type = response.headers.get("content-type", "")
        # 只有当 content-type 明确表示流式传输时才认为是流式响应
        # text/event-stream: SSE (Server-Sent Events)
        # application/x-ndjson: 换行分隔的 JSON 流
        # application/octet-stream: 二进制流（某些情况下）
        if any(
            stream_type in content_type
            for stream_type in ["text/event-stream", "application/x-ndjson"]
        ):
            is_streaming = True
    except Exception as e:
        struct_logger.console_logger.warn("检查流式响应失败", error=str(e))

    # 调试日志：打印is_streaming和响应类型信息（可选，生产环境可删除）
    # struct_logger.console_logger.debug(
    #     "is_streaming value",
    #     is_streaming=is_streaming,
    #     response_type=str(type(response)),
    #     content_type=response.headers.get("content-type", "N/A"),
    # )

    # 5. 响应处理
    # 5.1. 流式响应处理
    if is_streaming:
        response = handle_streaming_response(response, request_id, process_time)
    # 5.2. 非流式响应处理
    else:
        response = await _handle_non_streaming_response(
            response, request_id, process_time
        )

    return response


async def _handle_non_streaming_response(response, request_id, process_time):
    """处理非流式响应，记录响应信息和响应体（调试模式）

    Args:
        response: FastAPI响应对象
        request_id (str): 请求ID
        process_time (float): 处理时间（毫秒）

    Returns:
        Response: 处理完成的响应对象
    """
    # 获取响应头
    response_headers = dict(response.headers)

    # 记录非流式响应的详细信息
    response_info = {
        "status_code": response.status_code,
        "is_streaming": False,
        "process_time_ms": f"{process_time:.2f}",
        "response_headers": response_headers,
    }

    # 记录响应体（仅在调试模式下）
    if Config.is_debug_mode():
        try:
            # 获取响应体内容
            response_body = b""
            async for chunk in response.body_iterator:
                response_body += chunk

            # 重新构造响应
            from starlette.responses import Response

            response = Response(
                content=response_body,
                status_code=response.status_code,
                headers=dict(response.headers),
                media_type=response.media_type,
            )

            response_info["response_body"] = response_body.decode(
                "utf-8", errors="ignore"
            )
        except Exception as e:
            struct_logger.console_logger.warn(f"Failed to capture response body: {e}")

    # 使用换行格式记录响应信息
    log_lines = [
        f"🟢 [{request_id}] 请求处理完成:",
        f"🟢 响应状态码: {response_info['status_code']}",
        f"🟢 处理时间: {response_info['process_time_ms']}ms",
        f"🟢 响应头: {response_info['response_headers']}",
    ]
    if "response_body" in response_info:
        log_lines.append(f"🟢 响应体: {response_info['response_body']}")

    request_logger.info("\n".join(log_lines))
    return response
