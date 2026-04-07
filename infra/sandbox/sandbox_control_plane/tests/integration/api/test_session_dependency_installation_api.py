"""
Session dependency installation integration tests.

完整验证链路：
1. 创建 session
2. 等待 session 进入 running
3. 调用依赖安装接口安装指定仓库源和依赖
4. 校验 session 中的依赖安装状态与结果
5. 调用同步执行接口运行依赖相关代码
6. 校验执行成功并返回预期结果
7. 删除 session
"""
import asyncio

import pytest
from httpx import AsyncClient

from tests.integration.conftest import track_session, untrack_session


PYTHON_PACKAGE_INDEX_URL = "https://pypi.org/simple/"
DEPENDENCY_NAME = "pyfiglet"
DEPENDENCY_VERSION = "==1.0.2"


async def _create_session(http_client: AsyncClient, template_id: str) -> str:
    response = await http_client.post(
        "/sessions",
        json={
            "template_id": template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {},
        },
    )
    assert response.status_code in (200, 201), response.text
    session_id = response.json()["id"]
    track_session(session_id)
    return session_id


async def _wait_for_session_running(
    http_client: AsyncClient,
    session_id: str,
    timeout_seconds: int = 60,
) -> dict:
    for _ in range(timeout_seconds):
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200, response.text
        session = response.json()
        if session["status"] in ("running", "ready"):
            return session
        if session["status"] == "failed":
            pytest.fail(f"Session failed to start: {session}")
        await asyncio.sleep(1)

    pytest.fail(f"Session {session_id} did not become running within {timeout_seconds}s")


async def _wait_for_dependency_install_completed(
    http_client: AsyncClient,
    session_id: str,
    timeout_seconds: int = 30,
) -> dict:
    for _ in range(timeout_seconds):
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200, response.text
        session = response.json()
        if session["dependency_install_status"] == "completed":
            return session
        if session["dependency_install_status"] == "failed":
            pytest.fail(f"Dependency installation failed: {session}")
        await asyncio.sleep(1)

    pytest.fail(
        f"Session {session_id} dependency installation did not complete within "
        f"{timeout_seconds}s"
    )


@pytest.mark.asyncio
class TestSessionDependencyInstallationAPI:
    async def test_install_dependency_then_execute_code(
        self,
        http_client: AsyncClient,
        test_template_id: str,
    ):
        session_id = await _create_session(http_client, test_template_id)

        try:
            session = await _wait_for_session_running(http_client, session_id)
            assert session["dependency_install_status"] == "completed"
            assert session["installed_dependencies"] == []

            install_response = await http_client.post(
                f"/sessions/{session_id}/dependencies/install",
                json={
                    "python_package_index_url": PYTHON_PACKAGE_INDEX_URL,
                    "dependencies": [
                        {
                            "name": DEPENDENCY_NAME,
                            "version": DEPENDENCY_VERSION,
                        }
                    ],
                },
            )
            assert install_response.status_code == 200, install_response.text

            installed_session = install_response.json()
            assert installed_session["id"] == session_id
            assert (
                installed_session["python_package_index_url"]
                == PYTHON_PACKAGE_INDEX_URL
            )
            assert installed_session["dependency_install_status"] == "completed"
            assert installed_session["dependency_install_error"] in (None, "")
            assert installed_session["dependency_install_started_at"] is not None
            assert installed_session["dependency_install_completed_at"] is not None
            assert installed_session["requested_dependencies"] == [
                {"name": DEPENDENCY_NAME, "version": DEPENDENCY_VERSION}
            ]

            installed_dep_names = {
                dep["name"]: dep for dep in installed_session["installed_dependencies"]
            }
            assert DEPENDENCY_NAME in installed_dep_names, installed_session
            assert installed_dep_names[DEPENDENCY_NAME]["version"] == "1.0.2"

            latest_session = await _wait_for_dependency_install_completed(
                http_client,
                session_id,
            )
            assert latest_session["dependency_install_status"] == "completed"

            execute_response = await http_client.post(
                f"/executions/sessions/{session_id}/execute-sync",
                params={"poll_interval": 0.5, "sync_timeout": 60},
                json={
                    "code": """
import pyfiglet

def handler(event):
    text = pyfiglet.figlet_format(event["text"]).splitlines()[0]
    print(pyfiglet.__version__)
    return {
        "dependency": "pyfiglet",
        "version": pyfiglet.__version__,
        "preview": text,
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
            assert execution["return_value"]["dependency"] == DEPENDENCY_NAME
            assert execution["return_value"]["version"] == "1.0.2"
            assert execution["return_value"]["preview"]
            assert "1.0.2" in (execution.get("stdout") or "")
        finally:
            delete_response = await http_client.delete(f"/sessions/{session_id}")
            assert delete_response.status_code == 204, delete_response.text
            untrack_session(session_id)

