import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, AsyncMock
from src.application.memory.retrieval_memory import RetrievalMemoryUseCase


class TestRetrievalMemoryUseCase:
    @pytest.fixture
    def use_case(self):
        """Create a use case instance for testing"""
        return RetrievalMemoryUseCase()

    @pytest.mark.asyncio
    async def test_initialization(self, use_case):
        """Test use case initialization"""
        assert use_case.memory_adapter is None

        with patch.object(use_case, "memory_adapter") as mock_adapter:
            result = await use_case.initialize()
            assert result is use_case

    @pytest.mark.asyncio
    async def test_execute_with_query(self, use_case):
        """Test execute with basic query"""
        query = "What is AI?"

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {
            "results": [
                {"id": "mem1", "memory": "AI is artificial intelligence"},
                {"id": "mem2", "memory": "Machine learning is part of AI"},
            ]
        }
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query)

        mock_adapter.search.assert_called_once()
        assert len(result["results"]) == 2

    @pytest.mark.asyncio
    async def test_execute_with_all_parameters(self, use_case):
        """Test execute with all parameters"""
        query = "Test query"
        user_id = "user123"
        agent_id = "agent456"
        run_id = "run789"
        limit = 10
        filters = {"tag": "important"}
        threshold = 0.8
        rerank_threshold = 0.7
        context = {"user_id": "user123", "visitor_type": "realname"}

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(
            query=query,
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            limit=limit,
            filters=filters,
            threshold=threshold,
            rerank_threshold=rerank_threshold,
            context=context,
        )

        mock_adapter.search.assert_called_once_with(
            query=query,
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            limit=limit,
            filters=filters,
            threshold=threshold,
            rerank_threshold=rerank_threshold,
            context=context,
        )

    @pytest.mark.asyncio
    async def test_execute_without_initialization(self, use_case):
        """Test execute without prior initialization"""
        query = "Test query"

        with patch(
            "src.application.memory.retrieval_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
        ) as mock_create:
            mock_adapter = AsyncMock()
            mock_adapter.search.return_value = {"results": []}
            mock_create.return_value = mock_adapter

            result = await use_case.execute(query=query)

            mock_create.assert_called_once()
            assert "results" in result

    @pytest.mark.asyncio
    async def test_execute_with_minimal_parameters(self, use_case):
        """Test execute with minimal parameters"""
        query = "Test query"

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query)

        assert "results" in result
        mock_adapter.search.assert_called_once()

    @pytest.mark.asyncio
    async def test_execute_with_none_filters(self, use_case):
        """Test execute with None filters"""
        query = "Test query"

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query, filters=None)

        call_args = mock_adapter.search.call_args
        assert call_args.kwargs.get("filters") is None

    @pytest.mark.asyncio
    async def test_execute_with_none_threshold(self, use_case):
        """Test execute with None threshold"""
        query = "Test query"

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query, threshold=None)

        call_args = mock_adapter.search.call_args
        assert call_args.kwargs.get("threshold") is None

    @pytest.mark.asyncio
    async def test_execute_with_none_rerank_threshold(self, use_case):
        """Test execute with None rerank_threshold"""
        query = "Test query"

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query, rerank_threshold=None)

        call_args = mock_adapter.search.call_args
        assert call_args.kwargs.get("rerank_threshold") is None

    @pytest.mark.asyncio
    async def test_execute_return_value(self, use_case):
        """Test execute returns adapter result"""
        query = "Test query"
        expected_result = {
            "results": [
                {"id": "mem1", "memory": "Result 1"},
                {"id": "mem2", "memory": "Result 2"},
            ]
        }

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = expected_result
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query)

        assert result == expected_result

    @pytest.mark.asyncio
    async def test_execute_adapter_error_handling(self, use_case):
        """Test execute handles adapter errors"""
        query = "Test query"

        mock_adapter = AsyncMock()
        mock_adapter.search.side_effect = Exception("Search error")
        use_case.memory_adapter = mock_adapter

        with pytest.raises(Exception) as exc:
            await use_case.execute(query=query)

        assert "Search error" in str(exc.value)

    @pytest.mark.asyncio
    async def test_initialize_called_multiple_times(self, use_case):
        """Test that initialization can be called multiple times safely"""
        mock_adapter = AsyncMock()

        with patch(
            "src.application.memory.retrieval_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
        ) as mock_create:
            call_count = [0]

            async def create_mock():
                call_count[0] += 1
                return mock_adapter

            mock_create.side_effect = create_mock

            await use_case.initialize()
            await use_case.initialize()

            assert call_count[0] == 1
            assert use_case.memory_adapter is mock_adapter

    @pytest.mark.asyncio
    async def test_execute_with_empty_query(self, use_case):
        """Test execute with empty query string"""
        query = ""

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query)

        mock_adapter.search.assert_called_once_with(
            query="",
            limit=5,
            filters=None,
            threshold=None,
            rerank_threshold=None,
            user_id=None,
            agent_id=None,
            run_id=None,
            context=None,
        )

    @pytest.mark.asyncio
    async def test_execute_with_context(self, use_case):
        """Test execute with context parameter"""
        query = "Test query"
        context = {"user_id": "user123", "visitor_type": "anonymous"}

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query, context=context)

        call_args = mock_adapter.search.call_args
        assert call_args.kwargs.get("context") == context

    @pytest.mark.asyncio
    async def test_execute_with_custom_limit(self, use_case):
        """Test execute with custom limit"""
        query = "Test query"
        limit = 20

        mock_adapter = AsyncMock()
        mock_adapter.search.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(query=query, limit=limit)

        call_args = mock_adapter.search.call_args
        assert call_args.kwargs.get("limit") == 20
