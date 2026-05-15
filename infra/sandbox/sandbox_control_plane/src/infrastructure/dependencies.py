"""
依赖注入配置

配置和提供应用所需的所有依赖项。
"""

import asyncio
import os
import time
from functools import lru_cache

from fastapi import FastAPI, Depends
from src.infrastructure.logging import get_logger

from src.application.services.session_service import SessionService
from src.application.services.template_service import TemplateService
from src.application.services.file_service import FileService

from src.domain.repositories.session_repository import ISessionRepository
from src.domain.repositories.execution_repository import IExecutionRepository
from src.domain.repositories.template_repository import ITemplateRepository
from src.domain.services.scheduler import IScheduler, RuntimeNode
from src.domain.services.storage import IStorageService
from src.domain.value_objects.execution_request import ExecutionRequest
from src.domain.value_objects.execution_status import SessionStatus

from src.infrastructure.persistence.database import db_manager
from src.infrastructure.executors import ExecutorClient
from src.infrastructure.config.settings import get_settings

# Configuration flag to switch between Mock and SQL repositories
USE_SQL_REPOSITORIES = True  # Set to False to use Mock repositories

# Auto-detect runtime environment
# In Kubernetes, KUBERNETES_SERVICE_HOST environment variable is automatically set
IS_IN_KUBERNETES = os.getenv("KUBERNETES_SERVICE_HOST") is not None

# Configuration flag to switch between schedulers
# - In Kubernetes: auto-detect and use K8s scheduler
# - In local development: use Docker scheduler
# - Set to False to use Mock scheduler
USE_MOCK_SCHEDULER = False  # Set to True to use Mock scheduler

logger = get_logger(__name__)
logger.info(f"Runtime environment: {'Kubernetes' if IS_IN_KUBERNETES else 'Local Docker'}")


def _get_docker_url() -> str:
    """获取 Docker socket URL"""
    settings = get_settings()
    # 确保 docker_host 有正确的协议前缀
    docker_host = settings.docker_host
    if not docker_host.startswith("unix://") and not docker_host.startswith("tcp://"):
        docker_host = f"unix://{docker_host}"
    logger.info(f"Using Docker URL: {docker_host}")
    return docker_host


# Mock implementations for development
class MockSessionRepository(ISessionRepository):
    """Mock 会话仓储（用于开发测试）"""

    def __init__(self):
        self._sessions = {}

    async def save(self, session):
        self._sessions[session.id] = session

    async def find_by_id(self, session_id: str):
        return self._sessions.get(session_id)

    async def find_by_container_id(self, container_id: str):
        for session in self._sessions.values():
            if getattr(session, "container_id", None) == container_id:
                return session
        return None

    async def find_by_status(self, status: str, limit: int = 100):
        return [s for s in self._sessions.values() if s.status == status][:limit]

    async def find_by_template(self, template_id: str):
        return [s for s in self._sessions.values() if s.template_id == template_id]

    async def find_idle_sessions(self, threshold):
        return []

    async def find_expired_sessions(self, threshold):
        return []

    async def delete(self, session_id: str) -> None:
        if session_id in self._sessions:
            del self._sessions[session_id]

    async def exists(self, session_id: str) -> bool:
        return session_id in self._sessions

    async def count_by_status(self, status: str) -> int:
        return sum(1 for s in self._sessions.values() if s.status == status)

    async def count_by_node(self, runtime_node: str) -> int:
        return sum(
            1 for s in self._sessions.values() if getattr(s, "node_id", None) == runtime_node
        )

    async def find_sessions(
        self,
        status: str | None = None,
        template_id: str | None = None,
        limit: int = 50,
        offset: int = 0,
    ):
        sessions = list(self._sessions.values())
        if status is not None:
            sessions = [
                s
                for s in sessions
                if getattr(getattr(s, "status", None), "value", getattr(s, "status", None))
                == status
            ]
        if template_id is not None:
            sessions = [s for s in sessions if s.template_id == template_id]
        return sessions[offset : offset + limit]

    async def count_sessions(
        self,
        status: str | None = None,
        template_id: str | None = None,
    ) -> int:
        sessions = await self.find_sessions(
            status=status,
            template_id=template_id,
            limit=len(self._sessions),
            offset=0,
        )
        return len(sessions)


