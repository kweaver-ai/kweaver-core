"""单元测试 - logic/agent_core_logic_v2/trace 模块"""

from unittest.mock import MagicMock


class TestSpanSetAttrs:
    """测试 span_set_attrs 函数"""

    def test_sets_all_attributes(self):
        """测试设置所有属性"""
        from app.logic.agent_core_logic_v2.trace import span_set_attrs
        from opentelemetry.trace import Span

        mock_span = MagicMock(spec=Span)
        mock_span.is_recording.return_value = True

        span_set_attrs(
            mock_span, agent_run_id="run123", agent_id="agent456", user_id="user789"
        )

        mock_span.set_attribute.assert_any_call("agent_run_id", "run123")
        mock_span.set_attribute.assert_any_call("agent_id", "agent456")
        mock_span.set_attribute.assert_any_call("user_id", "user789")

    def test_sets_only_agent_run_id(self):
        """测试只设置agent_run_id"""
        from app.logic.agent_core_logic_v2.trace import span_set_attrs
        from opentelemetry.trace import Span

        mock_span = MagicMock(spec=Span)
        mock_span.is_recording.return_value = True

        span_set_attrs(mock_span, agent_run_id="run123")

        mock_span.set_attribute.assert_called_once_with("agent_run_id", "run123")

    def test_with_none_span(self):
        """测试span为None时不报错"""
        from app.logic.agent_core_logic_v2.trace import span_set_attrs

        # Should not raise
        span_set_attrs(None, agent_run_id="run123")

    def test_with_non_recording_span(self):
        """测试span不记录时不设置属性"""
        from app.logic.agent_core_logic_v2.trace import span_set_attrs
        from opentelemetry.trace import Span

        mock_span = MagicMock(spec=Span)
        mock_span.is_recording.return_value = False

        span_set_attrs(mock_span, agent_run_id="run123")

        mock_span.set_attribute.assert_not_called()

    def test_with_none_values(self):
        """测试所有参数为None时不设置属性"""
        from app.logic.agent_core_logic_v2.trace import span_set_attrs
        from opentelemetry.trace import Span

        mock_span = MagicMock(spec=Span)
        mock_span.is_recording.return_value = True

        span_set_attrs(mock_span)

        mock_span.set_attribute.assert_not_called()
