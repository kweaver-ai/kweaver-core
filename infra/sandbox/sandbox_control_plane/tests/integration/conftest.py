"""
Integration Test Configuration

Shared fixtures and configuration for integration tests.
These tests run against a live docker-compose stack.
"""
import asyncio
import os
import pytest
import httpx
from typing import AsyncGenerator, Dict, Any, List, Optional, Set
from datetime import datetime


# ============== Configuration ==============

CONTROL_PLANE_URL = os.getenv(
    "CONTROL_PLANE_URL",
    "http://localhost:8000"
)
API_BASE_URL = f"{CONTROL_PLANE_URL}/api/v1"

# Test template configuration
TEST_TEMPLATE_ID = "test_template_python"
TEST_TEMPLATE_IMAGE = "sandbox-template-python-basic:latest"

# Module-level set to track all created sessions for cleanup
_created_sessions: Set[str] = set()


HAPPY_PATH_TESTS: Set[str] = {
    # Health and internal contract smoke tests
    "tests/integration/api/test_health_api.py::TestHealthAPI::test_detailed_health_check",
    "tests/integration/api/test_health_api.py::TestHealthAPI::test_manual_state_sync",
    "tests/integration/api/test_internal_api.py::TestInternalAPIContract::test_container_ready_callback_contract",
    "tests/integration/api/test_internal_api.py::TestInternalAPIContract::test_container_exited_callback_contract",
    "tests/integration/api/test_internal_api.py::TestInternalAPIContract::test_execution_heartbeat_contract",
    # Template API happy path
    "tests/integration/api/test_templates_api.py::TestTemplatesAPI::test_create_template",
    "tests/integration/api/test_templates_api.py::TestTemplatesAPI::test_list_templates",
    "tests/integration/api/test_templates_api.py::TestTemplatesAPI::test_get_template",
    "tests/integration/api/test_templates_api.py::TestTemplatesAPI::test_update_template",
    "tests/integration/api/test_templates_api.py::TestTemplatesAPI::test_delete_template",
    # Session API happy path
    "tests/integration/api/test_sessions_api.py::TestSessionsAPI::test_create_session",
    "tests/integration/api/test_sessions_api.py::TestSessionsAPI::test_create_session_with_defaults",
    "tests/integration/api/test_sessions_api.py::TestSessionsAPI::test_get_session",
    "tests/integration/api/test_sessions_api.py::TestSessionsAPI::test_list_sessions",
    "tests/integration/api/test_sessions_api.py::TestSessionsAPI::test_terminate_session",
    "tests/integration/api/test_sessions_api.py::TestSessionsAPI::test_health_check",
    "tests/integration/api/test_sessions_api.py::TestSessionStatusTransitions::test_session_status_running_when_startup_succeeds",
    # Async execution happy path
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_execute_python_code",
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_execute_python_with_event",
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_execute_shell_code",
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_get_execution_status",
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_get_execution_result",
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_execution_with_return_value",
    "tests/integration/api/test_executions_api.py::TestExecutionsAPI::test_list_session_executions",
    # Sync execution happy path
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_valid_parameters",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_default_parameters",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_with_event_data",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_shell_script",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_return_dict",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_use_json_stdlib",
    # Workspace happy path
    "tests/integration/test_e2e_s3_workspace.py::test_file_upload",
    "tests/integration/test_e2e_s3_workspace.py::test_file_download",
    "tests/integration/test_e2e_s3_workspace.py::test_execute_code_read_file",
    "tests/integration/test_e2e_s3_workspace.py::test_execute_code_write_file",
    "tests/integration/test_e2e_s3_workspace.py::test_nested_directory_upload",
    "tests/integration/test_e2e_s3_workspace.py::test_list_files",
    "tests/integration/test_e2e_s3_workspace.py::test_zip_upload_and_extract",
}


