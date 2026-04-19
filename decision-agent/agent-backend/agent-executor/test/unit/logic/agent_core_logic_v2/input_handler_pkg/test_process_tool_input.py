# -*- coding: utf-8 -*-
"""单元测试 - input_handler_pkg/process_tool_input 模块"""

import importlib

import pytest
from unittest.mock import MagicMock, patch


def get_process_tool_input_module():
    return importlib.import_module(
        "app.logic.agent_core_logic_v2.input_handler_pkg.process_tool_input"
    )


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
        module = get_process_tool_input_module()
        process_tool_input = module.process_tool_input

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):

                result, event_key = await process_tool_input(mock_agent_input)

                assert isinstance(result, dict)
                assert event_key is None
                assert "query" in result
                assert "context" in result

    @pytest.mark.asyncio
    async def test_process_tool_input_removes_tool_field(self, mock_agent_input):
        """测试移除tool字段"""
        module = get_process_tool_input_module()
        process_tool_input = module.process_tool_input

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):

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

        module = get_process_tool_input_module()
        process_tool_input = module.process_tool_input

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):

                result, _ = await process_tool_input(mock_input)

                assert "query" in result
                assert "context" in result

    @pytest.mark.asyncio
    async def test_process_tool_input_with_none_header(self):
        """测试header为None的情况"""
        mock_input = MagicMock()
        mock_input.header = None
        mock_input.model_dump.return_value = {"query": "test"}

        module = get_process_tool_input_module()
        process_tool_input = module.process_tool_input

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):

                result, _ = await process_tool_input(mock_input)

                assert isinstance(result, dict)

    @pytest.mark.asyncio
    async def test_process_tool_input_with_span(self, mock_agent_input):
        """测试带span参数的处理"""
        mock_span = MagicMock()

        module = get_process_tool_input_module()
        process_tool_input = module.process_tool_input

        with patch.object(module, "span_set_attrs") as mock_span_attrs:
            with patch.object(module, "get_user_account_id", return_value="user123"):

                await process_tool_input(mock_agent_input, span=mock_span)

                mock_span_attrs.assert_called_once()

    @pytest.mark.asyncio
    async def test_process_tool_input_empty_input(self):
        """测试空输入处理"""
        mock_input = MagicMock()
        mock_input.header = {}
        mock_input.model_dump.return_value = {}

        module = get_process_tool_input_module()
        process_tool_input = module.process_tool_input

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value=""):

                result, event_key = await process_tool_input(mock_input)

                assert result == {}
                assert event_key is None