class MockExecutionRepository(IExecutionRepository):
    """Mock 执行仓储（用于开发测试）"""

    def __init__(self):
        self._executions = {}

    async def save(self, execution):
        self._executions[execution.id] = execution

    async def commit(self):
        """Mock commit - no-op"""
        pass

    async def find_by_id(self, execution_id: str):
        return self._executions.get(execution_id)

    async def find_by_session_id(self, session_id: str, limit: int = 100):
        return [e for e in self._executions.values() if e.session_id == session_id][:limit]

    async def find_by_status(self, status: str, limit: int = 100):
        return [e for e in self._executions.values() if e.status == status][:limit]

    async def find_crashed_executions(self, max_retry_count: int):
        return []

    async def find_heartbeat_timeouts(self, timeout_threshold):
        return []

    async def delete(self, execution_id: str) -> None:
        if execution_id in self._executions:
            del self._executions[execution_id]

    async def delete_by_session_id(self, session_id: str) -> None:
        to_delete = [eid for eid, e in self._executions.items() if e.session_id == session_id]
        for eid in to_delete:
            del self._executions[eid]

    async def count_by_status(self, status: str) -> int:
        return sum(1 for e in self._executions.values() if e.status == status)


class MockTemplateRepository(ITemplateRepository):
    """Mock 模板仓储（用于开发测试）"""

    def __init__(self):
        from datetime import datetime
        from src.domain.entities.template import Template
        from src.domain.value_objects.resource_limit import ResourceLimit

        # 默认模板
        self._templates = {
            "python-basic": Template(
                id="python-basic",
                name="Python Basic",
                image="python:3.11-slim",
                base_image="python:3.11-slim",
                pre_installed_packages=[],
                default_resources=ResourceLimit.default(),
                created_at=datetime.now(),
                updated_at=datetime.now(),
            ),
            "string": Template(
                id="string",
                name="Test Template",
                image="python:3.11-slim",
                base_image="python:3.11-slim",
                pre_installed_packages=[],
                default_resources=ResourceLimit.default(),
                created_at=datetime.now(),
                updated_at=datetime.now(),
            ),
        }

    async def save(self, template):
        self._templates[template.id] = template

    async def find_by_id(self, template_id: str):
        return self._templates.get(template_id)

    async def find_by_name(self, name: str):
        for t in self._templates.values():
            if t.name == name:
                return t
        return None

    async def find_all(self, offset: int = 0, limit: int = 100):
        return list(self._templates.values())[offset : offset + limit]

    async def delete(self, template_id: str) -> None:
        if template_id in self._templates:
            del self._templates[template_id]

    async def exists(self, template_id: str) -> bool:
        return template_id in self._templates

    async def exists_by_name(self, name: str) -> bool:
        return any(t.name == name for t in self._templates.values())

    async def count(self) -> int:
        return len(self._templates)


class MockScheduler(IScheduler):
    """Mock 调度器（用于开发测试）"""

    async def schedule(self, request):
        # 返回一个 mock 运行时节点
        return RuntimeNode(
            id="node-1",
            type="docker",
            url="http://localhost:2375",
            status="healthy",
            cpu_usage=0.3,
            mem_usage=0.4,
            session_count=5,
            max_sessions=100,
            cached_templates=["python-basic", "string"],
        )

    async def get_node(self, node_id: str):
        return None

    async def get_healthy_nodes(self):
        return []

    async def mark_node_unhealthy(self, node_id: str) -> None:
        pass

    async def execute(
        self,
        session_id: str,
        container_id: str,
        execution_request: ExecutionRequest,
    ) -> str:
        """Mock 执行方法"""
        return execution_request.execution_id or "mock_execution_id"

    async def get_executor_url(self, container_id: str) -> str:
        return "http://localhost:8080"


