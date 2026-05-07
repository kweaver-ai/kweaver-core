"""Contract tests for internal API endpoints."""

import pytest
from httpx import AsyncClient


@pytest.mark.contract
class TestInternalAPIContract:
    """Contract tests for internal API endpoints."""

    async def test_container_ready_callback_contract(self, http_client: AsyncClient) -> None:
        """Test POST /internal/containers/ready contract."""
        response = await http_client.post(
            "/internal/containers/ready",
            json={
                "container_id": "abc123",
                "executor_port": 8080,
            },
        )

        assert response.status_code == 200
        assert "message" in response.json()

    async def test_container_exited_callback_contract(self, http_client: AsyncClient) -> None:
        """Test POST /internal/containers/exited contract."""
        response = await http_client.post("/internal/containers/exited")

        assert response.status_code == 200
        assert response.json()["message"] == "Container exited acknowledged"

    async def test_execution_heartbeat_contract(self, http_client: AsyncClient) -> None:
        """Test POST /internal/executions/{execution_id}/heartbeat contract."""
        response = await http_client.post("/internal/executions/exec_test123/heartbeat")

        assert response.status_code == 200
        assert response.json()["message"] == "Heartbeat acknowledged"
