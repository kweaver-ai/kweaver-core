"""Health API integration tests."""

import pytest
from httpx import AsyncClient


@pytest.mark.asyncio
class TestHealthAPI:
    async def test_detailed_health_check(self, http_client: AsyncClient) -> None:
        """Test GET /health/detailed happy path."""
        response = await http_client.get("/health/detailed")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "healthy"
        assert data["version"]
        assert data["uptime"] >= 0
        assert data["dependencies"]["database"] == "healthy"
        assert data["dependencies"]["storage"] == "healthy"
        assert data["dependencies"]["runtime_nodes"] == "healthy"

    async def test_manual_state_sync(self, http_client: AsyncClient) -> None:
        """Test POST /health/sync happy path."""
        response = await http_client.post("/health/sync")

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "success"
        assert data["message"] == "State sync completed"
        assert "result" in data