class MockRuntimeNodeRepository:
    """Mock 运行时节点仓储（用于开发测试）"""

    class _NodeModel:
        def __init__(self, id, type, url, status, **kwargs):
            self.id = id
            self.type = type
            self.url = url
            self.status = status
            self.cpu_usage = kwargs.get("cpu_usage", 0.3)
            self.mem_usage = kwargs.get("mem_usage", 0.4)
            self.session_count = kwargs.get("session_count", 0)
            self.max_sessions = kwargs.get("max_sessions", 100)
            self.cached_templates = kwargs.get("cached_templates", [])

        def to_runtime_node(self):
            from src.domain.services.scheduler import RuntimeNode

            return RuntimeNode(
                id=self.id,
                type=self.type,
                url=self.url,
                status=self.status,
                cpu_usage=self.cpu_usage,
                mem_usage=self.mem_usage,
                session_count=self.session_count,
                max_sessions=self.max_sessions,
                cached_templates=self.cached_templates,
            )

    def __init__(self):
        # 默认有一个 Docker 节点
        self._nodes = {
            "docker-local": self._NodeModel(
                id="docker-local",
                type="docker",
                url="unix:///Users/guochenguang/.docker/run/docker.sock",
                status="online",
                cpu_usage=0.3,
                mem_usage=0.4,
                session_count=0,
                max_sessions=100,
                cached_templates=[],
            )
        }

    async def find_by_id(self, node_id: str):
        return self._nodes.get(node_id)

    async def find_by_status(self, status: str):
        return [n for n in self._nodes.values() if n.status == status]

    async def save(self, node):
        self._nodes[node.id] = node

    async def update_status(self, node_id: str, status: str) -> None:
        if node_id in self._nodes:
            self._nodes[node_id].status = status

    async def increment_session_count(self, node_id: str) -> None:
        if node_id in self._nodes:
            self._nodes[node_id].session_count += 1

    async def decrement_session_count(self, node_id: str) -> None:
        if node_id in self._nodes:
            self._nodes[node_id].session_count = max(0, self._nodes[node_id].session_count - 1)

    async def add_cached_template(self, node_id: str, template_id: str) -> None:
        if node_id in self._nodes:
            if template_id not in self._nodes[node_id].cached_templates:
                self._nodes[node_id].cached_templates.append(template_id)


class MockStorageService(IStorageService):
    """Mock 存储服务（用于开发测试）"""

    async def upload_file(
        self, s3_path: str, content: bytes, content_type: str = "application/octet-stream"
    ):
        pass

    async def download_file(self, s3_path: str) -> bytes:
        return b""

    async def file_exists(self, s3_path: str) -> bool:
        return False

    async def get_file_info(self, s3_path: str):
        return {"size": 0, "content_type": "application/octet-stream"}

    async def generate_presigned_url(self, s3_path: str, expiration_seconds: int = 3600) -> str:
        return f"http://localhost:9000/{s3_path}?presigned=true"

    async def delete_file(self, s3_path: str) -> None:
        pass

    async def list_files(self, prefix: str, limit: int = 1000):
        return []

    async def delete_prefix(self, prefix: str) -> int:
        return 0


# Module-level singletons for shared components
_container_scheduler_singleton = None
_scheduler_singleton = None


def initialize_dependencies(app: FastAPI):
    """初始化所有依赖项并存储到应用状态中"""

    # 注意：数据库管理器的初始化现在是异步的，需要在 lifespan 中调用
    # 这里不再同步调用 initialize()
    settings = get_settings()

    # 创建仓储实例（根据配置选择 Mock 或 SQL）
    if USE_SQL_REPOSITORIES:
        # SQL 模式：仓储在请求时通过 Depends() 注入
        # 存储仓储工厂函数引用到 app.state
        app.state.get_session_repository = get_session_repository
        app.state.get_execution_repository = get_execution_repository
        app.state.get_template_repository = get_template_repository

        # SessionService 也需要动态创建
        app.state.get_session_service = get_session_service_db
        app.state.get_template_service = get_template_service_db
        app.state.get_file_service = get_file_service_db

        # 对于向后兼容，也设置仓储实例（用于可能直接访问的情况）
        app.state.session_repo = None  # 使用工厂函数
        app.state.execution_repo = None
        app.state.template_repo = None

        # 服务也使用工厂函数
        app.state.session_service = None
        app.state.template_service = None
        app.state.file_service = None
    else:
        # Mock 模式：直接创建实例
        session_repo = MockSessionRepository()
        execution_repo = MockExecutionRepository()
        template_repo = MockTemplateRepository()

        # 创建领域服务实例
        storage_service = MockStorageService()

        # 创建 Mock 调度器
        scheduler = MockScheduler()

        # 创建应用服务实例
        session_service = SessionService(
            session_repo=session_repo,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=scheduler,
        )

        template_service = TemplateService(
            template_repo=template_repo,
        )

        file_service = FileService(
            session_repo=session_repo,
            storage_service=storage_service,
            max_extracted_file_count=settings.max_extracted_file_count,
            max_extracted_total_size_mb=settings.max_extracted_total_size_mb,
        )

        # 存储到应用状态
        app.state.session_service = session_service
        app.state.template_service = template_service
        app.state.file_service = file_service

        # 也存储仓储（可能需要）
        app.state.session_repo = session_repo
        app.state.execution_repo = execution_repo
        app.state.template_repo = template_repo

    # 创建容器调度器（模块级单例）
    global _container_scheduler_singleton, _scheduler_singleton

    if USE_MOCK_SCHEDULER:
        # Mock 模式：不创建真实调度器
        _container_scheduler_singleton = None
        _scheduler_singleton = MockScheduler()
    elif IS_IN_KUBERNETES:
        # Kubernetes 环境：使用 K8s 调度器
        from src.infrastructure.container_scheduler.k8s_scheduler import K8sScheduler

        settings = get_settings()

        _container_scheduler_singleton = K8sScheduler(
            namespace=settings.kubernetes_namespace,
        )
        logger.info(f"Initialized K8s scheduler with namespace: {settings.kubernetes_namespace}")

        # 调度器服务延迟初始化
        _scheduler_singleton = None
    else:
        # 本地开发环境：使用 Docker 调度器
        from src.infrastructure.container_scheduler.docker_scheduler import DockerScheduler

        _container_scheduler_singleton = DockerScheduler(docker_url=_get_docker_url())
        logger.info(f"Initialized Docker scheduler with URL: {_get_docker_url()}")

        # 调度器服务延迟初始化
        _scheduler_singleton = None


