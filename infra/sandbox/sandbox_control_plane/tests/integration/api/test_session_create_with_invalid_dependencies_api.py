"""
Integration tests for creating a session with invalid initial dependencies.
"""
import asyncio

import pytest
from httpx import AsyncClient

from tests.integration.conftest import track_session, untrack_session


async def _wait_for_initial_dependency_failure(
    http_client: AsyncClient,
    session_id: str,
    timeout_seconds: int = 90,
) -> dict:
    for _ in range(timeout_seconds):
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200, response.text
        session = response.json()
        if session["dependency_install_status"] == "failed":
            return session
        await asyncio.sleep(1)

    pytest.fail(
        f"Session {session_id} initial dependency sync did not fail within "
        f"{timeout_seconds}s"
    )


@pytest.mark.asyncio
class TestSessionCreateWithInvalidDependenciesAPI:
    async def test_create_session_with_invalid_initial_dependencies_marks_session_failed(
        self,
        http_client: AsyncClient,
        test_template_id: str,
    ) -> None:
        create_response = await http_client.post(
            "/sessions",
            json={
                "id": "test_invalid_dependency_install",
                "template_id": test_template_id,
                "timeout": 300,
                "cpu": "1",
                "memory": "512Mi",
                "disk": "1Gi",
                "env_vars": {},
                "python_package_index_url": "",
                "install_timeout": 120,
                "dependencies": [
                    {
                        "name": "fastapi",
                        "version": "==0.13.5",
                    }
                ],
            },
        )
        assert create_response.status_code in (200, 201), create_response.text

        session = create_response.json()
        session_id = session["id"]
        track_session(session_id)

        try:
            assert session["dependency_install_status"] == "installing"

            failed_session = await _wait_for_initial_dependency_failure(http_client, session_id)

            assert failed_session["dependency_install_status"] == "failed"
            assert failed_session["dependency_install_started_at"] is not None
            assert failed_session["dependency_install_completed_at"] is not None
            assert failed_session["installed_dependencies"] == []
            assert failed_session["dependency_install_error"]
            assert "fastapi==0.13.5" in failed_session["dependency_install_error"]
        finally:
            delete_response = await http_client.delete(f"/sessions/{session_id}")
            assert delete_response.status_code == 204, delete_response.text
            untrack_session(session_id)
