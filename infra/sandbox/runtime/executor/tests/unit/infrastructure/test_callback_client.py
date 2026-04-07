"""
Unit tests for CallbackClient.

Tests the HTTP callback client for reporting execution results.
"""

import math
import pytest
import httpx
import json
import tempfile
from datetime import datetime
from pathlib import Path
from unittest.mock import AsyncMock, Mock, patch, MagicMock

from executor.infrastructure.http.callback_client import (
    CallbackClient,
    get_callback_client,
)
from executor.domain.value_objects import (
    ExecutionResult,
    ExecutionStatus,
    HeartbeatSignal,
    ContainerLifecycleEvent,
    ExitReason,
    ExecutionMetrics,
)


class TestCallbackClientInit:
    """Tests for CallbackClient initialization."""

    def test_init_with_defaults(self):
        """Test initialization with default values."""
        with patch("executor.infrastructure.http.callback_client.settings") as mock_settings:
            mock_settings.control_plane_url = "http://localhost:8000"
            mock_settings.internal_api_token = "test-token"

            client = CallbackClient()

            assert client.control_plane_url == "http://localhost:8000"
            assert client.api_token == "test-token"
            assert client.max_retries == 5
            assert client.base_retry_delay == 1.0
            assert client.max_retry_delay == 10.0

    def test_init_with_custom_values(self):
        """Test initialization with custom values."""
        client = CallbackClient(
            control_plane_url="http://custom:9000",
            api_token="custom-token",
            max_retries=3,
            base_retry_delay=2.0,
            max_retry_delay=20.0,
        )

        assert client.control_plane_url == "http://custom:9000"
        assert client.api_token == "custom-token"
        assert client.max_retries == 3
        assert client.base_retry_delay == 2.0
        assert client.max_retry_delay == 20.0


class TestSanitizeForJson:
    """Test the _sanitize_for_json method."""

    def setup_method(self):
        """Create a CallbackClient instance for testing."""
        self.client = CallbackClient(
            control_plane_url="http://test.invalid",
            api_token="test-token",
        )

    def test_sanitize_valid_float(self):
        """Test that valid float values are preserved."""
        result = self.client._sanitize_for_json(3.14)
        assert result == 3.14

    def test_sanitize_nan(self):
        """Test that NaN is converted to None."""
        result = self.client._sanitize_for_json(float('nan'))
        assert result is None

    def test_sanitize_positive_infinity(self):
        """Test that positive Infinity is converted to None."""
        result = self.client._sanitize_for_json(float('inf'))
        assert result is None

    def test_sanitize_negative_infinity(self):
        """Test that negative Infinity is converted to None."""
        result = self.client._sanitize_for_json(float('-inf'))
        assert result is None

    def test_sanitize_nested_dict_with_nan(self):
        """Test sanitization of nested dictionaries with NaN values."""
        data = {
            "valid": 1.5,
            "nested": {
                "value": float('nan'),
                "other": "string",
            },
            "list": [float('inf'), 2.0],
        }
        result = self.client._sanitize_for_json(data)

        assert result["valid"] == 1.5
        assert result["nested"]["value"] is None
        assert result["nested"]["other"] == "string"
        assert result["list"][0] is None
        assert result["list"][1] == 2.0

    def test_sanitize_list_with_invalid_floats(self):
        """Test sanitization of lists with invalid float values."""
        data = [float('nan'), 1.0, float('inf'), float('-inf'), 2.5]
        result = self.client._sanitize_for_json(data)

        assert result == [None, 1.0, None, None, 2.5]

    def test_sanitize_none_value(self):
        """Test that None values are preserved."""
        result = self.client._sanitize_for_json(None)
        assert result is None

    def test_sanitize_string(self):
        """Test that string values are preserved."""
        result = self.client._sanitize_for_json("test")
        assert result == "test"

    def test_sanitize_integer(self):
        """Test that integer values are preserved."""
        result = self.client._sanitize_for_json(42)
        assert result == 42

    def test_sanitize_boolean(self):
        """Test that boolean values are preserved."""
        result = self.client._sanitize_for_json(True)
        assert result is True
        result = self.client._sanitize_for_json(False)
        assert result is False

    def test_sanitize_empty_dict(self):
        """Test that empty dictionaries are preserved."""
        result = self.client._sanitize_for_json({})
        assert result == {}

    def test_sanitize_empty_list(self):
        """Test that empty lists are preserved."""
        result = self.client._sanitize_for_json([])
        assert result == []


