"""
Control Plane callback client for reporting execution results.

Implements ICallbackPort for Control Plane communication.
Handles asynchronous HTTP callbacks with:
- Retry logic with exponential backoff
- Local persistence fallback
- Authorization headers
- Idempotency support
"""

import httpx
import asyncio
import json
import os
import math
from pathlib import Path
from typing import Optional, Any
from datetime import datetime

from executor.domain.ports import ICallbackPort
from executor.domain.value_objects import ExecutionResult, HeartbeatSignal, ContainerLifecycleEvent
from executor.infrastructure.config import settings
from executor.infrastructure.logging import get_logger


logger = get_logger()


class CallbackClient(ICallbackPort):
    """
    Async HTTP client for reporting execution results to Control Plane.

    Implements ICallbackPort interface.

    Features:
    - Exponential backoff retry (1s, 2s, 4s, 8s, max 10s)
    - Local persistence on final failure
    - Bearer token authentication
    - Idempotency keys for retry deduplication
    """

    def __init__(
        self,
        control_plane_url: Optional[str] = None,
        api_token: Optional[str] = None,
        max_retries: int = 5,
        base_retry_delay: float = 1.0,
        max_retry_delay: float = 10.0,
    ):
        """
        Initialize callback client.

        Args:
            control_plane_url: Control Plane base URL
            api_token: API token for authentication
            max_retries: Maximum number of retry attempts
            base_retry_delay: Base delay for exponential backoff (seconds)
            max_retry_delay: Maximum delay between retries (seconds)
        """
        self.control_plane_url = control_plane_url or settings.control_plane_url
        self.api_token = api_token or settings.internal_api_token
        self.max_retries = max_retries
        self.base_retry_delay = base_retry_delay
        self.max_retry_delay = max_retry_delay

        # Timeout settings (5s connect, 30s read)
        self.timeout = httpx.Timeout(5.0, connect=5.0, read=30.0)

        # Results directory for local persistence
        self.results_dir = Path("/tmp/results")
        self.results_dir.mkdir(parents=True, exist_ok=True)

        # Lazy-initialized async client
        self._client: Optional[httpx.AsyncClient] = None

    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create async HTTP client."""
        if self._client is None:
            self._client = httpx.AsyncClient(timeout=self.timeout)
        return self._client

    async def close(self) -> None:
        """
        Close the HTTP client.

        Implementation of ICallbackPort.close().
        """
        if self._client:
            await self._client.aclose()
            self._client = None

    def _sanitize_for_json(self, data: Any) -> Any:
        """
        Sanitize data for JSON serialization.

        Converts non-JSON-compliant float values (NaN, Infinity, -Infinity) to None.
        Recursively handles nested dictionaries and lists.

        Args:
            data: Data to sanitize

        Returns:
            Sanitized data with JSON-compliant values
        """
        if isinstance(data, float):
            # Check for non-JSON-compliant float values
            if math.isnan(data) or math.isinf(data):
                return None
            return data
        elif isinstance(data, dict):
            return {k: self._sanitize_for_json(v) for k, v in data.items()}
        elif isinstance(data, list):
            return [self._sanitize_for_json(item) for item in data]
        else:
            return data

    async def report_result(
        self,
        execution_id: str,
        result: ExecutionResult,
    ) -> bool:
        """
        Report execution result to Control Plane.

        Implementation of ICallbackPort.report_result().

        Args:
            execution_id: Unique execution identifier
            result: Execution result to report

        Returns:
            True if report succeeded, False otherwise

        Retries:
            - Exponential backoff: 1s, 2s, 4s, 8s, max 10s
            - Max 5 retry attempts
            - Retries on: network errors, 5xx responses
        """
        url = f"{self.control_plane_url}/api/v1/internal/executions/{execution_id}/result"

        # Build request payload
        # Convert execution_time_ms to execution_time (seconds)
        execution_time = result.execution_time_ms / 1000.0 if result.execution_time_ms else 0.0

        # Map internal ExecutionStatus to Control Plane API status values
        # Internal: "completed" → API: "success"
        # Internal: "failed" → API: "failed"
        # Internal: "timeout" → API: "timeout"
        # Internal: "crashed" → API: "crashed"
        status_mapping = {
            "completed": "success",
            "failed": "failed",
            "timeout": "timeout",
            "crashed": "crashed",
            "error": "failed",  # Map error status to failed
        }
        api_status = status_mapping.get(result.status.value, result.status.value)

        # Sanitize return_value and metrics for JSON serialization
        sanitized_return_value = self._sanitize_for_json(result.return_value) if result.return_value else None
        sanitized_metrics = self._sanitize_for_json(result.metrics.to_dict()) if result.metrics else None

        payload = {
            "status": api_status,
            "stdout": result.stdout,
            "stderr": result.stderr,
            "exit_code": result.exit_code,
            "execution_time": execution_time,
            "return_value": sanitized_return_value,
            "metrics": sanitized_metrics,
            # Send file paths (strings) instead of full artifact objects
            "artifacts": [str(a.path) for a in result.artifacts],
        }
        print("report body:",payload)

        # Headers with auth and idempotency
        headers = {
            "Authorization": f"Bearer {self.api_token}",
            "Content-Type": "application/json",
            "Idempotency-Key": execution_id,  # Idempotency support
        }

        # Implement retry logic with exponential backoff
        last_error = None

        for attempt in range(self.max_retries):
            try:
                client = await self._get_client()

                logger.debug(
                    "Reporting result to Control Plane",
                    execution_id=execution_id,
                    attempt=attempt + 1,
                    max_retries=self.max_retries,
                )

                # Log the payload for debugging
                logger.info(
                    "Sending execution result to Control Plane",
                    execution_id=execution_id,
                    internal_status=result.status.value,
                    api_status=api_status,
                    exit_code=result.exit_code,
                )

                response = await client.post(url, json=payload, headers=headers)

                # Success on 2xx response
                if 200 <= response.status_code < 300:
                    logger.info(
                        "Result reported successfully",
                        execution_id=execution_id,
                        status_code=response.status_code,
                    )
                    return True

                # Retry on 401 Unauthorized
                if response.status_code == 401:
                    logger.warning(
                        "Callback unauthorized - will retry",
                        execution_id=execution_id,
                        attempt=attempt + 1,
                        status_code=response.status_code,
                    )
                    last_error = f"Unauthorized (401)"
                    await self._backoff(attempt)
                    continue

                # Retry on 5xx errors
                if response.status_code >= 500:
                    logger.warning(
                        "Callback server error - will retry",
                        execution_id=execution_id,
                        attempt=attempt + 1,
                        status_code=response.status_code,
                    )
                    last_error = f"Server error ({response.status_code})"
                    await self._backoff(attempt)
                    continue

                # Don't retry on 4xx (except 401)
                logger.error(
                    "Callback failed with client error",
                    execution_id=execution_id,
                    status_code=response.status_code,
                    response=response.text,
                )
                return False

            except httpx.TimeoutException as e:
                # Retry on network timeout
                logger.warning(
                    "Callback timeout - will retry",
                    execution_id=execution_id,
                    attempt=attempt + 1,
                    error=str(e),
                )
                last_error = f"Timeout: {e}"
                await self._backoff(attempt)
                continue

            except httpx.NetworkError as e:
                # Retry on network errors
                logger.warning(
                    "Callback network error - will retry",
                    execution_id=execution_id,
                    attempt=attempt + 1,
                    error=str(e),
                )
                last_error = f"Network error: {e}"
                await self._backoff(attempt)
                continue

            except Exception as e:
                logger.error(
                    "Callback unexpected error",
                    execution_id=execution_id,
                    attempt=attempt + 1,
                    error=str(e),
                    error_type=type(e).__name__,
                )
                last_error = f"Unexpected error: {e}"
                # Don't retry on unexpected errors
                break

        # All retries exhausted
        logger.error(
            "Callback failed after all retries",
            execution_id=execution_id,
            last_error=last_error,
            retries_attempted=self.max_retries,
        )

        # Implement local persistence fallback
        await self._persist_result(execution_id, result)

        return False

    async def report_heartbeat(
        self,
        execution_id: str,
        signal: HeartbeatSignal,
    ) -> bool:
        """
        Report heartbeat signal to Control Plane.

        Implementation of ICallbackPort.report_heartbeat().

        Args:
            execution_id: Unique execution identifier
            signal: Heartbeat signal to report

        Returns:
            True if report succeeded, False otherwise
        """
        url = f"{self.control_plane_url}/api/v1/internal/executions/{execution_id}/heartbeat"

        # Build heartbeat payload
        payload = {
            "timestamp": signal.timestamp.isoformat() if signal.timestamp else datetime.now().isoformat(),
            "progress": signal.progress,
        }

        headers = {
            "Authorization": f"Bearer {self.api_token}",
            "Content-Type": "application/json",
        }

        try:
            client = await self._get_client()
            response = await client.post(url, json=payload, headers=headers)

            if 200 <= response.status_code < 300:
                logger.debug(
                    "Heartbeat reported successfully",
                    execution_id=execution_id,
                )
                return True

            logger.warning(
                "Heartbeat callback failed",
                execution_id=execution_id,
                status_code=response.status_code,
            )
            return False

        except Exception as e:
            # Log heartbeat errors but don't fail execution
            logger.warning(
                "Heartbeat callback error (non-fatal)",
                execution_id=execution_id,
                error=str(e),
            )
            return False

    async def report_lifecycle(
        self,
        event: ContainerLifecycleEvent,
    ) -> bool:
        """
        Report container lifecycle event to Control Plane.

        Implementation of ICallbackPort.report_lifecycle().

        Args:
            event: Lifecycle event (ready or exited)

        Returns:
            True if report succeeded, False otherwise
        """
        event_type = event.event_type
        url = f"{self.control_plane_url}/api/v1/internal/containers/{event_type}"

        # Build lifecycle payload
        payload = event.to_dict()

        headers = {
            "Authorization": f"Bearer {self.api_token}",
            "Content-Type": "application/json",
        }

        try:
            client = await self._get_client()
            response = await client.post(url, json=payload, headers=headers)

            if 200 <= response.status_code < 300:
                logger.info(
                    f"Lifecycle event sent: {event_type}",
                    status_code=response.status_code,
                )
                return True

            logger.warning(
                f"Lifecycle event failed: {event_type}",
                status_code=response.status_code,
                response=response.text,
            )
            return False

        except Exception as e:
            logger.error(
                f"Failed to send lifecycle event: {event_type}",
                error=str(e),
            )
            return False

    async def _backoff(self, attempt: int):
        """
        Calculate and wait exponential backoff delay.

        Delays: 1s, 2s, 4s, 8s, max 10s

        Args:
            attempt: Current attempt number (0-indexed)
        """
        delay = min(self.base_retry_delay * (2 ** attempt), self.max_retry_delay)

        logger.debug(
            "Backing off before retry",
            attempt=attempt + 1,
            delay_seconds=delay,
        )

        await asyncio.sleep(delay)

    async def _persist_result(self, execution_id: str, result: ExecutionResult):
        """
        Persist execution result locally after callback failure.

        Args:
            execution_id: Unique execution identifier
            result: Execution result to persist
        """
        try:
            file_path = self.results_dir / f"{execution_id}.json"

            # Build persistence payload with sanitization
            payload = {
                "execution_id": execution_id,
                "result": self._sanitize_for_json(result.to_dict()),
                "persisted_at": datetime.now().isoformat(),
                "control_plane_url": self.control_plane_url,
            }

            # Write to file
            with open(file_path, "w") as f:
                json.dump(payload, f, indent=2)

            logger.info(
                "Result persisted locally after callback failure",
                execution_id=execution_id,
                file_path=str(file_path),
            )

        except Exception as e:
            logger.error(
                "Failed to persist result locally",
                execution_id=execution_id,
                error=str(e),
            )

    async def get_persisted_result(self, execution_id: str) -> Optional[dict]:
        """
        Retrieve locally persisted result for an execution.

        Args:
            execution_id: Unique execution identifier

        Returns:
            Persisted result dict, or None if not found
        """
        try:
            file_path = self.results_dir / f"{execution_id}.json"

            if not file_path.exists():
                return None

            with open(file_path, "r") as f:
                payload = json.load(f)

            logger.debug(
                "Retrieved persisted result",
                execution_id=execution_id,
                file_path=str(file_path),
            )

            return payload

        except Exception as e:
            logger.error(
                "Failed to retrieve persisted result",
                execution_id=execution_id,
                error=str(e),
            )
            return None

    async def delete_persisted_result(self, execution_id: str):
        """
        Delete locally persisted result after successful retry.

        Args:
            execution_id: Unique execution identifier
        """
        try:
            file_path = self.results_dir / f"{execution_id}.json"

            if file_path.exists():
                file_path.unlink()

                logger.debug(
                    "Deleted persisted result",
                    execution_id=execution_id,
                )

        except Exception as e:
            logger.warning(
                "Failed to delete persisted result",
                execution_id=execution_id,
                error=str(e),
            )


# Global callback client instance
_callback_client: Optional[CallbackClient] = None


def get_callback_client() -> CallbackClient:
    """
    Get global callback client instance.

    Returns:
        CallbackClient instance
    """
    global _callback_client

    if _callback_client is None:
        _callback_client = CallbackClient()

    return _callback_client
