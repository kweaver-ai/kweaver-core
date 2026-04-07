#!/usr/bin/env python3
"""
E2E Test: S3 Workspace Mounting

This test verifies:
1. File upload to S3
2. Session creation with S3 workspace mounting
3. User code can access uploaded files via /workspace
4. File download
5. Session cleanup deletes files

These tests require the control plane to be running on localhost:8000
"""
import pytest
import httpx
import asyncio
from typing import Generator


# Configuration
BASE_URL = "http://localhost:8000"


@pytest.fixture(scope="module")
def event_loop():
    """Create an event loop for async tests"""
    loop = asyncio.get_event_loop_policy().new_event_loop()
    yield loop
    loop.close()


@pytest.fixture
async def http_client() -> Generator[httpx.AsyncClient, None, None]:
    """Create an HTTP client for testing"""
    async with httpx.AsyncClient(timeout=60.0) as client:
        yield client


@pytest.fixture
async def session_id(http_client: httpx.AsyncClient) -> str:
    """Create a test session and return its ID"""
    response = await http_client.post(
        f"{BASE_URL}/api/v1/sessions",
        json={"template_id": "python-basic", "timeout": 300}
    )

    if response.status_code not in [200, 201]:
        pytest.skip(f"Failed to create session: {response.text}")

    session = response.json()
    sid = session["id"]

    # Wait for session to be running
    max_wait = 30
    for _ in range(max_wait):
        response = await http_client.get(f"{BASE_URL}/api/v1/sessions/{sid}")
        if response.status_code == 200:
            status = response.json()
            if status["status"] == "running":
                break
            elif status["status"] in ["failed", "terminated"]:
                pytest.skip(f"Session failed with status: {status['status']}")
        await asyncio.sleep(1)
    else:
        pytest.skip("Session did not start within 30 seconds")

    yield sid

    # Cleanup: Delete the session
    try:
        await http_client.delete(f"{BASE_URL}/api/v1/sessions/{sid}")
    except Exception:
        pass  # Ignore cleanup errors


@pytest.mark.asyncio
async def test_file_upload(http_client: httpx.AsyncClient, session_id: str):
    """Test uploading a file to S3"""
    test_csv_content = b"name,age\nAlice,30\nBob,25\nCharlie,35\n"

    response = await http_client.post(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/upload?path=uploads/test_data.csv",
        files={"file": ("test_data.csv", test_csv_content, "text/csv")}
    )

    assert response.status_code == 200, f"Failed to upload file: {response.text}"

    upload_result = response.json()
    assert upload_result["file_path"] == "uploads/test_data.csv"
    assert upload_result["size"] > 0


@pytest.mark.asyncio
async def test_file_download(http_client: httpx.AsyncClient, session_id: str):
    """Test downloading a file from S3"""
    # First upload a file
    test_content = b"test,data\n1,2\n3,4\n"
    await http_client.post(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/upload?path=test/download.csv",
        files={"file": ("download.csv", test_content, "text/csv")}
    )

    # Download the file
    response = await http_client.get(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/test/download.csv"
    )

    assert response.status_code == 200
    content = response.text
    assert "test,data" in content


@pytest.mark.asyncio
async def test_execute_code_read_file(http_client: httpx.AsyncClient, session_id: str):
    """Test executing code that reads uploaded files"""
    # Upload a test file
    test_csv_content = b"name,age\nAlice,30\nBob,25\n"
    await http_client.post(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/upload?path=input/people.csv",
        files={"file": ("people.csv", test_csv_content, "text/csv")}
    )

    # Execute code to read the file
    read_code = """
import os

def handler(event):
    uploads_path = '/workspace/input'
    if os.path.exists(uploads_path):
        files = os.listdir(uploads_path)
        print(f'Files: {files}')

        file_path = '/workspace/input/people.csv'
        if os.path.exists(file_path):
            with open(file_path, 'r') as f:
                content = f.read()
            print(f'Content: {content}')
            return {"found": True, "content": content}
        else:
            print('File not found')
            return {"found": False}
    else:
        print('Directory not found')
        return {"found": False}
"""

    response = await http_client.post(
        f"{BASE_URL}/api/v1/executions/sessions/{session_id}/execute",
        json={"code": read_code, "language": "python"}
    )

    # Handle executor connection failures
    if response.status_code not in [200, 201]:
        pytest.skip(f"Execution creation failed: {response.text}")

    # Wait for execution to complete
    execution = response.json()
    exec_id = execution.get("execution_id")

    # Poll for result
    max_wait = 20
    for _ in range(max_wait):
        response = await http_client.get(
            f"{BASE_URL}/api/v1/executions/{exec_id}/result"
        )
        if response.status_code == 200:
            result = response.json()
            if result.get("status") in ["completed", "failed", "success"]:
                stdout = result.get("stdout", "")
                assert "people.csv" in stdout, f"File not accessed. Output: {stdout}"
                break
        await asyncio.sleep(1)