class TestReportResult:
    """Tests for report_result method."""

    @pytest.fixture
    def client(self):
        """Create a CallbackClient instance for testing."""
        return CallbackClient(
            control_plane_url="http://test.invalid",
            api_token="test-token",
            max_retries=2,
            base_retry_delay=0.1,  # Fast for testing
        )

    @pytest.fixture
    def success_result(self):
        """Create a successful execution result."""
        return ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=1000,
        )

    @pytest.mark.asyncio
    async def test_report_result_success(self, client, success_result):
        """Test successful result reporting."""
        mock_response = Mock()
        mock_response.status_code = 200

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is True
            mock_client.post.assert_called_once()

    @pytest.mark.asyncio
    async def test_report_result_with_metrics(self, client):
        """Test result reporting with metrics containing NaN."""
        metrics = ExecutionMetrics(
            duration_ms=float('nan'),  # Should be sanitized
            cpu_time_ms=80.0,
        )
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=1000,
            metrics=metrics,
        )

        mock_response = Mock()
        mock_response.status_code = 200

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            success = await client.report_result("exec_001", result)

            assert success is True
            # Verify the call was made with sanitized data
            call_args = mock_client.post.call_args
            payload = call_args.kwargs['json']
            assert payload['metrics']['duration_ms'] is None

    @pytest.mark.asyncio
    async def test_report_result_server_error_retry(self, client, success_result):
        """Test retry on server error."""
        mock_error_response = Mock()
        mock_error_response.status_code = 500

        mock_success_response = Mock()
        mock_success_response.status_code = 200

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            # First call fails, second succeeds
            mock_client.post = AsyncMock(
                side_effect=[mock_error_response, mock_success_response]
            )
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is True
            assert mock_client.post.call_count == 2

    @pytest.mark.asyncio
    async def test_report_result_unauthorized_retry(self, client, success_result):
        """Test retry on 401 Unauthorized."""
        mock_unauth_response = Mock()
        mock_unauth_response.status_code = 401

        mock_success_response = Mock()
        mock_success_response.status_code = 200

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(
                side_effect=[mock_unauth_response, mock_success_response]
            )
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is True
            assert mock_client.post.call_count == 2

    @pytest.mark.asyncio
    async def test_report_result_client_error_no_retry(self, client, success_result):
        """Test no retry on 4xx client error (except 401)."""
        mock_response = Mock()
        mock_response.status_code = 400
        mock_response.text = "Bad Request"

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is False
            # Should only try once
            assert mock_client.post.call_count == 1

    @pytest.mark.asyncio
    async def test_report_result_timeout_retry(self, client, success_result):
        """Test retry on timeout."""
        mock_success_response = Mock()
        mock_success_response.status_code = 200

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(
                side_effect=[
                    httpx.TimeoutException("Timeout"),
                    mock_success_response,
                ]
            )
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is True
            assert mock_client.post.call_count == 2

    @pytest.mark.asyncio
    async def test_report_result_network_error_retry(self, client, success_result):
        """Test retry on network error."""
        mock_success_response = Mock()
        mock_success_response.status_code = 200

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(
                side_effect=[
                    httpx.NetworkError("Connection refused"),
                    mock_success_response,
                ]
            )
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is True
            assert mock_client.post.call_count == 2

    @pytest.mark.asyncio
    async def test_report_result_all_retries_exhausted(self, client, success_result):
        """Test local persistence when all retries fail."""
        mock_response = Mock()
        mock_response.status_code = 500

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_result("exec_001", success_result)

            assert result is False
            assert mock_client.post.call_count == client.max_retries


