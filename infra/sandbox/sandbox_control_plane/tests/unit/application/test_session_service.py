"""
会话应用服务单元测试

测试 SessionService 的用例编排逻辑。
"""

from types import SimpleNamespace
from unittest.mock import AsyncMock, Mock

import pytest

from src.application.commands.create_session import CreateSessionCommand
from src.application.commands.execute_code import ExecuteCodeCommand
from src.application.commands.install_session_dependencies import (
    InstallSessionDependenciesCommand,
)
from src.application.queries.get_execution import GetExecutionQuery
from src.application.services.session_service import SessionService
from src.domain.entities.session import Session
from src.domain.entities.template import Template
from src.domain.services.scheduler import RuntimeNode
from src.domain.value_objects.execution_status import ExecutionStatus, SessionStatus
from src.domain.value_objects.resource_limit import ResourceLimit
from src.infrastructure.executors.dto import (
    ExecutorInstalledDependency,
    ExecutorSyncSessionConfigResponse,
)
from src.infrastructure.executors.errors import ExecutorConnectionError
from src.shared.errors.domain import ConflictError, NotFoundError, ValidationError


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
            base_image="python:3.11-slim",
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
            cached_templates=["python-datascience"],
        )
        scheduler.schedule.return_value = runtime_node

        # 执行命令
        command = CreateSessionCommand(
            template_id="python-datascience", timeout=300, resource_limit=ResourceLimit.default()
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

        command = CreateSessionCommand(template_id="non-existent", timeout=300)

        with pytest.raises(NotFoundError, match="Template not found"):
            await service.create_session(command)

    @pytest.mark.asyncio
    async def test_create_session_uses_default_template_when_omitted(
        self,
        service,
        template_repo,
        scheduler,
        session_repo,
        monkeypatch,
    ):
        """测试未指定模板时使用默认模板配置"""
        monkeypatch.setattr(
            "src.application.services.session_service.get_settings",
            lambda: SimpleNamespace(
                default_template_id="multi-language",
                s3_bucket="sandbox-workspace",
            ),
        )
        template = Template(
            id="multi-language",
            name="Multi Language Basic",
            image="sandbox-template-multi-language:test",
            base_image="sandbox-multi-executor-base:go1.25-python3.11-v1",
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
            cached_templates=["multi-language"],
        )
        scheduler.schedule.return_value = runtime_node

        command = CreateSessionCommand(timeout=300, resource_limit=ResourceLimit.default())

        result = await service.create_session(command)

        assert result.template_id == "multi-language"
        template_repo.find_by_id.assert_awaited_once_with("multi-language")
        assert command.template_id == "multi-language"

    @pytest.mark.asyncio
    async def test_get_session_success(self, service, session_repo):
        """测试成功获取会话"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
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
            runtime_type="docker",
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
            runtime_type="docker",
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
            runtime_type="docker",
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
    async def test_create_session_with_manual_id(
        self, service, template_repo, scheduler, session_repo
    ):
        """测试使用手动指定 ID 创建会话"""
        template = Template(
            id="python-test", name="Python Test", image="python:3.11", base_image="python:3.11-slim"
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
            cached_templates=["python-test"],
        )
        scheduler.schedule.return_value = runtime_node

        # 第一个调用返回 None（检查 ID 是否存在），后续调用返回会话
        session_repo.find_by_id.side_effect = [None, None]

        command = CreateSessionCommand(
            id="custom-session-id",
            template_id="python-test",
            timeout=300,
            resource_limit=ResourceLimit.default(),
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
    async def test_install_session_dependencies_uses_default_install_timeout(
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
        )
        session_repo.find_by_id.return_value = session
        executor_client.sync_session_config.return_value = ExecutorSyncSessionConfigResponse(
            status="completed",
            installed_dependencies=[],
            started_at="2026-03-09T12:00:00+00:00",
            completed_at="2026-03-09T12:00:05+00:00",
        )

        await service.install_session_dependencies(
            InstallSessionDependenciesCommand(
                session_id="sess_1",
                dependencies=["requests==2.31.0"],
            )
        )

        assert executor_client.sync_session_config.call_args.kwargs["executor_timeout"] == 300

    @pytest.mark.asyncio
    async def test_install_session_dependencies_uses_custom_install_timeout(
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
        )
        session_repo.find_by_id.return_value = session
        executor_client.sync_session_config.return_value = ExecutorSyncSessionConfigResponse(
            status="completed",
            installed_dependencies=[],
            started_at="2026-03-09T12:00:00+00:00",
            completed_at="2026-03-09T12:00:05+00:00",
        )

        await service.install_session_dependencies(
            InstallSessionDependenciesCommand(
                session_id="sess_1",
                dependencies=["requests==2.31.0"],
                install_timeout=900,
            )
        )

        assert executor_client.sync_session_config.call_args.kwargs["executor_timeout"] == 900

    @pytest.mark.asyncio
    async def test_create_session_with_duplicate_id(
        self, service, template_repo, scheduler, session_repo
    ):
        """测试使用重复 ID 创建会话"""
        template = Template(
            id="python-test", name="Python Test", image="python:3.11", base_image="python:3.11-slim"
        )
        template_repo.find_by_id.return_value = template

        existing_session = Session(
            id="existing-session-id",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/existing",
            runtime_type="docker",
        )
        session_repo.find_by_id.return_value = existing_session

        command = CreateSessionCommand(
            id="existing-session-id", template_id="python-test", timeout=300
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
            id="python-test", name="Python Test", image="python:3.11", base_image="python:3.11-slim"
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
            cached_templates=["python-test"],
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
                runtime_type="docker",
            ),
            Session(
                id="sess_2",
                template_id="python-test",
                status=SessionStatus.TERMINATED,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://bucket/sessions/sess_2",
                runtime_type="docker",
            ),
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
                runtime_type="docker",
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
            container_id="container-123",
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
            runtime_type="docker",
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
                language="python",
            )
        ]
        execution_repo.find_by_session_id.return_value = executions

        result = await service.list_executions("sess_123")

        assert len(result) == 1
        execution_repo.find_by_session_id.assert_called_once()

    @pytest.mark.asyncio
    async def test_get_session_executions_session_not_found(
        self, service, session_repo, execution_repo
    ):
        """测试获取不存在会话的执行记录"""
        # list_executions 不检查会话是否存在，直接查询执行记录
        execution_repo.find_by_session_id.return_value = []

        result = await service.list_executions("non-existent")

        # 应该返回空列表
        assert result == []

    @pytest.mark.asyncio
    async def test_create_session_without_container_creation_support(
        self, session_repo, template_repo, scheduler
    ):
        """调度器不支持容器创建时保留创建中的会话记录。"""
        scheduler_without_container_create = SimpleNamespace(schedule=scheduler.schedule)
        service = SessionService(
            session_repo=session_repo,
            execution_repo=Mock(),
            template_repo=template_repo,
            scheduler=scheduler_without_container_create,
        )
        template_repo.find_by_id.return_value = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim",
        )
        scheduler.schedule.return_value = RuntimeNode(
            id="node-1",
            type="docker",
            url="http://node-1:2375",
            status="healthy",
            cpu_usage=0.1,
            mem_usage=0.1,
            session_count=1,
            max_sessions=10,
            cached_templates=[],
        )

        result = await service.create_session(
            CreateSessionCommand(template_id="python-test", timeout=300)
        )

        assert result.status == SessionStatus.CREATING.value
        assert result.container_id is None

    @pytest.mark.asyncio
    async def test_create_session_marks_failed_when_container_creation_fails(
        self, service, template_repo, scheduler, session_repo
    ):
        template_repo.find_by_id.return_value = Template(
            id="python-test",
            name="Python Test",
            image="python:3.11",
            base_image="python:3.11-slim",
        )
        scheduler.schedule.return_value = RuntimeNode(
            id="node-1",
            type="docker",
            url="http://node-1:2375",
            status="healthy",
            cpu_usage=0.1,
            mem_usage=0.1,
            session_count=1,
            max_sessions=10,
            cached_templates=[],
        )
        scheduler.create_container_for_session.side_effect = RuntimeError("docker failed")

        with pytest.raises(ValidationError, match="Failed to create container"):
            await service.create_session(
                CreateSessionCommand(
                    template_id="python-test",
                    timeout=300,
                    dependencies=["requests==2.31.0"],
                )
            )

        saved_session = session_repo.save.call_args.args[0]
        assert saved_session.status == SessionStatus.FAILED
        assert saved_session.dependency_install_status == "failed"

    @pytest.mark.asyncio
    async def test_handle_container_creation_failure_cleans_existing_container(
        self, service, session_repo, scheduler
    ):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="docker",
        )

        with pytest.raises(ValidationError):
            await service._handle_container_creation_failure(
                session=session,
                container_id="container-1",
                error=RuntimeError("boom"),
            )

        scheduler.destroy_container.assert_awaited_once_with("container-1")
        session_repo.save.assert_awaited()
        assert session.status == SessionStatus.FAILED

    @pytest.mark.asyncio
    async def test_destroy_container_ignores_scheduler_errors(self, service, scheduler):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="docker",
            container_id="container-1",
        )
        scheduler.destroy_container.side_effect = RuntimeError("cleanup failed")

        await service._destroy_container(session)

        scheduler.destroy_container.assert_awaited_once_with(container_id="container-1")

    @pytest.mark.asyncio
    async def test_cleanup_storage_success_and_failure(self, session_repo, template_repo, scheduler):
        storage = Mock()
        storage.delete_prefix = AsyncMock(return_value=3)
        service = SessionService(
            session_repo=session_repo,
            execution_repo=Mock(),
            template_repo=template_repo,
            scheduler=scheduler,
            storage_service=storage,
        )
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="docker",
        )

        await service._cleanup_storage(session)
        storage.delete_prefix.assert_awaited_once_with("s3://bucket/sessions/sess_1")

        storage.delete_prefix.reset_mock()
        storage.delete_prefix.side_effect = RuntimeError("s3 failed")
        await service._cleanup_storage(session)
        storage.delete_prefix.assert_awaited_once()

    @pytest.mark.asyncio
    async def test_execute_code_success(self, service, session_repo, execution_repo, scheduler):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
            container_id="container-1",
            env_vars={"A": "B"},
        )
        session_repo.find_by_id.return_value = session
        execution_repo.commit = AsyncMock()
        scheduler.execute = AsyncMock(return_value="exec_1")

        result = await service.execute_code(
            ExecuteCodeCommand(
                session_id="sess_1",
                code="print('ok')",
                language="python",
                timeout=10,
                event_data={"name": "Ada"},
                working_directory="src",
            )
        )

        assert result.session_id == "sess_1"
        execution_repo.save.assert_awaited_once()
        execution_repo.commit.assert_awaited_once()
        scheduler.execute.assert_awaited_once()
        request = scheduler.execute.call_args.kwargs["execution_request"]
        assert request.working_directory == "src"

    @pytest.mark.asyncio
    async def test_execute_code_rejects_missing_inactive_and_containerless_sessions(
        self, service, session_repo, execution_repo
    ):
        command = ExecuteCodeCommand(session_id="sess_1", code="print(1)", language="python")
        session_repo.find_by_id.return_value = None
        with pytest.raises(NotFoundError):
            await service.execute_code(command)

        session_repo.find_by_id.return_value = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
        )
        with pytest.raises(ValidationError, match="Session is not active"):
            await service.execute_code(command)

        session_repo.find_by_id.return_value = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
        )
        execution_repo.commit = AsyncMock()
        with pytest.raises(ValidationError, match="has no container"):
            await service.execute_code(command)

    @pytest.mark.asyncio
    async def test_get_execution_success_and_not_found(self, service, execution_repo):
        from src.domain.entities.execution import Execution
        from src.domain.value_objects.execution_status import ExecutionState

        execution_repo.find_by_id.return_value = Execution(
            id="exec_1",
            session_id="sess_1",
            state=ExecutionState(status=ExecutionStatus.COMPLETED),
            code="print('hello')",
            language="python",
        )

        result = await service.get_execution(GetExecutionQuery(execution_id="exec_1"))
        assert result.id == "exec_1"

        execution_repo.find_by_id.return_value = None
        with pytest.raises(NotFoundError):
            await service.get_execution(GetExecutionQuery(execution_id="missing"))

    @pytest.mark.asyncio
    async def test_cleanup_idle_sessions_counts_only_cleaned_sessions(
        self, service, session_repo, scheduler
    ):
        class HashableSession(Session):
            def __hash__(self):
                return hash(self.id)

        active = HashableSession(
            id="sess_active",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_active",
            runtime_type="docker",
            container_id="container-1",
        )
        terminated = HashableSession(
            id="sess_done",
            template_id="python-test",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_done",
            runtime_type="docker",
        )
        session_repo.find_idle_sessions = AsyncMock(return_value=[active])
        session_repo.find_expired_sessions = AsyncMock(return_value=[active, terminated])

        cleaned = await service.cleanup_idle_sessions()

        assert cleaned == 1
        assert active.status == SessionStatus.TERMINATED
        scheduler.destroy_container.assert_awaited_with(container_id="container-1")

    @pytest.mark.asyncio
    async def test_cleanup_session_ignores_container_destroy_error(
        self, service, session_repo, scheduler
    ):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="docker",
            container_id="container-1",
        )
        scheduler.destroy_container.side_effect = RuntimeError("docker failed")

        assert await service._cleanup_session(session) is True
        assert session.status == SessionStatus.TERMINATED
        session_repo.save.assert_awaited_once_with(session)

    @pytest.mark.asyncio
    async def test_sync_session_dependencies_validation_failures(self, service):
        inactive = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
            container_id="container-1",
        )
        with pytest.raises(ValidationError, match="not active"):
            await service._sync_session_dependencies(inactive, sync_mode="replace")

        no_container = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
        )
        with pytest.raises(ValidationError, match="has no container"):
            await service._sync_session_dependencies(no_container, sync_mode="replace")

    @pytest.mark.asyncio
    async def test_sync_session_dependencies_marks_failed_on_executor_error(
        self, service, session_repo, executor_client
    ):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
            container_id="container-1",
            requested_dependencies=["requests==2.31.0"],
        )
        executor_client.sync_session_config.side_effect = ExecutorConnectionError(
            "http://sandbox", "refused"
        )

        with pytest.raises(ExecutorConnectionError):
            await service._sync_session_dependencies(session, sync_mode="merge")

        assert session.dependency_install_status == "failed"
        assert session_repo.save.await_count == 2

    @pytest.mark.asyncio
    async def test_sync_session_dependencies_without_completed_or_started_timestamps(
        self, service, session_repo, executor_client
    ):
        session = Session(
            id="sess_1",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://bucket/sessions/sess_1",
            runtime_type="python3.11",
            container_id="container-1",
            requested_dependencies=["requests==2.31.0"],
        )
        executor_client.sync_session_config.return_value = ExecutorSyncSessionConfigResponse(
            status="completed",
            installed_dependencies=[
                ExecutorInstalledDependency(
                    name="requests",
                    version="2.31.0",
                    install_location="/workspace/.venv/",
                    install_time="2026-03-09T12:00:05Z",
                    is_from_template=False,
                )
            ],
            started_at=None,
            completed_at=None,
        )

        result = await service._sync_session_dependencies(session, sync_mode="replace")

        assert result.dependency_install_status == "completed"
        assert result.installed_dependencies[0]["name"] == "requests"

    def test_schedule_initial_dependency_sync_without_scheduler(
        self, session_repo, template_repo, scheduler
    ):
        service = SessionService(
            session_repo=session_repo,
            execution_repo=Mock(),
            template_repo=template_repo,
            scheduler=scheduler,
        )

        service._schedule_initial_dependency_sync("sess_1", 300, 1)

    def test_infer_runtime_type_variants(self, service):
        assert service._infer_runtime_type("node:20") == "nodejs20"
        assert service._infer_runtime_type("java:17") == "java17"
        assert service._infer_runtime_type("golang:1.21") == "go1.21"
        assert service._infer_runtime_type("ubuntu:22.04") == "python3.11"
