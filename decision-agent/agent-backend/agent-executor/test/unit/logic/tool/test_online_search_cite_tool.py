"""Tests for app.logictool.online_search_cite_tool module."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


@pytest.mark.asyncio
class TestGetSearchResults:
    """Tests for get_search_results function."""

    async def test_get_search_results_zhipu_success(self):
        """Test successful search with zhipu_search_tool."""
        from app.logic.tool.online_search_cite_tool import get_search_results

        request = {
            "search_tool": "zhipu_search_tool",
            "query": "test query",
            "api_key": "test_key",
        }
        headers = {"userid": "test_user"}

        mock_response = {
            "choices": [{"message": {"tool_calls": [{"search_result": []}]}}],
            "id": "test_id",
        }

        # Create proper async context manager mocks
        mock_response_obj = AsyncMock()
        mock_response_obj.status = 200
        mock_response_obj.json = AsyncMock(return_value=mock_response)
        mock_response_obj.text = AsyncMock(return_value="")

        mock_post = MagicMock()
        mock_post.__aenter__ = AsyncMock(return_value=mock_response_obj)
        mock_post.__aexit__ = AsyncMock(return_value=None)

        mock_session = MagicMock()
        mock_session.post = MagicMock(return_value=mock_post)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await get_search_results(request, headers)
            assert result == mock_response

    async def test_get_search_results_unsupported_tool(self):
        """Test get_search_results with unsupported search tool."""
        from app.logic.tool.online_search_cite_tool import get_search_results

        request = {"search_tool": "unsupported_tool", "query": "test query"}
        headers = {}

        with pytest.raises(ValueError, match="不支持的搜索工具"):
            await get_search_results(request, headers)


@pytest.mark.asyncio
class TestGetAnswer:
    """Tests for get_answer function."""

    async def test_get_answer_success(self):
        """Test successful answer generation."""
        from app.logic.tool.online_search_cite_tool import get_answer

        request = {"query": "test query", "model_name": "test_model"}
        headers = {"userid": "test_user"}

        search_results = {
            "choices": [
                {
                    "message": {
                        "tool_calls": [
                            {},  # First tool_call (ignored)
                            {
                                "search_result": [
                                    {
                                        "title": "Test Title",
                                        "content": "Test Content",
                                        "link": "http://test.com",
                                    }
                                ]
                            },
                        ]
                    }
                }
            ]
        }

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(return_value="Test answer")

            result, references = await get_answer(request, headers, search_results)

            assert result == "Test answer"
            assert len(references) == 1
            assert references[0]["title"] == "Test Title"
            assert references[0]["content"] == "Test Content"
            assert references[0]["index"] == 0

    async def test_get_answer_missing_fields(self):
        """Test get_answer with missing fields in search results."""
        from app.logic.tool.online_search_cite_tool import get_answer

        request = {"query": "test query", "model_name": "test_model"}
        headers = {"userid": "test_user"}

        search_results = {
            "choices": [
                {
                    "message": {
                        "tool_calls": [
                            {},
                            {
                                "search_result": [
                                    {
                                        "content": "Test Content"
                                    }  # Missing title and link
                                ]
                            },
                        ]
                    }
                }
            ]
        }

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(return_value="Test answer")

            result, references = await get_answer(request, headers, search_results)

            assert len(references) == 1
            assert references[0]["title"] == "未知标题"
            assert references[0]["link"] == ""

    async def test_get_answer_invalid_structure(self):
        """Test get_answer with invalid search_results structure."""
        from app.logic.tool.online_search_cite_tool import get_answer

        request = {"query": "test", "model_name": "test_model"}
        headers = {}

        # Missing choices key
        search_results = {}

        with pytest.raises(ValueError, match="Invalid search_results structure"):
            await get_answer(request, headers, search_results)


@pytest.mark.asyncio
class TestDataPreAnswer:
    """Tests for data_pre_answer function."""

    async def test_data_pre_answer_success(self):
        """Test successful data pre-processing."""
        from app.logic.tool.online_search_cite_tool import data_pre_answer

        answer = "Test answer"
        references = [
            {
                "title": "Test",
                "content": "Content",
                "link": "http://test.com",
                "index": 0,
            }
        ]

        result, refs_str = await data_pre_answer(answer, references)

        assert result == "Test answer"
        assert isinstance(refs_str, str)
        assert "Test" in refs_str


@pytest.mark.asyncio
class TestGetCompletion:
    """Tests for get_completion function."""

    async def test_get_completion_success(self):
        """Test successful completion generation."""
        from app.logic.tool.online_search_cite_tool import get_completion

        request = {"model_name": "test_model"}
        headers = {"userid": "test_user"}
        answer = "Test answer"
        references = [{"title": "Test", "content": "Content"}]

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(
                return_value="Completed answer with citations"
            )

            result = await get_completion(request, headers, answer, references)

            assert result == "Completed answer with citations"
            mock_service.call.assert_called_once()


@pytest.mark.asyncio
class TestGetCompletionStream:
    """Tests for get_completion_stream function."""

    async def test_get_completion_stream_success(self):
        """Test successful streaming completion generation."""
        from app.logic.tool.online_search_cite_tool import get_completion_stream

        request = {"model_name": "test_model"}
        headers = {"userid": "test_user"}
        answer = "Test answer"
        references = [{"title": "Test", "content": "Content"}]

        async def mock_stream(**kwargs):
            yield "chunk1"
            yield "chunk2"

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call_stream = mock_stream

            chunks = []
            async for chunk in get_completion_stream(
                request, headers, answer, references
            ):
                chunks.append(chunk)

            assert chunks == ["chunk1", "chunk2"]


@pytest.mark.asyncio
class TestOnlineSearchCiteTool:
    """Tests for online_search_cite_tool function."""

    async def test_online_search_cite_tool_success(self):
        """Test successful online search with citations."""
        from app.logic.tool.online_search_cite_tool import online_search_cite_tool

        request = {
            "query": "test query",
            "model_name": "test_model",
            "search_tool": "zhipu_search_tool",
            "api_key": "test_key",
        }
        headers = {"userid": "test_user"}

        mock_search_results = {
            "choices": [
                {
                    "message": {
                        "tool_calls": [
                            {},
                            {
                                "search_result": [
                                    {"title": "Test", "content": "Content"}
                                ]
                            },
                        ]
                    }
                }
            ]
        }

        with (
            patch(
                "app.logic.tool.online_search_cite_tool.get_search_results",
                new_callable=AsyncMock,
            ) as mock_get_search,
            patch(
                "app.logic.tool.online_search_cite_tool.get_answer",
                new_callable=AsyncMock,
            ) as mock_get_ans,
            patch(
                "app.logic.tool.online_search_cite_tool.get_completion",
                new_callable=AsyncMock,
            ) as mock_get_comp,
        ):
            mock_get_search.return_value = mock_search_results
            mock_get_ans.return_value = ("answer", [{"title": "Test"}])
            mock_get_comp.return_value = "completed answer"

            result = await online_search_cite_tool(request, headers)

            assert "answer" in result
            assert "references" in result
            assert result["answer"] == "completed answer"


class TestPrompts:
    """Tests for prompt constants."""

    def test_answer_prompt_exists(self):
        """Test that answer_prompt is defined."""
        from app.logic.tool.online_search_cite_tool import answer_prompt

        assert isinstance(answer_prompt, str)
        assert "references" in answer_prompt
        assert "query" in answer_prompt

    def test_sys_prompt_exists(self):
        """Test that sys_prompt is defined."""
        from app.logic.tool.online_search_cite_tool import sys_prompt

        assert isinstance(sys_prompt, str)
        assert "example" in sys_prompt
        assert "references" in sys_prompt
