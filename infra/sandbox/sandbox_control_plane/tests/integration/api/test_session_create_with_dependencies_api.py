"""
Integration tests for creating a session with initial dependencies.
"""
import asyncio

import pytest
from httpx import AsyncClient

from tests.integration.conftest import track_session, untrack_session


PYTHON_PACKAGE_INDEX_URL = "https://pypi.org/simple/"


async def _wait_for_initial_dependency_sync(
    http_client: AsyncClient,
    session_id: str,
    timeout_seconds: int = 90,
) -> dict:
    for _ in range(timeout_seconds):
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200, response.text
        session = response.json()
        if session["dependency_install_status"] == "completed":
            return session
        if session["dependency_install_status"] == "failed":
            pytest.fail(f"Initial dependency sync failed: {session}")
        await asyncio.sleep(1)

    pytest.fail(
        f"Session {session_id} initial dependency sync did not finish within "
        f"{timeout_seconds}s"
    )


@pytest.mark.asyncio
class TestSessionCreateWithDependenciesAPI:
    async def test_create_session_with_initial_dependencies(
        self,
        http_client: AsyncClient,
        test_template_id: str,
    ) -> None:
        create_response = await http_client.post(
            "/sessions",
            json={
                "template_id": test_template_id,
                "timeout": 300,
                "cpu": "1",
                "memory": "512Mi",
                "disk": "1Gi",
                "env_vars": {},
                "python_package_index_url": PYTHON_PACKAGE_INDEX_URL,
                "dependencies": [
                    {
                        "name": "pyfiglet",
                        "version": "==1.0.2",
                    }
                ],
            },
        )
        assert create_response.status_code in (200, 201), create_response.text

        session = create_response.json()
        session_id = session["id"]
        track_session(session_id)

        try:
            assert session["python_package_index_url"] == PYTHON_PACKAGE_INDEX_URL
            assert session["requested_dependencies"] == [
                {"name": "pyfiglet", "version": "==1.0.2"}
            ]
            assert session["dependency_install_status"] == "installing"

            synced_session = await _wait_for_initial_dependency_sync(http_client, session_id)

            assert synced_session["status"] in ("running", "ready")
            assert synced_session["dependency_install_status"] == "completed"
            assert synced_session["dependency_install_error"] in (None, "")
            assert synced_session["dependency_install_started_at"] is not None
            assert synced_session["dependency_install_completed_at"] is not None

            installed = {dep["name"]: dep for dep in synced_session["installed_dependencies"]}
            assert "pyfiglet" in installed, synced_session
            assert installed["pyfiglet"]["version"] == "1.0.2"

            execute_response = await http_client.post(
                f"/executions/sessions/{session_id}/execute-sync",
                params={"poll_interval": 0.5, "sync_timeout": 60},
                json={
                    "code": """
import pyfiglet

def handler(event):
    return {
        "version": pyfiglet.__version__,
        "preview": pyfiglet.figlet_format(event["text"]).splitlines()[0],
    }
""",
                    "language": "python",
                    "timeout": 20,
                    "event": {"text": "OK"},
                },
            )
            assert execute_response.status_code == 200, execute_response.text

            execution = execute_response.json()
            assert execution["status"] in ("success", "completed"), execution
            assert execution["return_value"]["version"] == "1.0.2"
            assert execution["return_value"]["preview"]
        finally:
            delete_response = await http_client.delete(f"/sessions/{session_id}")
            assert delete_response.status_code == 204, delete_response.text
            untrack_session(session_id)