async def cleanup_dependencies(app: FastAPI):
    """清理依赖项"""
    await db_manager.close()


def get_session_service(app: FastAPI) -> SessionService:
    """获取会话服务"""
    return app.state.session_service


def get_template_service(app: FastAPI) -> TemplateService:
    """获取模板服务"""
    return app.state.template_service


def get_file_service(app: FastAPI) -> FileService:
    """获取文件服务"""
    return app.state.file_service


# ============================================================================
# Database-based dependency injection (request-scoped)
# ============================================================================


async def get_db_session():
    """获取数据库会话（FastAPI 依赖）"""
    async with db_manager.get_session() as session:
        yield session


def get_execution_repository(session=Depends(get_db_session)) -> IExecutionRepository:
    """获取执行仓储（SQL 或 Mock）"""
    if USE_SQL_REPOSITORIES:
        from src.infrastructure.persistence.repositories.sql_execution_repository import (
            SqlExecutionRepository,
        )

        return SqlExecutionRepository(session)
    return MockExecutionRepository()


def get_session_repository(
    session=Depends(get_db_session),
    execution_repo: IExecutionRepository = Depends(get_execution_repository),
) -> ISessionRepository:
    """获取会话仓储（SQL 或 Mock）"""
    if USE_SQL_REPOSITORIES:
        from src.infrastructure.persistence.repositories.sql_session_repository import (
            SqlSessionRepository,
        )

        return SqlSessionRepository(session, execution_repo)
    return MockSessionRepository()


def get_template_repository(session=Depends(get_db_session)) -> ITemplateRepository:
    """获取模板仓储（SQL 或 Mock）"""
    if USE_SQL_REPOSITORIES:
        from src.infrastructure.persistence.repositories.sql_template_repository import (
            SqlTemplateRepository,
        )

        return SqlTemplateRepository(session)
    return MockTemplateRepository()


def get_scheduler() -> IScheduler:
    """获取调度器（Mock、Docker 或 K8s）"""
    if USE_MOCK_SCHEDULER:
        return MockScheduler()

    # 实际使用时需要通过 session-scoped 依赖获取
    return MockScheduler()


def get_runtime_node_repository(session=Depends(get_db_session)):
    """获取运行时节点仓储（SQL 或 Mock）"""
    if USE_SQL_REPOSITORIES:
        from src.infrastructure.persistence.repositories.sql_runtime_node_repository import (
            SqlRuntimeNodeRepository,
        )

        return SqlRuntimeNodeRepository(session)
    return MockRuntimeNodeRepository()


def get_container_scheduler():
    """获取容器调度器"""
    from src.infrastructure.container_scheduler.docker_scheduler import DockerScheduler

    return DockerScheduler(docker_url=_get_docker_url())


def get_docker_scheduler_service(
    runtime_node_repo=Depends(get_runtime_node_repository),
    template_repo=Depends(get_template_repository),
) -> IScheduler:
    """获取调度服务（Docker 或 K8s）"""
    return _create_scheduler_service(
        runtime_node_repo=runtime_node_repo,
        template_repo=template_repo,
    )


