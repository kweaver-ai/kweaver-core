"""
Request logging middleware for tracing and context management.

Adds request_id to all logs within a request context for better tracing.
"""

import time
import uuid
from typing import Callable

from fastapi import Request, Response
from starlette.middleware.base import BaseHTTPMiddleware

from src.infrastructure.logging import get_logger, bind_context, clear_context

logger = get_logger(__name__)


class RequestLoggingMiddleware(BaseHTTPMiddleware):
    """
    Middleware that adds request context to all logs.

    Features:
    - Generates unique request_id for each request
    - Binds request context to all logs within the request
    - Logs request start/end with timing information
    - Adds X-Request-ID header to response
    """

    async def dispatch(self, request: Request, call_next: Callable) -> Response:
        """
        Process request with logging context.

        Args:
            request: Incoming HTTP request
            call_next: Next middleware/handler in chain

        Returns:
            HTTP response with X-Request-ID header
        """
        # Generate unique request ID
        request_id = str(uuid.uuid4())
        request.state.request_id = request_id

        # Bind context for all logs in this request
        bind_context(
            request_id=request_id,
            method=request.method,
            path=request.url.path,
        )

        # Log request start
        logger.debug(
            "Request started",
            request_id=request_id,
            method=request.method,
            path=request.url.path,
            client=request.client.host if request.client else None,
        )

        # Track start time
        start_time = time.time()

        try:
            # Process request
            response = await call_next(request)

            # Calculate processing time
            process_time = time.time() - start_time

            # Add request ID to response headers
            response.headers["X-Request-ID"] = request_id
            response.headers["X-Process-Time"] = str(f"{process_time:.3f}")

            # Log request completion (only for INFO level and above)
            logger.info(
                "Request completed",
                request_id=request_id,
                method=request.method,
                path=request.url.path,
                status_code=response.status_code,
                process_time=f"{process_time:.3f}s",
            )

            return response

        except Exception as e:
            # Log error with context
            process_time = time.time() - start_time
            logger.exception(
                "Request failed",
                request_id=request_id,
                method=request.method,
                path=request.url.path,
                error=str(e),
                process_time=f"{process_time:.3f}s",
            )
            raise

        finally:
            # Clear context for next request
            clear_context()
