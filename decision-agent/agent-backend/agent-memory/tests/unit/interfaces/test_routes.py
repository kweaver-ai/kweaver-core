import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi.testclient import TestClient
from fastapi import status


class TestRoutes:
    @pytest.fixture
    def client(self):
        from src.main import app

        return TestClient(app)

    @pytest.fixture
    def mock_build_use_case(self):
        from src.application import BuildMemoryUseCase

        with patch.object(
            BuildMemoryUseCase, "execute", new_callable=AsyncMock
        ) as mock:
            yield mock

    @pytest.fixture
    def mock_retrieval_use_case(self):
        from src.application import RetrievalMemoryUseCase

        with patch.object(
            RetrievalMemoryUseCase, "execute", new_callable=AsyncMock
        ) as mock:
            yield mock

    @pytest.fixture
    def mock_manage_use_case(self):
        from src.application import ManageMemoryUseCase

        with (
            patch.object(ManageMemoryUseCase, "get_memory", new_callable=AsyncMock),
            patch.object(
                ManageMemoryUseCase, "get_all_memories", new_callable=AsyncMock
            ),
            patch.object(ManageMemoryUseCase, "update_memory", new_callable=AsyncMock),
            patch.object(ManageMemoryUseCase, "delete_memory", new_callable=AsyncMock),
            patch.object(
                ManageMemoryUseCase, "get_memory_history", new_callable=AsyncMock
            ),
        ):
            from src.application import ManageMemoryUseCase

            yield ManageMemoryUseCase

    def test_build_memory_success(self, client, mock_build_use_case):
        """Test successful memory building"""
        mock_build_use_case.return_value = {"id": "mem123"}

        response = client.post(
            "/api/agent-memory/internal/v1/memory",
            json={
                "messages": [{"role": "user", "content": "Test message"}],
                "user_id": "user123",
            },
            headers={"x-account-id": "test-account", "x-account-type": "test-type"},
        )

        assert response.status_code == status.HTTP_204_NO_CONTENT
        mock_build_use_case.assert_called_once()

    def test_build_memory_with_all_params(self, client, mock_build_use_case):
        """Test memory building with all parameters"""
        mock_build_use_case.return_value = {"id": "mem123"}

        response = client.post(
            "/api/agent-memory/internal/v1/memory",
            json={
                "messages": [
                    {"role": "user", "content": "Test message"},
                    {"role": "assistant", "content": "Test response"},
                ],
                "user_id": "user123",
                "agent_id": "agent456",
                "run_id": "run789",
                "metadata": {"key": "value"},
                "infer": True,
                "memory_type": "episodic",
                "prompt": "Test prompt",
            },
            headers={"x-account-id": "test-account"},
        )

        assert response.status_code == status.HTTP_204_NO_CONTENT
        mock_build_use_case.assert_called_once()

    def test_retrieval_memory_success(self, client, mock_retrieval_use_case):
        """Test successful memory retrieval"""
        mock_retrieval_use_case.return_value = {
            "results": [
                {
                    "id": "mem123",
                    "memory": "Test memory",
                    "created_at": "2024-01-01T00:00:00Z",
                }
            ]
        }

        response = client.post(
            "/api/agent-memory/internal/v1/search",
            json={
                "query": "test query",
                "user_id": "user123",
                "limit": 10,
            },
            headers={"x-account-id": "test-account"},
        )

        assert response.status_code == status.HTTP_200_OK
        assert "result" in response.json()
        assert len(response.json()["result"]) == 1

    def test_retrieval_memory_with_filters(self, client, mock_retrieval_use_case):
        """Test memory retrieval with filters and thresholds"""
        mock_retrieval_use_case.return_value = {"results": []}

        response = client.post(
            "/api/agent-memory/internal/v1/search",
            json={
                "query": "test query",
                "user_id": "user123",
                "agent_id": "agent456",
                "run_id": "run789",
                "limit": 5,
                "filters": {"key": "value"},
                "threshold": 0.8,
                "rerank_threshold": 0.7,
            },
            headers={"x-account-id": "test-account"},
        )

        assert response.status_code == status.HTTP_200_OK
        mock_retrieval_use_case.assert_called_once()

    def test_get_memory_success(self, client, mock_manage_use_case):
        """Test successful get memory"""
        with patch.object(
            mock_manage_use_case, "get_memory", new_callable=AsyncMock
        ) as mock_get:
            mock_get.return_value = {
                "id": "mem123",
                "memory": "Test memory",
                "created_at": "2024-01-01T00:00:00Z",
            }

            response = client.get("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == status.HTTP_200_OK
            assert response.json()["id"] == "mem123"

    def test_get_memory_not_found(self, client, mock_manage_use_case):
        """Test get memory when not found"""
        with patch.object(
            mock_manage_use_case, "get_memory", new_callable=AsyncMock
        ) as mock_get:
            mock_get.return_value = None

            response = client.get("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == status.HTTP_404_NOT_FOUND

    def test_update_memory_success(self, client):
        """Test successful memory update"""
        from src.application import ManageMemoryUseCase

        with (
            patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value={"id": "mem123"},
            ),
            patch.object(
                ManageMemoryUseCase,
                "update_memory",
                new_callable=AsyncMock,
                return_value={"id": "mem123", "memory": "Updated"},
            ),
        ):
            response = client.put(
                "/api/agent-memory/v1/memory/mem123",
                json={"data": "Updated memory"},
            )

            assert response.status_code == status.HTTP_204_NO_CONTENT

    def test_update_memory_not_found(self, client):
        """Test update memory when not found"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.put(
                "/api/agent-memory/v1/memory/mem123",
                json={"data": "Updated memory"},
            )

            assert response.status_code == status.HTTP_404_NOT_FOUND

    def test_update_memory_operation_failed(self, client):
        """Test update memory when operation fails"""
        from src.application import ManageMemoryUseCase

        with (
            patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value={"id": "mem123"},
            ),
            patch.object(
                ManageMemoryUseCase,
                "update_memory",
                new_callable=AsyncMock,
                return_value=None,
            ),
        ):
            response = client.put(
                "/api/agent-memory/v1/memory/mem123",
                json={"data": "Updated memory"},
            )

            assert response.status_code == status.HTTP_500_INTERNAL_SERVER_ERROR

    def test_delete_memory_success(self, client):
        """Test successful memory deletion"""
        from src.application import ManageMemoryUseCase

        with (
            patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value={"id": "mem123"},
            ),
            patch.object(
                ManageMemoryUseCase,
                "delete_memory",
                new_callable=AsyncMock,
                return_value=True,
            ),
        ):
            response = client.delete("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == status.HTTP_204_NO_CONTENT

    def test_delete_memory_not_found(self, client):
        """Test delete memory when not found"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.delete("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == status.HTTP_404_NOT_FOUND

    def test_delete_memory_operation_failed(self, client):
        """Test delete memory when operation fails"""
        from src.application import ManageMemoryUseCase

        with (
            patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value={"id": "mem123"},
            ),
            patch.object(
                ManageMemoryUseCase,
                "delete_memory",
                new_callable=AsyncMock,
                return_value=False,
            ),
        ):
            response = client.delete("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == status.HTTP_500_INTERNAL_SERVER_ERROR

    def test_get_memory_history_success(self, client):
        """Test successful get memory history"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase,
            "get_memory_history",
            new_callable=AsyncMock,
            return_value=[
                {"action": "created", "timestamp": "2024-01-01T00:00:00Z"},
                {"action": "updated", "timestamp": "2024-01-02T00:00:00Z"},
            ],
        ):
            response = client.get("/api/agent-memory/v1/memory/mem123/history")

            assert response.status_code == status.HTTP_200_OK
            assert "history" in response.json()
            assert len(response.json()["history"]) == 2

    def test_convert_to_memory_response(self):
        """Test _convert_to_memory_response function"""
        from src.interfaces.api.routes import _convert_to_memory_response

        memory_data = {
            "id": "mem123",
            "memory": "Test memory",
            "hash": "abc123",
            "metadata": {"key": "value"},
            "score": 0.9,
            "rerank_score": 0.95,
            "created_at": "2024-01-01T00:00:00Z",
            "updated_at": "2024-01-02T00:00:00Z",
            "user_id": "user123",
            "agent_id": "agent456",
            "run_id": "run789",
        }

        result = _convert_to_memory_response(memory_data)

        assert result["id"] == "mem123"
        assert result["memory"] == "Test memory"
        assert result["hash"] == "abc123"
        assert result["score"] == 0.9
        assert result["rerank_score"] == 0.95

    def test_convert_to_memory_response_minimal(self):
        """Test _convert_to_memory_response with minimal data"""
        from src.interfaces.api.routes import _convert_to_memory_response

        memory_data = {
            "id": "mem123",
            "memory": "Test memory",
            "created_at": "2024-01-01T00:00:00Z",
        }

        result = _convert_to_memory_response(memory_data)

        assert result["id"] == "mem123"
        assert result["hash"] is None
        assert result["metadata"] is None
        assert result["score"] is None
        assert result["rerank_score"] == 0