def _create_scheduler_service(
    runtime_node_repo,
    template_repo,
    executor_timeout: float = 30.0,
) -> IScheduler:
    """创建调度服务实例。"""
    if USE_MOCK_SCHEDULER:
        return MockScheduler()

    # 使用模块级单例
    container_scheduler = _container_scheduler_singleton

    # 创建 ExecutorClient 实例
    executor_client = ExecutorClient(
        timeout=executor_timeout,
        max_retries=3,
        retry_delay=0.5,
    )

    # 为每个请求创建新的调度服务实例
    settings = get_settings()

    if IS_IN_KUBERNETES:
        # K8s 环境：使用 K8sSchedulerService
        from src.infrastructure.schedulers.k8s_scheduler_service import K8sSchedulerService

        # Build CONTROL_PLANE_URL based on kubernetes_namespace
        control_plane_url = (
            settings.control_plane_url
            if settings.control_plane_url is not None
            else f"http://sandbox-control-plane.{settings.kubernetes_namespace}.svc.cluster.local:8000"
        )

        return K8sSchedulerService(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
            executor_client=executor_client,
            executor_port=8080,
            control_plane_url=control_plane_url,
            disable_bwrap=settings.disable_bwrap,
        )
    else:
        # 本地环境：使用 DockerSchedulerService
        from src.infrastructure.schedulers.docker_scheduler_service import DockerSchedulerService

        return DockerSchedulerService(
            runtime_node_repo=runtime_node_repo,
            container_scheduler=container_scheduler,
            template_repo=template_repo,
            executor_client=executor_client,
            executor_port=8080,
            control_plane_url=settings.control_plane_url,
            disable_bwrap=settings.disable_bwrap,
        )


# Storage service singleton (cached at module level for use with Depends)
_storage_service_singleton = None


def get_storage_service():
    """
    获取存储服务（S3 或 Mock）

    架构说明：
    - Control Plane 通过 S3 API 将文件写入 MinIO 的 /sessions/{session_id}/ 路径
    - Executor Pod 在启动脚本中挂载 s3fs，将 S3 bucket 的 session 子目录挂载到 /workspace
    """
    global _storage_service_singleton

    if _storage_service_singleton is not None:
        return _storage_service_singleton

    settings = get_settings()

    # 直接使用 S3
    if settings.s3_access_key_id:
        from src.infrastructure.storage.s3_storage import S3Storage

        _storage_service_singleton = S3Storage()
        logger.info(f"Using S3 storage: endpoint={settings.s3_endpoint_url}")
        return _storage_service_singleton

    # 降级到 Mock
    logger.warning("No storage backend configured, using MockStorageService")
    _storage_service_singleton = MockStorageService()
    return _storage_service_singleton


def get_executor_client() -> ExecutorClient:
    """获取 ExecutorClient。"""
    return ExecutorClient(
        timeout=30.0,
        max_retries=3,
        retry_delay=0.5,
    )


def get_session_service_db(
    session_repo: ISessionRepository = Depends(get_session_repository),
    execution_repo: IExecutionRepository = Depends(get_execution_repository),
    template_repo: ITemplateRepository = Depends(get_template_repository),
    scheduler: IScheduler = Depends(get_docker_scheduler_service),
    storage_service=Depends(get_storage_service),
    executor_client: ExecutorClient = Depends(get_executor_client),
) -> SessionService:
    """获取会话服务（使用数据库仓储和 Docker 调度器）"""
    return SessionService(
        session_repo=session_repo,
        execution_repo=execution_repo,
        template_repo=template_repo,
        scheduler=scheduler,
        storage_service=storage_service,
        executor_client=executor_client,
        initial_dependency_sync_scheduler=get_initial_dependency_sync_scheduler(),
    )


def get_initial_dependency_sync_scheduler():
    """获取首次依赖同步后台调度器。"""

    def schedule(session_id: str, install_timeout: int) -> None:
        async def _run() -> None:
            try:
                await _run_initial_dependency_sync(session_id, install_timeout)
            except Exception as exc:
                logger.exception(
                    "Initial dependency sync task failed unexpectedly",
                    session_id=session_id,
                )
                await _mark_initial_dependency_sync_failed(
                    session_id=session_id,
                    error=f"Initial dependency sync failed unexpectedly: {exc}",
                )

        asyncio.create_task(_run())

    return schedule


