"""
Background Tasks Integration Tests

Tests for background task management including health checks
and session cleanup service.
"""
import pytest
import asyncio
from httpx import AsyncClient


@pytest.mark.asyncio
class TestBackgroundTaskManager:
    """Background task manager integration tests."""

    async def test_background_tasks_are_running(
        self,
        http_client: AsyncClient
    ):
        """
        Test that background tasks are properly started.

        Verifies that the control plane has started its background
        tasks for health checks and session cleanup.
        """
        # Check health endpoint - should indicate system is running
        response = await http_client.get("/health")
        assert response.status_code == 200

        health_data = response.json()
        assert health_data.get("status") in ("healthy", "operational", "running")

    async def test_periodic_health_check_execution(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that periodic health checks are executed.

        Creates a session and waits for longer than the health
        check interval to verify periodic checks are running.
        """
        # Create a session
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {}
        }

        response = await http_client.post("/sessions", json=session_data)
        assert response.status_code in (201, 200)
        session = response.json()
        session_id = session["id"]

        # Wait for session to be ready
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            if response.status_code == 200:
                session_data = response.json()
                if session_data.get("status") in ("running", "ready"):
                    break
            await asyncio.sleep(1)
        else:
            pytest.fail("Session did not become ready")

        # Wait for at least one health check cycle (30 seconds)
        # We'll wait a bit longer to be safe
        await asyncio.sleep(35)

        # Verify session is still healthy
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200
        session_data = response.json()
        assert session_data.get("status") in ("running", "ready")

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")

    async def test_background_tasks_continue_after_session_operations(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that background tasks continue working after session operations.

        Verifies that creating, using, and terminating sessions doesn't
        interfere with background task execution.
        """
        # Create and terminate multiple sessions
        for i in range(2):
            session_data = {
                "template_id": test_template_id,
                "timeout": 300,
                "cpu": "1",
                "memory": "512Mi",
                "disk": "1Gi",
                "env_vars": {}
            }

            response = await http_client.post("/sessions", json=session_data)
            assert response.status_code in (201, 200)
            session = response.json()
            session_id = session["id"]

            # Wait for session to be ready
            for _ in range(30):
                response = await http_client.get(f"/sessions/{session_id}")
                if response.status_code == 200:
                    session_data = response.json()
                    if session_data.get("status") in ("running", "ready"):
                        break
                await asyncio.sleep(1)

            # Terminate
            await http_client.delete(f"/sessions/{session_id}")

        # Wait a bit and verify system is still healthy
        await asyncio.sleep(5)

        response = await http_client.get("/health")
        assert response.status_code == 200


@pytest.mark.asyncio
class TestGracefulShutdown:
    """Graceful shutdown integration tests."""

    async def test_sessions_clean_shutdown(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that sessions are handled properly during shutdown.

        Creates sessions and verifies they maintain state
        that can be recovered after restart (simulated).
        """
        # Create a session
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {}
        }

        response = await http_client.post("/sessions", json=session_data)
        assert response.status_code in (201, 200)
        session = response.json()
        session_id = session["id"]
        container_id = session.get("container_id")

        # Wait for session to be ready
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            if response.status_code == 200:
                session_data = response.json()
                if session_data.get("status") in ("running", "ready"):
                    container_id = session_data.get("container_id")
                    break
            await asyncio.sleep(1)

        # Verify session has container ID (for state sync recovery)
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200
        session_data = response.json()
        assert session_data.get("container_id") is not None

        # Note: We can't actually restart the control plane in integration tests,
        # but we verify the session state is properly set for recovery

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
class TestBackgroundTaskErrorHandling:
    """Background task error handling integration tests."""

    async def test_health_check_handles_terminated_sessions(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that health check properly handles terminated sessions.

        Creates a session, terminates it, and verifies health check
        doesn't report errors for properly terminated sessions.
        """
        # Create and terminate a session
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {}
        }

        response = await http_client.post("/sessions", json=session_data)
        assert response.status_code in (201, 200)
        session = response.json()
        session_id = session["id"]

        # Wait for session to be ready
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            if response.status_code == 200:
                session_data = response.json()
                if session_data.get("status") in ("running", "ready"):
                    break
            await asyncio.sleep(1)

        # Terminate session
        await http_client.delete(f"/sessions/{session_id}")

        # Wait for health check cycle
        await asyncio.sleep(35)

        # Verify health endpoint is still accessible
        response = await http_client.get("/health")
        assert response.status_code == 200

    async def test_background_tasks_continue_after_errors(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that background tasks continue working after encountering errors.

        Tries to create sessions with invalid templates, then verifies
        background tasks still work.
        """
        # Try to create sessions with invalid template
        for i in range(2):
            session_data = {
                "template_id": "invalid_template_xyz",
                "timeout": 300,
                "cpu": "1",
                "memory": "512Mi",
                "disk": "1Gi",
                "env_vars": {}
            }

            response = await http_client.post("/sessions", json=session_data)
            # Expected to fail
            assert response.status_code in (400, 404)

        # Now create a valid session
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {}
        }

        response = await http_client.post("/sessions", json=session_data)
        assert response.status_code in (201, 200)
        session = response.json()
        session_id = session["id"]

        # Wait for session to be ready
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            if response.status_code == 200:
                session_data = response.json()
                if session_data.get("status") in ("running", "ready"):
                    break
            await asyncio.sleep(1)

        # Verify system is still healthy
        response = await http_client.get("/health")
        assert response.status_code == 200

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
class TestBackgroundTaskPerformance:
    """Background task performance integration tests."""

    async def test_background_tasks_dont_block_api(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that background tasks don't block API requests.

        Makes API requests while background tasks are running
        to verify responsiveness.
        """
        # Create multiple sessions rapidly
        session_ids = []
        start_time = asyncio.get_event_loop().time()

        for i in range(3):
            session_data = {
                "template_id": test_template_id,
                "timeout": 300,
                "cpu": "1",
                "memory": "512Mi",
                "disk": "1Gi",
                "env_vars": {}
            }

            response = await http_client.post("/sessions", json=session_data)
            assert response.status_code in (201, 200)
            session = response.json()
            session_ids.append(session["id"])

        creation_time = asyncio.get_event_loop().time() - start_time

        # Requests should complete quickly (not blocked by background tasks)
        # Allow generous time for container creation
        assert creation_time < 60, "API requests should not be blocked by background tasks"

        # Cleanup
        for session_id in session_ids:
            try:
                await http_client.delete(f"/sessions/{session_id}")
            except Exception:
                pass

    async def test_concurrent_health_checks_and_requests(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test concurrent health checks and API requests.

        Makes API requests while waiting for health check cycles
        to verify they don't interfere.
        """
        # Create a session
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {}
        }

        response = await http_client.post("/sessions", json=session_data)
        assert response.status_code in (201, 200)
        session = response.json()
        session_id = session["id"]

        # Make concurrent requests while health checks might be running
        tasks = []
        for i in range(5):
            tasks.append(http_client.get(f"/sessions/{session_id}"))
            tasks.append(http_client.get("/health"))

        responses = await asyncio.gather(*tasks)

        # All requests should succeed
        for resp in responses:
            assert resp.status_code == 200

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")
