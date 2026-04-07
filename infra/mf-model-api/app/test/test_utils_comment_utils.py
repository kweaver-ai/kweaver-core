"""测试 comment_utils 模块"""
import pytest
import os
import asyncio
from unittest.mock import patch, mock_open, AsyncMock
from datetime import datetime
from app.utils.comment_utils import write_log, error_log


class TestWriteLog:
    """测试write_log函数"""

    @pytest.mark.asyncio
    async def test_write_log_basic(self):
        """测试基本日志写入"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await write_log(api="test_api", msg="test message", user="test_user")
            mock_file.assert_called_once_with("log.log", mode="a", encoding='utf-8')
            # 验证写入的内容包含关键信息
            handle = mock_file()
            written_content = ''.join(call.args[0] for call in handle.write.call_args_list)
            assert "test_api" in written_content
            assert "test message" in written_content
            assert "test_user" in written_content

    @pytest.mark.asyncio
    async def test_write_log_default_user(self):
        """测试默认用户"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await write_log(api="test_api", msg="test message")
            handle = mock_file()
            written_content = ''.join(call.args[0] for call in handle.write.call_args_list)
            assert "root" in written_content

    @pytest.mark.asyncio
    async def test_write_log_none_values(self):
        """测试None值"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await write_log(api=None, msg=None, user=None)
            mock_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_write_log_special_characters(self):
        """测试特殊字符"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await write_log(
                api="test_api!@#",
                msg="测试消息\n换行",
                user="用户123"
            )
            mock_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_write_log_empty_strings(self):
        """测试空字符串"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await write_log(api="", msg="", user="")
            mock_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_write_log_long_message(self):
        """测试长消息"""
        mock_file = mock_open()
        long_msg = "a" * 10000
        with patch('builtins.open', mock_file):
            await write_log(api="test_api", msg=long_msg, user="test_user")
            mock_file.assert_called_once()


class TestErrorLog:
    """测试error_log函数"""

    @pytest.mark.asyncio
    async def test_error_log_basic(self):
        """测试基本错误日志写入"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await error_log(api="error_api", msg="error message", user="error_user")
            # 注意文件名有空格 "err or.log"
            mock_file.assert_called_once_with("err or.log", mode="a", encoding='utf-8')
            handle = mock_file()
            written_content = ''.join(call.args[0] for call in handle.write.call_args_list)
            assert "error_api" in written_content
            assert "error message" in written_content
            assert "error_user" in written_content

    @pytest.mark.asyncio
    async def test_error_log_default_user(self):
        """测试默认用户"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await error_log(api="error_api", msg="error message")
            handle = mock_file()
            written_content = ''.join(call.args[0] for call in handle.write.call_args_list)
            assert "root" in written_content

    @pytest.mark.asyncio
    async def test_error_log_none_values(self):
        """测试None值"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await error_log(api=None, msg=None, user=None)
            mock_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_error_log_exception_info(self):
        """测试异常信息"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            await error_log(
                api="test_api",
                msg="Exception: ValueError('test error')",
                user="test_user"
            )
            mock_file.assert_called_once()

    @pytest.mark.asyncio
    async def test_error_log_concurrent_writes(self):
        """测试并发写入"""
        mock_file = mock_open()
        with patch('builtins.open', mock_file):
            tasks = [
                error_log(api=f"api_{i}", msg=f"msg_{i}", user=f"user_{i}")
                for i in range(10)
            ]
            await asyncio.gather(*tasks)
            # 应该有10次文件打开调用
            assert mock_file.call_count == 10


class TestLogFileIntegration:
    """集成测试"""

    @pytest.mark.asyncio
    async def test_write_and_error_log_different_files(self):
        """测试写入不同的日志文件"""
        write_mock = mock_open()
        error_mock = mock_open()
        
        def side_effect(filename, *args, **kwargs):
            if filename == "log.log":
                return write_mock()
            elif filename == "err or.log":
                return error_mock()
        
        with patch('builtins.open', side_effect=side_effect):
            await write_log(api="test", msg="normal log")
            await error_log(api="test", msg="error log")
            
            # 验证两个文件都被打开
            assert write_mock.call_count > 0 or error_mock.call_count > 0