async def _run_initial_dependency_sync(session_id: str, install_timeout: int) -> None:
    """执行首次依赖同步后台任务。"""
    deadline = time.monotonic() + install_timeout
    poll_interval = 1.0
    last_status = "unknown"
    session_seen = False

    while time.monotonic() < deadline:
        async with db_manager.get_session() as session:
            from src.infrastructure.persistence.repositories.sql_execution_repository import (
                SqlExecutionRepository,
            )
            from src.infrastructure.persistence.repositories.sql_runtime_node_repository import (
                SqlRuntimeNodeRepository,
            )
            from src.infrastructure.persistence.repositories.sql_session_repository import (
                SqlSessionRepository,
            )
            from src.infrastructure.persistence.repositories.sql_template_repository import (
                SqlTemplateRepository,
            )

            execution_repo = SqlExecutionRepository(session)
            session_repo = SqlSessionRepository(session, execution_repo)
            current_session = await session_repo.find_by_id(session_id)

            if current_session is None:
                if session_seen:
                    logger.warning(
                        "Initial dependency sync skipped because session was deleted",
                        session_id=session_id,
                    )
                    return
            else:
                session_seen = True

                last_status = current_session.status.value

                if current_session.status in {
                    SessionStatus.FAILED,
                    SessionStatus.TERMINATED,
                    SessionStatus.COMPLETED,
                    SessionStatus.TIMEOUT,
                }:
                    current_session.mark_dependency_install_failed(
                        f"Session became {current_session.status.value} before initial dependency sync"
                    )
                    await session_repo.save(current_session)
                    return

                if (
                    current_session.status == SessionStatus.RUNNING
                    and current_session.container_id
                    and current_session.requested_dependencies
                ):
                    template_repo = SqlTemplateRepository(session)
                    runtime_node_repo = SqlRuntimeNodeRepository(session)
                    scheduler = _create_scheduler_service(
                        runtime_node_repo=runtime_node_repo,
                        template_repo=template_repo,
                        executor_timeout=float(install_timeout),
                    )
                    service = SessionService(
                        session_repo=session_repo,
                        execution_repo=execution_repo,
                        template_repo=template_repo,
                        scheduler=scheduler,
                        storage_service=get_storage_service(),
                        executor_client=ExecutorClient(timeout=float(install_timeout)),
                    )
                    await service.sync_session_dependencies_for_session(
                        session_id=session_id,
                        sync_mode="replace",
                    )
                    return

        await asyncio.sleep(poll_interval)

    async with db_manager.get_session() as session:
        from src.infrastructure.persistence.repositories.sql_execution_repository import (
            SqlExecutionRepository,
        )
        from src.infrastructure.persistence.repositories.sql_session_repository import (
            SqlSessionRepository,
        )

        execution_repo = SqlExecutionRepository(session)
        session_repo = SqlSessionRepository(session, execution_repo)
        current_session = await session_repo.find_by_id(session_id)
        if current_session is None:
            return

        current_session.mark_dependency_install_failed(
            "Timed out waiting for session to become ready for initial dependency sync"
        )
        await session_repo.save(current_session)

    logger.error(
        "Initial dependency sync timed out waiting for running session",
        session_id=session_id,
        install_timeout=install_timeout,
        last_status=last_status,
    )


async def _mark_initial_dependency_sync_failed(session_id: str, error: str) -> None:
    """兜底回写首次依赖同步失败状态。"""
    async with db_manager.get_session() as session:
        from src.infrastructure.persistence.repositories.sql_execution_repository import (
            SqlExecutionRepository,
        )
        from src.infrastructure.persistence.repositories.sql_session_repository import (
            SqlSessionRepository,
        )

        execution_repo = SqlExecutionRepository(session)
        session_repo = SqlSessionRepository(session, execution_repo)
        current_session = await session_repo.find_by_id(session_id)
        if current_session is None:
            logger.warning(
                "Skipping initial dependency sync failure persistence because session was not found",
                session_id=session_id,
            )
            return

        current_session.mark_dependency_install_failed(error)
        await session_repo.save(current_session)


def get_template_service_db(
    template_repo: ITemplateRepository = Depends(get_template_repository),
) -> TemplateService:
    """获取模板服务（使用数据库仓储）"""
    return TemplateService(template_repo=template_repo)


def get_file_service_db(
    session_repo: ISessionRepository = Depends(get_session_repository),
    storage_service=Depends(get_storage_service),
) -> FileService:
    """获取文件服务（使用数据库仓储）"""
    settings = get_settings()
    return FileService(
        session_repo=session_repo,
        storage_service=storage_service,
        max_extracted_file_count=settings.max_extracted_file_count,
        max_extracted_total_size_mb=settings.max_extracted_total_size_mb,
    )


# ============================================================================
# State Sync Service (shared singleton)
# ============================================================================

_state_sync_service_singleton = None


