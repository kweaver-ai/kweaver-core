import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from abc import ABC
from src.domain.memory.repositories import MemoryRepository, MemoryChunkRepository
from src.domain.memory.entities import Memory, MemoryChunk


class TestMemoryRepository:
    def test_memory_repository_is_abstract(self):
        """Test that MemoryRepository is an abstract base class"""
        assert issubclass(MemoryRepository, ABC)

    def test_memory_repository_has_save_method(self):
        """Test that MemoryRepository has save method"""
        assert hasattr(MemoryRepository, "save")

    def test_memory_repository_has_get_by_id_method(self):
        """Test that MemoryRepository has get_by_id method"""
        assert hasattr(MemoryRepository, "get_by_id")

    def test_memory_repository_has_delete_method(self):
        """Test that MemoryRepository has delete method"""
        assert hasattr(MemoryRepository, "delete")

    def test_memory_repository_has_update_method(self):
        """Test that MemoryRepository has update method"""
        assert hasattr(MemoryRepository, "update")

    def test_memory_repository_has_list_by_tags_method(self):
        """Test that MemoryRepository has list_by_tags method"""
        assert hasattr(MemoryRepository, "list_by_tags")

    def test_memory_repository_cannot_be_instantiated(self):
        """Test that MemoryRepository cannot be instantiated directly"""
        with pytest.raises(TypeError):
            MemoryRepository()


class TestMemoryChunkRepository:
    def test_memory_chunk_repository_is_abstract(self):
        """Test that MemoryChunkRepository is an abstract base class"""
        assert issubclass(MemoryChunkRepository, ABC)

    def test_memory_chunk_repository_has_save_method(self):
        """Test that MemoryChunkRepository has save method"""
        assert hasattr(MemoryChunkRepository, "save")

    def test_memory_chunk_repository_has_get_by_memory_id_method(self):
        """Test that MemoryChunkRepository has get_by_memory_id method"""
        assert hasattr(MemoryChunkRepository, "get_by_memory_id")

    def test_memory_chunk_repository_has_delete_by_memory_id_method(self):
        """Test that MemoryChunkRepository has delete_by_memory_id method"""
        assert hasattr(MemoryChunkRepository, "delete_by_memory_id")

    def test_memory_chunk_repository_cannot_be_instantiated(self):
        """Test that MemoryChunkRepository cannot be instantiated directly"""
        with pytest.raises(TypeError):
            MemoryChunkRepository()


class TestMemoryRepositoryConcreteImplementation:
    def test_concrete_implementation(self):
        """Test creating a concrete implementation of MemoryRepository"""

        class ConcreteMemoryRepository(MemoryRepository):
            async def save(self, memory: Memory) -> Memory:
                return memory

            async def get_by_id(self, memory_id: str):
                return None

            async def delete(self, memory_id: str) -> bool:
                return True

            async def update(self, memory: Memory) -> Memory:
                return memory

            async def list_by_tags(self, tags: list) -> list:
                return []

        repo = ConcreteMemoryRepository()
        assert isinstance(repo, MemoryRepository)

    @pytest.mark.asyncio
    async def test_concrete_repository_methods(self):
        """Test concrete repository methods are callable"""

        class ConcreteMemoryRepository(MemoryRepository):
            async def save(self, memory: Memory) -> Memory:
                memory.id = "saved_id"
                return memory

            async def get_by_id(self, memory_id: str):
                if memory_id == "saved_id":
                    return Memory(
                        id=memory_id,
                        content="Test",
                        embedding=[0.1, 0.2],
                        source="test",
                    )
                return None

            async def delete(self, memory_id: str) -> bool:
                return memory_id == "saved_id"

            async def update(self, memory: Memory) -> Memory:
                return memory

            async def list_by_tags(self, tags: list) -> list:
                return []

        repo = ConcreteMemoryRepository()

        memory = Memory(
            id="mem123", content="Test", embedding=[0.1, 0.2], source="test"
        )

        saved = await repo.save(memory)
        assert saved.id == "saved_id"

        found = await repo.get_by_id("saved_id")
        assert found is not None
        assert found.id == "saved_id"

        deleted = await repo.delete("saved_id")
        assert deleted is True

        listed = await repo.list_by_tags(["tag1"])
        assert listed == []


class TestMemoryChunkRepositoryConcreteImplementation:
    def test_concrete_implementation(self):
        """Test creating a concrete implementation of MemoryChunkRepository"""

        class ConcreteMemoryChunkRepository(MemoryChunkRepository):
            async def save(self, chunk: MemoryChunk) -> MemoryChunk:
                return chunk

            async def get_by_memory_id(self, memory_id: str) -> list:
                return []

            async def delete_by_memory_id(self, memory_id: str) -> bool:
                return True

        repo = ConcreteMemoryChunkRepository()
        assert isinstance(repo, MemoryChunkRepository)

    @pytest.mark.asyncio
    async def test_concrete_chunk_repository_methods(self):
        """Test concrete chunk repository methods are callable"""

        class ConcreteMemoryChunkRepository(MemoryChunkRepository):
            def __init__(self):
                self.chunks = []

            async def save(self, chunk: MemoryChunk) -> MemoryChunk:
                self.chunks.append(chunk)
                return chunk

            async def get_by_memory_id(self, memory_id: str) -> list:
                return [c for c in self.chunks if c.memory_id == memory_id]

            async def delete_by_memory_id(self, memory_id: str) -> bool:
                initial_count = len(self.chunks)
                self.chunks = [c for c in self.chunks if c.memory_id != memory_id]
                return len(self.chunks) < initial_count

        repo = ConcreteMemoryChunkRepository()

        chunk1 = MemoryChunk(
            id="chunk1",
            memory_id="mem123",
            content="Part 1",
            embedding=[0.1, 0.2],
            chunk_index=0,
        )

        chunk2 = MemoryChunk(
            id="chunk2",
            memory_id="mem123",
            content="Part 2",
            embedding=[0.1, 0.2],
            chunk_index=1,
        )

        await repo.save(chunk1)
        await repo.save(chunk2)

        found = await repo.get_by_memory_id("mem123")
        assert len(found) == 2

        deleted = await repo.delete_by_memory_id("mem123")
        assert deleted is True

        found_after = await repo.get_by_memory_id("mem123")
        assert len(found_after) == 0
