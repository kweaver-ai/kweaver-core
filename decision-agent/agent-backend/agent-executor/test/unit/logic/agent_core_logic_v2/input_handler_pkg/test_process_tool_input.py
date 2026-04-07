# -*- coding: utf-8 -*-
"""单元测试 - input_handler_pkg/process_tool_input 模块"""

import pytest
from unittest.mock import MagicMock, patch


class TestProcessToolInput:
    """测试 process_tool_input 函数"""

    @pytest.fixture
    def mock_agent_input(self):
        """创建 mock AgentInputVo"""
        mock = MagicMock()
        mock.header = {"x-user-id": "user123"}
        mock.model_dump.return_value = {
            "query": "test query",
            "context": {"key": "value"},
            "tool": {"tool_id": "tool123"},
        }
        return mock

    @pytest.mark.asyncio
    async def test_process_tool_input_basic(self, mock_agent_input):
        """测试基本工具输入处理"""
        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.span_set_attrs"
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.get_user_account_id",
                return_value="user123",
            ):
                from app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input import (
                    process_tool_input,
                )

                result, event_key = await process_tool_input(mock_agent_input)

                assert isinstance(result, dict)
                assert event_key is None
                assert "query" in result
                assert "context" in result

    @pytest.mark.asyncio
    async def test_process_tool_input_removes_tool_field(self, mock_agent_input):
        """测试移除tool字段"""
        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.span_set_attrs"
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.get_user_account_id",
                return_value="user123",
            ):
                from app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input import (
                    process_tool_input,
                )

                result, _ = await process_tool_input(mock_agent_input)

                # tool字段应该被移除
                assert "tool" not in result

    @pytest.mark.asyncio
    async def test_process_tool_input_without_tool_field(self):
        """测试输入中没有tool字段"""
        mock_input = MagicMock()
        mock_input.header = {"x-user-id": "user123"}
        mock_input.model_dump.return_value = {
            "query": "test query",
            "context": {"key": "value"},
        }

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.span_set_attrs"
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.get_user_account_id",
                return_value="user123",
            ):
                from app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input import (
                    process_tool_input,
                )

                result, _ = await process_tool_input(mock_input)

                assert "query" in result
                assert "context" in result

    @pytest.mark.asyncio
    async def test_process_tool_input_with_none_header(self):
        """测试header为None的情况"""
        mock_input = MagicMock()
        mock_input.header = None
        mock_input.model_dump.return_value = {"query": "test"}

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.span_set_attrs"
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.get_user_account_id",
                return_value="user123",
            ):
                from app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input import (
                    process_tool_input,
                )

                result, _ = await process_tool_input(mock_input)

                assert isinstance(result, dict)

    @pytest.mark.asyncio
    async def test_process_tool_input_with_span(self, mock_agent_input):
        """测试带span参数的处理"""
        mock_span = MagicMock()

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.span_set_attrs"
        ) as mock_span_attrs:
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.get_user_account_id",
                return_value="user123",
            ):
                from app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input import (
                    process_tool_input,
                )

                await process_tool_input(mock_agent_input, span=mock_span)

                mock_span_attrs.assert_called_once()

    @pytest.mark.asyncio
    async def test_process_tool_input_empty_input(self):
        """测试空输入处理"""
        mock_input = MagicMock()
        mock_input.header = {}
        mock_input.model_dump.return_value = {}

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.span_set_attrs"
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input.get_user_account_id",
                return_value="",
            ):
                from app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input import (
                    process_tool_input,
                )

                result, event_key = await process_tool_input(mock_input)

                assert result == {}
                assert event_key is None