def _create_direct_session_repository(db_mgr):
    """
    创建直接使用数据库的会话仓储

    用于状态同步服务，避免仓储层开销。
    """
    from src.infrastructure.persistence.models.session_model import SessionModel
    from src.domain.entities.session import Session, SessionStatus
    from src.domain.value_objects.resource_limit import ResourceLimit
    from sqlalchemy import select

    class DirectSessionRepository:
        """直接使用数据库的仓储，用于状态同步"""

        def __init__(self, db_mgr):
            self._db_mgr = db_mgr

        async def find_by_status(self, status: str, limit: int = 100):
            """直接查询数据库"""
            result = []
            async with self._db_mgr.get_session() as session:
                stmt = select(SessionModel).filter(SessionModel.f_status == status).limit(limit)
                models_result = await session.execute(stmt)
                for model in models_result.scalars():
                    session_entity = Session(
                        id=model.f_id,
                        template_id=model.f_template_id,
                        status=SessionStatus(model.f_status),
                        resource_limit=ResourceLimit(
                            cpu=model.f_resources_cpu,
                            memory=model.f_resources_memory,
                            disk=model.f_resources_disk,
                            max_processes=128,
                        ),
                        workspace_path=model.f_workspace_path,
                        runtime_type=model.f_runtime_type,
                        runtime_node=model.f_runtime_node or None,
                        container_id=model.f_container_id or None,
                        pod_name=model.f_pod_name or None,
                        env_vars=model._parse_json(model.f_env_vars) or {},
                        timeout=model.f_timeout,
                        created_at=model._millis_to_datetime(model.f_created_at) or datetime.now(),
                        updated_at=model._millis_to_datetime(model.f_updated_at) or datetime.now(),
                        last_activity_at=model._millis_to_datetime(model.f_last_activity_at)
                        or datetime.now(),
                    )
                    result.append(session_entity)
            return result

        async def find_by_id(self, session_id: str):
            """通过 ID 查找"""
            async with self._db_mgr.get_session() as session:
                model = await session.get(SessionModel, session_id)
                if model:
                    return Session(
                        id=model.f_id,
                        template_id=model.f_template_id,
                        status=SessionStatus(model.f_status),
                        resource_limit=ResourceLimit(
                            cpu=model.f_resources_cpu,
                            memory=model.f_resources_memory,
                            disk=model.f_resources_disk,
                            max_processes=128,
                        ),
                        workspace_path=model.f_workspace_path,
                        runtime_type=model.f_runtime_type,
                        runtime_node=model.f_runtime_node or None,
                        container_id=model.f_container_id or None,
                        pod_name=model.f_pod_name or None,
                        env_vars=model._parse_json(model.f_env_vars) or {},
                        timeout=model.f_timeout,
                        created_at=model._millis_to_datetime(model.f_created_at) or datetime.now(),
                        updated_at=model._millis_to_datetime(model.f_updated_at) or datetime.now(),
                        last_activity_at=model._millis_to_datetime(model.f_last_activity_at)
                        or datetime.now(),
                    )
                return None

        async def save(self, session):
            """保存 session"""
            import time

            async with self._db_mgr.get_session() as db:
                model = await db.get(SessionModel, session.id)
                if model:
                    # 处理 status 可能是枚举或字符串的情况
                    if hasattr(session.status, "value"):
                        model.f_status = session.status.value
                    else:
                        model.f_status = session.status
                    model.f_container_id = session.container_id or ""
                    model.f_runtime_node = session.runtime_node or ""
                    model.f_updated_at = int(time.time() * 1000)
                    await db.commit()

        async def find_sessions(
            self,
            status: str | None = None,
            template_id: str | None = None,
            limit: int = 50,
            offset: int = 0,
        ):
            async with self._db_mgr.get_session() as session:
                stmt = select(SessionModel)
                if status is not None:
                    stmt = stmt.where(SessionModel.f_status == status)
                if template_id is not None:
                    stmt = stmt.where(SessionModel.f_template_id == template_id)
                stmt = stmt.order_by(SessionModel.f_created_at.desc()).limit(limit).offset(offset)
                models_result = await session.execute(stmt)
                return [model.to_entity() for model in models_result.scalars().all()]

        async def count_sessions(
            self,
            status: str | None = None,
            template_id: str | None = None,
        ) -> int:
            async with self._db_mgr.get_session() as session:
                stmt = select(func.count()).select_from(SessionModel)
                if status is not None:
                    stmt = stmt.where(SessionModel.f_status == status)
                if template_id is not None:
                    stmt = stmt.where(SessionModel.f_template_id == template_id)
                result = await session.execute(stmt)
                return result.scalar() or 0

    return DirectSessionRepository(db_mgr)


