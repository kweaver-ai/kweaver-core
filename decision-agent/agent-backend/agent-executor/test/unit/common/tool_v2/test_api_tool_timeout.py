# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/api_tool.py 超时策略"""

from unittest.mock import MagicMock, patch

import pytest


class _FakeResponseContext:
    async def __aenter__(self):
        return object()

    async def __aexit__(self, exc_type, exc, tb):
        return False


def _build_tool(tool_timeout):
    from app.common.tool_v2.api_tool import APITool

    tool_info = {"name": "test_tool", "metadata": {"api_spec": {}}}
    tool_config = {
        "tool_input": [],
        "tool_timeout": tool_timeout,
        "tool_box_id": "box-1",
        "tool_id": "tool-1",
        "HOST_AGENT_OPERATOR": "localhost",
        "PORT_AGENT_OPERATOR": "9000",
    }

    tool = APITool(tool_info, tool_config)
    tool.process_params = MagicMock(
        return_value=({}, {"q": "1"}, {"foo": "bar"}, {"Auth": "x"})
    )

    return tool


async def _run_arun_stream(tool):
    captured = {}

    class FakeClientSession:
        def __init__(self, *args, **kwargs):
            captured["client_timeout"] = kwargs["timeout"]

        async def __aenter__(self):
            return self

        async def __aexit__(self, exc_type, exc, tb):
            return False

        def request(self, method, url, headers=None, json=None, verify_ssl=None):
            captured["method"] = method
            captured["url"] = url
            captured["headers"] = headers
            captured["json"] = json
            captured["verify_ssl"] = verify_ssl
            return _FakeResponseContext()

    async def fake_handle_response(response, total_timeout):
        captured["handle_response_timeout"] = total_timeout
        yield {"answer": "ok", "block_answer": ""}

    tool.handle_response = fake_handle_response

    with (
        patch("app.common.tool_v2.api_tool.StandLogger"),
        patch("app.common.tool_v2.api_tool.aiohttp.ClientSession", FakeClientSession),
        patch("app.common.tool_v2.api_tool.is_aaron_local_dev", return_value=False),
    ):
        result = [
            item
            async for item in tool.arun_stream(
                tool_input={"query": "hello"},
                props={"gvp": None},
            )
        ]

    return captured, result


class TestAPIToolTimeout:
    @pytest.mark.asyncio
    async def test_arun_stream_extends_client_timeout_only(self):
        tool = _build_tool(120)

        captured, result = await _run_arun_stream(tool)

        assert result == [{"answer": "ok", "block_answer": ""}]
        assert captured["json"]["timeout"] == 120
        assert captured["client_timeout"].total == 150
        assert captured["client_timeout"].sock_read == 150
        assert captured["handle_response_timeout"] == 150

    @pytest.mark.asyncio
    async def test_arun_stream_invalid_timeout_falls_back_then_extends(self):
        tool = _build_tool(0)

        captured, result = await _run_arun_stream(tool)

        assert result == [{"answer": "ok", "block_answer": ""}]
        assert captured["json"]["timeout"] == 300
        assert captured["client_timeout"].total == 330
        assert captured["client_timeout"].sock_read == 330
        assert captured["handle_response_timeout"] == 330
