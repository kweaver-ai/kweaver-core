import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, AsyncMock, MagicMock
from src.application.memory.build_memory import BuildMemoryUseCase


class TestBuildMemoryUseCase:
    @pytest.fixture
    def use_case(self):
        """Create a use case instance for testing"""
        return BuildMemoryUseCase()

    @pytest.mark.asyncio
    async def test_initialization(self, use_case):
        """Test use case initialization"""
        assert use_case.memory_adapter is None

        with patch.object(use_case, "memory_adapter") as mock_adapter:
            result = await use_case.initialize()
            assert result is use_case

    @pytest.mark.asyncio
    async def test_execute_with_messages(self, use_case):
        """Test execute with messages"""
        messages = [
            {"role": "user", "content": "Hello"},
            {"role": "assistant", "content": "Hi there"},
        ]

        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = {"id": "mem123"}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(messages=messages)

        mock_adapter.add.assert_called_once()
        assert result["id"] == "mem123"

    @pytest.mark.asyncio
    async def test_execute_with_all_parameters(self, use_case):
        """Test execute with all parameters"""
        messages = [{"role": "user", "content": "Test"}]
        user_id = "user123"
        agent_id = "agent456"
        run_id = "run789"
        metadata = {"key": "value"}
        infer = True
        memory_type = "episodic"
        prompt = "Test prompt"
        context = {"user_id": "user123", "visitor_type": "realname"}

        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = {"id": "mem123"}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(
            messages=messages,
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            metadata=metadata,
            infer=infer,
            memory_type=memory_type,
            prompt=prompt,
            context=context,
        )

        mock_adapter.add.assert_called_once_with(
            messages=messages,
            user_id=user_id,
            agent_id=agent_id,
            run_id=run_id,
            metadata=metadata,
            infer=infer,
            memory_type=memory_type,
            prompt=prompt,
            context=context,
        )

    @pytest.mark.asyncio
    async def test_execute_without_initialization(self, use_case):
        """Test execute without prior initialization"""
        messages = [{"role": "user", "content": "Test"}]

        with patch(
            "src.application.memory.build_memory.Mem0MemoryAdapter.create",
            new_callable=AsyncMock,
        ) as mock_create:
            mock_adapter = AsyncMock()
            mock_adapter.add.return_value = {"id": "mem123"}
            mock_create.return_value = mock_adapter

            result = await use_case.execute(messages=messages)

            mock_create.assert_called_once()
            assert result["id"] == "mem123"

    @pytest.mark.asyncio
    async def test_execute_with_minimal_parameters(self, use_case):
        """Test execute with minimal parameters"""
        messages = [{"role": "user", "content": "Test"}]

        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = {"id": "mem123"}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(messages=messages)

        assert "id" in result
        mock_adapter.add.assert_called_once()

    @pytest.mark.asyncio
    async def test_initialize_called_multiple_times(self, use_case):
        """Test that initialization can be called multiple times safely"""
        mock_adapter = AsyncMock()

        with patch(
            "src.application.memory.build_memory.Mem0MemoryAdapter.create",
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
    async def test_execute_adapter_error_handling(self, use_case):
        """Test execute handles adapter errors"""
        messages = [{"role": "user", "content": "Test"}]

        mock_adapter = AsyncMock()
        mock_adapter.add.side_effect = Exception("Adapter error")
        use_case.memory_adapter = mock_adapter

        with pytest.raises(Exception) as exc:
            await use_case.execute(messages=messages)

        assert "Adapter error" in str(exc.value)

    @pytest.mark.asyncio
    async def test_execute_with_none_metadata(self, use_case):
        """Test execute with None metadata"""
        messages = [{"role": "user", "content": "Test"}]

        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = {"id": "mem123"}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(messages=messages, metadata=None)

        call_args = mock_adapter.add.call_args
        assert call_args.kwargs.get("metadata") is None

    @pytest.mark.asyncio
    async def test_execute_with_empty_messages(self, use_case):
        """Test execute with empty messages list"""
        messages = []

        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = {"id": "mem123"}
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(messages=messages)

        assert mock_adapter.add.called
        assert result["id"] == "mem123"

    @pytest.mark.asyncio
    async def test_execute_return_value(self, use_case):
        """Test execute returns adapter result"""
        messages = [{"role": "user", "content": "Test"}]
        expected_result = {
            "id": "mem123",
            "memory": "Test memory",
            "created_at": "2024-01-01T00:00:00Z",
        }

        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = expected_result
        use_case.memory_adapter = mock_adapter

        result = await use_case.execute(messages=messages)

        assert result == expected_result