def _create_direct_execution_repository(db_mgr):
    """
    创建直接使用数据库的执行仓储。

    用于状态同步服务在 takeover 场景回写中断 execution 状态。
    """
    from src.infrastructure.persistence.models.execution_model import ExecutionModel
    from src.domain.value_objects.execution_status import ExecutionStatus
    from sqlalchemy import select

    class DirectExecutionRepository:
        def __init__(self, db_mgr):
            self._db_mgr = db_mgr

        async def find_by_session_id(self, session_id: str, limit: int = 100):
            async with self._db_mgr.get_session() as session:
                stmt = (
                    select(ExecutionModel)
                    .where(ExecutionModel.f_session_id == session_id)
                    .order_by(ExecutionModel.f_created_at.desc())
                    .limit(limit)
                )
                result = await session.execute(stmt)
                return [model.to_entity() for model in result.scalars().all()]

        async def save(self, execution):
            import time

            async with self._db_mgr.get_session() as session:
                model = await session.get(ExecutionModel, execution.id)
                if model is None:
                    return

                model.f_status = execution.state.status.value
                model.f_stdout = execution.stdout
                model.f_stderr = execution.stderr
                model.f_exit_code = execution.state.exit_code or 0
                model.f_error_message = execution.state.error_message or ""
                model.f_completed_at = (
                    int(execution.completed_at.timestamp() * 1000) if execution.completed_at else 0
                )
                model.f_updated_at = int(time.time() * 1000)
                await session.commit()

        async def commit(self):
            """Direct repository 每次 save 已提交，保留兼容空实现。"""
            return None

    return DirectExecutionRepository(db_mgr)


def _create_direct_template_repository(db_mgr):
    """
    创建直接使用数据库的模板仓储。

    用于状态同步服务恢复 session 时解析真实模板镜像。
    """
    from src.infrastructure.persistence.models.template_model import TemplateModel

    class DirectTemplateRepository:
        def __init__(self, db_mgr):
            self._db_mgr = db_mgr

        async def find_by_id(self, template_id: str):
            async with self._db_mgr.get_session() as session:
                model = await session.get(TemplateModel, template_id)
                return model.to_entity() if model else None

    return DirectTemplateRepository(db_mgr)


def _create_scheduler_for_state_sync(container_scheduler, template_repo=None):
    """为状态同步服务创建调度器"""
    settings = get_settings()

    if USE_MOCK_SCHEDULER:
        return MockScheduler()

    if IS_IN_KUBERNETES:
        from src.infrastructure.schedulers.k8s_scheduler_service import K8sSchedulerService

        # Build CONTROL_PLANE_URL based on kubernetes_namespace
        control_plane_url = (
            settings.control_plane_url
            if settings.control_plane_url is not None
            else f"http://sandbox-control-plane.{settings.kubernetes_namespace}.svc.cluster.local:8000"
        )

        return K8sSchedulerService(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
            executor_client=None,
            executor_port=8080,
            control_plane_url=control_plane_url,
            disable_bwrap=settings.disable_bwrap,
        )

    # 本地 Docker 环境
    return MockScheduler()


def get_state_sync_service():
    """
    获取状态同步服务（共享单例）

    此服务在启动时调用，需要使用已初始化的单例。
    使用 SQL 数据库直接查询，不依赖仓储模式。
    """
    global _scheduler_singleton
    from src.application.services.state_sync_service import StateSyncService

    settings = get_settings()
    container_scheduler = _container_scheduler_singleton

    # 创建会话仓储
    if USE_SQL_REPOSITORIES:
        session_repo = _create_direct_session_repository(db_manager)
        execution_repo = _create_direct_execution_repository(db_manager)
        template_repo = _create_direct_template_repository(db_manager)
    else:
        session_repo = MockSessionRepository()
        execution_repo = MockExecutionRepository()
        template_repo = MockTemplateRepository()

    # 创建或复用调度器
    scheduler = _scheduler_singleton
    if scheduler is None:
        scheduler = _create_scheduler_for_state_sync(
            container_scheduler=container_scheduler,
            template_repo=template_repo,
        )
        _scheduler_singleton = scheduler

    # Build CONTROL_PLANE_URL based on environment
    if IS_IN_KUBERNETES:
        control_plane_url = (
            settings.control_plane_url
            if settings.control_plane_url is not None
            else f"http://sandbox-control-plane.{settings.kubernetes_namespace}.svc.cluster.local:8000"
        )
    else:
        control_plane_url = settings.control_plane_url

    return StateSyncService(
        session_repo=session_repo,
        container_scheduler=container_scheduler,
        execution_repo=execution_repo,
        template_repo=template_repo,
        scheduler=scheduler,
        control_plane_url=control_plane_url,
    )
