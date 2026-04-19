"""Unit tests for app/utils/observability/trace_wrapper.py."""

from unittest.mock import MagicMock, patch

import pytest


class TestInternalSpan:
    def test_returns_original_function_when_trace_disabled(self):
        with patch("app.common.config.Config") as mock_config:
            mock_config.is_o11y_trace_enabled.return_value = False
            from app.utils.observability.trace_wrapper import internal_span

            @internal_span()
            def test_func():
                return "result"

            assert test_func() == "result"

    @patch("app.utils.observability.trace_wrapper.func_judgment", return_value=(False, False))
    def test_sync_function_uses_standard_otel_tracer(self, _mock_func_judgment):
        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True
        mock_span.get_span_context.return_value.trace_id = 1
        mock_span.get_span_context.return_value.span_id = 2

        with patch("app.common.config.Config") as mock_config:
            mock_config.is_o11y_trace_enabled.return_value = True
            with patch(
                "app.utils.observability.trace_wrapper.trace.get_tracer",
                return_value=mock_tracer,
            ) as mock_get_tracer:
                with patch("app.common.stand_log.StandLogger.info_log") as mock_info_log:
                    from app.utils.observability.trace_wrapper import internal_span

                    @internal_span(name="test_span")
                    def test_func(span=None):
                        return "ok"

                    assert test_func() == "ok"
                    mock_info_log.assert_not_called()

        mock_get_tracer.assert_called_once_with("agent-executor.internal")
        mock_tracer.start_span.assert_called_once()
        mock_span.end.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.func_judgment", return_value=(True, False))
    async def test_async_function_records_exception(self, _mock_func_judgment):
        mock_tracer = MagicMock()
        mock_span = MagicMock()
        mock_tracer.start_span.return_value = mock_span
        mock_span.is_recording.return_value = True
        mock_span.get_span_context.return_value.trace_id = 1
        mock_span.get_span_context.return_value.span_id = 2

        with patch("app.common.config.Config") as mock_config:
            mock_config.is_o11y_trace_enabled.return_value = True
            with patch(
                "app.utils.observability.trace_wrapper.trace.get_tracer",
                return_value=mock_tracer,
            ):
                with patch("app.common.stand_log.StandLogger.info_log") as mock_info_log:
                    from app.utils.observability.trace_wrapper import internal_span

                    @internal_span(name="async_span")
                    async def test_func(span=None):
                        raise RuntimeError("boom")

                    with pytest.raises(RuntimeError, match="boom"):
                        await test_func()

                    mock_info_log.assert_not_called()

        mock_span.record_exception.assert_called_once()
        mock_span.end.assert_called_once()

    @patch("app.utils.observability.trace_wrapper.func_judgment", return_value=(False, False))
    def test_sync_function_preserves_explicit_positional_span(self, _mock_func_judgment):
        mock_tracer = MagicMock()
        internal_span_mock = MagicMock()
        mock_tracer.start_span.return_value = internal_span_mock
        internal_span_mock.is_recording.return_value = True

        with patch("app.common.config.Config") as mock_config:
            mock_config.is_o11y_trace_enabled.return_value = True
            with patch(
                "app.utils.observability.trace_wrapper.trace.get_tracer",
                return_value=mock_tracer,
            ):
                from app.utils.observability.trace_wrapper import internal_span

                provided_span = MagicMock()

                @internal_span(name="positional_span")
                def test_func(value, span=None):
                    return value, span

                result = test_func("ok", provided_span)

        assert result == ("ok", provided_span)
        internal_span_mock.end.assert_called_once()

    @pytest.mark.asyncio
    @patch("app.utils.observability.trace_wrapper.func_judgment", return_value=(True, False))
    async def test_async_function_preserves_explicit_kwarg_span(self, _mock_func_judgment):
        mock_tracer = MagicMock()
        internal_span_mock = MagicMock()
        mock_tracer.start_span.return_value = internal_span_mock
        internal_span_mock.is_recording.return_value = True

        with patch("app.common.config.Config") as mock_config:
            mock_config.is_o11y_trace_enabled.return_value = True
            with patch(
                "app.utils.observability.trace_wrapper.trace.get_tracer",
                return_value=mock_tracer,
            ):
                from app.utils.observability.trace_wrapper import internal_span

                provided_span = MagicMock()

                @internal_span(name="kwarg_span")
                async def test_func(span=None):
                    return span

                result = await test_func(span=provided_span)

        assert result is provided_span
        internal_span_mock.end.assert_called_once()
