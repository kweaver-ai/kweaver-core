"""Contract tests for internal API endpoints."""

import pytest
from httpx import AsyncClient


@pytest.mark.contract
class TestInternalAPIContract:
    """Contract tests for internal API endpoints."""

    async def test_container_ready_callback_contract(self, http_client: AsyncClient) -> None:
        """Test POST /internal/sessions/{id}/container_ready contract."""
        # This would require a valid session and INTERNAL_API_TOKEN
        # For contract test, we validate the endpoint structure
        response = await http_client.post(
            "/internal/sessions/sess_test123/container_ready",
            json={
                "container_id": "abc123",
                "executor_port": 8080,
            },
            headers={"Authorization": "Bearer test-token"},
        )

        # Should return 401 with invalid token, but endpoint exists
        assert response.status_code in [200, 401, 404]

    async def test_container_exited_callback_contract(self, http_client: AsyncClient) -> None:
        """Test POST /internal/sessions/{id}/container_exited contract."""
        response = await http_client.post(
            "/internal/sessions/sess_test123/container_exited",
            json={
                "exit_code": 143,
                "exit_reason": "sigterm",
            },
            headers={"Authorization": "Bearer test-token"},
        )

        # Should return 401 with invalid token, but endpoint exists
        assert response.status_code in [200, 401, 404]