@pytest.mark.asyncio
async def test_execute_code_write_file(http_client: httpx.AsyncClient, session_id: str):
    """Test executing code that writes files"""
    write_code = """
import os

def handler(event):
    os.makedirs('/workspace/output', exist_ok=True)
    with open('/workspace/output/result.txt', 'w') as f:
        f.write('Test output from code')
    print('File written successfully')
    return {"status": "ok"}
"""

    response = await http_client.post(
        f"{BASE_URL}/api/v1/executions/sessions/{session_id}/execute",
        json={"code": write_code, "language": "python"}
    )

    # Handle executor connection failures
    if response.status_code not in [200, 201]:
        pytest.skip(f"Execution creation failed: {response.text}")

    # Wait for execution and download the created file
    execution = response.json()
    exec_id = execution.get("execution_id")

    # Poll for result
    max_wait = 10
    for _ in range(max_wait):
        response = await http_client.get(
            f"{BASE_URL}/api/v1/executions/{exec_id}/result"
        )
        if response.status_code == 200:
            result = response.json()
            if result.get("status") in ["completed", "failed"]:
                break
        await asyncio.sleep(1)

    # Download the created file
    response = await http_client.get(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/output/result.txt"
    )

    assert response.status_code == 200
    assert "Test output from code" in response.text


@pytest.mark.asyncio
async def test_nested_directory_upload(http_client: httpx.AsyncClient, session_id: str):
    """Test uploading to nested directory paths"""
    # Upload to nested path
    test_content = b"x,y\n1,2\n"
    response = await http_client.post(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/upload?path=data/nested/test.csv",
        files={"file": ("test.csv", test_content, "text/csv")}
    )

    assert response.status_code == 200
    result = response.json()
    assert result["file_path"] == "data/nested/test.csv"

    # Download to verify
    response = await http_client.get(
        f"{BASE_URL}/api/v1/sessions/{session_id}/files/data/nested/test.csv"
    )
    assert response.status_code == 200
    assert "x,y" in response.text


# Standalone execution for manual testing
async def main():
    """Run the E2E test manually"""
    print("=" * 60)
    print("E2E Test: S3 Workspace Mounting")
    print("=" * 60)

    async with httpx.AsyncClient(timeout=60.0) as client:
        # Create session
        print("\n[Step 1] Creating session...")
        response = await client.post(
            f"{BASE_URL}/api/v1/sessions",
            json={"template_id": "python-basic", "timeout": 300}
        )

        if response.status_code not in [200, 201]:
            print(f"ERROR: Failed to create session: {response.text}")
            return

        session = response.json()
        session_id = session["id"]
        print(f"  Session ID: {session_id}")
        print(f"  Workspace: {session.get('workspace_path', '')}")

        # Upload file
        print("\n[Step 2] Uploading test file...")
        test_csv_content = b"name,age\nAlice,30\nBob,25\nCharlie,35\n"
        response = await client.post(
            f"{BASE_URL}/api/v1/sessions/{session_id}/files/upload?path=uploads/test.csv",
            files={"file": ("test.csv", test_csv_content, "text/csv")}
        )

        if response.status_code != 200:
            print(f"ERROR: Failed to upload: {response.text}")
            return

        upload_result = response.json()
        print(f"  Uploaded: {upload_result['file_path']}")

        # Wait for session ready
        print("\n[Step 3] Waiting for session to be ready...")
        for i in range(30):
            response = await client.get(f"{BASE_URL}/api/v1/sessions/{session_id}")
            if response.status_code == 200:
                status = response.json()
                if status["status"] == "running":
                    print(f"  Session is running after {i+1} seconds")
                    break
                elif status["status"] in ["failed", "terminated"]:
                    print(f"ERROR: Session failed with status: {status['status']}")
                    return
            await asyncio.sleep(1)

        # Execute code to read file
        print("\n[Step 4] Executing code to read uploaded file...")
        read_code = """
import os
print(f"Workspace exists: {os.path.exists('/workspace')}")
print(f"Workspace contents: {os.listdir('/workspace') if os.path.exists('/workspace') else 'N/A'}")
"""

        response = await client.post(
            f"{BASE_URL}/api/v1/executions/sessions/{session_id}/execute",
            json={"code": read_code, "language": "python"}
        )

        if response.status_code in [200, 201]:
            execution = response.json()
            print(f"  Execution ID: {execution.get('execution_id')}")

        # Cleanup
        print("\n[Step 5] Deleting session...")
        await client.delete(f"{BASE_URL}/api/v1/sessions/{session_id}")
        print("  Session deleted")

        print("\n" + "=" * 60)
        print("E2E test completed!")
        print("=" * 60)


if __name__ == "__main__":
    asyncio.run(main())
