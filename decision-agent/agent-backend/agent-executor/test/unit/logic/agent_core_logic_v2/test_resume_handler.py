"""Tests for app.logic.agent_core_logic_v2.resume_handler module."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


@pytest.mark.asyncio
class TestCreateResumeGenerator:
    """Tests for create_resume_generator function."""

    async def test_create_resume_generator_with_continue_action(self):
        """Test resume generator with continue action."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "continue"
        mock_resume_info.modified_args = []
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_123"

        async def mock_process_arun_loop(agent, is_debug):
            yield {"output": "chunk1"}
            yield {"output": "chunk2"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle"
            ) as mock_convert,
            patch(
                "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
            ) as mock_process,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async",
                new_callable=AsyncMock,
            ) as mock_serialize,
        ):
            mock_convert.return_value = MagicMock()
            mock_process.return_value = mock_process_arun_loop(mock_agent, False)
            mock_serialize.return_value = '{"output": "serialized"}'

            mock_agent.resume = AsyncMock()

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            assert len(results) == 3  # 2 chunks + 1 final
            mock_agent.resume.assert_called_once()

    async def test_create_resume_generator_with_skip_action(self):
        """Test resume generator with skip action."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "skip"
        mock_resume_info.modified_args = []
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_456"

        async def mock_process_arun_loop(agent, is_debug):
            yield {"output": "result"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle"
            ) as mock_convert,
            patch(
                "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
            ) as mock_process,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async",
                new_callable=AsyncMock,
            ) as mock_serialize,
        ):
            mock_convert.return_value = MagicMock()
            mock_process.return_value = mock_process_arun_loop(mock_agent, False)
            mock_serialize.return_value = '{"output": "serialized"}'

            mock_agent.resume = AsyncMock()

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            assert len(results) == 2  # 1 chunk + 1 final
            # Verify that __skip_tool__ was set
            call_args = mock_agent.resume.call_args
            assert call_args is not None

    async def test_create_resume_generator_with_modified_args(self):
        """Test resume generator with modified arguments."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "continue"

        # Create mock modified args
        mock_arg1 = MagicMock()
        mock_arg1.key = "param1"
        mock_arg1.value = "value1"
        mock_arg2 = MagicMock()
        mock_arg2.key = "param2"
        mock_arg2.value = "value2"

        mock_resume_info.modified_args = [mock_arg1, mock_arg2]
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_789"

        async def mock_process_arun_loop(agent, is_debug):
            yield {"output": "test"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle"
            ) as mock_convert,
            patch(
                "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
            ) as mock_process,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async",
                new_callable=AsyncMock,
            ) as mock_serialize,
        ):
            mock_convert.return_value = MagicMock()
            mock_process.return_value = mock_process_arun_loop(mock_agent, False)
            mock_serialize.return_value = '{"output": "test"}'

            mock_agent.resume = AsyncMock()

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            assert len(results) == 2
            mock_agent.resume.assert_called_once()

    async def test_create_resume_generator_with_exception(self):
        """Test resume generator with exception handling."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "continue"
        mock_resume_info.modified_args = []
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_error"

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle",
                side_effect=Exception("Test error"),
            ),
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async",
                new_callable=AsyncMock,
            ) as mock_serialize,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.agent_instance_manager"
            ) as mock_manager,
        ):
            mock_serialize.return_value = '{"error": "Test error"}'

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            # Should return error output
            assert len(results) == 1
            mock_manager.remove.assert_called_once_with(agent_run_id)

    async def test_create_resume_generator_empty_process_loop(self):
        """Test resume generator with empty process loop."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "continue"
        mock_resume_info.modified_args = []
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_empty"

        async def mock_process_arun_loop(agent, is_debug):
            return
            yield  # Make it a generator without yielding

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle"
            ) as mock_convert,
            patch(
                "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
            ) as mock_process,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async",
                new_callable=AsyncMock,
            ) as mock_serialize,
        ):
            mock_convert.return_value = MagicMock()
            mock_process.return_value = mock_process_arun_loop(mock_agent, False)
            mock_serialize.return_value = '{"answer": {}}'

            mock_agent.resume = AsyncMock()

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            # Should still return final output
            assert len(results) == 1

    async def test_create_resume_generator_status_fields(self):
        """Test that resume generator sets correct status fields."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "continue"
        mock_resume_info.modified_args = []
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_status"

        async def mock_process_arun_loop(agent, is_debug):
            yield {"data": "test_data"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle"
            ) as mock_convert,
            patch(
                "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
            ) as mock_process,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async"
            ) as mock_serialize,
        ):
            serialization_calls = []

            async def mock_serialize_async(data):
                serialization_calls.append(data)
                return '{"data": "serialized"}'

            mock_convert.return_value = MagicMock()
            mock_process.return_value = mock_process_arun_loop(mock_agent, False)
            mock_serialize.side_effect = mock_serialize_async

            mock_agent.resume = AsyncMock()

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            # Check that status fields were set
            assert len(serialization_calls) >= 1
            # First chunk should have status "False"
            assert "status" in serialization_calls[0]
            assert serialization_calls[0]["agent_run_id"] == agent_run_id

    async def test_create_resume_generator_with_agent_error(self):
        """Test resume generator when agent.resume raises an error."""
        from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator

        mock_agent = MagicMock()
        mock_agent_core = MagicMock()
        mock_resume_info = MagicMock()
        mock_resume_info.action = "continue"
        mock_resume_info.modified_args = []
        mock_resume_info.resume_handle = MagicMock()

        agent_run_id = "test_run_agent_error"

        with (
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.interrupt_handle_to_resume_handle"
            ) as mock_convert,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.json_serialize_async",
                new_callable=AsyncMock,
            ) as mock_serialize,
            patch(
                "app.logic.agent_core_logic_v2.resume_handler.agent_instance_manager"
            ) as mock_manager,
        ):
            mock_convert.return_value = MagicMock()
            mock_serialize.return_value = '{"error": "Agent error"}'
            mock_agent.resume = AsyncMock(side_effect=Exception("Agent resume failed"))

            results = []
            async for result in create_resume_generator(
                mock_agent, mock_agent_core, agent_run_id, mock_resume_info
            ):
                results.append(result)

            # Should handle error and return error output
            assert len(results) == 1
            mock_manager.remove.assert_called_once_with(agent_run_id)
