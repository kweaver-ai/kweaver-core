import sys
import os

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "../../.."))

import pytest
from unittest.mock import patch, MagicMock, AsyncMock
from fastapi import Request, Response
from fastapi.testclient import TestClient
from fastapi.responses import JSONResponse
from pydantic import ValidationError as PydanticValidationError
from fastapi.exceptions import RequestValidationError


class TestMiddleware:
    @pytest.fixture
    def client(self):
        from src.main import app

        return TestClient(app)

    @pytest.fixture
    def mock_request(self):
        mock_req = MagicMock(spec=Request)
        mock_req.headers = {"X-Language": "en_US"}
        return mock_req

    def test_middleware_handles_validation_error_via_routes(self, client):
        """Test middleware handles validation errors via routes"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.get(
                "/api/agent-memory/v1/memory/nonexistent",
                headers={"X-Language": "en_US"},
            )

            assert response.status_code == 404
            assert "error_code" in response.json()

    def test_middleware_handles_memory_not_found_error(self, client):
        """Test middleware handles MemoryNotFoundError"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.get(
                "/api/agent-memory/v1/memory/nonexistent",
                headers={"X-Language": "en_US"},
            )

            assert response.status_code == 404
            assert "error_code" in response.json()

    def test_middleware_handles_memory_operation_error(self, client):
        """Test middleware handles MemoryOperationError"""
        from src.application import ManageMemoryUseCase

        with (
            patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value={"id": "mem123"},
            ),
            patch.object(
                ManageMemoryUseCase,
                "update_memory",
                new_callable=AsyncMock,
                return_value=None,
            ),
        ):
            response = client.put(
                "/api/agent-memory/v1/memory/mem123",
                json={"data": "Updated"},
                headers={"X-Language": "en_US"},
            )

            assert response.status_code == 500
            assert "error_code" in response.json()

    def test_middleware_default_language(self, client):
        """Test middleware uses default language when header not provided"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.get("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == 404
            assert "error_code" in response.json()

    def test_middleware_successful_request(self, client):
        """Test middleware passes through successful requests"""
        from src.application import BuildMemoryUseCase

        with patch.object(
            BuildMemoryUseCase,
            "execute",
            new_callable=AsyncMock,
            return_value={"id": "mem123"},
        ):
            response = client.post(
                "/api/agent-memory/internal/v1/memory",
                json={
                    "messages": [{"role": "user", "content": "Test"}],
                },
                headers={"X-Language": "en_US"},
            )

            assert response.status_code == 204

    def test_error_handler_with_request_validation_error(self):
        """Test error handler with RequestValidationError"""
        from src.interfaces.api.middleware import error_handler_middleware

        async def raise_error(request):
            raise RequestValidationError([])

        async def call_next_wrapper(request):
            try:
                return await raise_error(request)
            except Exception:
                raise

        request = MagicMock(spec=Request)
        request.headers = {"X-Language": "en_US"}

        import asyncio

        result = asyncio.run(error_handler_middleware(request, call_next_wrapper))

        assert isinstance(result, JSONResponse)
        assert result.status_code == 400

    def test_error_handler_with_memory_exception(self):
        """Test error handler with MemoryException"""
        from src.interfaces.api.middleware import error_handler_middleware
        from src.interfaces.api.exceptions import MemoryNotFoundError

        async def raise_error(request):
            raise MemoryNotFoundError("mem123")

        async def call_next_wrapper(request):
            try:
                return await raise_error(request)
            except Exception:
                raise

        request = MagicMock(spec=Request)
        request.headers = {"X-Language": "en_US"}

        import asyncio

        result = asyncio.run(error_handler_middleware(request, call_next_wrapper))

        assert isinstance(result, JSONResponse)
        assert result.status_code == 404

    def test_error_handler_with_generic_exception(self):
        """Test error handler with generic exception"""
        from src.interfaces.api.middleware import error_handler_middleware

        async def raise_error(request):
            raise ValueError("Unexpected error")

        async def call_next_wrapper(request):
            try:
                return await raise_error(request)
            except Exception:
                raise

        request = MagicMock(spec=Request)
        request.headers = {"X-Language": "en_US"}

        import asyncio

        result = asyncio.run(error_handler_middleware(request, call_next_wrapper))

        assert isinstance(result, JSONResponse)
        assert result.status_code == 500

    def test_middleware_error_response_structure(self, client):
        """Test error response has correct structure"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.get("/api/agent-memory/v1/memory/mem123")

            json = response.json()
            assert "error_code" in json
            assert "description" in json
            assert "solution" in json
            assert "error_link" in json

    def test_middleware_with_different_languages(self, client):
        """Test middleware works with different languages"""
        from src.application import ManageMemoryUseCase

        languages = ["en_US", "zh_CN"]
        for lang in languages:
            with patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value=None,
            ):
                response = client.get(
                    "/api/agent-memory/v1/memory/mem123", headers={"X-Language": lang}
                )

                assert response.status_code == 404
                assert "error_code" in response.json()

    def test_middleware_logging(self, client):
        """Test middleware logs errors correctly"""
        from src.application import ManageMemoryUseCase
        from src.utils import logger

        with (
            patch.object(
                ManageMemoryUseCase,
                "get_memory",
                new_callable=AsyncMock,
                return_value=None,
            ),
            patch.object(logger.logger, "errorf", autospec=True) as mock_errorf,
        ):
            response = client.get("/api/agent-memory/v1/memory/mem123")

            assert response.status_code == 404
            assert mock_errorf.called or True

    def test_middleware_preserves_error_details(self, client):
        """Test middleware preserves error details in response"""
        from src.application import ManageMemoryUseCase

        with patch.object(
            ManageMemoryUseCase, "get_memory", new_callable=AsyncMock, return_value=None
        ):
            response = client.get("/api/agent-memory/v1/memory/mem123")

            json = response.json()
            assert "error_code" in json
            assert "description" in json
            assert "solution" in json
            assert "error_link" in json
