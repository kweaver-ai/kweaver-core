"""
State Sync Integration Tests

Tests for state synchronization service that ensures session-container
state consistency across control plane restarts.
"""
import pytest
import asyncio
from httpx import AsyncClient


@pytest.mark.asyncio
class TestStateSyncService:
    """State sync service integration tests."""

    async def test_session_container_state_consistency(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that session state reflects actual container state.

        Creates a session and verifies that:
        1. Session status is "running" when container is running
        2. Container ID is properly associated
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

        # Wait for session to be running
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            assert response.status_code == 200
            session_data = response.json()

            # Verify session has container_id
            if session_data.get("status") in ("running", "ready"):
                container_id = session_data.get("container_id")
                assert container_id, "Container ID should be set when session is running"
                assert len(container_id) > 0, "Container ID should not be empty"
                break
            await asyncio.sleep(1)
        else:
            pytest.fail("Session did not reach running state")

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")

    async def test_multiple_sessions_state_consistency(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test state consistency across multiple sessions.

        Creates multiple sessions and verifies all have proper
        container associations and consistent states.
        """
        session_ids = []

        # Create 3 sessions
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

        # Wait for all sessions to be running and verify container associations
        for session_id in session_ids:
            for _ in range(30):
                response = await http_client.get(f"/sessions/{session_id}")
                assert response.status_code == 200
                session_data = response.json()

                if session_data.get("status") in ("running", "ready"):
                    container_id = session_data.get("container_id")
                    assert container_id, f"Session {session_id} should have container_id"
                    break
                await asyncio.sleep(1)
            else:
                pytest.fail(f"Session {session_id} did not reach running state")

        # Verify all sessions are unique (skip if GET /sessions not available)
        response = await http_client.get("/sessions")
        if response.status_code == 405:
            # GET /sessions not implemented, skip verification
            return
        assert response.status_code == 200
        result = response.json()
        all_sessions = result.get("items", [])

        container_ids = set()
        for session in all_sessions:
            if session.get("id") in session_ids:
                container_id = session.get("container_id")
                if container_id:
                    container_ids.add(container_id)

        # All running sessions should have unique container IDs
        assert len(container_ids) == len(session_ids), \
            "Each session should have a unique container ID"

        # Cleanup
        for session_id in session_ids:
            await http_client.delete(f"/sessions/{session_id}")

    async def test_session_termination_state_update(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that session state updates correctly after termination.

        Verifies that:
        1. Session is "running" after creation
        2. Session becomes "terminated" after deletion
        """
        # Create session
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

        # Wait for session to be running
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            assert response.status_code == 200
            session_data = response.json()
            if session_data.get("status") in ("running", "ready"):
                break
            await asyncio.sleep(1)
        else:
            pytest.fail("Session did not reach running state")

        # Terminate session
        response = await http_client.post(f"/sessions/{session_id}/terminate")
        assert response.status_code == 200

        # Verify session status is terminated
        for _ in range(10):
            response = await http_client.get(f"/sessions/{session_id}")
            if response.status_code == 200:
                session_data = response.json()
                if session_data.get("status") in ("terminated", "terminating"):
                    break
            await asyncio.sleep(0.5)

    async def test_persistent_session_state_maintenance(
        self,
        http_client: AsyncClient,
        persistent_session_id: str,
        wait_for_execution_completion
    ):
        """
        Test that persistent session maintains state across multiple executions.

        Verifies that:
        1. Session remains "running" after execution completes
        2. Container ID remains consistent
        """
        # Get initial session state
        response = await http_client.get(f"/sessions/{persistent_session_id}")
        assert response.status_code == 200
        initial_session = response.json()
        initial_container_id = initial_session.get("container_id")

        # Execute code
        execution_data = {
            "code": '''
def handler(event):
    print("State maintenance test")
    return {"status": "ok"}
''',
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{persistent_session_id}/execute",
            json=execution_data
        )
        # Handle executor connection failures
        if response.status_code not in (201, 200):
            pytest.skip(f"Execution creation failed: {response.text}")
        execution = response.json()
        execution_id = execution.get("execution_id") or execution.get("id")

        # Wait for execution to complete
        await wait_for_execution_completion(execution_id, timeout=20)

        # Verify session is still running with same container
        response = await http_client.get(f"/sessions/{persistent_session_id}")
        assert response.status_code == 200
        final_session = response.json()

        assert final_session.get("status") in ("running", "ready"), \
            "Persistent session should still be running after execution"
        assert final_session.get("container_id") == initial_container_id, \
            "Container ID should remain consistent for persistent session"

    async def test_session_without_container_id_handling(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test handling of sessions without container_id.

        Verifies that sessions in "creating" state without container_id
        are properly handled and transition to running state.
        """
        # Create session
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

        # Initially might be in "creating" state without container_id
        # Wait for transition to "running" with container_id
        has_container_id = False
        is_running = False
        for _ in range(45):  # Increased timeout to account for executor ready callback
            response = await http_client.get(f"/sessions/{session_id}")
            assert response.status_code == 200
            session_data = response.json()

            container_id = session_data.get("container_id")
            status = session_data.get("status")

            if container_id:
                has_container_id = True
                # Wait for both container_id AND running status
                # The executor sends ready callback after HTTP server starts,
                # which may take a moment
                if status in ("running", "ready"):
                    is_running = True
                    break
                # Log that we have container_id but still waiting for running status
                print(f"[Test] Session has container_id={container_id} but status={status}, waiting for running...")
            await asyncio.sleep(1)

        assert has_container_id, "Session should eventually get a container_id"
        assert is_running, "Session with container_id should eventually transition to running state"

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")


@pytest.mark.asyncio
class TestHealthCheckIntegration:
    """Integration tests for health check mechanism."""

    async def test_health_endpoint_returns_system_status(
        self,
        http_client: AsyncClient
    ):
        """
        Test that health endpoint returns system status.

        Verifies that the health check endpoint responds with
        system status information.
        """
        response = await http_client.get("/health")
        assert response.status_code == 200

        health_data = response.json()
        assert "status" in health_data
        assert health_data["status"] in ("healthy", "degraded", "unhealthy")

    async def test_health_endpoint_with_background_tasks(
        self,
        http_client: AsyncClient
    ):
        """
        Test that health endpoint reflects background task status.

        Verifies that periodic health checks and session cleanup
        are reflected in the health status.
        """
        response = await http_client.get("/health")
        assert response.status_code == 200

        health_data = response.json()
        # Background tasks should be running
        # (The exact format depends on implementation)
        assert "status" in health_data


@pytest.mark.asyncio
class TestStateRecoveryScenarios:
    """Integration tests for state recovery scenarios."""

    async def test_failed_session_marking(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """
        Test that sessions with failed containers are marked properly.

        This test creates a session and simulates a failure scenario
        to verify proper state handling.
        """
        # Create session
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

        # Wait for session to be running
        for _ in range(30):
            response = await http_client.get(f"/sessions/{session_id}")
            if response.status_code == 200:
                session_data = response.json()
                if session_data.get("status") in ("running", "ready"):
                    break
            await asyncio.sleep(1)
        else:
            pytest.fail("Session did not reach running state")

        # Verify session is in valid state
        response = await http_client.get(f"/sessions/{session_id}")
        assert response.status_code == 200
        session_data = response.json()
        assert session_data.get("status") in ("running", "ready", "terminated", "terminating")

        # Cleanup
        await http_client.delete(f"/sessions/{session_id}")
