from dataclasses import dataclass
from datetime import datetime, timedelta
from decimal import Decimal

import pytest

from src.domain.entities.execution import Execution
from src.domain.entities.session import InstalledDependency, Session
from src.domain.entities.template import Template
from src.domain.services.scheduler import RuntimeNode
from src.domain.value_objects.execution_status import ExecutionState, ExecutionStatus, SessionStatus
from src.domain.value_objects.resource_limit import ResourceLimit
from src.infrastructure.persistence.database import Base  # noqa: F401
from src.infrastructure.persistence.models.execution_model import ExecutionModel
from src.infrastructure.persistence.models.runtime_node_model import RuntimeNodeModel
from src.infrastructure.persistence.models.session_model import SessionModel
from src.infrastructure.persistence.models.template_model import TemplateModel
from src.infrastructure.persistence.repositories.sql_execution_repository import (
    SqlExecutionRepository,
)
from src.infrastructure.persistence.repositories.sql_runtime_node_repository import (
    SqlRuntimeNodeRepository,
)
from src.infrastructure.persistence.repositories.sql_session_repository import SqlSessionRepository
from src.infrastructure.persistence.repositories.sql_template_repository import (
    SqlTemplateRepository,
)


class FakeScalarResult:
    def __init__(self, items):
        self._items = items

    def all(self):
        return self._items


class FakeResult:
    def __init__(self, items=None, scalar_value=None):
        self._items = items or []
        self._scalar_value = scalar_value

    def scalar_one_or_none(self):
        return self._items[0] if self._items else None

    def scalar(self):
        return self._scalar_value

    def scalars(self):
        return FakeScalarResult(self._items)


class FakeAsyncSession:
    def __init__(self, get_map=None, execute_results=None):
        self.get_map = get_map or {}
        self.execute_results = list(execute_results or [])
        self.added = []
        self.executed = []
        self.flushed = 0
        self.committed = 0

    async def get(self, model, key):
        return self.get_map.get((model, key))

    async def execute(self, stmt):
        self.executed.append(stmt)
        return self.execute_results.pop(0) if self.execute_results else FakeResult()

    def add(self, model):
        self.added.append(model)

    async def flush(self):
        self.flushed += 1

    async def commit(self):
        self.committed += 1


def make_session(session_id="sess-1", status=SessionStatus.RUNNING):
    return Session(
        id=session_id,
        template_id="tpl-1",
        status=status,
        resource_limit=ResourceLimit(cpu="1", memory="1Gi", disk="2Gi"),
        workspace_path=f"s3://bucket/sessions/{session_id}",
        runtime_type="docker",
        runtime_node="node-1",
        container_id="container-1",
        pod_name="pod-1",
        env_vars={"A": "B"},
        timeout=600,
        requested_dependencies=["requests==2.32.0"],
        installed_dependencies=[
            InstalledDependency(
                name="requests",
                version="2.32.0",
                install_location="/workspace/.venv/",
                install_time=datetime(2024, 1, 1, 0, 0, 0),
                is_from_template=False,
            )
        ],
        dependency_install_status="completed",
        dependency_install_started_at=datetime(2024, 1, 1, 0, 0, 0),
        dependency_install_completed_at=datetime(2024, 1, 1, 0, 1, 0),
    )


def make_template(template_id="tpl-1"):
    return Template(
        id=template_id,
        name="Python",
        image="python:3.11",
        base_image="python:3.11-slim",
        pre_installed_packages=["numpy"],
        default_resources=ResourceLimit(cpu="0.5", memory="1Gi", disk="2048Mi"),
        default_timeout_sec=120,
        security_context={"network": "none"},
    )


def make_execution(execution_id="exec-1", status=ExecutionStatus.PENDING):
    return Execution(
        id=execution_id,
        session_id="sess-1",
        code="print('ok')",
        language="python",
        timeout=30,
        event_data={"name": "Ada"},
        state=ExecutionState(status=status, exit_code=0),
        stdout="out",
        stderr="",
        return_value={"ok": True},
        metrics={"time": 1},
    )


@dataclass
class PersistableRuntimeNode:
    node_id: str = "node-1"
    hostname: str = "host-1"
    type: str = "docker"
    ip_address: str = "10.0.0.1"
    url: str = "tcp://10.0.0.1:2375"
    total_cpu_cores: float = 8.0
    total_memory_mb: int = 16384
    max_sessions: int = 20
    cached_templates: list[str] = None

    def __post_init__(self):
        if self.cached_templates is None:
            self.cached_templates = ["tpl-1"]


