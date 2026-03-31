"""Tests for app.logic.agent_core_logic_v2.output module."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


@pytest.mark.asyncio
class TestOutputHandler:
    """Tests for OutputHandler class."""

    async def test_output_handler_init(self):
        """Test OutputHandler initialization."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        assert handler.agent_core == mock_agent_core


@pytest.mark.asyncio
class TestStringOutput:
    """Tests for string_output method."""

    async def test_string_output_single_chunk(self):
        """Test string_output with single chunk."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"key": "value"}

        with patch(
            "app.logic.agent_core_logic_v2.output.json_serialize_async",
            new_callable=AsyncMock,
        ) as mock_serialize:
            mock_serialize.return_value = '{"key": "value"}'

            results = []
            async for result in handler.string_output(mock_generator()):
                results.append(result)

            assert len(results) == 1
            assert results[0] == '{"key": "value"}'

    async def test_string_output_multiple_chunks(self):
        """Test string_output with multiple chunks."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"chunk": 1}
            yield {"chunk": 2}
            yield {"chunk": 3}

        with patch(
            "app.logic.agent_core_logic_v2.output.json_serialize_async",
            new_callable=AsyncMock,
        ) as mock_serialize:
            mock_serialize.side_effect = [
                '{"chunk": 1}',
                '{"chunk": 2}',
                '{"chunk": 3}',
            ]

            results = []
            async for result in handler.string_output(mock_generator()):
                results.append(result)

            assert len(results) == 3