SLOW_TEST_FILES: Set[str] = {
    "tests/integration/test_background_tasks.py",
    "tests/integration/test_dependency_installation.py",
    "tests/integration/test_e2e_workflow.py",
    "tests/integration/test_state_sync.py",
    "tests/integration/test_stdlib_execution.py",
    "tests/integration/api/test_session_create_with_dependencies_api.py",
    "tests/integration/api/test_session_dependency_installation_api.py",
}


SLOW_TESTS: Set[str] = {
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_use_requests_third_party",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_use_click_third_party",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_execution_timeout",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_large_response",
    "tests/integration/api/test_execute_sync_api.py::TestExecuteSyncAPI::test_execute_sync_multiple_executions_same_session",
}


@pytest.hookimpl(tryfirst=True)
def pytest_collection_modifyitems(items):
    """Attach integration selection markers used by local and CI commands."""
    for item in items:
        if item.nodeid in HAPPY_PATH_TESTS:
            item.add_marker(pytest.mark.happy_path)

        if item.nodeid in SLOW_TESTS or any(item.nodeid.startswith(path) for path in SLOW_TEST_FILES):
            item.add_marker(pytest.mark.slow)


def track_session(session_id: str) -> None:
    """Track a session ID for cleanup."""
    _created_sessions.add(session_id)


def untrack_session(session_id: str) -> None:
    """Untrack a session ID (e.g., if already cleaned up)."""
    _created_sessions.discard(session_id)


async def _list_session_ids(http_client: httpx.AsyncClient) -> Optional[Set[str]]:
    """
    Return all visible session IDs, or None if the control plane is unavailable.

    Use absolute URLs so this cleanup still works for modules that override
    the http_client fixture without a base_url.
    """
    session_ids: Set[str] = set()
    limit = 200
    offset = 0

    try:
        while True:
            response = await http_client.get(
                f"{API_BASE_URL}/sessions",
                params={"limit": limit, "offset": offset},
            )
            if response.status_code != 200:
                return None

            payload = response.json()
            if isinstance(payload, dict):
                sessions = payload.get("items", [])
                total = payload.get("total", len(sessions))
                has_more = payload.get("has_more", offset + len(sessions) < total)
            else:
                # Backward compatibility with older list API shape.
                sessions = payload
                has_more = False

            for session in sessions:
                session_id = session.get("id")
                if session_id:
                    session_ids.add(session_id)

            if not has_more or not sessions:
                break
            offset += limit
    except Exception as e:
        print(f"[Cleanup] Could not list sessions: {e}")
        return None

    return session_ids


async def _delete_sessions(http_client: httpx.AsyncClient, session_ids: Set[str]) -> None:
    """Best-effort hard-delete sessions and untrack them."""
    if not session_ids:
        return

    print(f"[Cleanup] Cleaning up {len(session_ids)} session(s): {session_ids}")
    for session_id in sorted(session_ids):
        try:
            response = await http_client.delete(f"{API_BASE_URL}/sessions/{session_id}")
            if response.status_code in (200, 202, 204, 404):
                print(f"[Cleanup] Deleted session: {session_id}")
            else:
                print(
                    f"[Cleanup] Failed to delete {session_id}: "
                    f"HTTP {response.status_code} - {response.text[:100]}"
                )
        except Exception as e:
            print(f"[Cleanup] Error deleting {session_id}: {e}")
        finally:
            untrack_session(session_id)


# ============== Fixtures ==============

@pytest.fixture(scope="session")
def event_loop_policy():
    """
    为整个测试会话创建一个事件循环策略。

    这有助于避免异步测试中的事件循环问题。
    """
    import asyncio
    policy = asyncio.get_event_loop_policy()
    yield policy


