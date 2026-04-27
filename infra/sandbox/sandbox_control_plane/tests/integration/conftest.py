"""
Integration Test Configuration

Shared fixtures and configuration for integration tests.
These tests run against a live docker-compose stack.
"""
import asyncio
import os
import pytest
import httpx
from typing import AsyncGenerator, Dict, Any, List, Set
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


def track_session(session_id: str) -> None:
    """Track a session ID for cleanup."""
    _created_sessions.add(session_id)


def untrack_session(session_id: str) -> None:
    """Untrack a session ID (e.g., if already cleaned up)."""
    _created_sessions.discard(session_id)


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
    # 记录测试开始前已存在的 sessions
    sessions_before_test = set(_created_sessions)

    yield  # 测试运行

    # 获取测试期间创建的新 sessions
    new_sessions = _created_sessions - sessions_before_test

    # 清理测试中创建的 session
    if new_sessions:
        print(f"[Cleanup] Cleaning up {len(new_sessions)} session(s): {new_sessions}")
        for session_id in list(new_sessions):
            try:
                response = await http_client.delete(f"/sessions/{session_id}")
                if response.status_code in (200, 202, 204):
                    print(f"[Cleanup] Terminated session: {session_id}")
                else:
                    print(f"[Cleanup] Failed to terminate {session_id}: HTTP {response.status_code} - {response.text[:100]}")
                # Always untrack to avoid accumulation in the set
                untrack_session(session_id)
            except Exception as e:
                print(f"[Cleanup] Error terminating {session_id}: {e}")
                # Still untrack even on error to avoid accumulation
                untrack_session(session_id)


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
        response = await http_client.get("/sessions")
        if response.status_code == 200:
            sessions = response.json()
            for session in sessions:
                session_id = session.get("id", "")
                if session_id.startswith("test_"):
                    await http_client.delete(f"/sessions/{session_id}")
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