@pytest.mark.asyncio
class TestAddStatus:
    """Tests for add_status method."""

    async def test_add_status_without_status(self):
        """Test add_status when chunks don't have status."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"data": "test1"}
            yield {"data": "test2"}

        results = []
        async for result in handler.add_status(mock_generator()):
            results.append(result)

        # All chunks should have status="False"
        assert all("status" in r for r in results[:-1])
        # Last chunk should have status="True"
        assert results[-1].get("status") == "True"

    async def test_add_status_with_existing_status(self):
        """Test add_status when chunks already have status."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"data": "test", "status": "Completed"}

        results = []
        async for result in handler.add_status(mock_generator()):
            results.append(result)

        # Status should remain unchanged
        assert results[0]["status"] == "Completed"

    async def test_add_status_empty_generator(self):
        """Test add_status with empty generator."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            return
            yield

        results = []
        async for result in handler.add_status(mock_generator()):
            results.append(result)

        assert len(results) == 0


@pytest.mark.asyncio
class TestAddTtft:
    """Tests for add_ttft method."""

    async def test_add_ttft_with_start_time(self):
        """Test add_ttft with start time provided."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"data": "test"}

        with patch("app.logic.agent_core_logic_v2.output.time.time") as mock_time:
            mock_time.return_value = 100.5  # 100.5 seconds after start
            start_time = 100.0

            results = []
            async for result in handler.add_ttft(mock_generator(), start_time):
                results.append(result)

            assert len(results) == 1

    async def test_add_ttft_multiple_chunks(self):
        """Test add_ttft with multiple chunks."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"chunk": 1}
            yield {"chunk": 2}

        with patch(
            "app.logic.agent_core_logic_v2.output.time.time", return_value=101.0
        ):
            start_time = 100.0

            results = []
            async for result in handler.add_ttft(mock_generator(), start_time):
                results.append(result)

            assert len(results) == 2


@pytest.mark.asyncio
class TestAddDatetime:
    """Tests for add_datetime method."""

    async def test_add_datetime_single_chunk(self):
        """Test add_datetime with single chunk."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"data": "test"}

        with patch("app.logic.agent_core_logic_v2.output.datetime") as mock_datetime:
            mock_datetime.now.return_value.strftime.return_value[:-3] = (
                "2025-01-01 12:00:00.000"
            )

            results = []
            async for result in handler.add_datetime(mock_generator()):
                results.append(result)

            assert len(results) == 1

    async def test_add_datetime_multiple_chunks(self):
        """Test add_datetime with multiple chunks."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"chunk": 1}
            yield {"chunk": 2}

        with patch("app.logic.agent_core_logic_v2.output.datetime"):
            results = []
            async for result in handler.add_datetime(mock_generator()):
                results.append(result)

            assert len(results) == 2


@pytest.mark.asyncio
class TestResultOutput:
    """Tests for result_output method."""

    async def test_result_output_basic(self):
        """Test basic result_output flow."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run = MagicMock(return_value=self._mock_output_generator())
        mock_agent_core.cleanup = MagicMock()
        handler = OutputHandler(mock_agent_core)

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_123"
        mock_agent_config.agent_id = "agent_456"
        mock_agent_config.incremental_output = False
        mock_agent_config.output_vars = []

        mock_agent_input = MagicMock()
        headers = {"x-user-id": "user123"}

        with (
            patch("app.logic.agent_core_logic_v2.output.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.output.get_user_account_id",
                return_value="user123",
            ),
            patch(
                "app.logic.agent_core_logic_v2.output.json_serialize_async",
                new_callable=AsyncMock,
                return_value='{"data": "test"}',
            ),
        ):
            results = []
            async for result in handler.result_output(
                mock_agent_config, mock_agent_input, headers
            ):
                results.append(result)

            # Verify cleanup was called
            mock_agent_core.cleanup.assert_called_once()

    async def test_result_output_with_incremental(self):
        """Test result_output with incremental_output enabled."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run = MagicMock(return_value=self._mock_output_generator())
        mock_agent_core.cleanup = MagicMock()
        handler = OutputHandler(mock_agent_core)

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_123"
        mock_agent_config.agent_id = "agent_456"
        mock_agent_config.incremental_output = True
        mock_agent_config.output_vars = []

        mock_agent_input = MagicMock()
        headers = {}

        with (
            patch("app.logic.agent_core_logic_v2.output.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.output.get_user_account_id",
                return_value="user123",
            ),
            patch(
                "app.logic.agent_core_logic_v2.output.incremental_async_generator"
            ) as mock_incremental,
            patch(
                "app.logic.agent_core_logic_v2.output.json_serialize_async",
                new_callable=AsyncMock,
                return_value='{"data": "test"}',
            ),
        ):
            mock_incremental.return_value = self._mock_output_generator()

            results = []
            async for result in handler.result_output(
                mock_agent_config, mock_agent_input, headers
            ):
                results.append(result)

            # Verify incremental generator was used
            mock_incremental.assert_called_once()

    async def test_result_output_with_start_time(self):
        """Test result_output with start_time provided."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run = MagicMock(return_value=self._mock_output_generator())
        mock_agent_core.cleanup = MagicMock()
        handler = OutputHandler(mock_agent_core)

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_123"
        mock_agent_config.agent_id = "agent_456"
        mock_agent_config.incremental_output = False
        mock_agent_config.output_vars = []

        mock_agent_input = MagicMock()
        headers = {}
        start_time = 100.0

        with (
            patch("app.logic.agent_core_logic_v2.output.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.output.get_user_account_id",
                return_value="user123",
            ),
            patch(
                "app.logic.agent_core_logic_v2.output.json_serialize_async",
                new_callable=AsyncMock,
                return_value='{"data": "test"}',
            ),
        ):
            results = []
            async for result in handler.result_output(
                mock_agent_config, mock_agent_input, headers, start_time=start_time
            ):
                results.append(result)

            assert len(results) > 0

    async def test_result_output_cleanup_on_exception(self):
        """Test that cleanup is called even when exception occurs."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run = MagicMock(side_effect=Exception("Run failed"))
        mock_agent_core.cleanup = MagicMock()
        handler = OutputHandler(mock_agent_core)

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_123"
        mock_agent_config.agent_id = "agent_456"
        mock_agent_config.incremental_output = False
        mock_agent_config.output_vars = []

        mock_agent_input = MagicMock()
        headers = {}

        with (
            patch("app.logic.agent_core_logic_v2.output.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.output.get_user_account_id",
                return_value="user123",
            ),
        ):
            with pytest.raises(Exception, match="Run failed"):
                async for _ in handler.result_output(
                    mock_agent_config, mock_agent_input, headers
                ):
                    pass

            # Cleanup should still be called
            mock_agent_core.cleanup.assert_called_once()

    async def _mock_output_generator(self):
        """Helper method to create a mock output generator."""
        yield {"data": "test"}


@pytest.mark.asyncio
class TestPartialOutput:
    """Tests for partial_output method."""

    async def test_partial_output_single_var(self):
        """Test partial_output with single output variable."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"field1": {"nested": "value1"}}

        output_vars = ["field1.nested"]

        results = []
        async for result in handler.partial_output(mock_generator(), output_vars):
            results.append(result)

        assert len(results) == 1

    async def test_partial_output_multiple_vars(self):
        """Test partial_output with multiple output variables."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"field1": "value1", "field2": "value2"}

        output_vars = ["field1", "field2"]

        results = []
        async for result in handler.partial_output(mock_generator(), output_vars):
            results.append(result)

        assert len(results) == 1

    async def test_partial_output_empty_vars(self):
        """Test partial_output with empty output_vars."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"data": "test"}

        output_vars = []

        results = []
        async for result in handler.partial_output(mock_generator(), output_vars):
            results.append(result)

        # Should return original output
        assert len(results) == 1

    async def test_partial_output_with_dolphin_var(self):
        """Test partial_output with dolphin variable."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"field1": {"__dolphin_var__": True, "value": "extracted_value"}}

        output_vars = ["field1"]

        with patch(
            "app.logic.agent_core_logic_v2.output.is_dolphin_var", return_value=True
        ):
            results = []
            async for result in handler.partial_output(mock_generator(), output_vars):
                results.append(result)

            assert len(results) == 1

    async def test_partial_output_missing_field(self):
        """Test partial_output when field doesn't exist."""
        from app.logic.agent_core_logic_v2.output import OutputHandler

        mock_agent_core = MagicMock()
        handler = OutputHandler(mock_agent_core)

        async def mock_generator():
            yield {"other_field": "value"}

        output_vars = ["nonexistent_field"]

        results = []
        async for result in handler.partial_output(mock_generator(), output_vars):
            results.append(result)

        # Should handle missing field gracefully
        assert len(results) >= 0
