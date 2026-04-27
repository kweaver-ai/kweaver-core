"""Tests for app.logic.agent_core_logic_v2.resume_dolphin_agent_run."""

from unittest.mock import AsyncMock, MagicMock, patch

import pytest


@pytest.mark.asyncio
async def test_resume_dolphin_agent_run_accepts_injected_span():
    """恢复执行链路应兼容 internal_span 注入的 span 参数。"""
    from app.logic.agent_core_logic_v2.resume_dolphin_agent_run import (
        resume_dolphin_agent_run,
    )

    mock_agent_core = MagicMock()
    mock_agent_core.memory_handler.start_memory_build_thread = MagicMock()

    mock_agent = MagicMock()
    mock_agent.resume = AsyncMock()

    mock_resume_info = MagicMock()
    mock_resume_info.data = {"tool_name": "demo_tool", "tool_args": []}
    mock_resume_info.modified_args = []
    mock_resume_info.action = "continue"
    mock_resume_info.resume_handle = MagicMock()

    async def mock_process_arun_loop(_agent, _is_debug):
        if False:
            yield {}

    with (
        patch(
            "app.logic.agent_core_logic_v2.resume_dolphin_agent_run.interrupt_handle_to_resume_handle",
            return_value="resume-handle",
        ),
        patch(
            "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop",
            side_effect=mock_process_arun_loop,
        ),
        patch(
            "app.logic.agent_core_logic_v2.resume_dolphin_agent_run.DialogLogHandler"
        ) as mock_dialog_log_handler,
    ):
        results = []
        async for item in resume_dolphin_agent_run(
            ac=mock_agent_core,
            agent=mock_agent,
            agent_run_id="run-123",
            resume_info=mock_resume_info,
            config=MagicMock(),
            context_variables={},
            headers={},
            is_debug=False,
            span=MagicMock(),
        ):
            results.append(item)

    assert results == [{}]
    mock_agent.resume.assert_awaited_once()
    mock_dialog_log_handler.return_value.save_dialog_logs.assert_called_once()
    mock_agent_core.memory_handler.start_memory_build_thread.assert_called_once()
