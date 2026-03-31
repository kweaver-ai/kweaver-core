# -*- coding:utf-8 -*-
"""
流式响应处理模块
负责流式响应的拦截、记录和文件存储
"""

import os
import time
from datetime import datetime

from app.common.config import Config
from app.common.stand_log import StandLogger
from app.common.struct_logger import struct_logger
from .streaming_rate_limiter import create_rate_limited_iterator

# 流式响应日志存储目录
STREAMING_RESPONSE_LOG_DIR = "log/streaming_responses"

# 获取请求日志记录器
request_logger = StandLogger.get_request_logger()


def _ensure_streaming_log_dir():
    """确保流式响应日志目录存在"""
    if not os.path.exists(STREAMING_RESPONSE_LOG_DIR):
        os.makedirs(STREAMING_RESPONSE_LOG_DIR, exist_ok=True)


def _get_streaming_log_file_path(request_id: str) -> str:
    """获取流式响应日志文件路径

    Args:
        request_id (str): 请求ID

    Returns:
        str: 日志文件的完整路径
    """
    # 使用时间戳和request_id创建文件名，避免冲突
    timestamp = datetime.now().strftime("%Y%m%d_%H%M%S")
    filename = f"{timestamp}_{request_id}.log"
    return os.path.join(STREAMING_RESPONSE_LOG_DIR, filename)


def _write_chunk_to_file(
    file_path: str, chunk_content: str, chunk_number: int, chunk_size: int
):
    """将流式响应块写入文件

    Args:
        file_path (str): 日志文件路径
        chunk_content (str): 块内容
        chunk_number (int): 块序号
        chunk_size (int): 块大小（字节数）
    """
    try:
        with open(file_path, "a", encoding="utf-8") as f:
            f.write(
                f"[{datetime.now().isoformat()}] Chunk {chunk_number} ({chunk_size} bytes):\n"
            )
            f.write(chunk_content)
            f.write("\n" + "=" * 50 + "\n")
    except Exception as e:
        struct_logger.console_logger.error(
            f"写入流式响应日志文件失败: {e}",
            file_path=file_path,
            chunk_number=chunk_number,
        )


def _write_stream_completion_info(file_path: str, chunks_count: int, total_bytes: int):
    """写入流式响应完成信息

    Args:
        file_path (str): 日志文件路径
        chunks_count (int): 总块数
        total_bytes (int): 总字节数
    """
    try:
        with open(file_path, "a", encoding="utf-8") as f:
            f.write(
                f"\n[{datetime.now().isoformat()}] Stream completed: {chunks_count} chunks, {total_bytes} bytes total\n"
            )
    except Exception as e:
        struct_logger.console_logger.error(
            f"写入流式响应完成信息失败: {e}", file_path=file_path
        )


def _create_streaming_wrapper(original_iterator, request_id):
    """创建流式响应包装器，用于拦截和记录每个数据块

    Args:
        original_iterator: 原始响应迭代器
        request_id (str): 请求ID

    Returns:
        async generator: 包装后的异步生成器
    """
    chunks_count = 0
    total_bytes = 0
    streaming_log_file = None

    async def wrapped_iterator():
        nonlocal chunks_count, total_bytes, streaming_log_file
        start_stream_time = time.time()

        # 仅在debug模式下创建流式响应日志文件
        if Config.is_debug_mode():
            _ensure_streaming_log_dir()
            streaming_log_file = _get_streaming_log_file_path(request_id)

            # 在日志中记录文件路径
            request_logger.info(f"🟢 [{request_id}] 流式响应文件: {streaming_log_file}")

        try:
            # 根据配置决定是否使用速率限制迭代器
            if Config.local_dev.enable_streaming_response_rate_limit:
                request_logger.info(
                    f"🚦 [{request_id}] 启用流式响应速率限制（前10个不限制）"
                )
                iterator = create_rate_limited_iterator(original_iterator)
            else:
                iterator = original_iterator

            # 统一处理所有chunk
            async for chunk in iterator:
                chunks_count += 1
                total_bytes += len(chunk)

                # 仅在debug模式下将块内容写入文件
                if Config.is_debug_mode() and streaming_log_file:
                    chunk_content = chunk.decode("utf-8", errors="ignore")
                    _write_chunk_to_file(
                        streaming_log_file, chunk_content, chunks_count, len(chunk)
                    )

                yield chunk

        finally:
            # 流式响应完成后记录总时间和详细信息
            stream_process_time = (time.time() - start_stream_time) * 1000
            request_logger.info(
                f"🟢 [{request_id}] 流式响应完成: 块数={chunks_count}, 总大小={total_bytes}B, 流处理时间={stream_process_time:.2f}ms"
            )

            # 如果有日志文件，记录文件完成信息
            if streaming_log_file and Config.is_debug_mode():
                _write_stream_completion_info(
                    streaming_log_file, chunks_count, total_bytes
                )

    return wrapped_iterator()


def handle_streaming_response(response, request_id, process_time):
    """处理流式响应，包装响应迭代器并记录初始信息

    Args:
        response: FastAPI响应对象
        request_id (str): 请求ID
        process_time (float): 初始处理时间（毫秒）

    Returns:
        Response: 包装后的响应对象
    """
    # 替换原始的响应迭代器为包装后的迭代器
    response.body_iterator = _create_streaming_wrapper(
        response.body_iterator, request_id
    )

    # 记录流式响应的初始信息
    response_info = {
        "status_code": response.status_code,
        "is_streaming": True,
        "initial_process_time_ms": f"{process_time:.2f}",
    }
    request_logger.info(f"🟢 [{request_id}] 流式请求处理中: {response_info}")

    return response
