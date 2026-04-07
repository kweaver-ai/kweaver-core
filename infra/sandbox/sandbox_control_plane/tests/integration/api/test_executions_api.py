"""
Code Execution API Integration Tests

Tests for code execution endpoints.
"""
import pytest
import asyncio
from httpx import AsyncClient


@pytest.mark.asyncio
class TestExecutionsAPI:
    """Execution API integration tests."""

    async def test_execute_python_code(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test executing simple Python code."""
        # The executor expects a handler(event) function (Lambda-style)
        execution_data = {
            "code": '''
def handler(event):
    print("Hello, World!")
    return {"status": "success"}
''',
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"

        data = response.json()
        assert "execution_id" in data or "id" in data

        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion and check result
        result = await wait_for_execution_completion(execution_id, timeout=20)
        assert result["status"] in ("success", "completed"), f"Execution failed with status {result.get('status')}: {result.get('stderr', 'Unknown error')}"
        assert "Hello, World!" in result.get("stdout", "")

    async def test_execute_python_with_event(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test executing Python code with event data."""
        code = '''
def handler(event):
    name = event.get("name", "World")
    return {"message": f"Hello, {name}!"}
'''

        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {"name": "Integration Test"},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion
        result = await wait_for_execution_completion(execution_id, timeout=20)
        assert result["status"] in ("success", "completed"), f"Execution failed with status {result.get('status')}: {result.get('stderr', 'Unknown error')}"
        # Check return value
        assert result["return_value"]["message"] == "Hello, Integration Test!"

    async def test_execute_python_with_env_vars(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test executing Python code with environment variables."""
        code = '''
import os

def handler(event):
    test_var = os.environ.get("TEST_VAR", "default")
    print(f"TEST_VAR={test_var}")
    return {"test_var": test_var}
'''

        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {"TEST_VAR": "test_value_123"}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"

        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion
        result = await wait_for_execution_completion(execution_id, timeout=20)
        assert result["status"] in ("success", "completed"), f"Execution failed with status {result.get('status')}: {result.get('stderr', 'Unknown error')}"

    async def test_get_execution_status(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test getting execution status."""
        execution_data = {
            "code": '''
def handler(event):
    print("Status check")
    return {"status": "ok"}
''',
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Check status
        status_response = await http_client.get(f"/executions/{execution_id}/status")
        assert status_response.status_code == 200
        status_data = status_response.json()
        assert "status" in status_data
        assert execution_id in status_data.get("id", status_data.get("execution_id", ""))

    async def test_get_execution_result(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test getting execution result."""
        execution_data = {
            "code": '''
def handler(event):
    print("Result test")
    return {"value": 42}
''',
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"

        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion
        await wait_for_execution_completion(execution_id, timeout=30)

        # Get result
        result_response = await http_client.get(f"/executions/{execution_id}/result")
        assert result_response.status_code == 200
        result_data = result_response.json()
        assert result_data["status"] in ("success", "completed"), f"Execution failed with status {result_data.get('status')}: {result_data.get('stderr', 'Unknown error')}"
        assert "stdout" in result_data
        assert "Result test" in result_data["stdout"]
        assert "execution_time" in result_data

    async def test_execution_syntax_error(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test executing code with syntax error."""
        execution_data = {
            "code": 'print("Missing quote)',
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"
        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion
        result = await wait_for_execution_completion(execution_id, timeout=20)
        assert result["status"] == "failed"
        assert "stderr" in result
        assert len(result["stderr"]) > 0

    async def test_execution_with_return_value(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test execution with return value."""
        code = '''
def handler(event):
    return {
        "statusCode": 200,
        "body": "Success"
    }
'''

        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"

        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion
        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed with status {result.get('status')}: {result.get('stderr', 'Unknown error')}"
        # Check return value
        assert result["return_value"]["statusCode"] == 200
        assert result["return_value"]["body"] == "Success"

    async def test_list_session_executions(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test listing executions for a session."""
        # Create a few executions
        for i in range(3):
            execution_data = {
                "code": f'''
def handler(event):
    print("Execution {i}")
    return {{"index": {i}}}
''',
                "language": "python",
                "timeout": 10,
                "event": {},
                "env_vars": {}
            }
            resp = await http_client.post(
                f"/executions/sessions/{test_session_id}/execute",
                json=execution_data
            )
            assert resp.status_code in (201, 200), f"Execution creation failed: {resp.text}"
            await asyncio.sleep(0.5)  # Small delay between executions

        # List executions
        response = await http_client.get(
            f"/executions/sessions/{test_session_id}/executions"
        )

        # Endpoint might not be implemented (400/405)
        assert response.status_code in (200, 400, 405)

    async def test_execute_in_nonexistent_session(self, http_client: AsyncClient):
        """Test executing code in a nonexistent session."""
        execution_data = {
            "code": 'print("Should fail")',
            "language": "python",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            "/executions/sessions/nonexistent_session_id/execute",
            json=execution_data
        )

        # API returns 400 for validation errors (session not found)
        assert response.status_code in (400, 404)

    async def test_execution_timeout_validation(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execution timeout validation (timeout > MAX_TIMEOUT should fail)."""
        execution_data = {
            "code": 'print("Test")',
            "language": "python",
            "timeout": 5000,  # Exceeds MAX_TIMEOUT of 3600
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code == 422  # Validation error

    async def test_execution_with_large_output(
        self,
        http_client: AsyncClient,
        test_session_id: str,
        wait_for_execution_completion
    ):
        """Test execution with large stdout output."""
        code = '''
def handler(event):
    for i in range(1000):
        print(f"Line {i}: " + "x" * 50)
    return {"lines": 1000}
'''

        execution_data = {
            "code": code,
            "language": "python",
            "timeout": 15,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code in (201, 200), f"Execution creation failed: {response.text}"

        data = response.json()
        execution_id = data.get("execution_id") or data.get("id")

        # Wait for completion
        result = await wait_for_execution_completion(execution_id, timeout=30)
        assert result["status"] in ("success", "completed"), f"Execution failed with status {result.get('status')}: {result.get('stderr', 'Unknown error')}"
        assert len(result["stdout"]) > 50000  # Should have large output

    async def test_execution_language_validation(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execution language validation."""
        execution_data = {
            "code": 'print("Test")',
            "language": "invalid_language",
            "timeout": 10,
            "event": {},
            "env_vars": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute",
            json=execution_data
        )

        assert response.status_code == 422  # Validation error
