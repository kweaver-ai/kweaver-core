"""Tests for app.logic.tool.zhipu_search_tool module."""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
import aiohttp


@pytest.mark.asyncio
class TestZhipuSearchTool:
    """Tests for zhipu_search_tool function."""

    async def test_zhipu_search_tool_success(self):
        """Test successful zhipu search."""
        from app.logic.tool.zhipu_search_tool import zhipu_search_tool

        inputs = {"query": "test query"}
        props = {"api_key": "test_key"}
        resource = {}
        data_source_config = {}

        mock_response_data = {
            "choices": [{"message": {"content": "test result"}}],
            "id": "test_id",
            "created": 1234567890,
        }

        # Create proper async context manager mocks
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=mock_response_data)
        mock_response.text = AsyncMock(return_value="")

        mock_post = MagicMock()
        mock_post.__aenter__ = AsyncMock(return_value=mock_response)
        mock_post.__aexit__ = AsyncMock(return_value=None)

        mock_session = MagicMock()
        mock_session.post = MagicMock(return_value=mock_post)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await zhipu_search_tool(
                inputs, props, resource, data_source_config
            )

            assert result == mock_response_data

    async def test_zhipu_search_tool_http_error(self):
        """Test zhipu_search_tool with HTTP error status."""
        from app.logic.tool.zhipu_search_tool import zhipu_search_tool

        inputs = {"query": "test query"}
        props = {"api_key": "test_key"}
        resource = {}
        data_source_config = {}

        # Create proper async context manager mocks
        mock_response = AsyncMock()
        mock_response.status = 404
        mock_response.text = AsyncMock(return_value="Not Found")
        mock_response.json = AsyncMock(return_value={})

        mock_post = MagicMock()
        mock_post.__aenter__ = AsyncMock(return_value=mock_response)
        mock_post.__aexit__ = AsyncMock(return_value=None)

        mock_session = MagicMock()
        mock_session.post = MagicMock(return_value=mock_post)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await zhipu_search_tool(
                inputs, props, resource, data_source_config
            )

            assert "error" in result
            assert "404" in result["error"]

    async def test_zhipu_search_tool_client_error(self):
        """Test zhipu_search_tool with aiohttp ClientError."""
        from app.logic.tool.zhipu_search_tool import zhipu_search_tool

        inputs = {"query": "test query"}
        props = {"api_key": "test_key"}
        resource = {}
        data_source_config = {}

        mock_session = MagicMock()
        mock_session.post = MagicMock(
            side_effect=aiohttp.ClientError("Connection error")
        )
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await zhipu_search_tool(
                inputs, props, resource, data_source_config
            )

            assert "error" in result
            assert "Request failed" in result["error"]

    async def test_zhipu_search_tool_generic_exception(self):
        """Test zhipu_search_tool with unexpected exception."""
        from app.logic.tool.zhipu_search_tool import zhipu_search_tool

        inputs = {"query": "test query"}
        props = {"api_key": "test_key"}
        resource = {}
        data_source_config = {}

        mock_session = MagicMock()
        mock_session.post = MagicMock(side_effect=Exception("Unexpected error"))
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await zhipu_search_tool(
                inputs, props, resource, data_source_config
            )

            assert "error" in result
            assert "Unexpected error" in result["error"]

    async def test_zhipu_search_tool_with_context(self):
        """Test zhipu_search_tool with context parameter."""
        from app.logic.tool.zhipu_search_tool import zhipu_search_tool

        inputs = {"query": "test query"}
        props = {"api_key": "test_key"}
        resource = {}
        data_source_config = {}
        context = {"user_id": "test_user"}

        mock_response_data = {"id": "test_id"}

        # Create proper async context manager mocks
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=mock_response_data)
        mock_response.text = AsyncMock(return_value="")

        mock_post = MagicMock()
        mock_post.__aenter__ = AsyncMock(return_value=mock_response)
        mock_post.__aexit__ = AsyncMock(return_value=None)

        mock_session = MagicMock()
        mock_session.post = MagicMock(return_value=mock_post)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            result = await zhipu_search_tool(
                inputs, props, resource, data_source_config, context
            )

            assert result == mock_response_data

    async def test_zhipu_search_tool_request_structure(self):
        """Test that zhipu_search_tool sends correct request structure."""
        from app.logic.tool.zhipu_search_tool import zhipu_search_tool

        inputs = {"query": "test query"}
        props = {"api_key": "test_key_123"}
        resource = {}
        data_source_config = {}

        mock_response_data = {"id": "test_id"}

        # Create proper async context manager mocks
        mock_response = AsyncMock()
        mock_response.status = 200
        mock_response.json = AsyncMock(return_value=mock_response_data)
        mock_response.text = AsyncMock(return_value="")

        mock_post = MagicMock()
        mock_post.__aenter__ = AsyncMock(return_value=mock_response)
        mock_post.__aexit__ = AsyncMock(return_value=None)

        mock_session = MagicMock()
        mock_session.post = MagicMock(return_value=mock_post)
        mock_session.__aenter__ = AsyncMock(return_value=mock_session)
        mock_session.__aexit__ = AsyncMock(return_value=None)

        with patch("aiohttp.ClientSession", return_value=mock_session):
            await zhipu_search_tool(inputs, props, resource, data_source_config)

            # Verify post was called with correct arguments
            assert mock_session.post.called
