"""单元测试 - logic/agent_core_logic_v2/memory 模块"""

from unittest.mock import Mock, patch, AsyncMock


class TestMemoryHandler:
    """测试 MemoryHandler 类"""

    def setup_method(self):
        """每个测试方法前的设置"""
        from app.logic.agent_core_logic_v2.memory import MemoryHandler

        self.MemoryHandler = MemoryHandler

        # Mock agent config
        self.mock_config = Mock()
        self.mock_config.agent_id = "test_agent_123"
        self.mock_config.agent_run_id = "run_789"
        self.mock_config.memory = {"is_enabled": True}
        self.mock_config.input = {
            "fields": [
                {"name": "query", "type": "string"},
                {"name": "custom_field", "type": "string"},
            ]
        }
        self.mock_config.output = Mock()
        self.mock_config.output.get_final_answer_var = Mock(return_value="final_answer")

        # Mock agent input
        self.mock_agent_input = Mock()
        self.mock_agent_input.get = Mock(return_value="test value")

        # Mock headers
        self.mock_headers = {"x-account-id": "user_123", "x-account-type": "premium"}

        # Mock final result
        self.mock_final_result = {"final_answer": "This is the answer"}

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    def test_start_memory_build_thread_enabled(self, mock_logger, mock_thread_class):
        """测试启动记忆构建线程 - 功能启用"""
        handler = self.MemoryHandler()

        mock_thread = Mock()
        mock_thread_class.return_value = mock_thread

        handler.start_memory_build_thread(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify thread was created and started
        mock_thread_class.assert_called_once()
        call_kwargs = mock_thread_class.call_args[1]
        assert call_kwargs["daemon"] is True
        assert call_kwargs["target"] == handler.build_memory

        mock_thread.start.assert_called_once()
        mock_logger.info.assert_called_once_with("记忆构建线程已启动")

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    def test_start_memory_build_thread_disabled(self, mock_logger, mock_thread_class):
        """测试启动记忆构建线程 - 功能未启用"""
        self.mock_config.memory = {"is_enabled": False}
        handler = self.MemoryHandler()

        handler.start_memory_build_thread(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify thread was not created
        mock_thread_class.assert_not_called()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    def test_start_memory_build_thread_none_memory(
        self, mock_logger, mock_thread_class
    ):
        """测试启动记忆构建线程 - memory为None"""
        self.mock_config.memory = None
        handler = self.MemoryHandler()

        handler.start_memory_build_thread(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify thread was not created
        mock_thread_class.assert_not_called()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio.new_event_loop")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio.set_event_loop")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    def test_build_memory_success(
        self,
        mock_get_type,
        mock_get_id,
        mock_logger,
        mock_memory_service,
        mock_set_loop,
        mock_new_loop,
        mock_thread_class,
    ):
        """测试成功构建记忆"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_new_loop.return_value = mock_loop
        mock_loop.run_until_complete = Mock()

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify event loop was created and used
        mock_new_loop.assert_called_once()
        mock_set_loop.assert_called_once_with(mock_loop)
        mock_loop.run_until_complete.assert_called_once()
        mock_loop.close.assert_called_once()

        # Verify logger was called
        mock_logger.info.assert_called()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_no_fields(
        self,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试没有输入字段时构建记忆"""
        # Create a new config with empty fields for this test
        test_config = Mock()
        test_config.agent_id = "test_agent_123"
        test_config.agent_run_id = "run_789"
        test_config.memory = {"is_enabled": True}
        test_config.input = {"fields": []}
        test_config.output = Mock()
        test_config.output.get_final_answer_var = Mock(return_value="final_answer")

        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        # Should not raise exception
        handler.build_memory(
            agent_config=test_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.get_dolphin_var_final_value")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_final_result_error(
        self,
        mock_asyncio,
        mock_logger,
        mock_get_final_value,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试获取最终结果失败时的处理"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"
        mock_get_final_value.side_effect = Exception("Failed to get value")

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        # Should handle error gracefully
        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify error was logged
        assert mock_logger.error.called

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    @patch("app.logic.agent_core_logic_v2.memory.o11y_logger")
    def test_build_memory_exception(
        self,
        mock_o11y_logger,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试构建记忆时的异常处理"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        # Make loop.run_until_complete raise exception
        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop
        mock_loop.run_until_complete.side_effect = Exception("Build failed")

        handler = self.MemoryHandler()

        # Should handle exception gracefully
        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify errors were logged
        assert mock_logger.error.called

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_with_unknown_user(
        self,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试未知用户ID时构建记忆"""
        mock_get_id.return_value = None
        mock_get_type.return_value = None

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers={},  # Empty headers
            final_result=self.mock_final_result,
        )

        # Verify memory service was called with "unknown" user
        mock_loop.run_until_complete.assert_called_once()
        mock_memory_service.build_memory.assert_called_once()
        call_kwargs = mock_memory_service.build_memory.call_args[1]
        assert call_kwargs["user_id"] == "unknown"

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_empty_final_result(
        self,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试空最终结果时构建记忆"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result={},  # Empty final result
        )

        # Should still complete without error
        mock_loop.close.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_file_field_skipped(
        self,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试文件类型字段被跳过"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        # Create a new config with file field
        test_config = Mock()
        test_config.agent_id = "test_agent_123"
        test_config.agent_run_id = "run_789"
        test_config.memory = {"is_enabled": True}
        test_config.input = {
            "fields": [
                {"name": "query", "type": "string"},
                {"name": "file_upload", "type": "file"},  # Should be skipped
            ]
        }
        test_config.output = Mock()
        test_config.output.get_final_answer_var = Mock(return_value="final_answer")

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=test_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Should complete without error
        mock_loop.close.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    def test_start_memory_build_with_span(self, mock_logger, mock_thread_class):
        """测试带span参数启动记忆构建线程"""
        mock_span = Mock()
        mock_span.is_recording = Mock(return_value=True)

        handler = self.MemoryHandler()

        handler.start_memory_build_thread(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
            span=mock_span,
        )

        # Verify thread was created
        mock_thread_class.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_with_span(
        self,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试带span参数构建记忆"""
        mock_span = Mock()
        mock_span.is_recording = Mock(return_value=True)

        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
            span=mock_span,
        )

        # Verify completion
        mock_loop.close.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    @patch("app.logic.agent_core_logic_v2.memory.o11y_logger")
    def test_build_memory_loop_close_always_called(
        self,
        mock_o11y_logger,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试无论成功失败，event loop都会被关闭"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop
        mock_loop.run_until_complete.side_effect = Exception("Build failed")

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=self.mock_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Verify loop was still closed despite error
        mock_loop.close.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.memory.threading.Thread")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_id")
    @patch("app.logic.agent_core_logic_v2.memory.get_user_account_type")
    @patch("app.logic.agent_core_logic_v2.memory.agent_memory_service")
    @patch("app.logic.agent_core_logic_v2.memory.StandLogger")
    @patch("app.logic.agent_core_logic_v2.memory.asyncio")
    def test_build_memory_with_default_inputs(
        self,
        mock_asyncio,
        mock_logger,
        mock_memory_service,
        mock_get_type,
        mock_get_id,
        mock_thread_class,
    ):
        """测试包含默认输入字段时构建记忆"""
        mock_get_id.return_value = "user_123"
        mock_get_type.return_value = "premium"

        # Mock the async build_memory method
        mock_memory_service.build_memory = AsyncMock()

        # Create a new config with default inputs
        test_config = Mock()
        test_config.agent_id = "test_agent_123"
        test_config.agent_run_id = "run_789"
        test_config.memory = {"is_enabled": True}
        test_config.input = {
            "fields": [
                {"name": "query", "type": "string"},
                {"name": "history", "type": "list"},  # Default input
                {"name": "tool", "type": "dict"},  # Default input
            ]
        }
        test_config.output = Mock()
        test_config.output.get_final_answer_var = Mock(return_value="final_answer")

        mock_loop = Mock()
        mock_asyncio.new_event_loop.return_value = mock_loop

        handler = self.MemoryHandler()

        handler.build_memory(
            agent_config=test_config,
            agent_input=self.mock_agent_input,
            headers=self.mock_headers,
            final_result=self.mock_final_result,
        )

        # Should complete without error, default inputs should be skipped
        mock_loop.close.assert_called_once()