@pytest.fixture(scope="function", autouse=True)
async def auto_cleanup_sessions(http_client: httpx.AsyncClient, request):
    """
    每个测试函数完成后自动清理所有测试 session。

    autouse=True 确保此 fixture 在每个测试函数后自动运行。
    """
    # 记录测试开始前已存在的 sessions。清理时只删除本测试新增的
    # sessions，避免误删运行环境中已有的手工 session。
    sessions_before_test = await _list_session_ids(http_client)
    tracked_before_test = set(_created_sessions)

    yield  # 测试运行

    tracked_new_sessions = _created_sessions - tracked_before_test

    sessions_after_test = await _list_session_ids(http_client)
    if sessions_before_test is not None and sessions_after_test is not None:
        api_new_sessions = sessions_after_test - sessions_before_test
    else:
        api_new_sessions = set()

    await _delete_sessions(http_client, tracked_new_sessions | api_new_sessions)


@pytest.fixture(scope="function")
async def http_client() -> AsyncGenerator[httpx.AsyncClient, None]:
    """
    Create HTTP client for API calls.

    This client is used for all integration tests to communicate
    with the control plane API.

    Note: trust_env=False is set to bypass macOS system proxy
    that would otherwise cause 502 errors. localhost resolves to IPv6
    on this system, which is required for connectivity.
    """
    async with httpx.AsyncClient(
        base_url=API_BASE_URL,
        timeout=httpx.Timeout(30.0, connect=10.0),
        trust_env=False,  # Disable system proxy for localhost testing
    ) as client:
        yield client


@pytest.fixture(scope="function")
async def cleanup_test_data(http_client: httpx.AsyncClient) -> None:
    """
    Cleanup test data after each test.

    This fixture runs after each test to clean up:
    - Test sessions
    - Test executions
    - Test templates (if created during test)
    """
    yield  # Run the test first

    # Cleanup: Delete test sessions
    try:
        response = await http_client.get(f"{API_BASE_URL}/sessions")
        if response.status_code == 200:
            payload = response.json()
            sessions = payload.get("items", []) if isinstance(payload, dict) else payload
            for session in sessions:
                session_id = session.get("id", "")
                if session_id.startswith("test_"):
                    await http_client.delete(f"{API_BASE_URL}/sessions/{session_id}")
    except Exception:
        pass  # Best effort cleanup


@pytest.fixture(scope="function")
async def test_template_id(http_client: httpx.AsyncClient) -> str:
    """
    Create or get test template.

    Returns the template ID for testing.
    """
    # Try to get existing template
    response = await http_client.get(f"/templates/{TEST_TEMPLATE_ID}")
    if response.status_code == 200:
        return TEST_TEMPLATE_ID

    # Create template if it doesn't exist
    template_data = {
        "id": TEST_TEMPLATE_ID,
        "name": "Python Basic (Test)",
        "image_url": TEST_TEMPLATE_IMAGE,
        "runtime_type": "python3.11",
        "default_cpu_cores": 1.0,
        "default_memory_mb": 512,
        "default_disk_mb": 1024,
        "default_timeout_sec": 300,
        "is_active": True
    }

    response = await http_client.post("/templates", json=template_data)
    if response.status_code in (201, 200):
        return TEST_TEMPLATE_ID

    # If creation failed, try to get again (might have been created concurrently)
    response = await http_client.get(f"/templates/{TEST_TEMPLATE_ID}")
    if response.status_code == 200:
        return TEST_TEMPLATE_ID

    pytest.fail(f"Failed to create/get test template: {TEST_TEMPLATE_ID}")


async def _create_session_and_track(
    http_client: httpx.AsyncClient,
    template_id: str,
    mode: str = None
) -> str:
    """
    Helper function to create a session and track it for cleanup.

    This ensures all created sessions are tracked for automatic cleanup.
    """
    session_data = {
        "template_id": template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {}
    }

    if mode:
        session_data["mode"] = mode

    response = await http_client.post("/sessions", json=session_data)
    assert response.status_code in (201, 200), f"Failed to create session: {response.text}"

    data = response.json()
    session_id = data.get("id")
    assert session_id, "Session ID not found in response"

    # Track session for cleanup
    track_session(session_id)

    # Wait for session to be ready
    # Session becomes RUNNING only after executor sends ready callback
    max_wait = 45  # Increased timeout since we now wait for executor ready callback
    for _ in range(max_wait):
        response = await http_client.get(f"/sessions/{session_id}")
        if response.status_code == 200:
            session = response.json()
            status = session.get("status")
            if status in ("running", "ready"):
                # Session is now RUNNING after executor sent ready callback
                return session_id
            elif status == "failed":
                pytest.fail(f"Session failed to start: {session}")
        await asyncio.sleep(1)

    pytest.fail(f"Session did not become ready in {max_wait} seconds (executor ready callback not received)")