class TestSqlSessionRepository:
    @pytest.mark.asyncio
    async def test_save_creates_new_session_model(self):
        db = FakeAsyncSession()
        repo = SqlSessionRepository(db)

        await repo.save(make_session())

        assert isinstance(db.added[0], SessionModel)
        assert db.added[0].f_id == "sess-1"
        assert db.flushed == 1

    @pytest.mark.asyncio
    async def test_save_updates_existing_session_model_with_dependency_fields(self):
        model = SessionModel.from_entity(make_session(status=SessionStatus.CREATING))
        db = FakeAsyncSession({(SessionModel, "sess-1"): model})
        repo = SqlSessionRepository(db)

        await repo.save(make_session(status=SessionStatus.RUNNING))

        assert model.f_status == "running"
        assert model.f_requested_dependencies == '["requests==2.32.0"]'
        assert "requests" in model.f_installed_dependencies
        assert model.f_dependency_install_status == "completed"
        assert db.flushed == 1

    @pytest.mark.asyncio
    async def test_find_methods_convert_models_to_entities(self):
        model = SessionModel.from_entity(make_session())
        db = FakeAsyncSession(
            {(SessionModel, "sess-1"): model},
            [FakeResult([model]), FakeResult([model]), FakeResult([model]), FakeResult([model])],
        )
        repo = SqlSessionRepository(db)

        by_id = await repo.find_by_id("sess-1")
        by_container = await repo.find_by_container_id("container-1")
        by_status = await repo.find_by_status("running")
        by_template = await repo.find_by_template("tpl-1")

        assert by_id.id == "sess-1"
        assert by_container.id == "sess-1"
        assert [item.id for item in by_status] == ["sess-1"]
        assert [item.id for item in by_template] == ["sess-1"]

    @pytest.mark.asyncio
    async def test_find_idle_expired_paginated_and_counts(self):
        model = SessionModel.from_entity(make_session())
        db = FakeAsyncSession(
            execute_results=[
                FakeResult([model]),
                FakeResult([model]),
                FakeResult([model]),
                FakeResult(scalar_value=3),
                FakeResult(scalar_value=2),
                FakeResult(scalar_value=5),
            ]
        )
        repo = SqlSessionRepository(db)

        idle = await repo.find_idle_sessions(datetime.now() - timedelta(hours=1))
        expired = await repo.find_expired_sessions(datetime.now() - timedelta(hours=6))
        sessions = await repo.find_sessions(status="running", template_id="tpl-1", limit=999, offset=-1)
        running_count = await repo.count_by_status("running")
        node_count = await repo.count_by_node("node-1")
        total = await repo.count_sessions(status="running", template_id="tpl-1")

        assert [item.id for item in idle] == ["sess-1"]
        assert [item.id for item in expired] == ["sess-1"]
        assert [item.id for item in sessions] == ["sess-1"]
        assert (running_count, node_count, total) == (3, 2, 5)

    @pytest.mark.asyncio
    async def test_delete_cascades_to_execution_repo_and_flushes(self):
        execution_repo = type("ExecutionRepo", (), {"delete_by_session_id": None})()

        async def delete_by_session_id(session_id):
            execution_repo.deleted = session_id

        execution_repo.delete_by_session_id = delete_by_session_id
        db = FakeAsyncSession()
        repo = SqlSessionRepository(db, execution_repo=execution_repo)

        await repo.delete("sess-1")

        assert execution_repo.deleted == "sess-1"
        assert len(db.executed) == 1
        assert db.flushed == 1

    @pytest.mark.asyncio
    async def test_exists_returns_boolean(self):
        db = FakeAsyncSession({(SessionModel, "sess-1"): SessionModel.from_entity(make_session())})
        repo = SqlSessionRepository(db)

        assert await repo.exists("sess-1") is True
        assert await repo.exists("missing") is False


class TestSqlTemplateRepository:
    @pytest.mark.asyncio
    async def test_save_creates_and_updates_template_model(self):
        db = FakeAsyncSession()
        repo = SqlTemplateRepository(db)
        template = make_template()

        await repo.save(template)
        created = db.added[0]
        assert isinstance(created, TemplateModel)
        assert created.f_default_memory_mb == 1024

        existing = TemplateModel.from_entity(template)
        update_db = FakeAsyncSession({(TemplateModel, "tpl-1"): existing})
        await SqlTemplateRepository(update_db).save(template)
        assert existing.f_pre_installed_packages == '["numpy"]'
        assert existing.f_security_context == '{"network": "none"}'
        assert update_db.flushed == 1

    @pytest.mark.asyncio
    async def test_queries_delete_exists_and_count(self):
        model = TemplateModel.from_entity(make_template())
        db = FakeAsyncSession(
            {(TemplateModel, "tpl-1"): model},
            [
                FakeResult([model]),
                FakeResult([model]),
                FakeResult(),
                FakeResult(scalar_value=1),
                FakeResult(scalar_value=7),
            ],
        )
        repo = SqlTemplateRepository(db)

        assert (await repo.find_by_id("tpl-1")).id == "tpl-1"
        assert (await repo.find_by_name("Python")).id == "tpl-1"
        assert [item.id for item in await repo.find_all(offset=1, limit=5)] == ["tpl-1"]
        await repo.delete("tpl-1")
        assert await repo.exists("tpl-1") is True
        assert await repo.exists_by_name("Python") is True
        assert await repo.count() == 7
        assert db.flushed == 1