class TestReportHeartbeat:
    """Tests for report_heartbeat method."""

    @pytest.fixture
    def client(self):
        """Create a CallbackClient instance for testing."""
        return CallbackClient(
            control_plane_url="http://test.invalid",
            api_token="test-token",
        )

    @pytest.mark.asyncio
    async def test_report_heartbeat_success(self, client):
        """Test successful heartbeat reporting."""
        mock_response = Mock()
        mock_response.status_code = 200

        signal = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={"status": "running"},
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_heartbeat("exec_001", signal)

            assert result is True

    @pytest.mark.asyncio
    async def test_report_heartbeat_failure(self, client):
        """Test heartbeat reporting failure."""
        mock_response = Mock()
        mock_response.status_code = 500

        signal = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={"status": "running"},
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_heartbeat("exec_001", signal)

            assert result is False

    @pytest.mark.asyncio
    async def test_report_heartbeat_exception(self, client):
        """Test heartbeat reporting with exception."""
        signal = HeartbeatSignal(
            timestamp=datetime.now(),
            progress={"status": "running"},
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(side_effect=Exception("Network error"))
            mock_get_client.return_value = mock_client

            result = await client.report_heartbeat("exec_001", signal)

            assert result is False


class TestReportLifecycle:
    """Tests for report_lifecycle method."""

    @pytest.fixture
    def client(self):
        """Create a CallbackClient instance for testing."""
        return CallbackClient(
            control_plane_url="http://test.invalid",
            api_token="test-token",
        )

    @pytest.mark.asyncio
    async def test_report_lifecycle_ready(self, client):
        """Test reporting container ready event."""
        mock_response = Mock()
        mock_response.status_code = 200

        event = ContainerLifecycleEvent(
            event_type="ready",
            container_id="container-123",
            pod_name="pod-456",
            executor_port=8080,
            ready_at=datetime.now(),
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_lifecycle(event)

            assert result is True

    @pytest.mark.asyncio
    async def test_report_lifecycle_exited(self, client):
        """Test reporting container exited event."""
        mock_response = Mock()
        mock_response.status_code = 200

        event = ContainerLifecycleEvent(
            event_type="exited",
            container_id="container-123",
            pod_name="pod-456",
            executor_port=8080,
            exit_code=0,
            exit_reason=ExitReason.NORMAL,
            exited_at=datetime.now(),
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_lifecycle(event)

            assert result is True

    @pytest.mark.asyncio
    async def test_report_lifecycle_failure(self, client):
        """Test lifecycle reporting failure."""
        mock_response = Mock()
        mock_response.status_code = 500
        mock_response.text = "Server error"

        event = ContainerLifecycleEvent(
            event_type="ready",
            container_id="container-123",
            executor_port=8080,
            ready_at=datetime.now(),
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(return_value=mock_response)
            mock_get_client.return_value = mock_client

            result = await client.report_lifecycle(event)

            assert result is False

    @pytest.mark.asyncio
    async def test_report_lifecycle_exception(self, client):
        """Test lifecycle reporting with exception."""
        event = ContainerLifecycleEvent(
            event_type="ready",
            container_id="container-123",
            executor_port=8080,
            ready_at=datetime.now(),
        )

        with patch.object(client, '_get_client') as mock_get_client:
            mock_client = AsyncMock()
            mock_client.post = AsyncMock(side_effect=Exception("Network error"))
            mock_get_client.return_value = mock_client

            result = await client.report_lifecycle(event)

            assert result is False


class TestPersistResult:
    """Tests for local persistence methods."""

    @pytest.fixture
    def client(self):
        """Create a CallbackClient with temp directory."""
        with tempfile.TemporaryDirectory() as tmpdir:
            client = CallbackClient(
                control_plane_url="http://test.invalid",
                api_token="test-token",
            )
            client.results_dir = Path(tmpdir)
            yield client

    @pytest.mark.asyncio
    async def test_persist_result(self, client):
        """Test persisting result locally."""
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        await client._persist_result("exec_001", result)

        # Verify file was created
        file_path = client.results_dir / "exec_001.json"
        assert file_path.exists()

        # Verify content
        with open(file_path) as f:
            data = json.load(f)

        assert data["execution_id"] == "exec_001"
        assert data["result"]["status"] == "success"

    @pytest.mark.asyncio
    async def test_get_persisted_result(self, client):
        """Test retrieving persisted result."""
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        await client._persist_result("exec_001", result)

        data = await client.get_persisted_result("exec_001")

        assert data is not None
        assert data["execution_id"] == "exec_001"

    @pytest.mark.asyncio
    async def test_get_persisted_result_not_found(self, client):
        """Test retrieving non-existent result."""
        data = await client.get_persisted_result("nonexistent")
        assert data is None

    @pytest.mark.asyncio
    async def test_delete_persisted_result(self, client):
        """Test deleting persisted result."""
        result = ExecutionResult(
            status=ExecutionStatus.SUCCESS,
            stdout="output",
            stderr="",
            exit_code=0,
            execution_time_ms=100,
        )

        await client._persist_result("exec_001", result)

        # Verify file exists
        file_path = client.results_dir / "exec_001.json"
        assert file_path.exists()

        # Delete
        await client.delete_persisted_result("exec_001")

        # Verify file is gone
        assert not file_path.exists()


class TestClose:
    """Tests for close method."""

    def test_close_with_client(self):
        """Test closing with initialized client."""
        client = CallbackClient(
            control_plane_url="http://test.invalid",
            api_token="test-token",
        )

        async def run_test():
            mock_client = AsyncMock()
            client._client = mock_client

            await client.close()

            mock_client.aclose.assert_called_once()
            assert client._client is None

        import asyncio
        asyncio.run(run_test())

    @pytest.mark.asyncio
    async def test_close_without_client(self):
        """Test closing without initialized client."""
        client = CallbackClient(
            control_plane_url="http://test.invalid",
            api_token="test-token",
        )

        # Should not raise error
        await client.close()


class TestGlobalCallbackClient:
    """Tests for global callback client functions."""

    def test_get_callback_client_creates_instance(self):
        """Test that get_callback_client creates instance."""
        import executor.infrastructure.http.callback_client as module
        module._callback_client = None

        with patch("executor.infrastructure.http.callback_client.settings") as mock_settings:
            mock_settings.control_plane_url = "http://localhost:8000"
            mock_settings.internal_api_token = "test-token"

            client = get_callback_client()
            assert client is not None

        # Reset
        module._callback_client = None
