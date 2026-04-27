"""
Result Reporter Adapter

Handles reporting execution results back to the Control Plane via HTTP callbacks.
"""

import httpx
from typing import Optional
import structlog

from executor.domain.value_objects import ExecutionResult


logger = structlog.get_logger(__name__)


class ResultReporter:
    """
    Reports execution results to the Control Plane.

    Implements the callback mechanism for the executor to notify
    the Control Plane of execution completion.
    """

    def __init__(self, control_plane_url: str, internal_api_token: Optional[str] = None):
        """
        Initialize the result reporter.

        Args:
            control_plane_url: Base URL of the Control Plane
            internal_api_token: Optional API token for internal API authentication
        """
        self.control_plane_url = control_plane_url.rstrip("/")
        self.internal_api_token = internal_api_token
        self._client: Optional[httpx.AsyncClient] = None

    async def report(self, execution_id: str, result: ExecutionResult) -> None:
        """
        Report execution result to Control Plane.

        Args:
            execution_id: Unique execution identifier
            result: Execution result to report

        Raises:
            httpx.HTTPError: If the HTTP request fails
        """
        url = f"{self.control_plane_url}/internal/executions/{execution_id}/result"

        headers = {"Content-Type": "application/json"}
        if self.internal_api_token:
            headers["Authorization"] = f"Bearer {self.internal_api_token}"

        payload = result.to_dict()

        logger.info(
            "Reporting execution result",
            execution_id=execution_id,
            url=url,
            status=result.status.value,
        )

        client = await self._get_client()
        response = await client.post(url, json=payload, headers=headers, timeout=10.0)

        response.raise_for_status()

        logger.info(
            "Result reported successfully",
            execution_id=execution_id,
            status_code=response.status_code,
        )

    async def _get_client(self) -> httpx.AsyncClient:
        """Get or create HTTP client."""
        if self._client is None:
            self._client = httpx.AsyncClient(
                limits=httpx.Limits(max_connections=100, max_keepalive_connections=20),
                timeout=httpx.Timeout(10.0),
            )
        return self._client

    async def close(self) -> None:
        """Close the HTTP client."""
        if self._client:
            await self._client.aclose()
            self._client = None

    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.close()
