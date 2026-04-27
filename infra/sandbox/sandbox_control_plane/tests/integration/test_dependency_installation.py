"""
Integration Tests for Python Dependency Installation

Tests the dependency installation feature as described in
sandbox-design-v2.1.md Chapter 5.

These tests require:
- Running docker-compose stack (control-plane, executor)
- S3/MinIO service available
- Test template available
"""
import asyncio
import pytest
import httpx
from typing import AsyncGenerator


# ============== Fixtures ==============

@pytest.fixture(scope="function")
async def http_client() -> AsyncGenerator[httpx.AsyncClient, None]:
    """Create HTTP client for API calls."""
    import os
    CONTROL_PLANE_URL = os.getenv("CONTROL_PLANE_URL", "http://localhost:8000")
    API_BASE_URL = f"{CONTROL_PLANE_URL}/api/v1"

    async with httpx.AsyncClient(
        base_url=API_BASE_URL,
        timeout=httpx.Timeout(60.0, connect=10.0),
        trust_env=False,
    ) as client:
        yield client


@pytest.fixture(scope="function")
async def test_template_id(http_client: httpx.AsyncClient) -> str:
    """Create or get test template."""
    TEST_TEMPLATE_ID = "test_template_python"

    # Try to get existing template
    response = await http_client.get(f"/templates/{TEST_TEMPLATE_ID}")
    if response.status_code == 200:
        return TEST_TEMPLATE_ID

    # Create template if it doesn't exist
    template_data = {
        "id": TEST_TEMPLATE_ID,
        "name": "Python Basic (Test)",
        "image_url": "sandbox-template-python-basic:latest",
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


async def create_session_with_dependencies(
    http_client: httpx.AsyncClient,
    template_id: str,
    dependencies: list,
    timeout: int = 60
) -> tuple[str, dict]:
    """
    Helper function to create a session with dependencies.

    Returns:
        Tuple of (session_id, session_data)
    """
    session_data = {
        "template_id": template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": dependencies,
        "install_timeout": timeout,
    }

    response = await http_client.post("/sessions", json=session_data)
    assert response.status_code in (201, 200), f"Failed to create session: {response.text}"

    data = response.json()
    session_id = data.get("id")
    assert session_id, "Session ID not found in response"

    return session_id, data


async def wait_for_session_ready(
    http_client: httpx.AsyncClient,
    session_id: str,
    max_wait: int = 120
) -> dict:
    """
    Wait for session to be ready (running status).

    Returns:
        Final session data
    """
    for i in range(max_wait):
        response = await http_client.get(f"/sessions/{session_id}")
        if response.status_code == 200:
            session = response.json()
            status = session.get("status")
            if status == "running":
                return session
            elif status == "failed":
                pytest.fail(f"Session failed to start: {session}")
        await asyncio.sleep(1)

    pytest.fail(f"Session did not become ready in {max_wait} seconds")


# ============== Tests ==============

@pytest.mark.asyncio
async def test_create_session_without_dependencies(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test creating a session without dependencies (baseline).
    """
    session_data = {
        "template_id": test_template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": [],  # Empty dependencies list
    }

    response = await http_client.post("/sessions", json=session_data)
    assert response.status_code in (201, 200), f"Failed to create session: {response.text}"

    data = response.json()
    session_id = data.get("id")
    assert session_id, "Session ID not found in response"

    # Wait for session to be ready
    session = await wait_for_session_ready(http_client, session_id)

    # Verify session is running
    assert session.get("status") == "running"

    # Cleanup
    await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
async def test_create_session_with_single_dependency(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test creating a session with a single dependency.

    This test installs a small package (requests) and verifies
    that the session starts successfully.
    """
    dependencies = [
        {"name": "requests", "version": "==2.31.0"}
    ]

    session_id, _ = await create_session_with_dependencies(
        http_client,
        test_template_id,
        dependencies,
        timeout=120  # Give more time for dependency installation
    )

    try:
        # Wait for session to be ready
        session = await wait_for_session_ready(http_client, session_id, max_wait=120)

        # Verify session is running
        assert session.get("status") == "running"

        # Execute code that uses the installed dependency
        execution_data = {
            "code": """
import requests
def handler(event):
    return {"requests_version": requests.__version__}
""",
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{session_id}/execute",
            json=execution_data
        )
        assert response.status_code in (201, 200), f"Failed to create execution: {response.text}"

        execution = response.json()
        execution_id = execution.get("execution_id") or execution.get("id")
        assert execution_id, "Execution ID not found in response"

        # Wait for execution to complete
        for _ in range(30):
            response = await http_client.get(f"/executions/{execution_id}/status")
            if response.status_code == 200:
                exec_data = response.json()
                status = exec_data.get("status")
                if status in ("success", "completed", "failed", "timeout"):
                    # Get result
                    result_response = await http_client.get(f"/executions/{execution_id}/result")
                    if result_response.status_code == 200:
                        result = result_response.json()
                        # Verify requests was imported successfully
                        if status == "success":
                            output = result.get("stdout", "")
                            assert "2.31.0" in output or "requests_version" in output, \
                                f"requests==2.31.0 not found in output: {output}"
                        return
            await asyncio.sleep(1)

        pytest.fail("Execution did not complete in 30 seconds")

    finally:
        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
async def test_create_session_with_multiple_dependencies(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test creating a session with multiple dependencies.

    This test installs multiple packages and verifies
    that the session starts successfully.
    """
    dependencies = [
        {"name": "requests", "version": "==2.31.0"},
        {"name": "urllib3", "version": "==2.0.0"},
    ]

    session_id, _ = await create_session_with_dependencies(
        http_client,
        test_template_id,
        dependencies,
        timeout=120
    )

    try:
        # Wait for session to be ready
        session = await wait_for_session_ready(http_client, session_id, max_wait=120)

        # Verify session is running
        assert session.get("status") == "running"

        # Execute code that uses both installed dependencies
        execution_data = {
            "code": """
import requests
import urllib3
def handler(event):
    return {
        "requests_version": requests.__version__,
        "urllib3_version": urllib3.__version__
    }
""",
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{session_id}/execute",
            json=execution_data
        )
        assert response.status_code in (201, 200), f"Failed to create execution: {response.text}"

    finally:
        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
async def test_create_session_with_dependency_without_version(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test creating a session with a dependency without version constraint.

    This test installs a package without specifying a version,
    allowing pip to install the latest compatible version.
    """
    dependencies = [
        {"name": "requests"}  # No version specified
    ]

    session_id, _ = await create_session_with_dependencies(
        http_client,
        test_template_id,
        dependencies,
        timeout=120
    )

    try:
        # Wait for session to be ready
        session = await wait_for_session_ready(http_client, session_id, max_wait=120)

        # Verify session is running
        assert session.get("status") == "running"

        # Execute code that uses the installed dependency
        execution_data = {
            "code": """
import requests
def handler(event):
    return {"requests_version": requests.__version__}
""",
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{session_id}/execute",
            json=execution_data
        )
        assert response.status_code in (201, 200), f"Failed to create execution: {response.text}"

    finally:
        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
async def test_dependency_spec_validation_path_traversal(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test that dependency spec validation rejects path traversal attempts.
    """
    # Try to create a session with a malicious package name
    dependencies = [
        {"name": "../../../etc/passwd", "version": ""}
    ]

    session_data = {
        "template_id": test_template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": dependencies,
    }

    response = await http_client.post("/sessions", json=session_data)

    # Should reject the request due to validation error
    assert response.status_code == 422, "Expected validation error for path traversal attempt"


@pytest.mark.asyncio
async def test_dependency_spec_validation_url_injection(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test that dependency spec validation rejects URL injection attempts.
    """
    # Try to create a session with a URL instead of package name
    dependencies = [
        {"name": "http://evil.com/malware", "version": ""}
    ]

    session_data = {
        "template_id": test_template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": dependencies,
    }

    response = await http_client.post("/sessions", json=session_data)

    # Should reject the request due to validation error
    assert response.status_code == 422, "Expected validation error for URL injection attempt"


@pytest.mark.asyncio
async def test_invalid_package_name_rejection(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test that invalid package names are rejected.
    """
    # Try to create a session with an invalid package name (shell injection attempt)
    dependencies = [
        {"name": "requests; rm -rf /", "version": ""}
    ]

    session_data = {
        "template_id": test_template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": dependencies,
    }

    response = await http_client.post("/sessions", json=session_data)

    # Should reject the request due to validation error
    assert response.status_code == 422, "Expected validation error for invalid package name"


@pytest.mark.asyncio
async def test_install_timeout_validation(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test that install_timeout is validated correctly.
    """
    # Try with timeout too low (< 30 seconds)
    session_data = {
        "template_id": test_template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": [{"name": "requests"}],
        "install_timeout": 10,  # Too low
    }

    response = await http_client.post("/sessions", json=session_data)
    assert response.status_code == 422, "Expected validation error for install_timeout < 30"

    # Try with timeout too high (> 1800 seconds)
    session_data["install_timeout"] = 2000  # Too high
    response = await http_client.post("/sessions", json=session_data)
    assert response.status_code == 422, "Expected validation error for install_timeout > 1800"


@pytest.mark.asyncio
async def test_max_dependencies_limit(
    http_client: httpx.AsyncClient,
    test_template_id: str
):
    """
    Test that the maximum number of dependencies is enforced.
    """
    # Create 51 dependencies (exceeds the limit of 50)
    dependencies = [
        {"name": f"package{i}", "version": ""}
        for i in range(51)
    ]

    session_data = {
        "template_id": test_template_id,
        "timeout": 300,
        "cpu": "1",
        "memory": "512Mi",
        "disk": "1Gi",
        "env_vars": {},
        "dependencies": dependencies,
    }

    response = await http_client.post("/sessions", json=session_data)

    # Should reject the request due to too many dependencies
    assert response.status_code == 422, "Expected validation error for exceeding max dependencies"


# ============== Helper Functions ==============

def test_module_imports():
    """Test that all required modules can be imported."""
    # Test request schema imports
    from src.interfaces.rest.schemas.request import DependencySpec, CreateSessionRequest

    # Test command imports
    from src.application.commands.create_session import CreateSessionCommand

    # Test entity imports
    from src.domain.entities.session import Session, InstalledDependency

    # Verify that DependencySpec has validation methods
    assert hasattr(DependencySpec, "validate_package_name")
    assert hasattr(DependencySpec, "to_pip_spec")

    # Verify that Session has dependency management methods
    assert hasattr(Session, "set_dependencies_installing")
    assert hasattr(Session, "set_dependencies_completed")
    assert hasattr(Session, "set_dependencies_failed")
    assert hasattr(Session, "has_dependencies")
    assert hasattr(Session, "is_dependency_install_successful")

    print("All module imports successful!")


if __name__ == "__main__":
    # Run module import test
    test_module_imports()
