"""
Execute-Sync API Integration Tests

Tests for synchronous code execution endpoint.
覆盖正确传参、异常传参、代码正确执行、代码抛出异常、代码输出标准错误、代码正确返回结果、代码执行使用标准库、代码执行使用第三方库等情况。
"""
import pytest
import asyncio
from httpx import AsyncClient


@pytest.mark.asyncio
class TestExecuteSyncAPI:
    """Execute-Sync API integration tests."""

    # ==================== 正确传参测试 ====================

    async def test_execute_sync_valid_parameters(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with valid parameters."""
        request_data = {
            "code": '''
def handler(event):
    return {"message": "Hello, World!"}
''',
            "language": "python",
            "timeout": 10,
            "event": {}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"poll_interval": 0.5, "sync_timeout": 30}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["message"] == "Hello, World!"

    async def test_execute_sync_default_parameters(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with default query parameters."""
        request_data = {
            "code": '''
def handler(event):
    return {"test": "default_params"}
''',
            "language": "python",
            "timeout": 10
        }

        # 使用默认的 poll_interval 和 sync_timeout
        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["test"] == "default_params"

    async def test_execute_sync_with_event_data(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with event data passed correctly."""
        request_data = {
            "code": '''
def handler(event):
    name = event.get("name", "Anonymous")
    age = event.get("age", 0)
    return {"greeting": f"Hello, {name}!", "age_doubled": age * 2}
''',
            "language": "python",
            "timeout": 10,
            "event": {"name": "Alice", "age": 25}
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["greeting"] == "Hello, Alice!"
        assert data["return_value"]["age_doubled"] == 50

    async def test_execute_sync_poll_interval_boundary(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with poll_interval at boundaries (0.1s and 10.0s)."""
        request_data = {
            "code": "def handler(event):\n    return {\"fast\": \"polling\"}",
            "language": "python",
            "timeout": 10
        }

        # 测试最小 poll_interval
        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"poll_interval": 0.1}
        )
        assert response.status_code == 200

        # 测试最大 poll_interval
        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"poll_interval": 10.0}
        )
        assert response.status_code == 200

    # ==================== 异常传参测试 ====================

    async def test_execute_sync_invalid_poll_interval_too_low(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with poll_interval below minimum (0.1s)."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"poll_interval": 0.05}  # 低于最小值 0.1
        )

        assert response.status_code == 422

    async def test_execute_sync_invalid_poll_interval_too_high(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with poll_interval above maximum (10.0s)."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"poll_interval": 15.0}  # 高于最大值 10.0
        )

        assert response.status_code == 422

    async def test_execute_sync_invalid_sync_timeout_too_low(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with sync_timeout below minimum (10s)."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"sync_timeout": 5}  # 低于最小值 10
        )

        assert response.status_code == 422

    async def test_execute_sync_invalid_sync_timeout_too_high(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with sync_timeout above maximum (3600s)."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"sync_timeout": 4000}  # 高于最大值 3600
        )

        assert response.status_code == 422

    async def test_execute_sync_missing_required_field_code(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync without required 'code' field."""
        request_data = {
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 422

    async def test_execute_sync_missing_required_field_language(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync without required 'language' field."""
        request_data = {
            "code": "print('test')",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 422

    async def test_execute_sync_invalid_language(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with invalid language value."""
        request_data = {
            "code": "print('test')",
            "language": "invalid_language",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 422

    async def test_execute_sync_invalid_timeout_too_low(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with timeout below minimum (1s)."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 0  # 低于最小值 1
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 422

    async def test_execute_sync_invalid_timeout_too_high(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with timeout above maximum (3600s)."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 4000  # 高于最大值 3600
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 422

    async def test_execute_sync_nonexistent_session(
        self,
        http_client: AsyncClient
    ):
        """Test execute-sync with nonexistent session ID."""
        request_data = {
            "code": "print('test')",
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            "/executions/sessions/nonexistent_session_id/execute-sync",
            json=request_data
        )

        assert response.status_code in (400, 404)

    # ==================== 代码正确执行测试 ====================

    async def test_execute_sync_simple_execution(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test simple code execution completes successfully."""
        request_data = {
            "code": '''
def handler(event):
    print("Starting execution...")
    result = 2 + 2
    print(f"Calculation complete: {result}")
    return {"result": result}
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["result"] == 4
        assert "Starting execution..." in data["stdout"]
        assert "Calculation complete: 4" in data["stdout"]
        assert data["exit_code"] == 0

    async def test_execute_sync_empty_handler(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with empty handler that returns None."""
        request_data = {
            "code": '''
def handler(event):
    pass
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["exit_code"] == 0

    # ==================== 代码抛出异常测试 ====================

    async def test_execute_sync_runtime_exception(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with code that raises runtime exception."""
        request_data = {
            "code": '''
def handler(event):
    x = 10
    y = 0
    return {"result": x / y}  # ZeroDivisionError
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "failed"
        assert data["exit_code"] != 0
        assert "stderr" in data
        assert len(data["stderr"]) > 0
        assert "ZeroDivisionError" in data["stderr"]

    async def test_execute_sync_type_error(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with code that raises TypeError."""
        request_data = {
            "code": '''
def handler(event):
    return {"result": sum("invalid")}  # TypeError: sum() expects iterable of numbers
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "failed"
        assert data["exit_code"] != 0
        assert "TypeError" in data["stderr"]

    async def test_execute_sync_attribute_error(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with code that raises AttributeError."""
        request_data = {
            "code": '''
def handler(event):
    return {"result": "string".nonexistent_method()}
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "failed"
        assert data["exit_code"] != 0
        assert "AttributeError" in data["stderr"]

    async def test_execute_sync_custom_exception(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with code that raises custom exception."""
        request_data = {
            "code": '''
class CustomError(Exception):
    pass

def handler(event):
    raise CustomError("This is a custom error")
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] == "failed"
        assert data["exit_code"] != 0
        assert "CustomError" in data["stderr"]
        assert "custom error" in data["stderr"].lower()

    # ==================== 代码输出标准错误测试 ====================

    async def test_execute_sync_stderr_output(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with code that outputs to stderr."""
        request_data = {
            "code": '''
import sys

def handler(event):
    print("This goes to stdout")
    sys.stderr.write("This goes to stderr\\n")
    sys.stderr.flush()
    return {"status": "done"}
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert "This goes to stdout" in data["stdout"]
        assert "This goes to stderr" in data["stderr"]

    async def test_execute_sync_warning_to_stderr(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with warnings written to stderr."""
        request_data = {
            "code": '''
import warnings

def handler(event):
    warnings.warn("This is a warning message", UserWarning)
    return {"status": "completed"}
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        # 警告信息通常输出到 stderr
        assert "stderr" in data

    # ==================== 代码正确返回结果测试 ====================

    async def test_execute_sync_return_dict(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with handler returning dictionary."""
        request_data = {
            "code": '''
def handler(event):
    return {
        "statusCode": 200,
        "body": {"message": "Success"},
        "headers": {"Content-Type": "application/json"}
    }
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["statusCode"] == 200
        assert data["return_value"]["body"]["message"] == "Success"

    async def test_execute_sync_return_list(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with handler returning list."""
        request_data = {
            "code": '''
def handler(event):
    return [1, 2, 3, 4, 5]
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"] == [1, 2, 3, 4, 5]

    async def test_execute_sync_return_string(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with handler returning string."""
        request_data = {
            "code": '''
def handler(event):
    return "Hello, World!"
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"] == "Hello, World!"

    async def test_execute_sync_return_number(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with handler returning number."""
        request_data = {
            "code": '''
def handler(event):
    return 42.195
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"] == 42.195

    async def test_execute_sync_return_nested_structure(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with handler returning nested structure."""
        request_data = {
            "code": '''
def handler(event):
    return {
        "user": {
            "id": 123,
            "name": "Test User",
            "tags": ["admin", "tester"],
            "metadata": {
                "created": "2024-01-01",
                "active": True
            }
        },
        "items": [
            {"id": 1, "value": "first"},
            {"id": 2, "value": "second"}
        ]
    }
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["user"]["name"] == "Test User"
        assert data["return_value"]["user"]["tags"] == ["admin", "tester"]
        assert data["return_value"]["items"][0]["value"] == "first"

    # ==================== 代码执行使用标准库测试 ====================

    async def test_execute_sync_use_json_stdlib(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync using json standard library."""
        request_data = {
            "code": '''
import json

def handler(event):
    data = {"name": "Alice", "age": 30, "city": "NYC"}
    json_str = json.dumps(data)
    parsed = json.loads(json_str)
    return {"original": data, "parsed": parsed}
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["parsed"]["name"] == "Alice"

    async def test_execute_sync_use_datetime_stdlib(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync using datetime standard library."""
        request_data = {
            "code": '''
from datetime import datetime, timedelta

def handler(event):
    now = datetime.utcnow()
    tomorrow = now + timedelta(days=1)
    return {
        "now": now.isoformat(),
        "tomorrow": tomorrow.isoformat(),
        "year": now.year,
        "month": now.month
    }
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert "now" in data["return_value"]
        assert "year" in data["return_value"]

    async def test_execute_sync_use_re_stdlib(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync using re standard library."""
        request_data = {
            "code": r'''
import re

def handler(event):
    text = "The price is $123.45 and $67.89"
    prices = re.findall(r'\$\d+\.\d+', text)
    return {
        "original": text,
        "prices_found": prices,
        "count": len(prices)
    }
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["count"] == 2

    async def test_execute_sync_use_collections_stdlib(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync using collections standard library."""
        request_data = {
            "code": '''
from collections import Counter, defaultdict

def handler(event):
    words = ["apple", "banana", "apple", "cherry", "banana", "apple"]
    counter = Counter(words)

    dd = defaultdict(int)
    for word in words:
        dd[word] += 1

    return {
        "counter": dict(counter),
        "defaultdict": dict(dd)
    }
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["counter"]["apple"] == 3

    async def test_execute_sync_use_math_stdlib(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync using math standard library."""
        request_data = {
            "code": '''
import math

def handler(event):
    return {
        "pi": math.pi,
        "e": math.e,
        "sqrt_2": math.sqrt(2),
        "sin_pi_2": math.sin(math.pi / 2),
        "log_10": math.log10(100)
    }
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert 3.14 < data["return_value"]["pi"] < 3.15
        assert data["return_value"]["sqrt_2"] == 2 ** 0.5

    # ==================== 代码执行使用第三方库测试 ====================

    async def test_execute_sync_use_requests_third_party(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """Test execute-sync using requests third-party library.

        This test creates a session with requests dependency pre-installed,
        then executes code that uses the requests library.
        """
        # Step 1: 创建会话时传入 dependencies 参数安装 requests 库
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {},
            "dependencies": [
                {"name": "requests", "version": "==2.31.0"}
            ],
            "install_timeout": 120
        }

        create_response = await http_client.post("/sessions", json=session_data)
        assert create_response.status_code in (201, 200), f"Failed to create session: {create_response.text}"

        create_data = create_response.json()
        session_id = create_data.get("id")
        assert session_id, "Session ID not found in response"

        # Track session for cleanup
        from tests.integration.conftest import track_session
        track_session(session_id)

        try:
            # Step 2: 等待会话就绪（包括依赖安装完成）
            max_wait = 120
            for i in range(max_wait):
                response = await http_client.get(f"/sessions/{session_id}")
                if response.status_code == 200:
                    session = response.json()
                    status = session.get("status")
                    dependency_install_status = session.get("dependency_install_status")
                    if status == "running" and dependency_install_status == "completed":
                        break
                    elif status == "failed":
                        pytest.fail(f"Session failed to start: {session}")
                await asyncio.sleep(1)
            else:
                pytest.fail(f"Session did not become ready with dependencies in {max_wait} seconds")

            # Step 3: 使用已安装 requests 库的会话执行代码
            request_data = {
                "code": '''
import requests

def handler(event):
    # 测试 requests 库可导入且核心对象可用，不依赖外网
    request = requests.Request("GET", "https://example.com/api", params={"q": "sandbox"})
    prepared = request.prepare()
    return {
        "method": prepared.method,
        "url": prepared.url,
        "has_query_string": "q=sandbox" in prepared.url
    }
''',
                "language": "python",
                "timeout": 30
            }

            response = await http_client.post(
                f"/executions/sessions/{session_id}/execute-sync",
                json=request_data
            )

            assert response.status_code == 200
            data = response.json()
            assert data["status"] in ("success", "completed")
            assert data["return_value"]["method"] == "GET"
            assert data["return_value"]["url"].startswith("https://example.com/api")
            assert data["return_value"]["has_query_string"] == True
        finally:
            # Cleanup session manually (also tracked by auto_cleanup)
            await http_client.delete(f"/sessions/{session_id}")
            from tests.integration.conftest import untrack_session
            untrack_session(session_id)

    async def test_execute_sync_use_click_third_party(
        self,
        http_client: AsyncClient,
        test_template_id: str
    ):
        """Test execute-sync using click third-party library.

        This test creates a session with click dependency pre-installed,
        then executes code that uses the click library.
        Note: Using click instead of numpy as it's much lighter (<2MB vs >100MB).
        """
        # Step 1: 创建会话时传入 dependencies 参数安装 click 库
        session_data = {
            "template_id": test_template_id,
            "timeout": 300,
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "env_vars": {},
            "dependencies": [
                {"name": "click"}
            ],
            "install_timeout": 60  # click 安装很快
        }

        create_response = await http_client.post("/sessions", json=session_data)
        assert create_response.status_code in (201, 200), f"Failed to create session: {create_response.text}"

        create_data = create_response.json()
        session_id = create_data.get("id")
        assert session_id, "Session ID not found in response"

        # Track session for cleanup
        from tests.integration.conftest import track_session
        track_session(session_id)

        try:
            # Step 2: 等待会话就绪（包括依赖安装完成）
            max_wait = 60
            for i in range(max_wait):
                response = await http_client.get(f"/sessions/{session_id}")
                if response.status_code == 200:
                    session = response.json()
                    status = session.get("status")
                    dependency_install_status = session.get("dependency_install_status")
                    if status == "running" and dependency_install_status == "completed":
                        break
                    elif status == "failed":
                        pytest.fail(f"Session failed to start: {session}")
                await asyncio.sleep(1)
            else:
                pytest.fail(f"Session did not become ready with dependencies in {max_wait} seconds")

            # Step 3: 使用已安装 click 库的会话执行代码
            request_data = {
                "code": '''
import click

def handler(event):
    # Test click library functionality - simply verify it's importable
    return {
        "click_version": click.__version__,
        "is_installed": True,
        "test_passed": True
    }
''',
                "language": "python",
                "timeout": 10
            }

            response = await http_client.post(
                f"/executions/sessions/{session_id}/execute-sync",
                json=request_data
            )

            assert response.status_code == 200
            data = response.json()
            assert data["status"] in ("success", "completed")
            assert data["return_value"]["is_installed"] == True
            assert data["return_value"]["test_passed"] == True
        finally:
            # Cleanup session manually (also tracked by auto_cleanup)
            await http_client.delete(f"/sessions/{session_id}")
            from tests.integration.conftest import untrack_session
            untrack_session(session_id)

    # ==================== 其他边界情况测试 ====================

    async def test_execute_sync_execution_timeout(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync when execution times out."""
        request_data = {
            "code": '''
import time

def handler(event):
    time.sleep(15)  # 超过 timeout 设置
    return {"should_not_reach": "here"}
''',
            "language": "python",
            "timeout": 5  # 5秒超时
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data,
            params={"sync_timeout": 20}
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("timeout", "failed")

    async def test_execute_sync_large_response(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with large return value (within TEXT column limit of 64KB)."""
        # MySQL TEXT type has a limit of 65,535 bytes (64KB)
        # To stay safely within limit: 400 items × 100 chars = ~40KB
        request_data = {
            "code": '''
def handler(event):
    # 返回一个较大的数据结构（限制在TEXT列范围内）
    return {
        "items": [{"id": i, "data": "x" * 100} for i in range(400)],
        "metadata": {
            "total": 400,
            "description": "Large dataset (within TEXT limit)"
        }
    }
''',
            "language": "python",
            "timeout": 15
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert len(data["return_value"]["items"]) == 400

    async def test_execute_sync_unicode_output(
        self,
        http_client: AsyncClient,
        test_session_id: str
    ):
        """Test execute-sync with unicode characters in output."""
        request_data = {
            "code": '''
def handler(event):
    messages = [
        "Hello, 世界!",  # Chinese
        "Привет, мир!",  # Russian
        "こんにちは!",   # Japanese
        "مرحبا!",        # Arabic
                "🚀🌟✨"     # Emojis
    ]
    return {"messages": messages, "count": len(messages)}
''',
            "language": "python",
            "timeout": 10
        }

        response = await http_client.post(
            f"/executions/sessions/{test_session_id}/execute-sync",
            json=request_data
        )

        assert response.status_code == 200
        data = response.json()
        assert data["status"] in ("success", "completed")
        assert data["return_value"]["count"] == 5
        assert "🚀" in data["return_value"]["messages"][4]

    async def test_execute_sync_multiple_executions_same_session(
        self,
        http_client: AsyncClient,
        persistent_session_id: str
    ):
        """Test multiple execute-sync calls on the same persistent session."""
        request_data_template = '''
def handler(event):
    counter = event.get("counter", 0)
    return {"execution": counter, "message": f"Run {counter}"}
'''

        results = []
        for i in range(3):
            request_data = {
                "code": request_data_template,
                "language": "python",
                "timeout": 10,
                "event": {"counter": i + 1}
            }

            response = await http_client.post(
                f"/executions/sessions/{persistent_session_id}/execute-sync",
                json=request_data
            )

            assert response.status_code == 200
            data = response.json()
            assert data["status"] in ("success", "completed")
            results.append(data["return_value"])

        assert results[0]["execution"] == 1
        assert results[1]["execution"] == 2
        assert results[2]["execution"] == 3
