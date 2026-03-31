import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

import pytest
from unittest.mock import MagicMock


@pytest.fixture
def mock_logger():
    """Mock logger fixture"""
    return MagicMock()


@pytest.fixture
def sample_memory_data():
    """Sample memory data for testing"""
    return {
        "id": "mem123",
        "memory": "Test memory content",
        "hash": "abc123",
        "metadata": {"user_id": "user123", "session_id": "session456"},
        "score": 0.9,
        "rerank_score": 0.95,
        "created_at": "2024-01-01T00:00:00Z",
        "updated_at": "2024-01-02T00:00:00Z",
        "user_id": "user123",
        "agent_id": "agent456",
        "run_id": "run789",
    }


@pytest.fixture
def sample_messages():
    """Sample messages for testing"""
    return [
        {"role": "user", "content": "Hello"},
        {"role": "assistant", "content": "Hi there"},
        {"role": "user", "content": "How are you?"},
    ]


@pytest.fixture
def sample_rerank_response():
    """Sample rerank response for testing"""
    from src.adaptee.mf_model_factory.rerank_model_client import (
        RerankResponseUsage,
        RerankResponseResult,
        RerankResponse,
    )

    usage = RerankResponseUsage(prompt_tokens=10, total_tokens=20)
    results = [
        RerankResponseResult(relevance_score=0.9, index=0, document="Doc1"),
        RerankResponseResult(relevance_score=0.8, index=1, document="Doc2"),
        RerankResponseResult(relevance_score=0.7, index=2, document="Doc3"),
    ]

    return RerankResponse(
        id="test_id", usage=usage, results=results, created=1234567890
    )
