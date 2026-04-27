# -*- coding: utf-8 -*-
"""单元测试 - streaming_response_handler 模块"""

import pytest
from unittest.mock import MagicMock, patch, mock_open


class AsyncIterator:
    """异步迭代器辅助类"""

    def __init__(self, chunks):
        self.chunks = chunks

    def __aiter__(self):
        return self

    async def __anext__(self):
        if not self.chunks:
            raise StopAsyncIteration
        return self.chunks.pop(0)


class TestEnsureStreamingLogDir:
    """测试 _ensure_streaming_log_dir 函数"""

    def test_ensure_dir_not_exists(self):
        """测试目录不存在时创建"""
        with patch("os.path.exists", return_value=False):
            with patch("os.makedirs") as mock_makedirs:
                from app.router.middleware_pkg.streaming_response_handler import (
                    _ensure_streaming_log_dir,
                )

                _ensure_streaming_log_dir()

                mock_makedirs.assert_called_once()

    def test_ensure_dir_exists(self):
        """测试目录已存在"""
        with patch("os.path.exists", return_value=True):
            with patch("os.makedirs") as mock_makedirs:
                from app.router.middleware_pkg.streaming_response_handler import (
                    _ensure_streaming_log_dir,
                )

                _ensure_streaming_log_dir()

                mock_makedirs.assert_not_called()


class TestGetStreamingLogFilePath:
    """测试 _get_streaming_log_file_path 函数"""

    def test_get_log_file_path(self):
        """测试获取日志文件路径"""
        with patch(
            "app.router.middleware_pkg.streaming_response_handler.datetime"
        ) as mock_datetime:
            mock_datetime.now.return_value.strftime.return_value = "20240101_120000"

            from app.router.middleware_pkg.streaming_response_handler import (
                _get_streaming_log_file_path,
            )

            result = _get_streaming_log_file_path("test_request_id")

            assert "test_request_id" in result
            assert ".log" in result


class TestWriteChunkToFile:
    """测试 _write_chunk_to_file 函数"""

    def test_write_chunk_success(self):
        """测试成功写入块"""
        with patch("builtins.open", mock_open()) as mock_file:
            with patch(
                "app.router.middleware_pkg.streaming_response_handler.datetime"
            ) as mock_datetime:
                mock_datetime.now.return_value.isoformat.return_value = (
                    "2024-01-01T12:00:00"
                )

                from app.router.middleware_pkg.streaming_response_handler import (
                    _write_chunk_to_file,
                )

                _write_chunk_to_file("/test/path.log", "test content", 1, 12)

                mock_file.assert_called_once_with(
                    "/test/path.log", "a", encoding="utf-8"
                )

    def test_write_chunk_exception(self):
        """测试写入失败"""
        with patch("builtins.open", side_effect=Exception("Write error")):
            from app.router.middleware_pkg.streaming_response_handler import (
                _write_chunk_to_file,
            )

            # 不应该抛出异常
            _write_chunk_to_file("/test/path.log", "test content", 1, 12)


class TestWriteStreamCompletionInfo:
    """测试 _write_stream_completion_info 函数"""

    def test_write_completion_info_success(self):
        """测试成功写入完成信息"""
        with patch("builtins.open", mock_open()) as mock_file:
            with patch(
                "app.router.middleware_pkg.streaming_response_handler.datetime"
            ) as mock_datetime:
                mock_datetime.now.return_value.isoformat.return_value = (
                    "2024-01-01T12:00:00"
                )

                from app.router.middleware_pkg.streaming_response_handler import (
                    _write_stream_completion_info,
                )

                _write_stream_completion_info("/test/path.log", 10, 1024)

                mock_file.assert_called_once()

    def test_write_completion_info_exception(self):
        """测试写入失败"""
        with patch("builtins.open", side_effect=Exception("Write error")):
            from app.router.middleware_pkg.streaming_response_handler import (
                _write_stream_completion_info,
            )

            # 不应该抛出异常
            _write_stream_completion_info("/test/path.log", 10, 1024)


class TestCreateStreamingWrapper:
    """测试 _create_streaming_wrapper 函数"""

    @pytest.mark.asyncio
    async def test_create_wrapper_debug_mode(self):
        """测试debug模式下的包装器"""
        chunks = [b"chunk1", b"chunk2"]

        with patch(
            "app.router.middleware_pkg.streaming_response_handler.Config"
        ) as mock_config:
            mock_config.is_debug_mode.return_value = True
            mock_config.local_dev.enable_streaming_response_rate_limit = False

            with patch(
                "app.router.middleware_pkg.streaming_response_handler._ensure_streaming_log_dir"
            ):
                with patch(
                    "app.router.middleware_pkg.streaming_response_handler._get_streaming_log_file_path",
                    return_value="/test/path.log",
                ):
                    with patch(
                        "app.router.middleware_pkg.streaming_response_handler._write_chunk_to_file"
                    ):
                        with patch(
                            "app.router.middleware_pkg.streaming_response_handler._write_stream_completion_info"
                        ):
                            from app.router.middleware_pkg.streaming_response_handler import (
                                _create_streaming_wrapper,
                            )

                            wrapper = _create_streaming_wrapper(
                                AsyncIterator(chunks.copy()), "test_request_id"
                            )

                            results = []
                            async for chunk in wrapper:
                                results.append(chunk)

                            assert len(results) == 2

    @pytest.mark.asyncio
    async def test_create_wrapper_non_debug_mode(self):
        """测试非debug模式下的包装器"""
        chunks = [b"chunk1", b"chunk2"]

        with patch(
            "app.router.middleware_pkg.streaming_response_handler.Config"
        ) as mock_config:
            mock_config.is_debug_mode.return_value = False
            mock_config.local_dev.enable_streaming_response_rate_limit = False

            from app.router.middleware_pkg.streaming_response_handler import (
                _create_streaming_wrapper,
            )

            wrapper = _create_streaming_wrapper(
                AsyncIterator(chunks.copy()), "test_request_id"
            )

            results = []
            async for chunk in wrapper:
                results.append(chunk)

            assert len(results) == 2

    @pytest.mark.asyncio
    async def test_create_wrapper_with_rate_limit(self):
        """测试启用速率限制的包装器"""
        chunks = [b"chunk1", b"chunk2"]

        with patch(
            "app.router.middleware_pkg.streaming_response_handler.Config"
        ) as mock_config:
            mock_config.is_debug_mode.return_value = False
            mock_config.local_dev.enable_streaming_response_rate_limit = True

            with patch(
                "app.router.middleware_pkg.streaming_response_handler.create_rate_limited_iterator"
            ) as mock_create:
                mock_create.return_value = AsyncIterator(chunks.copy())

                from app.router.middleware_pkg.streaming_response_handler import (
                    _create_streaming_wrapper,
                )

                wrapper = _create_streaming_wrapper(
                    AsyncIterator([]), "test_request_id"
                )

                results = []
                async for chunk in wrapper:
                    results.append(chunk)

                mock_create.assert_called_once()


class TestHandleStreamingResponse:
    """测试 handle_streaming_response 函数"""

    def test_handle_streaming_response(self):
        """测试处理流式响应"""
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.body_iterator = AsyncIterator([b"chunk1"])

        with patch(
            "app.router.middleware_pkg.streaming_response_handler._create_streaming_wrapper"
        ) as mock_wrapper:
            mock_wrapper.return_value = AsyncIterator([b"chunk1"])

            from app.router.middleware_pkg.streaming_response_handler import (
                handle_streaming_response,
            )

            result = handle_streaming_response(mock_response, "test_id", 100.0)

            assert result is mock_response
            mock_wrapper.assert_called_once()
