import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from datetime import datetime
from src.domain.memory.entities import Memory, MemoryChunk


class TestMemory:
    def test_memory_creation(self):
        """Test creating a memory entity"""
        now = datetime.now()
        memory = Memory(
            id="mem123",
            content="Test memory content",
            embedding=[0.1, 0.2, 0.3],
            source="conversation",
            created_at=now,
            updated_at=now,
        )

        assert memory.id == "mem123"
        assert memory.content == "Test memory content"
        assert memory.embedding == [0.1, 0.2, 0.3]
        assert memory.source == "conversation"
        assert memory.created_at == now
        assert memory.updated_at == now

    def test_memory_with_defaults(self):
        """Test memory creation with default values"""
        memory = Memory(
            id="mem123", content="Test content", embedding=[0.1, 0.2], source="manual"
        )

        assert memory.metadata == {}
        assert memory.importance == 1.0
        assert memory.tags == []
        assert isinstance(memory.created_at, datetime)
        assert isinstance(memory.updated_at, datetime)

    def test_memory_with_metadata(self):
        """Test memory with custom metadata"""
        metadata = {"user_id": "user123", "session_id": "session456"}
        memory = Memory(
            id="mem123",
            content="Test",
            embedding=[0.1, 0.2],
            source="api",
            metadata=metadata,
        )

        assert memory.metadata == metadata
        assert memory.metadata["user_id"] == "user123"

    def test_memory_with_tags(self):
        """Test memory with custom tags"""
        tags = ["important", "urgent", "personal"]
        memory = Memory(
            id="mem123",
            content="Test",
            embedding=[0.1, 0.2],
            source="manual",
            tags=tags,
        )

        assert memory.tags == tags
        assert len(memory.tags) == 3

    def test_memory_with_importance(self):
        """Test memory with custom importance"""
        memory = Memory(
            id="mem123",
            content="Test",
            embedding=[0.1, 0.2],
            source="conversation",
            importance=0.8,
        )

        assert memory.importance == 0.8

    def test_memory_required_fields(self):
        """Test that required fields are validated"""
        with pytest.raises(Exception):
            Memory(id="mem123", content="Test")

    def test_memory_model_dump(self):
        """Test memory model serialization"""
        memory = Memory(id="mem123", content="Test", embedding=[0.1, 0.2], source="api")

        data = memory.model_dump()
        assert data["id"] == "mem123"
        assert data["content"] == "Test"
        assert data["embedding"] == [0.1, 0.2]

    def test_memory_json_serialization(self):
        """Test memory JSON serialization"""
        memory = Memory(id="mem123", content="Test", embedding=[0.1, 0.2], source="api")

        json_str = memory.model_dump_json()
        assert "mem123" in json_str
        assert "Test" in json_str


class TestMemoryChunk:
    def test_memory_chunk_creation(self):
        """Test creating a memory chunk"""
        chunk = MemoryChunk(
            id="chunk123",
            memory_id="mem123",
            content="Chunk content",
            embedding=[0.1, 0.2, 0.3],
            chunk_index=0,
        )

        assert chunk.id == "chunk123"
        assert chunk.memory_id == "mem123"
        assert chunk.content == "Chunk content"
        assert chunk.embedding == [0.1, 0.2, 0.3]
        assert chunk.chunk_index == 0

    def test_memory_chunk_with_defaults(self):
        """Test memory chunk with default metadata"""
        chunk = MemoryChunk(
            id="chunk123",
            memory_id="mem123",
            content="Test",
            embedding=[0.1, 0.2],
            chunk_index=1,
        )

        assert chunk.metadata == {}

    def test_memory_chunk_with_metadata(self):
        """Test memory chunk with custom metadata"""
        metadata = {"position": 0, "length": 100}
        chunk = MemoryChunk(
            id="chunk123",
            memory_id="mem123",
            content="Test",
            embedding=[0.1, 0.2],
            chunk_index=0,
            metadata=metadata,
        )

        assert chunk.metadata == metadata

    def test_memory_chunk_multiple_chunks(self):
        """Test creating multiple chunks for a memory"""
        chunks = [
            MemoryChunk(
                id=f"chunk{i}",
                memory_id="mem123",
                content=f"Part {i}",
                embedding=[0.1, 0.2],
                chunk_index=i,
            )
            for i in range(3)
        ]

        assert len(chunks) == 3
        assert all(chunk.memory_id == "mem123" for chunk in chunks)
        assert [chunk.chunk_index for chunk in chunks] == [0, 1, 2]

    def test_memory_chunk_required_fields(self):
        """Test that required fields are validated"""
        with pytest.raises(Exception):
            MemoryChunk(id="chunk123", memory_id="mem123", content="Test")

    def test_memory_chunk_model_dump(self):
        """Test memory chunk model serialization"""
        chunk = MemoryChunk(
            id="chunk123",
            memory_id="mem123",
            content="Test",
            embedding=[0.1, 0.2],
            chunk_index=0,
        )

        data = chunk.model_dump()
        assert data["id"] == "chunk123"
        assert data["memory_id"] == "mem123"
        assert data["chunk_index"] == 0

    def test_memory_chunk_embedding_dimensions(self):
        """Test memory chunk with different embedding dimensions"""
        chunk_3d = MemoryChunk(
            id="chunk1",
            memory_id="mem123",
            content="Test",
            embedding=[0.1, 0.2, 0.3],
            chunk_index=0,
        )
        chunk_5d = MemoryChunk(
            id="chunk2",
            memory_id="mem123",
            content="Test",
            embedding=[0.1, 0.2, 0.3, 0.4, 0.5],
            chunk_index=1,
        )

        assert len(chunk_3d.embedding) == 3
        assert len(chunk_5d.embedding) == 5
