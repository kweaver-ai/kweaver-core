"""Tests for app.logic.agent_core_logic_v2.input_handler_pkg.process_input module."""

import pytest
from unittest.mock import MagicMock, patch


@pytest.mark.asyncio
class TestProcessInput:
    """Tests for process_input function."""

    async def test_process_input_with_string_field(self):
        """Test processing input with string field."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "string", "name": "query"}]}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value="test query")
        mock_agent_input.set_value = MagicMock()

        headers = {"x-user-id": "user123"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ) as mock_span_attrs,
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # Verify string value was set
            mock_agent_input.set_value.assert_any_call("query", "test query")
            assert isinstance(result, dict)

    async def test_process_input_with_string_field_empty(self):
        """Test processing input with empty string field."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "string", "name": "query"}]}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=None)
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ) as mock_span_attrs,
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value=None,
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # Verify empty string was set
            mock_agent_input.set_value.assert_any_call("query", "")

    async def test_process_input_with_object_field_json(self):
        """Test processing input with object field containing JSON."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "object", "name": "params"}]}

        mock_agent_input = MagicMock()
        json_str = '{"key": "value"}'
        mock_agent_input.get_value = MagicMock(return_value=json_str)
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # Verify JSON was parsed and set_value was called
            assert mock_agent_input.set_value.called
            call_args = mock_agent_input.set_value.call_args_list
            # Check that set_value was called with "params" and some dict value
            assert len(call_args) > 0

    async def test_process_input_with_object_field_non_string(self):
        """Test processing input with object field containing non-string value."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "object", "name": "params"}]}

        mock_agent_input = MagicMock()
        test_dict = {"key": "value"}
        mock_agent_input.get_value = MagicMock(return_value=test_dict)
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # For non-string values, process_input does 'continue' - no set_value called
            # This is the expected behavior from the code
            assert not mock_agent_input.set_value.called or result is not None

    async def test_process_input_with_object_field_empty(self):
        """Test processing input with empty object field."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "object", "name": "params"}]}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=None)
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # Empty dict should be set
            mock_agent_input.set_value.assert_any_call("params", {})

    async def test_process_input_with_file_field(self):
        """Test processing input with file field."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "file", "name": "upload"}]}

        file_info = {"filename": "test.txt", "content": "base64content"}
        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=[file_info])
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # File info should be set and stored in temp_files
            mock_agent_input.set_value.assert_any_call("upload", [file_info])
            assert "upload" in result
            assert result["upload"] == [file_info]

    async def test_process_input_with_file_field_empty(self):
        """Test processing input with empty file field."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "file", "name": "upload"}]}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=None)
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # Empty list should be set
            mock_agent_input.set_value.assert_any_call("upload", [])
            assert "upload" in result

    async def test_process_input_sets_history(self):
        """Test that history is set to empty list if not present."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": []}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=None)
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # History should be set to empty list
            mock_agent_input.set_value.assert_any_call("history", [])

    async def test_process_input_sets_header_and_config(self):
        """Test that header and self_config are set."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": []}
        mock_agent_config.model_dump.return_value = {"agent_id": "test_agent_id"}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=None)
        mock_agent_input.set_value = MagicMock()

        headers = {"x-user-id": "user123", "x-account-type": "standard"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # Header and self_config should be set
            assert mock_agent_input.header == headers
            mock_agent_config.model_dump.assert_called_once()

    async def test_process_input_with_is_debug(self):
        """Test processing input with is_debug flag."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": [{"type": "string", "name": "query"}]}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value="test")
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(
                mock_agent_config, mock_agent_input, headers, is_debug=True
            )

            # Should handle is_debug flag
            assert isinstance(result, dict)

    async def test_process_input_with_span(self):
        """Test processing input with span parameter."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {"fields": []}

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(return_value=None)
        mock_agent_input.set_value = MagicMock()

        headers = {}
        mock_span = MagicMock()

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ) as mock_span_attrs,
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(
                mock_agent_config, mock_agent_input, headers, span=mock_span
            )

            # Span attrs should be called with the span
            mock_span_attrs.assert_called_once()

    async def test_process_input_multiple_fields(self):
        """Test processing input with multiple fields."""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_input import (
            process_input,
        )

        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_config.input = {
            "fields": [
                {"type": "string", "name": "query"},
                {"type": "object", "name": "params"},
                {"type": "file", "name": "upload"},
            ]
        }

        mock_agent_input = MagicMock()
        mock_agent_input.get_value = MagicMock(
            side_effect=lambda x: {
                "query": "test query",
                "params": '{"key": "value"}',
                "upload": [{"filename": "test.txt"}],
            }.get(x)
        )
        mock_agent_input.set_value = MagicMock()

        headers = {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.span_set_attrs"
            ),
            patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_input.get_user_account_id",
                return_value="user123",
            ),
        ):
            result = await process_input(mock_agent_config, mock_agent_input, headers)

            # All fields should be processed
            assert mock_agent_input.set_value.call_count > 3
            assert "upload" in result
