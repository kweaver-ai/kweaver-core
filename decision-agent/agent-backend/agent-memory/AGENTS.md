# AGENTS.md - Agent Memory Service Development Guide

This guide provides essential information for agentic coding agents working in the Agent Memory Service.

## Project Overview

Agent Memory is a Python FastAPI service using hexagonal architecture for building, retrieving, and managing AI agent memory data.

## Development Commands

```bash
# Install dependencies
pip install -r requirements.txt
pip install -r requirements-dev.txt

# Run the service
python src/main.py

# Testing
make test-unit           # Run unit tests with coverage
make test                # Run all tests
make coverage            # Generate coverage report (htmlcov/index.html)

# Single Test Execution
pytest tests/unit/test_file.py -v
pytest tests/unit/test_file.py::TestClass::test_function -v

# Code Quality
make lint                # Run flake8 with max-line-length=120
make format              # Format code with autopep8
make clean               # Clean test artifacts
```

## Directory Structure (Hexagonal Architecture)

```
src/
├── application/          # Application layer (use cases)
│   └── memory/          # Build, manage, retrieval memory use cases
├── config/              # Configuration (config.yaml, config.py)
├── domain/              # Domain layer (entities, repositories)
│   └── memory/          # Memory entities, mem0 adapter, repositories
├── infrastructure/      # Infrastructure (database access)
│   └── db/              # Database connection pool
├── interfaces/          # Interface layer (API)
│   └── api/             # Routes, schemas, exceptions, middleware
├── locale/              # i18n (en_US, zh_CN error messages)
└── utils/               # Utilities (logger, i18n)
```

## Code Style Guidelines

### Import Organization
```python
# Standard library imports
import sys
from pathlib import Path
from typing import List, Optional, Dict, Any

# Third-party imports
from fastapi import APIRouter, HTTPException, Depends
from pydantic import BaseModel, Field

# Local imports
from src.domain.memory.entities import Memory
from src.utils.logger import logger
```

### Naming Conventions
- **Classes/Pydantic Models**: PascalCase (e.g., `BuildMemoryUseCase`, `MemoryResponse`)
- **Functions/Variables**: snake_case (e.g., `build_memory`, `user_id`)
- **Constants**: UPPER_SNAKE_CASE (e.g., `MAX_RETRIES`)
- **Private members**: underscore prefix (e.g., `_initialized`)

### Pydantic Models
```python
class BuildMemoryRequest(BaseModel):
    messages: List[Message] = Field(..., description="Message list")
    user_id: Optional[str] = Field(None, description="User ID")
    metadata: Optional[Dict[str, Any]] = Field(None, description="Metadata")

    @model_validator(mode="after")
    def validate_fields(self):
        if not self.messages:
            raise ValueError("At least one message required")
        return self
```

### Error Handling
```python
# Custom exceptions in interfaces/api/exceptions.py
class MemoryNotFoundError(MemoryException):
    def __init__(self, memory_id: str):
        super().__init__(
            message=f"Memory not found: {memory_id}",
            status_code=404,
            error_code="AgentMemory.NotFound.Memory",
            details={"memory_id": memory_id}
        )
```

### Logging
```python
from src.utils.logger import logger
logger.info("Memory built successfully")
logger.infof("Retrieving memories, query:%s, user_id:%s", query, user_id)
logger.debugf("Building memory, params=%s", request.model_dump())
```

### FastAPI Routes
```python
internal_router = APIRouter(prefix="/api/agent-memory/internal/v1", tags=["memory"])

@internal_router.post("/memory", response_model=MemoryResponse)
async def build_memory(request: Request, build_request: BuildMemoryRequest):
    context = {"user_id": request.headers.get("x-account-id")}
    result = await use_case.execute(..., context=context)
    return Response(status_code=status.HTTP_204_NO_CONTENT)
```

## Testing Patterns

### Framework: pytest with asyncio
```python
import pytest
from unittest.mock import patch, AsyncMock

class TestBuildMemoryUseCase:
    @pytest.fixture
    def use_case(self):
        return BuildMemoryUseCase()

    @pytest.mark.asyncio
    async def test_execute_with_messages(self, use_case):
        mock_adapter = AsyncMock()
        mock_adapter.add.return_value = {"id": "mem123"}
        use_case.memory_adapter = mock_adapter
        result = await use_case.execute(messages=[{"role": "user", "content": "Test"}])
        mock_adapter.add.assert_called_once()
        assert result["id"] == "mem123"
```

### Test Markers: unit, integration, slow

## Configuration

Edit `src/config/config.yaml` for:
- LLM settings (provider, model, API URL)
- Embedding model configuration
- Vector database (OpenSearch) settings
- Server host/port

## Key Notes

- Test coverage must be >= 90% (configured in pyproject.toml)
- Max line length: 120 characters
- Use async/await for all FastAPI endpoints and use cases
- Exclude `mem0/` directory from linting (third-party code)
- Use dependency injection pattern with FastAPI Depends()
- All timestamps in ISO 8601 format (e.g., "2024-01-01T00:00:00Z")
- Always return `Response(status_code=status.HTTP_204_NO_CONTENT)` for successful mutations
