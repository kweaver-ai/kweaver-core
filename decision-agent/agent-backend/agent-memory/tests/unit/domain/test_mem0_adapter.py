import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, AsyncMock, MagicMock
from src.domain.memory.mem0_adapter import Mem0MemoryAdapter


@pytest.fixture(autouse=True)
def reset_singleton():
    """Reset singleton before each test"""
    Mem0MemoryAdapter._instance = None
    yield


class TestMem0MemoryAdapter:
    @pytest.fixture
    def mock_memory(self):
        """Mock AsyncMemory"""
        memory = MagicMock()
        memory.add = AsyncMock(return_value={"id": "mem123"})
        memory.search = AsyncMock(return_value={"results": []})
        memory.get = AsyncMock(return_value={"id": "mem123"})
        memory.get_all = AsyncMock(return_value={"results": []})
        memory.update = AsyncMock(return_value={"id": "mem123"})
        memory.delete = AsyncMock(return_value=True)
        memory.history = AsyncMock(return_value=[])
        return memory

    @pytest.fixture
    def mock_rerank_client(self):
        """Mock rerank client"""
        client = MagicMock()
        client.rerank = AsyncMock()
        client.rerank_with_threshold = AsyncMock()
        return client

    @pytest.mark.asyncio
    async def test_singleton_pattern(self):
        """Test that Mem0MemoryAdapter implements singleton"""
        from src.domain.memory.mem0_adapter import Mem0MemoryAdapter

        Mem0MemoryAdapter._instance = None
        with patch("src.domain.memory.mem0_adapter.memory_config", MagicMock()):
            with patch("src.domain.memory.mem0_adapter.rerank_config", MagicMock()):
                adapter1 = Mem0MemoryAdapter()
                adapter2 = Mem0MemoryAdapter()

                assert adapter1 is adapter2

    @pytest.mark.asyncio
    async def test_initialized_flag(self, mock_memory, mock_rerank_client):
        """Test that initialization flag prevents reinitialization"""
        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                from src.domain.memory.mem0_adapter import Mem0MemoryAdapter

                instance = await Mem0MemoryAdapter.create()
                initial_memory = instance.memory
                initial_rerank = instance.rerank_client

                assert initial_memory is mock_memory
                assert initial_rerank is mock_rerank_client

    @pytest.mark.asyncio
    async def test_add_without_context(self, mock_memory, mock_rerank_client):
        """Test add method without context"""
        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                messages = [{"role": "user", "content": "Hello"}]
                await adapter.add(messages)

                mock_memory.add.assert_called_once()
                assert mock_memory.add.call_args[0][0] == messages
                assert mock_memory.add.call_args[1]["context"] is None

    @pytest.mark.asyncio
    async def test_add_with_context(self, mock_memory, mock_rerank_client):
        """Test add method with context"""
        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                messages = [{"role": "user", "content": "Hello"}]
                context = {"user_id": "user123", "visitor_type": "realname"}
                await adapter.add(messages, context=context)

                mock_memory.add.assert_called_once()
                assert mock_memory.add.call_args[0][0] == messages
                assert mock_memory.add.call_args[1]["context"] == context

    @pytest.mark.asyncio
    async def test_add_with_all_parameters(self, mock_memory, mock_rerank_client):
        """Test add method with all parameters"""
        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                messages = [{"role": "user", "content": "Hello"}]
                context = {"user_id": "user123"}
                await adapter.add(
                    messages,
                    user_id="user123",
                    agent_id="agent456",
                    run_id="run789",
                    metadata={"key": "value"},
                    infer=True,
                    memory_type="episodic",
                    prompt="Test prompt",
                    context=context,
                )

                assert mock_memory.add.called

    @pytest.mark.asyncio
    async def test_search_without_results(self, mock_memory, mock_rerank_client):
        """Test search with no results"""
        mock_memory.search.return_value = {"results": []}

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.search("query")

                assert result == {"results": []}
                mock_rerank_client.rerank_with_threshold.assert_not_called()

    @pytest.mark.asyncio
    async def test_search_with_results(self, mock_memory, mock_rerank_client):
        """Test search with results and reranking"""
        from src.adaptee.mf_model_factory.rerank_model_client import (
            RerankResponseUsage,
            RerankResponseResult,
            RerankResponse,
        )

        search_results = {
            "results": [
                {"id": "mem1", "memory": "Memory 1", "score": 0.8},
                {"id": "mem2", "memory": "Memory 2", "score": 0.7},
                {"id": "mem3", "memory": "Memory 3", "score": 0.6},
            ]
        }

        mock_memory.search.return_value = search_results

        rerank_response = RerankResponse(
            id="rerank_id",
            object="rerank",
            model="reranker",
            usage=RerankResponseUsage(prompt_tokens=10, total_tokens=20),
            results=[
                RerankResponseResult(relevance_score=0.9, index=0, document="Memory 1"),
                RerankResponseResult(relevance_score=0.5, index=1, document="Memory 2"),
                RerankResponseResult(relevance_score=0.7, index=2, document="Memory 3"),
            ],
            created=1234567890,
        )
        mock_rerank_client.rerank_with_threshold = AsyncMock(
            return_value=rerank_response
        )

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.search("query", rerank_threshold=0.7)

                assert "results" in result
                assert len(result["results"]) == 3
                assert result["results"][0]["rerank_score"] == 0.9
                assert result["results"][1]["rerank_score"] == 0.7
                assert result["results"][2]["rerank_score"] == 0.5

    @pytest.mark.asyncio
    async def test_search_without_rerank(self, mock_memory, mock_rerank_client):
        """Test search without reranking (threshold=0)"""
        mock_memory.search.return_value = {
            "results": [{"id": "mem1", "memory": "Test"}]
        }

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.search("query", rerank_threshold=0.0)

                assert "results" in result

    @pytest.mark.asyncio
    async def test_search_with_context(self, mock_memory, mock_rerank_client):
        """Test search with context"""
        mock_memory.search.return_value = {"results": []}

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                context = {"user_id": "user123"}
                result = await adapter.search("query", context=context)

                mock_memory.search.assert_called_once()
                assert mock_memory.search.call_args[0][0] == "query"
                assert mock_memory.search.call_args[1]["context"] == context
                mock_rerank_client.rerank_with_threshold.assert_not_called()
                assert result == {"results": []}

    @pytest.mark.asyncio
    async def test_search_rerank_error_handling(self, mock_memory, mock_rerank_client):
        """Test search handles rerank errors gracefully"""
        search_results = {
            "results": [
                {"id": "mem1", "memory": "Memory 1"},
                {"id": "mem2", "memory": "Memory 2"},
            ]
        }

        mock_memory.search.return_value = search_results
        mock_rerank_client.rerank_with_threshold.side_effect = Exception(
            "Rerank failed"
        )

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.search("query", rerank_threshold=0.5)

                assert result == search_results

    @pytest.mark.asyncio
    async def test_get_memory(self, mock_memory, mock_rerank_client):
        """Test getting a memory by ID"""
        mock_memory.get.return_value = {"id": "mem123", "memory": "Test"}

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.get("mem123")

                mock_memory.get.assert_called_once_with("mem123")
                assert result["id"] == "mem123"

    @pytest.mark.asyncio
    async def test_get_all_memories(self, mock_memory, mock_rerank_client):
        """Test getting all memories"""
        mock_memory.get_all.return_value = {"results": [{"id": "mem1"}]}

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.get_all(user_id="user123", limit=10)

                mock_memory.get_all.assert_called_once_with(user_id="user123", limit=10)
                assert "results" in result

    @pytest.mark.asyncio
    async def test_update_memory(self, mock_memory, mock_rerank_client):
        """Test updating a memory"""
        mock_memory.update.return_value = {"id": "mem123", "memory": "Updated"}

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.update("mem123", "New content")

                mock_memory.update.assert_called_once_with("mem123", "New content")
                assert result["memory"] == "Updated"

    @pytest.mark.asyncio
    async def test_delete_memory(self, mock_memory, mock_rerank_client):
        """Test deleting a memory"""
        mock_memory.delete.return_value = True

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.delete("mem123")

                mock_memory.delete.assert_called_once_with("mem123")
                assert result is True

    @pytest.mark.asyncio
    async def test_get_memory_history(self, mock_memory, mock_rerank_client):
        """Test getting memory history"""
        history = [
            {"action": "create", "timestamp": "2024-01-01T00:00:00Z"},
            {"action": "update", "timestamp": "2024-01-02T00:00:00Z"},
        ]
        mock_memory.history.return_value = history

        with patch(
            "src.domain.memory.mem0_adapter.AsyncMemory", return_value=mock_memory
        ):
            with patch(
                "src.domain.memory.mem0_adapter.RerankModelClient",
                return_value=mock_rerank_client,
            ):
                adapter = await Mem0MemoryAdapter.create()

                result = await adapter.history("mem123")

                mock_memory.history.assert_called_once_with("mem123")
                assert len(result) == 2
                assert result[0]["action"] == "create"
