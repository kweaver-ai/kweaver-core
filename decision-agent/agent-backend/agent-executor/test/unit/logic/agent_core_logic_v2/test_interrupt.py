"""单元测试 - logic/agent_core_logic_v2/interrupt 模块"""

import pytest
from unittest.mock import MagicMock, patch
import sys


# Setup mock modules before import
mock_tracer = MagicMock()
mock_span = MagicMock()
mock_span.is_recording.return_value = False
mock_tracer.start_span.return_value = mock_span

mock_exporter_module = MagicMock()
mock_exporter_module.tracer = mock_tracer
sys.modules["exporter.ar_trace.trace_exporter"] = mock_exporter_module


@pytest.mark.asyncio
class TestInterruptHandler:
    """测试 InterruptHandler 类"""

    @patch("app.logic.agent_core_logic_v2.interrupt.json.dumps")
    @patch("app.logic.agent_core_logic_v2.interrupt.StandLogger")
    @patch("app.logic.agent_core_logic_v2.interrupt.span_set_attrs")
    async def test_handle_tool_interrupt_basic(
        self, m_span_set_attrs, m_logger, m_json_dumps
    ):
        """测试基本工具中断处理"""
        m_json_dumps.return_value = '{"test": "data"}'

        from app.logic.agent_core_logic_v2.interrupt import InterruptHandler

        # Mock ToolInterruptException
        mock_interrupt_info = MagicMock()
        mock_interrupt_info.data = {"test": "data"}
        mock_tool_interrupt = MagicMock()
        mock_tool_interrupt.interrupt_info = mock_interrupt_info

        res = {}
        context_variables = {"session_id": "session123", "agent_id": "agent456"}

        await InterruptHandler.handle_tool_interrupt(
            mock_tool_interrupt, res, context_variables
        )

        assert res["interrupt_info"] == mock_interrupt_info
        assert res["status"] == "True"
        m_span_set_attrs.assert_called_once()

    @patch("app.logic.agent_core_logic_v2.interrupt.json.dumps")
    @patch("app.logic.agent_core_logic_v2.interrupt.StandLogger")
    @patch("app.logic.agent_core_logic_v2.interrupt.span_set_attrs")
    async def test_handle_tool_interrupt_with_span(
        self, m_span_set_attrs, m_logger, m_json_dumps
    ):
        """测试带 span 的工具中断处理"""
        m_json_dumps.return_value = '{"test": "data"}'

        from app.logic.agent_core_logic_v2.interrupt import InterruptHandler
        from opentelemetry.trace import Span

        mock_interrupt_info = MagicMock()
        mock_tool_interrupt = MagicMock()
        mock_tool_interrupt.interrupt_info = mock_interrupt_info

        res = {}
        context_variables = {}

        test_span = MagicMock(spec=Span)
        test_span.is_recording.return_value = True

        await InterruptHandler.handle_tool_interrupt(
            mock_tool_interrupt, res, context_variables, span=test_span
        )

        assert res["interrupt_info"] == mock_interrupt_info
        assert res["status"] == "True"

    @patch("app.logic.agent_core_logic_v2.interrupt.json.dumps")
    @patch("app.logic.agent_core_logic_v2.interrupt.StandLogger")
    @patch("app.logic.agent_core_logic_v2.interrupt.span_set_attrs")
    async def test_handle_tool_interrupt_empty_context(
        self, m_span_set_attrs, m_logger, m_json_dumps
    ):
        """测试空上下文变量的工具中断处理"""
        m_json_dumps.return_value = '{"test": "data"}'

        from app.logic.agent_core_logic_v2.interrupt import InterruptHandler

        mock_interrupt_info = MagicMock()
        mock_tool_interrupt = MagicMock()
        mock_tool_interrupt.interrupt_info = mock_interrupt_info

        res = {}
        context_variables = {}

        await InterruptHandler.handle_tool_interrupt(
            mock_tool_interrupt, res, context_variables
        )

        assert res["interrupt_info"] == mock_interrupt_info
        assert res["status"] == "True"
