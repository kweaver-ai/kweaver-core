import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, AsyncMock
from src.application.memory.manage_memory import ManageMemoryUseCase


class TestManageMemoryUseCase:
    @pytest.fixture
    def use_case(self):
        """Create a use case instance for testing"""
        return ManageMemoryUseCase()

    @pytest.mark.asyncio
    async def test_initialization(self, use_case):
        """Test use case initialization"""
        assert use_case.memory_adapter is None

        with patch.object(use_case, "memory_adapter") as mock_adapter:
            result = await use_case.initialize()
            assert result is use_case

    @pytest.mark.asyncio
    async def test_get_memory_success(self, use_case):
        """Test getting a memory successfully"""
        memory_id = "mem123"
        expected_memory = {
            "id": memory_id,
            "memory": "Test memory content",
            "created_at": "2024-01-01T00:00:00Z",
        }

        mock_adapter = AsyncMock()
        mock_adapter.get.return_value = expected_memory
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_memory(memory_id)

        mock_adapter.get.assert_called_once_with(memory_id)
        assert result == expected_memory

    @pytest.mark.asyncio
    async def test_get_memory_not_found(self, use_case):
        """Test getting a memory that doesn't exist"""
        memory_id = "nonexistent"

        mock_adapter = AsyncMock()
        mock_adapter.get.return_value = None
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_memory(memory_id)

        assert result is None
        mock_adapter.get.assert_called_once_with(memory_id)

    @pytest.mark.asyncio
    async def test_get_all_memories_with_user_id(self, use_case):
        """Test getting all memories with user_id"""
        user_id = "user123"
        expected_result = {
            "results": [
                {"id": "mem1", "memory": "Memory 1"},
                {"id": "mem2", "memory": "Memory 2"},
            ]
        }

        mock_adapter = AsyncMock()
        mock_adapter.get_all.return_value = expected_result
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_all_memories(user_id=user_id)

        mock_adapter.get_all.assert_called_once_with(
            user_id=user_id, agent_id=None, run_id=None, filters=None, limit=100
        )
        assert result == expected_result

    @pytest.mark.asyncio
    async def test_get_all_memories_with_all_filters(self, use_case):
        """Test getting all memories with all filters"""
        user_id = "user123"
        agent_id = "agent456"
        run_id = "run789"
        filters = {"tag": "important"}
        limit = 50

        mock_adapter = AsyncMock()
        mock_adapter.get_all.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_all_memories(
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            filters=filters,
            limit=limit,
        )

        mock_adapter.get_all.assert_called_once_with(
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            filters=filters,
            limit=limit,
        )

    @pytest.mark.asyncio
    async def test_get_all_memories_default_limit(self, use_case):
        """Test getting all memories with default limit"""
        mock_adapter = AsyncMock()
        mock_adapter.get_all.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_all_memories(user_id="user123")

        call_args = mock_adapter.get_all.call_args
        assert call_args.kwargs.get("limit") == 100

    @pytest.mark.asyncio
    async def test_update_memory_success(self, use_case):
        """Test updating a memory successfully"""
        memory_id = "mem123"
        data = "Updated memory content"
        expected_result = {
            "id": memory_id,
            "memory": data,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-02T00:00:00Z",
        }

        mock_adapter = AsyncMock()
        mock_adapter.update.return_value = expected_result
        use_case.memory_adapter = mock_adapter

        result = await use_case.update_memory(memory_id, data)

        mock_adapter.update.assert_called_once_with(memory_id, data)
        assert result == expected_result

    @pytest.mark.asyncio
    async def test_update_memory_failure(self, use_case):
        """Test updating a memory that fails"""
        memory_id = "mem123"
        data = "Updated content"

        mock_adapter = AsyncMock()
        mock_adapter.update.return_value = None
        use_case.memory_adapter = mock_adapter

        result = await use_case.update_memory(memory_id, data)

        assert result is None
        mock_adapter.update.assert_called_once_with(memory_id, data)

    @pytest.mark.asyncio
    async def test_delete_memory_success(self, use_case):
        """Test deleting a memory successfully"""
        memory_id = "mem123"

        mock_adapter = AsyncMock()
        mock_adapter.delete.return_value = True
        use_case.memory_adapter = mock_adapter

        result = await use_case.delete_memory(memory_id)

        assert result is True
        mock_adapter.delete.assert_called_once_with(memory_id)

    @pytest.mark.asyncio
    async def test_delete_memory_failure(self, use_case):
        """Test deleting a memory that fails"""
        memory_id = "mem123"

        mock_adapter = AsyncMock()
        mock_adapter.delete.return_value = False
        use_case.memory_adapter = mock_adapter

        result = await use_case.delete_memory(memory_id)

        assert result is False
        mock_adapter.delete.assert_called_once_with(memory_id)

    @pytest.mark.asyncio
    async def test_get_memory_history_success(self, use_case):
        """Test getting memory history successfully"""
        memory_id = "mem123"
        expected_history = [
            {
                "action": "create",
                "timestamp": "2024-01-01T00:00:00Z",
                "data": "initial",
            },
            {
                "action": "update",
                "timestamp": "2024-01-02T00:00:00Z",
                "data": "updated",
            },
        ]

        mock_adapter = AsyncMock()
        mock_adapter.history.return_value = expected_history
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_memory_history(memory_id)

        mock_adapter.history.assert_called_once_with(memory_id)
        assert result == expected_history

    @pytest.mark.asyncio
    async def test_get_memory_history_empty(self, use_case):
        """Test getting memory history when empty"""
        memory_id = "mem123"

        mock_adapter = AsyncMock()
        mock_adapter.history.return_value = []
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_memory_history(memory_id)

        assert result == []
        mock_adapter.history.assert_called_once_with(memory_id)

    @pytest.mark.asyncio
    async def test_initialize_without_prior_initialization(self, use_case):
        """Test methods initialize adapter if not initialized"""
        mock_adapter = AsyncMock()

        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
        ) as mock_create:
            call_count = [0]

            async def create_mock():
                call_count[0] += 1
                return mock_adapter

            mock_create.side_effect = create_mock
            mock_adapter.get.return_value = {"id": "mem123"}

            await use_case.get_memory("mem123")

            assert call_count[0] == 1

    @pytest.mark.asyncio
    async def test_multiple_operations_without_reinitialization(self, use_case):
        """Test that multiple operations don't cause reinitialization"""
        mock_adapter = AsyncMock()
        mock_adapter.get.return_value = {"id": "mem123"}
        mock_adapter.update.return_value = {"id": "mem123"}
        mock_adapter.delete.return_value = True

        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
        ) as mock_create:
            call_count = [0]

            async def create_mock():
                call_count[0] += 1
                return mock_adapter

            mock_create.side_effect = create_mock

            await use_case.get_memory("mem123")
            await use_case.update_memory("mem123", "new content")
            await use_case.delete_memory("mem456")

            assert call_count[0] == 1

    @pytest.mark.asyncio
    async def test_get_all_memories_with_none_filters(self, use_case):
        """Test get_all_memories with None filters"""
        mock_adapter = AsyncMock()
        mock_adapter.get_all.return_value = {"results": []}
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_all_memories(user_id="user123", filters=None)

        call_args = mock_adapter.get_all.call_args
        assert call_args.kwargs.get("filters") is None

    @pytest.mark.asyncio
    async def test_get_memory_with_empty_id(self, use_case):
        """Test get_memory with empty memory_id"""
        memory_id = ""

        mock_adapter = AsyncMock()
        mock_adapter.get.return_value = None
        use_case.memory_adapter = mock_adapter

        result = await use_case.get_memory(memory_id)

        assert result is None
        mock_adapter.get.assert_called_once_with("")

    @pytest.mark.asyncio
    async def test_get_memory_initializes_adapter(self, use_case):
        """Test get_memory initializes adapter when None"""
        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
            return_value=AsyncMock(get=AsyncMock(return_value={"id": "mem123"})),
        ):
            use_case.memory_adapter = None
            result = await use_case.get_memory("mem123")
            assert result is not None

    @pytest.mark.asyncio
    async def test_get_all_memories_initializes_adapter(self, use_case):
        """Test get_all_memories initializes adapter when None"""
        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
            return_value=AsyncMock(get_all=AsyncMock(return_value={"results": []})),
        ):
            use_case.memory_adapter = None
            result = await use_case.get_all_memories(user_id="user123")
            assert "results" in result

    @pytest.mark.asyncio
    async def test_update_memory_initializes_adapter(self, use_case):
        """Test update_memory initializes adapter when None"""
        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
            return_value=AsyncMock(update=AsyncMock(return_value={"id": "mem123"})),
        ):
            use_case.memory_adapter = None
            result = await use_case.update_memory("mem123", "data")
            assert result is not None

    @pytest.mark.asyncio
    async def test_delete_memory_initializes_adapter(self, use_case):
        """Test delete_memory initializes adapter when None"""
        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
            return_value=AsyncMock(delete=AsyncMock(return_value=True)),
        ):
            use_case.memory_adapter = None
            result = await use_case.delete_memory("mem123")
            assert result is True

    @pytest.mark.asyncio
    async def test_get_memory_history_initializes_adapter(self, use_case):
        """Test get_memory_history initializes adapter when None"""
        with patch(
            "src.application.memory.manage_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
            return_value=AsyncMock(history=AsyncMock(return_value=[])),
        ):
            use_case.memory_adapter = None
            result = await use_case.get_memory_history("mem123")
            assert isinstance(result, list)
