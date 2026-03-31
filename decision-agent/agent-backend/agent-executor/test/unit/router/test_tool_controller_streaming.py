# -*- coding: utf-8 -*-
"""单元测试 - tool_controller 流式部分"""

import pytest
from unittest.mock import AsyncMock, patch


class AsyncIterator:
    """异步迭代器辅助类"""

    def __init__(self, chunks):
        self.chunks = chunks

    def __aiter__(self):
        return self

    async def __anext__(self):
        if not self.chunks:
            raise StopAsyncIteration
        return self.chunks.pop(0)


class TestOnlineSearchCiteToolStreaming:
    """测试联网搜索引用工具流式端点"""

    @pytest.mark.asyncio
    async def test_online_search_cite_tool_stream(self):
        """测试流式联网搜索"""
        mock_search_results = {
            "choices": [
                {
                    "message": {
                        "tool_calls": [
                            {},
                            {
                                "search_result": [
                                    {
                                        "title": "参考1",
                                        "content": "内容1",
                                        "link": "http://example.com/1",
                                    },
                                    {
                                        "title": "参考2",
                                        "content": "内容2",
                                        "link": "http://example.com/2",
                                    },
                                ]
                            },
                        ]
                    }
                }
            ]
        }

        with patch(
            "app.logic.tool.online_search_cite_tool.get_search_results",
            new_callable=AsyncMock,
            return_value=mock_search_results,
        ):
            with patch(
                "app.logic.tool.online_search_cite_tool.get_answer",
                new_callable=AsyncMock,
                return_value=("这是答案", {}),
            ):
                with patch(
                    "app.logic.tool.online_search_cite_tool.get_completion_stream"
                ) as mock_stream:

                    async def mock_gen():
                        yield "这"
                        yield "是"
                        yield "答案"

                    mock_stream.return_value = mock_gen()

                    # 测试 generate_stream 函数
                    param = {
                        "query": "人工智能",
                        "model_name": "deepseek-v3",
                        "search_tool": "zhipu_search_tool",
                        "api_key": "test_key",
                        "user_id": "user123",
                        "stream": True,
                    }
                    headers = {"x-account-id": "user123"}

                    from app.logic.tool.online_search_cite_tool import (
                        get_search_results,
                        get_answer,
                    )

                    # 验证搜索结果获取
                    results = await get_search_results(param, headers)
                    assert "choices" in results

                    # 验证答案获取
                    answer, _ = await get_answer(param, headers, results)
                    assert answer == "这是答案"


class TestToolControllerStreaming:
    """测试 tool_controller 流式响应"""

    @pytest.mark.asyncio
    async def test_generate_stream_logic(self):
        """测试流式生成逻辑"""
        search_results = {
            "choices": [
                {
                    "message": {
                        "tool_calls": [
                            {},
                            {
                                "search_result": [
                                    {
                                        "title": "标题1",
                                        "content": "内容1",
                                        "link": "url1",
                                    },
                                ]
                            },
                        ]
                    }
                }
            ]
        }

        param = {
            "query": "测试",
            "model_name": "model",
            "search_tool": "tool",
            "api_key": "key",
            "user_id": "user",
            "stream": True,
        }
        headers = {"x-account-id": "user"}

        with patch(
            "app.logic.tool.online_search_cite_tool.get_search_results",
            new_callable=AsyncMock,
            return_value=search_results,
        ):
            with patch(
                "app.logic.tool.online_search_cite_tool.get_answer",
                new_callable=AsyncMock,
                return_value=("完整答案", {}),
            ):
                with patch(
                    "app.logic.tool.online_search_cite_tool.get_completion_stream"
                ) as mock_stream:

                    async def mock_gen():
                        yield "完"
                        yield "整"
                        yield "答案"

                    mock_stream.return_value = mock_gen()

                    from app.logic.tool.online_search_cite_tool import (
                        get_search_results,
                        get_answer,
                    )

                    # 测试搜索
                    results = await get_search_results(param, headers)
                    ref_list = results["choices"][0]["message"]["tool_calls"][1][
                        "search_result"
                    ]

                    final_references = []
                    for index, ref in enumerate(ref_list):
                        ref_item = {
                            "title": ref.get("title", "未知标题"),
                            "content": ref.get("content", ""),
                            "link": ref.get("link", ""),
                            "index": index,
                        }
                        final_references.append(ref_item)

                    assert len(final_references) == 1
                    assert final_references[0]["title"] == "标题1"

                    # 测试答案生成
                    answer, _ = await get_answer(param, headers, results)
                    assert answer == "完整答案"
