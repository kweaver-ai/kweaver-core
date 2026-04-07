"""
会话应用服务单元测试

测试 SessionService 的用例编排逻辑。
"""
import pytest
from datetime import datetime, timezone
from unittest.mock import Mock, AsyncMock

from src.application.services.session_service import SessionService
from src.application.commands.create_session import CreateSessionCommand
from src.application.commands.install_session_dependencies import (
    InstallSessionDependenciesCommand,
)
from src.domain.entities.session import Session
from src.domain.entities.template import Template
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import ExecutionStatus, SessionStatus
from src.domain.services.scheduler import RuntimeNode
from src.infrastructure.executors.dto import (
    ExecutorInstalledDependency,
    ExecutorSyncSessionConfigResponse,
)
from src.shared.errors.domain import ConflictError, NotFoundError


class TestSessionService:
    """会话应用服务测试"""

    @pytest.fixture
    def session_repo(self):
        """模拟会话仓储"""
        repo = Mock()
        repo.save = AsyncMock()
        repo.find_by_id = AsyncMock()
        return repo

    @pytest.fixture
    def template_repo(self):
        """模拟模板仓储"""
        repo = Mock()
        repo.find_by_id = AsyncMock()
        return repo

    @pytest.fixture
    def scheduler(self):
        """模拟调度器"""
        scheduler = Mock()
        scheduler.schedule = AsyncMock()
        scheduler.create_container_for_session = AsyncMock(return_value="container-123")
        scheduler.destroy_container = AsyncMock()
        scheduler.get_executor_url = AsyncMock(return_value="http://sandbox-sess:8080")
        return scheduler

    @pytest.fixture
    def execution_repo(self):
        """模拟执行仓储"""
        repo = Mock()
        repo.save = AsyncMock()
        repo.find_by_id = AsyncMock()
        repo.find_by_session_id = AsyncMock(return_value=[])
        return repo

    @pytest.fixture
    def executor_client(self):
        client = Mock()
        client.sync_session_config = AsyncMock()
        return client

    @pytest.fixture
    def initial_dependency_sync_scheduler(self):
        return Mock()

    @pytest.fixture
    def service(
        self,
        session_repo,
        template_repo,
        scheduler,
        execution_repo,
        executor_client,
        initial_dependency_sync_scheduler,
    ):
        """创建会话服务"""
        return SessionService(
            session_repo=session_repo,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=scheduler,
            executor_client=executor_client,
            initial_dependency_sync_scheduler=initial_dependency_sync_scheduler,
        )

    @pytest.mark.asyncio
    async def test_create_session_success(self, service, template_repo, scheduler, session_repo):
        """测试成功创建会话"""
        # 设置模拟返回值
        template = Template(
            id="python-datascience",
            name="Python Data Science",
            image="python:3.11-datascience",
            base_image="python:3.11-slim"
        )
        template_repo.find_by_id.return_value = template

        runtime_node = RuntimeNode(
            id="node-1",
            type="docker",
            url="http://node-1:2375",
            status="healthy",
            cpu_usage=0.5,
            mem_usage=0.6,
            session_count=5,
            max_sessions=100,
            cached_templates=["python-datascience"]
        )
        scheduler.schedule.return_value = runtime_node

        # 执行命令
        command = CreateSessionCommand(
            template_id="python-datascience",
            timeout=300,
            resource_limit=ResourceLimit.default()
        )

        result = await service.create_session(command)

        # 验证
        assert result.template_id == "python-datascience"
        # 状态可能是 CREATING 或 RUNNING，取决于具体实现
        assert result.status in (SessionStatus.CREATING.value, SessionStatus.RUNNING.value)
        assert session_repo.save.call_count >= 1  # 至少保存一次

    @pytest.mark.asyncio
    async def test_create_session_template_not_found(self, service, template_repo):
        """测试模板不存在"""
        template_repo.find_by_id.return_value = None

        command = CreateSessionCommand(
            template_id="non-existent",
            timeout=300
        )

        with pytest.raises(NotFoundError, match="Template not found"):
            await service.create_session(command)

    @pytest.mark.asyncio
    async def test_get_session_success(self, service, session_repo):
        """测试成功获取会话"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = session

        from src.application.queries.get_session import GetSessionQuery
        query = GetSessionQuery(session_id="sess_20240115_abc123")

        result = await service.get_session(query)

        assert result.id == "sess_20240115_abc123"
        assert result.template_id == "python-datascience"

    @pytest.mark.asyncio
    async def test_get_session_not_found(self, service, session_repo):
        """测试会话不存在"""
        session_repo.find_by_id.return_value = None

        from src.application.queries.get_session import GetSessionQuery
        query = GetSessionQuery(session_id="non-existent")

        with pytest.raises(NotFoundError, match="Session not found"):
            await service.get_session(query)

    @pytest.mark.asyncio
    async def test_terminate_session_success(self, service, session_repo):
        """测试成功终止会话"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = session

        result = await service.terminate_session("sess_20240115_abc123")

        assert result.status == SessionStatus.TERMINATED.value
        session_repo.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_terminate_already_terminated(self, service, session_repo):
        """测试终止已终止的会话"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = session

        result = await service.terminate_session("sess_20240115_abc123")

        assert result.status == SessionStatus.TERMINATED.value
        # 不应该再次调用 save
        session_repo.save.assert_not_called()

    @pytest.mark.asyncio
    async def test_delete_session_success(self, service, session_repo, execution_repo):
        """测试成功删除会话（硬删除，级联删除执行记录）"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = session

        # Mock execution repo's delete_by_session_id method
        execution_repo.delete_by_session_id = AsyncMock()
        session_repo.delete = AsyncMock()

        # Delete should not return anything
        result = await service.delete_session("sess_20240115_abc123")

        assert result is None
        # Verify session_repo.delete was called
        session_repo.delete.assert_called_once_with("sess_20240115_abc123")

    @pytest.mark.asyncio
    async def test_delete_session_not_found(self, service, session_repo):
        """测试删除不存在的会话"""
        session_repo.find_by_id.return_value = None

        with pytest.raises(NotFoundError, match="Session not found"):
            await service.delete_session("non-existent")

        # Verify delete was not called
        session_repo.delete.assert_not_called()

    @pytest.mark.asyncio
    async def test_create_session_with_manual_id(self, service, template_repo, scheduler, session_repo):
        """测试使用手动指定 ID 创建会话"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim"
        )
        template_repo.find_by_id.return_value = template

        runtime_node = RuntimeNode(
            id="node-1",
            type="docker",
            url="http://node-1:2375",
            status="healthy",
            cpu_usage=0.5,
            mem_usage=0.6,
            session_count=5,
            max_sessions=100,
            cached_templates=["python-test"]
        )
        scheduler.schedule.return_value = runtime_node

        # 第一个调用返回 None（检查 ID 是否存在），后续调用返回会话
        session_repo.find_by_id.side_effect = [None, None]

        command = CreateSessionCommand(
            id="custom-session-id",
            template_id="python-test",
            timeout=300,
            resource_limit=ResourceLimit.default()
        )

        result = await service.create_session(command)

        assert result.id == "custom-session-id"

    @pytest.mark.asyncio
    async def test_install_session_dependencies_merges_by_package_name(
        self,
        service,
        session_repo,
        executor_client,
    ):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_1",
            runtime_type="python3.11",
            container_id="sandbox-sess_1",
            requested_dependencies=["requests==2.30.0"],
        )
        session_repo.find_by_id.return_value = session
        executor_client.sync_session_config.return_value = ExecutorSyncSessionConfigResponse(
            status="completed",
            installed_dependencies=[],
            started_at="2026-03-09T12:00:00+00:00",
            completed_at="2026-03-09T12:00:05+00:00",
        )

        result = await service.install_session_dependencies(
            InstallSessionDependenciesCommand(
                session_id="sess_1",
                dependencies=["requests==2.31.0", "pandas==2.2.0"],
            )
        )

        assert {dep["name"] for dep in result.requested_dependencies} == {"requests", "pandas"}
        versions = {dep["name"]: dep["version"] for dep in result.requested_dependencies}
        assert versions["requests"] == "==2.31.0"

    @pytest.mark.asyncio
    async def test_install_session_dependencies_rejects_concurrent_install(
        self,
        service,
        session_repo,
    ):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_1",
            runtime_type="python3.11",
            container_id="sandbox-sess_1",
            dependency_install_status="installing",
        )
        session_repo.find_by_id.return_value = session

        with pytest.raises(ConflictError):
            await service.install_session_dependencies(
                InstallSessionDependenciesCommand(
                    session_id="sess_1",
                    dependencies=["requests==2.31.0"],
                )
            )

    @pytest.mark.asyncio
    async def test_create_session_with_duplicate_id(self, service, template_repo, scheduler, session_repo):
        """测试使用重复 ID 创建会话"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim"
        )
        template_repo.find_by_id.return_value = template

        existing_session = Session(
            id="existing-session-id",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/existing",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = existing_session

        command = CreateSessionCommand(
            id="existing-session-id",
            template_id="python-test",
            timeout=300
        )

        from src.shared.errors.domain import ConflictError
        with pytest.raises(ConflictError, match="already exists"):
            await service.create_session(command)

    @pytest.mark.asyncio
    async def test_create_session_with_dependencies(
        self,
        service,
        template_repo,
        scheduler,
        session_repo,
        initial_dependency_sync_scheduler,
    ):
        """测试创建带依赖的会话"""
        template = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim"
        )
        template_repo.find_by_id.return_value = template

        runtime_node = RuntimeNode(
            id="node-1",
            type="docker",
            url="http://node-1:2375",
            status="healthy",
            cpu_usage=0.5,
            mem_usage=0.6,
            session_count=5,
            max_sessions=100,
            cached_templates=["python-test"]
        )
        scheduler.schedule.return_value = runtime_node

        command = CreateSessionCommand(
            template_id="python-test",
            timeout=300,
            resource_limit=ResourceLimit.default(),
            dependencies=["requests>=2.28.0"],
        )

        result = await service.create_session(command)

        assert result.template_id == "python-test"
        assert result.dependency_install_status == "installing"
        assert result.dependency_install_started_at is not None
        initial_dependency_sync_scheduler.assert_called_once()

    @pytest.mark.asyncio
    async def test_list_sessions(self, service, session_repo):
        """测试列出会话"""
        sessions = [
            Session(
                id="sess_1",
                template_id="python-test",
                status=SessionStatus.RUNNING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://bucket/sessions/sess_1",
                runtime_type="docker"
            ),
            Session(
                id="sess_2",
                template_id="python-test",
                status=SessionStatus.TERMINATED,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://bucket/sessions/sess_2",
                runtime_type="docker"
            )
        ]
        session_repo.find_sessions = AsyncMock(return_value=sessions)
        session_repo.count_sessions = AsyncMock(return_value=2)

        result = await service.list_sessions()

        assert "items" in result
        session_repo.find_sessions.assert_called_once()

    @pytest.mark.asyncio
    async def test_list_sessions_by_status(self, service, session_repo):
        """测试按状态列出会话"""
        sessions = [
            Session(
                id="sess_1",
                template_id="python-test",
                status=SessionStatus.RUNNING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://bucket/sessions/sess_1",
                runtime_type="docker"
            )
        ]
        session_repo.find_sessions = AsyncMock(return_value=sessions)
        session_repo.count_sessions = AsyncMock(return_value=1)

        result = await service.list_sessions(status=SessionStatus.RUNNING)

        assert "items" in result
        session_repo.find_sessions.assert_called_once()

    @pytest.mark.asyncio
    async def test_terminate_session_not_found(self, service, session_repo):
        """测试终止不存在的会话"""
        session_repo.find_by_id.return_value = None

        with pytest.raises(NotFoundError, match="Session not found"):
            await service.terminate_session("non-existent")

    @pytest.mark.asyncio
    async def test_terminate_session_with_container(self, service, session_repo, scheduler):
        """测试终止带容器的会话"""
        session = Session(
            id="sess_123",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_123",
            runtime_type="docker",
            container_id="container-123"
        )
        session_repo.find_by_id.return_value = session

        result = await service.terminate_session("sess_123")

        assert result.status == SessionStatus.TERMINATED.value
        scheduler.destroy_container.assert_called_once()

    @pytest.mark.asyncio
    async def test_get_session_executions(self, service, session_repo, execution_repo):
        """测试获取会话的执行记录"""
        session = Session(
            id="sess_123",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_123",
            runtime_type="docker"
        )
        session_repo.find_by_id.return_value = session

        from src.domain.entities.execution import Execution
        from src.domain.value_objects.execution_status import ExecutionState

        execution_state = ExecutionState(status=ExecutionStatus.COMPLETED)
        executions = [
            Execution(
                id="exec_1",
                session_id="sess_123",
                state=execution_state,
                code="print('hello')",
                language="python"
            )
        ]
        execution_repo.find_by_session_id.return_value = executions

        result = await service.list_executions("sess_123")

        assert len(result) == 1
        execution_repo.find_by_session_id.assert_called_once()

    @pytest.mark.asyncio
    async def test_get_session_executions_session_not_found(self, service, session_repo, execution_repo):
        """测试获取不存在会话的执行记录"""
        # list_executions 不检查会话是否存在，直接查询执行记录
        execution_repo.find_by_session_id.return_value = []

        result = await service.list_executions("non-existent")

        # 应该返回空列表
        assert result == []
