"""单元测试 - logic/agent_core_logic_v2/interrupt_utils 模块"""

import pytest
from unittest.mock import MagicMock


class TestCheckAndRaiseInterrupt:
    """测试 check_and_raise_interrupt 函数"""

    def test_returns_early_for_non_dict(self):
        """测试非字典输入时提前返回"""
        from app.logic.agent_core_logic_v2.interrupt_utils import (
            check_and_raise_interrupt,
        )

        # Should not raise
        check_and_raise_interrupt("not a dict")
        check_and_raise_interrupt(123)
        check_and_raise_interrupt(None)

    def test_returns_early_for_non_interrupted_status(self):
        """测试非中断状态时提前返回"""
        from app.logic.agent_core_logic_v2.interrupt_utils import (
            check_and_raise_interrupt,
        )

        check_and_raise_interrupt({"status": "running"})
        check_and_raise_interrupt({"status": "completed"})
        check_and_raise_interrupt({"status": "error"})
        check_and_raise_interrupt({})  # no status key

    def test_returns_early_for_non_tool_confirmation_interrupt(self):
        """测试非工具确认中断时提前返回"""
        from app.logic.agent_core_logic_v2.interrupt_utils import (
            check_and_raise_interrupt,
        )

        check_and_raise_interrupt({"status": "interrupted", "interrupt_type": "other"})
        check_and_raise_interrupt({"status": "interrupted", "interrupt_type": ""})
        check_and_raise_interrupt({"status": "interrupted"})  # no interrupt_type

    def test_raises_tool_interrupt_exception(self):
        """测试抛出工具中断异常"""
        from app.logic.agent_core_logic_v2.interrupt_utils import (
            check_and_raise_interrupt,
        )
        from app.common.exceptions.tool_interrupt import ToolInterruptException

        item = {
            "status": "interrupted",
            "interrupt_type": "tool_confirmation",
            "handle": "handle123",
            "data": {"key": "value"},
        }

        with pytest.raises(ToolInterruptException) as exc_info:
            check_and_raise_interrupt(item)

        assert exc_info.value.interrupt_info.handle == "handle123"
        assert exc_info.value.interrupt_info.data == {"key": "value"}

    def test_raises_with_empty_data(self):
        """测试抛出异常时使用空数据作为默认值"""
        from app.logic.agent_core_logic_v2.interrupt_utils import (
            check_and_raise_interrupt,
        )
        from app.common.exceptions.tool_interrupt import ToolInterruptException

        item = {
            "status": "interrupted",
            "interrupt_type": "tool_confirmation",
            "handle": "handle456",
        }

        with pytest.raises(ToolInterruptException) as exc_info:
            check_and_raise_interrupt(item)

        assert exc_info.value.interrupt_info.handle == "handle456"
        assert exc_info.value.interrupt_info.data == {}


@pytest.mark.asyncio
class TestProcessArunLoop:
    """测试 process_arun_loop 函数"""

    @pytest.mark.asyncio
    async def test_processes_items(self):
        """测试处理项目"""
        from app.logic.agent_core_logic_v2.interrupt_utils import process_arun_loop

        # Mock agent
        mock_agent = MagicMock()
        mock_executor = MagicMock()
        mock_context = MagicMock()

        async def mock_arun():
            yield {"answer": "response1"}
            yield {"answer": "response2"}

        mock_agent.arun = mock_arun
        mock_agent.executor.context = mock_context
        mock_context.get_all_variables_values.return_value = {"var1": "value1"}

        results = []
        async for item in process_arun_loop(mock_agent):
            results.append(item)

        assert len(results) == 2
        assert "answer" in results[0]
        assert "context" in results[0]

    @pytest.mark.asyncio
    async def test_filters_progress_in_non_debug_mode(self):
        """测试非调试模式下过滤进度信息"""
        from app.logic.agent_core_logic_v2.interrupt_utils import process_arun_loop

        mock_agent = MagicMock()
        mock_executor = MagicMock()
        mock_context = MagicMock()

        async def mock_arun():
            yield {
                "_progress": [
                    {"stage": "assign", "value": "1"},
                    {"stage": "execute", "value": "2"},
                ]
            }

        mock_agent.arun = mock_arun
        mock_agent.executor.context = mock_context
        mock_context.get_all_variables_values.return_value = {}

        results = []
        async for item in process_arun_loop(mock_agent, is_debug=False):
            results.append(item)

        # assign stage should be filtered out
        assert len(results[0]["answer"]["_progress"]) == 1
        assert results[0]["answer"]["_progress"][0]["stage"] == "execute"

    @pytest.mark.asyncio
    async def test_keeps_progress_in_debug_mode(self):
        """测试调试模式下保留进度信息"""
        from app.logic.agent_core_logic_v2.interrupt_utils import process_arun_loop

        mock_agent = MagicMock()
        mock_executor = MagicMock()
        mock_context = MagicMock()

        async def mock_arun():
            yield {
                "_progress": [
                    {"stage": "assign", "value": "1"},
                    {"stage": "execute", "value": "2"},
                ]
            }

        mock_agent.arun = mock_arun
        mock_agent.executor.context = mock_context
        mock_context.get_all_variables_values.return_value = {}

        results = []
        async for item in process_arun_loop(mock_agent, is_debug=True):
            results.append(item)

        # All progress should be kept
        assert len(results[0]["answer"]["_progress"]) == 2

    @pytest.mark.asyncio
    async def test_raises_on_interrupt(self):
        """测试检测到中断时抛出异常"""
        from app.logic.agent_core_logic_v2.interrupt_utils import process_arun_loop
        from app.common.exceptions.tool_interrupt import ToolInterruptException

        mock_agent = MagicMock()

        async def mock_arun():
            yield {
                "status": "interrupted",
                "interrupt_type": "tool_confirmation",
                "handle": "handle123",
            }

        mock_agent.arun = mock_arun

        with pytest.raises(ToolInterruptException):
            async for _ in process_arun_loop(mock_agent):
                pass