@pytest.fixture(scope="function")
async def test_session_id(
    http_client: httpx.AsyncClient,
    test_template_id: str
) -> str:
    """
    Create a test session and return its ID.

    The session is automatically tracked for cleanup after the test.
    """
    return await _create_session_and_track(http_client, test_template_id)


@pytest.fixture(scope="function")
async def persistent_session_id(
    http_client: httpx.AsyncClient,
    test_template_id: str
) -> str:
    """
    Create a persistent test session for multiple executions.

    Persistent sessions can accept multiple execution requests and
    maintain state between executions. The session is automatically
    tracked for cleanup after the test.
    """
    return await _create_session_and_track(http_client, test_template_id, mode="persistent")


@pytest.fixture(scope="function")
async def test_execution_id(
    http_client: httpx.AsyncClient,
    test_session_id: str
) -> str:
    """
    Create a test execution and return its ID.

    Executes simple Python code that prints "Hello, World!".
    """
    execution_data = {
        "code": 'print("Hello, World!")',
        "language": "python",
        "timeout": 10,
        "event": {},
        "env_vars": {}
    }

    response = await http_client.post(
        f"/executions/sessions/{test_session_id}/execute",
        json=execution_data
    )
    assert response.status_code in (201, 200), f"Failed to create execution: {response.text}"

    data = response.json()
    execution_id = data.get("execution_id") or data.get("id")
    assert execution_id, "Execution ID not found in response"

    return execution_id


@pytest.fixture(scope="function")
async def wait_for_execution_completion(
    http_client: httpx.AsyncClient
):
    """
    Return a function that waits for execution completion.

    Usage:
        execution_id = await test_execution_id(http_client, test_session_id)
        result = await wait_for_execution_completion(http_client, execution_id)
    """
    async def _wait(execution_id: str, timeout: int = 60) -> Dict[str, Any]:
        """Wait for execution to complete and return result."""
        for _ in range(timeout):
            response = await http_client.get(f"/executions/{execution_id}/status")
            if response.status_code == 200:
                execution = response.json()
                status = execution.get("status")
                if status in ("success", "completed", "failed", "timeout", "crashed"):
                    # Get final result
                    result_response = await http_client.get(f"/executions/{execution_id}/result")
                    if result_response.status_code == 200:
                        return result_response.json()
                    return execution
            await asyncio.sleep(1)

        pytest.fail(f"Execution did not complete in {timeout} seconds")

    return _wait


# ============== Helpers ==============

def generate_test_id(prefix: str = "test") -> str:
    """
    Generate a unique test ID.

    Args:
        prefix: Prefix for the ID (default: "test")

    Returns:
        Unique ID string
    """
    timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
    return f"{prefix}_{timestamp}"


async def wait_for_container_ready(
    http_client: httpx.AsyncClient,
    session_id: str,
    timeout: int = 30
) -> bool:
    """
    Wait for container to be ready.

    Args:
        http_client: HTTP client for API calls
        session_id: Session ID to check
        timeout: Maximum wait time in seconds

    Returns:
        True if container is ready, False otherwise
    """
    for _ in range(timeout):
        response = await http_client.get(f"/sessions/{session_id}")
        if response.status_code == 200:
            session = response.json()
            if session.get("status") in ("running", "ready"):
                container_id = session.get("container_id")
                if container_id:
                    return True
        await asyncio.sleep(1)

    return False
