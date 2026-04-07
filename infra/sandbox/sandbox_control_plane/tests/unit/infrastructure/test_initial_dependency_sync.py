"""
Tests for initial dependency sync background task orchestration.
"""
from contextlib import asynccontextmanager
from unittest.mock import AsyncMock, Mock

import pytest

from src.domain.value_objects.execution_status import SessionStatus
from src.infrastructure import dependencies as dependencies_module


class _FakeSession:
    def __init__(self, status, container_id=None, requested_dependencies=None):
        self.id = "sess_test"
        self.status = status
        self.container_id = container_id
        self.requested_dependencies = requested_dependencies or []

    def mark_dependency_install_failed(self, error: str) -> None:
        self.error = error


class _FakeSessionRepository:
    responses = []

    def __init__(self, session, execution_repo):
        self._session = session
        self._execution_repo = execution_repo

    async def find_by_id(self, session_id: str):
        if not self.responses:
            return None
        return self.responses.pop(0)

    async def save(self, session):
        return None


class _FakeExecutionRepository:
    def __init__(self, session):
        self._session = session


class _FakeTemplateRepository:
    def __init__(self, session):
        self._session = session


class _FakeRuntimeNodeRepository:
    def __init__(self, session):
        self._session = session


@pytest.mark.asyncio
async def test_initial_dependency_sync_retries_until_session_visible(monkeypatch):
    seen_sync = AsyncMock()

    class FakeSessionService:
        def __init__(self, **kwargs):
            pass

        async def sync_session_dependencies_for_session(self, session_id: str, sync_mode: str):
            await seen_sync(session_id, sync_mode)

    @asynccontextmanager
    async def fake_get_session():
        yield object()

    _FakeSessionRepository.responses = [
        None,
        _FakeSession(
            status=SessionStatus.RUNNING,
            container_id="container-1",
            requested_dependencies=["pyfiglet==1.0.2"],
        ),
    ]

    monkeypatch.setattr(dependencies_module.db_manager, "get_session", fake_get_session)
    monkeypatch.setattr(
        "src.infrastructure.persistence.repositories.sql_session_repository.SqlSessionRepository",
        _FakeSessionRepository,
    )
    monkeypatch.setattr(
        "src.infrastructure.persistence.repositories.sql_execution_repository.SqlExecutionRepository",
        _FakeExecutionRepository,
    )
    monkeypatch.setattr(
        "src.infrastructure.persistence.repositories.sql_template_repository.SqlTemplateRepository",
        _FakeTemplateRepository,
    )
    monkeypatch.setattr(
        "src.infrastructure.persistence.repositories.sql_runtime_node_repository.SqlRuntimeNodeRepository",
        _FakeRuntimeNodeRepository,
    )
    monkeypatch.setattr(dependencies_module, "SessionService", FakeSessionService)
    monkeypatch.setattr(dependencies_module, "_create_scheduler_service", lambda **kwargs: Mock())
    monkeypatch.setattr(dependencies_module, "get_storage_service", lambda: Mock())

    sleep_calls = []

    async def fake_sleep(seconds: float):
        sleep_calls.append(seconds)

    monkeypatch.setattr(dependencies_module.asyncio, "sleep", fake_sleep)

    await dependencies_module._run_initial_dependency_sync("sess_test", install_timeout=5)

    seen_sync.assert_awaited_once_with("sess_test", "replace")
    assert sleep_calls == [1.0]


@pytest.mark.asyncio
async def test_initial_dependency_sync_scheduler_marks_session_failed_on_unexpected_error(
    monkeypatch,
):
    seen_failure = AsyncMock()

    async def fake_run_initial_dependency_sync(session_id: str, install_timeout: int) -> None:
        raise RuntimeError("pip install failed for fastapi==0.13.5")

    async def fake_mark_initial_dependency_sync_failed(session_id: str, error: str) -> None:
        await seen_failure(session_id, error)

    created_tasks = []

    def fake_create_task(coro):
        created_tasks.append(coro)
        return Mock()

    monkeypatch.setattr(
        dependencies_module,
        "_run_initial_dependency_sync",
        fake_run_initial_dependency_sync,
    )
    monkeypatch.setattr(
        dependencies_module,
        "_mark_initial_dependency_sync_failed",
        fake_mark_initial_dependency_sync_failed,
    )
    monkeypatch.setattr(dependencies_module.asyncio, "create_task", fake_create_task)

    schedule = dependencies_module.get_initial_dependency_sync_scheduler()
    scheduled = schedule("sess_test", 30)

    assert scheduled is None
    assert len(created_tasks) == 1

    await created_tasks[0]

    seen_failure.assert_awaited_once()
    args = seen_failure.await_args.args
    assert args[0] == "sess_test"
    assert "pip install failed for fastapi==0.13.5" in args[1]
