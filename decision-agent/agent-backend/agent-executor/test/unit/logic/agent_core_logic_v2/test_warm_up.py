"""Tests for app.logic.agent_core_logic_v2.warm_up module."""

import pytest
from unittest.mock import MagicMock, patch
from typing import Dict


@pytest.mark.asyncio
class TestWarmUpHandler:
    """Tests for WarmUpHandler class."""

    async def test_warmup_handler_init(self):
        """Test WarmUpHandler initialization."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        handler = WarmUpHandler(mock_agent_core)

        assert handler.agentCore == mock_agent_core

    async def test_warnup_success(self):
        """Test successful warmup execution."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)

        headers = {"x-user-id": "test_user", "x-account-type": "standard"}

        async def mock_generator():
            yield {"data": "chunk1"}
            yield {"data": "chunk2"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.warm_up.span_set_attrs"
            ) as mock_span_attrs,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin"
            ) as mock_run_dolphin,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value="test_user",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value="standard",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.set_user_account_id"
            ) as mock_set_id,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.set_user_account_type"
            ) as mock_set_type,
        ):
            mock_run_dolphin.return_value = mock_generator()

            await handler.warnup(headers)

            mock_span_attrs.assert_called_once()
            mock_set_id.assert_called_once()
            mock_set_type.assert_called_once()

    async def test_warnup_with_exception(self):
        """Test warmup with exception handling."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)

        headers = {}

        async def failing_generator():
            raise Exception("Test exception")
            yield

        with (
            patch("app.logic.agent_core_logic_v2.warm_up.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin"
            ) as mock_run_dolphin,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value=None,
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value=None,
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.StandLogger.error"
            ) as mock_standard_error,
            patch("app.logic.agent_core_logic_v2.warm_up.o11y_logger") as mock_logger,
        ):
            mock_run_dolphin.return_value = failing_generator()

            # Should not raise exception, should log error
            await handler.warnup(headers)

            mock_standard_error.assert_called_once()
            mock_logger().error.assert_called_once()

    async def test_warnup_with_span(self):
        """Test warmup with span parameter."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run_id"
        mock_agent_config.agent_id = "test_agent_id"
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)

        headers = {}
        mock_span = MagicMock()

        async def mock_generator():
            yield {"data": "test"}

        with (
            patch(
                "app.logic.agent_core_logic_v2.warm_up.span_set_attrs"
            ) as mock_span_attrs,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin",
                return_value=mock_generator(),
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value="user123",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value="premium",
            ),
            patch("app.logic.agent_core_logic_v2.warm_up.set_user_account_id"),
            patch("app.logic.agent_core_logic_v2.warm_up.set_user_account_type"),
        ):
            await handler.warnup(headers, span=mock_span)

            mock_span_attrs.assert_called_once()

    async def test_warnup_empty_headers(self):
        """Test warmup with empty headers."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_123"
        mock_agent_config.agent_id = "agent_456"
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)

        headers: Dict[str, str] = {}

        async def mock_generator():
            yield {}

        with (
            patch(
                "app.logic.agent_core_logic_v2.warm_up.span_set_attrs"
            ) as mock_span_attrs,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin",
                return_value=mock_generator(),
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value=None,
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value=None,
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.set_user_account_id"
            ) as mock_set_id,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.set_user_account_type"
            ) as mock_set_type,
        ):
            await handler.warnup(headers)

            # Should still call span_set_attrs with empty user_id
            mock_span_attrs.assert_called_once()
            mock_set_id.assert_called_once()
            mock_set_type.assert_called_once()

    async def test_warnup_generator_consumption(self):
        """Test that warmup properly consumes the generator."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "test_run"
        mock_agent_config.agent_id = "test_agent"
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)

        headers = {"x-user-id": "user1"}

        call_count = [0]

        async def mock_generator():
            for i in range(5):
                call_count[0] += 1
                yield {"chunk": i}

        with (
            patch("app.logic.agent_core_logic_v2.warm_up.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin",
                return_value=mock_generator(),
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value="user1",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value="standard",
            ),
            patch("app.logic.agent_core_logic_v2.warm_up.set_user_account_id"),
            patch("app.logic.agent_core_logic_v2.warm_up.set_user_account_type"),
        ):
            await handler.warnup(headers)

            # Generator should be fully consumed
            assert call_count[0] == 5

    async def test_warnup_injects_self_config_into_context_variables(self):
        """Warmup should provide self_config for downstream skill processing."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_with_self_config"
        mock_agent_config.agent_id = "agent_with_self_config"
        mock_agent_config.model_dump.return_value = {
            "agent_id": "agent_with_self_config",
            "skills": {"tools": [{"tool_id": "doc_qa"}]},
        }
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)
        captured_context = {}

        async def mock_generator():
            yield {}

        def capture_run_dolphin(
            agent_core, agent_config, context_vars, hdr, is_debug, temp_files
        ):
            captured_context["variables"] = context_vars
            return mock_generator()

        with (
            patch("app.logic.agent_core_logic_v2.warm_up.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin",
                side_effect=capture_run_dolphin,
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value="user_123",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value="enterprise",
            ),
            patch("app.logic.agent_core_logic_v2.warm_up.set_user_account_id"),
            patch("app.logic.agent_core_logic_v2.warm_up.set_user_account_type"),
        ):
            await handler.warnup({"x-user-id": "user_123"})

        assert captured_context["variables"]["self_config"] == (
            mock_agent_config.model_dump.return_value
        )
        assert "history" not in captured_context["variables"]
        assert "header" not in captured_context["variables"]


@pytest.mark.asyncio
class TestWarmUpIntegration:
    """Integration tests for warm_up module."""

    async def test_warmup_with_context_variables(self):
        """Test that context_variables are properly set."""
        from app.logic.agent_core_logic_v2.warm_up import WarmUpHandler

        mock_agent_core = MagicMock()
        mock_agent_config = MagicMock()
        mock_agent_config.agent_run_id = "run_001"
        mock_agent_config.agent_id = "agent_001"
        mock_agent_core.agent_config = mock_agent_config

        handler = WarmUpHandler(mock_agent_core)

        headers = {"x-user-id": "user_123", "x-account-type": "enterprise"}

        captured_context = {}

        async def mock_generator():
            yield {}

        def capture_run_dolphin(
            agent_core, agent_config, context_vars, hdr, is_debug, temp_files
        ):
            captured_context["variables"] = context_vars
            captured_context["headers"] = hdr
            return mock_generator()

        with (
            patch("app.logic.agent_core_logic_v2.warm_up.span_set_attrs"),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.run_dolphin",
                side_effect=capture_run_dolphin,
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_id",
                return_value="user_123",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.get_user_account_type",
                return_value="enterprise",
            ),
            patch(
                "app.logic.agent_core_logic_v2.warm_up.set_user_account_id"
            ) as mock_set_id,
            patch(
                "app.logic.agent_core_logic_v2.warm_up.set_user_account_type"
            ) as mock_set_type,
        ):
            await handler.warnup(headers)

            # Verify context variables were set
            assert mock_set_id.call_count > 0
            assert mock_set_type.call_count > 0
