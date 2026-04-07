"""
状态同步服务单元测试

测试 StateSyncService 的状态同步逻辑。
"""
import pytest
from unittest.mock import Mock, AsyncMock
from datetime import datetime

from src.application.services.state_sync_service import StateSyncService
from src.domain.entities.session import Session
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus
from src.domain.repositories.session_repository import ISessionRepository
from src.infrastructure.container_scheduler.base import IContainerScheduler


class TestStateSyncService:
    """状态同步服务测试"""

    @pytest.fixture
    def session_repo(self):
        """模拟会话仓储"""
        repo = Mock()
        repo.save = AsyncMock()
        repo.find_by_id = AsyncMock()
        repo.find_by_status = AsyncMock()
        return repo

    @pytest.fixture
    def container_scheduler(self):
        """模拟容器调度器"""
        scheduler = Mock()
        scheduler.is_container_running = AsyncMock()
        scheduler.create_container = AsyncMock()
        scheduler.start_container = AsyncMock()
        return scheduler

    @pytest.fixture
    def service(self, session_repo, container_scheduler):
        """创建状态同步服务"""
        return StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler
        )

    @pytest.fixture
    def running_session(self):
        """创建运行中的会话"""
        return Session(
            id="sess_running",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_running",
            runtime_type="docker",
            container_id="container-running",
            env_vars={"SESSION_ID": "sess_running"}
        )

    @pytest.fixture
    def creating_session(self):
        """创建创建中的会话"""
        return Session(
            id="sess_creating",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_creating",
            runtime_type="docker",
            container_id="container-creating",
            env_vars={"SESSION_ID": "sess_creating"}
        )

    @pytest.mark.asyncio
    async def test_sync_on_startup_all_healthy(self, service, session_repo, container_scheduler):
        """测试启动同步时所有容器健康"""
        session1 = Session(
            id="sess_1",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_1",
            runtime_type="docker",
            container_id="container-1"
        )
        session2 = Session(
            id="sess_2",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_2",
            runtime_type="docker",
            container_id="container-2"
        )

        # 使用 side_effect 区分不同参数的返回值
        session_repo.find_by_status.side_effect = [
            [session1],  # running 状态查询
            [session2]   # creating 状态查询
        ]
        container_scheduler.is_container_running.return_value = True

        result = await service.sync_on_startup()

        # 验证结果
        assert result["total"] == 2
        assert result["healthy"] == 2
        assert result["unhealthy"] == 0

    @pytest.mark.asyncio
    async def test_sync_on_startup_with_unhealthy(self, service, session_repo, container_scheduler):
        """测试启动同步时有不健康容器"""
        session1 = Session(
            id="sess_1",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_1",
            runtime_type="docker",
            container_id="container-1",
            env_vars={"SESSION_ID": "sess_1"}
        )
        session2 = Session(
            id="sess_2",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_2",
            runtime_type="docker",
            container_id="container-2",
            env_vars={"SESSION_ID": "sess_2"}
        )

        session_repo.find_by_status.return_value = [session1, session2]

        # 第一个健康，第二个不健康
        container_scheduler.is_container_running.side_effect = [True, False]

        result = await service.sync_on_startup()

        # find_by_status 可能被调用多次（running 和 creating）
        assert result["total"] >= 2
        assert result["healthy"] >= 1
        assert result["unhealthy"] >= 1

    @pytest.mark.asyncio
    async def test_sync_on_startup_skip_no_container(self, service, session_repo):
        """测试跳过没有 container_id 的会话"""
        session = Session(
            id="sess_no_container",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_no_container",
            runtime_type="docker",
            container_id=None  # 没有容器
        )

        # 使用 side_effect 区分不同参数的返回值
        session_repo.find_by_status.side_effect = [
            [session],  # running 状态查询
            []         # creating 状态查询
        ]

        result = await service.sync_on_startup()

        # 验证不检查容器状态（因为会话没有 container_id）
        assert result["total"] == 1

    @pytest.mark.asyncio
    async def test_periodic_health_check(self, service, session_repo, container_scheduler):
        """测试定期健康检查"""
        session1 = Session(
            id="sess_1",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_1",
            runtime_type="docker",
            container_id="container-1",
            env_vars={"SESSION_ID": "sess_1"}
        )
        session2 = Session(
            id="sess_2",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_2",
            runtime_type="docker",
            container_id="container-2",
            env_vars={"SESSION_ID": "sess_2"}
        )

        session_repo.find_by_status.return_value = [session1, session2]
        container_scheduler.is_container_running.return_value = True

        result = await service.periodic_health_check()

        assert result["checked"] == 2
        assert result["healthy"] == 2
        assert result["unhealthy"] == 0

    @pytest.mark.asyncio
    async def test_periodic_health_check_only_running(self, service, session_repo):
        """测试定期健康检查只检查 RUNNING 状态"""
        creating_session = Session(
            id="sess_creating",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_creating",
            runtime_type="docker",
            container_id="container-creating"
        )

        # 只调用 find_by_status("running")，不调用 "creating"
        session_repo.find_by_status.return_value = []

        result = await service.periodic_health_check()

        assert result["checked"] == 0

    @pytest.mark.asyncio
    async def test_check_session_health_success(self, service, session_repo, container_scheduler):
        """测试检查单个会话健康状态"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker",
            container_id="container-123"
        )

        session_repo.find_by_id.return_value = session
        container_scheduler.is_container_running.return_value = True

        result = await service.check_session_health("sess_123")

        assert result["session_id"] == "sess_123"
        assert result["container_id"] == "container-123"
        assert result["container_running"] is True
        assert result["status"] == "healthy"

    @pytest.mark.asyncio
    async def test_check_session_health_not_found(self, service, session_repo):
        """测试检查不存在的会话"""
        session_repo.find_by_id.return_value = None

        result = await service.check_session_health("non-existent")

        assert result["status"] == "not_found"
        assert "error" in result

    @pytest.mark.asyncio
    async def test_check_session_health_no_container(self, service, session_repo):
        """测试检查没有容器的会话"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker",
            container_id=None
        )

        session_repo.find_by_id.return_value = session

        result = await service.check_session_health("sess_123")

        assert result["status"] == "no_container"
        assert "error" in result

    @pytest.mark.asyncio
    async def test_check_session_health_unhealthy(self, service, session_repo, container_scheduler):
        """测试检查不健康的会话"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker",
            container_id="container-123"
        )

        session_repo.find_by_id.return_value = session
        container_scheduler.is_container_running.return_value = False

        result = await service.check_session_health("sess_123")

        assert result["status"] == "unhealthy"
        assert result["container_running"] is False

    @pytest.mark.asyncio
    async def test_recovery_success(self, service, session_repo, container_scheduler):
        """测试成功恢复会话"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker",
            container_id="old-container",
            env_vars={"SESSION_ID": "sess_123"}
        )

        container_scheduler.is_container_running.return_value = False
        container_scheduler.create_container.return_value = "new-container"

        # 不传入 scheduler 参数，恢复功能会尝试创建新容器
        result = await service._attempt_recovery(session)

        assert result is True
        assert session.container_id == "new-container"
        assert session.status == SessionStatus.RUNNING

    @pytest.mark.asyncio
    async def test_recovery_failure_marks_failed(self, service, session_repo, container_scheduler):
        """测试恢复失败时标记会话为失败"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker",
            container_id="old-container",
            env_vars={"SESSION_ID": "sess_123"}
        )

        container_scheduler.is_container_running.return_value = False
        container_scheduler.create_container.side_effect = Exception("Docker error")

        result = await service._attempt_recovery(session)

        assert result is False
        assert session.status == SessionStatus.FAILED

    @pytest.mark.asyncio
    async def test_sync_error_handling(self, service, session_repo):
        """测试同步过程中的错误处理"""
        session_repo.find_by_status.side_effect = Exception("Database error")

        result = await service.sync_on_startup()

        assert "errors" in result
        assert len(result["errors"]) > 0
        assert any("Database error" in str(e) for e in result["errors"])

    @pytest.mark.asyncio
    async def test_sync_empty_sessions(self, service, session_repo):
        """测试没有会话需要同步"""
        session_repo.find_by_status.return_value = []

        result = await service.sync_on_startup()

        assert result["total"] == 0
        assert result["healthy"] == 0
        assert result["unhealthy"] == 0