class TestSqlExecutionRepository:
    @pytest.mark.asyncio
    async def test_save_create_update_and_commit(self):
        execution = make_execution()
        db = FakeAsyncSession()
        repo = SqlExecutionRepository(db)

        await repo.save(execution)
        await repo.commit()

        assert isinstance(db.added[0], ExecutionModel)
        assert db.committed == 1

        existing = ExecutionModel.from_entity(make_execution(status=ExecutionStatus.RUNNING))
        update_db = FakeAsyncSession({(ExecutionModel, "exec-1"): existing})
        await SqlExecutionRepository(update_db).save(make_execution(status=ExecutionStatus.COMPLETED))
        assert existing.f_status == "completed"
        assert existing.f_return_value == '{"ok": true}'
        assert existing.f_metrics == '{"time": 1}'

    @pytest.mark.asyncio
    async def test_queries_deletes_counts_and_empty_retry_helpers(self):
        model = ExecutionModel.from_entity(make_execution())
        db = FakeAsyncSession(
            execute_results=[
                FakeResult([model]),
                FakeResult([model]),
                FakeResult([model]),
                FakeResult(),
                FakeResult(),
                FakeResult(scalar_value=4),
            ]
        )
        repo = SqlExecutionRepository(db)

        assert (await repo.find_by_id("exec-1")).id == "exec-1"
        assert [item.id for item in await repo.find_by_session_id("sess-1")] == ["exec-1"]
        assert [item.id for item in await repo.find_by_status("pending")] == ["exec-1"]
        assert await repo.find_crashed_executions(max_retry_count=3) == []
        assert await repo.find_heartbeat_timeouts(datetime.now()) == []
        await repo.delete("exec-1")
        await repo.delete_by_session_id("sess-1")
        assert await repo.count_by_status("pending") == 4
        assert db.flushed == 2


class TestSqlRuntimeNodeRepository:
    @pytest.mark.asyncio
    async def test_save_creates_and_updates_runtime_node_model(self):
        node = PersistableRuntimeNode()
        db = FakeAsyncSession()
        repo = SqlRuntimeNodeRepository(db)

        await repo.save(node)
        created = db.added[0]
        assert isinstance(created, RuntimeNodeModel)
        assert created.f_cached_images == '["tpl-1"]'

        existing = RuntimeNodeModel(
            f_node_id="node-1",
            f_hostname="old",
            f_runtime_type="docker",
            f_ip_address="127.0.0.1",
            f_api_endpoint="",
            f_status="offline",
            f_total_cpu_cores=Decimal("1"),
            f_total_memory_mb=512,
            f_max_containers=1,
            f_cached_images="[]",
            f_labels="{}",
            f_running_containers=0,
            f_allocated_cpu_cores=Decimal("0"),
            f_allocated_memory_mb=0,
            f_last_heartbeat_at=0,
            f_created_at=0,
            f_created_by="system",
            f_updated_at=0,
            f_updated_by="system",
            f_deleted_at=0,
            f_deleted_by="",
        )
        update_db = FakeAsyncSession({(RuntimeNodeModel, "node-1"): existing})
        await SqlRuntimeNodeRepository(update_db).save(node)
        assert existing.f_hostname == "host-1"
        assert existing.f_status == "online"
        assert existing.f_total_cpu_cores == Decimal("8.0")

    @pytest.mark.asyncio
    async def test_queries_and_update_methods(self):
        model = RuntimeNodeModel(
            f_node_id="node-1",
            f_hostname="host-1",
            f_runtime_type="docker",
            f_ip_address="10.0.0.1",
            f_api_endpoint="",
            f_status="online",
            f_total_cpu_cores=Decimal("8"),
            f_total_memory_mb=16000,
            f_allocated_cpu_cores=Decimal("4"),
            f_allocated_memory_mb=8000,
            f_running_containers=3,
            f_max_containers=10,
            f_cached_images='["tpl-1"]',
            f_labels="{}",
            f_last_heartbeat_at=0,
            f_created_at=0,
            f_created_by="system",
            f_updated_at=0,
            f_updated_by="system",
            f_deleted_at=0,
            f_deleted_by="",
        )
        db = FakeAsyncSession(
            {(RuntimeNodeModel, "node-1"): model},
            [FakeResult([model]), FakeResult([model]), FakeResult([model])],
        )
        repo = SqlRuntimeNodeRepository(db)

        assert await repo.find_by_id("node-1") is model
        assert await repo.find_by_hostname("host-1") is model
        assert await repo.find_by_status("online") == [model]
        assert await repo.find_all(offset=1, limit=10) == [model]

        await repo.update_status("node-1", "offline")
        await repo.update_heartbeat("node-1")
        await repo.allocate_resources("node-1", 0.5, 256)
        await repo.release_resources("node-1", 0.5, 256)
        await repo.increment_container_count("node-1")
        await repo.decrement_container_count("node-1")

        runtime_node = model.to_runtime_node()
        assert isinstance(runtime_node, RuntimeNode)
        assert runtime_node.cpu_usage == 0.5
        assert runtime_node.mem_usage == 0.5
        assert runtime_node.cached_templates == ["tpl-1"]
        assert db.flushed == 6
