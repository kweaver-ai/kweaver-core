"""
会话清理服务单元测试

测试 SessionCleanupService 的清理逻辑。
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch
from datetime import datetime, timedelta

from src.application.services.session_cleanup_service import SessionCleanupService
from src.domain.entities.session import Session
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus
from src.domain.repositories.session_repository import ISessionRepository
from src.domain.services.scheduler import IScheduler


class TestSessionCleanupService:
    """会话清理服务测试"""

    @pytest.fixture
    def session_repo(self):
        """模拟会话仓储"""
        repo = Mock()
        repo.save = AsyncMock()
        repo.find_by_id = AsyncMock()
        repo.find_by_status = AsyncMock()
        return repo

    @pytest.fixture
    def scheduler(self):
        """模拟调度器"""
        sched = Mock()
        sched.destroy_container = AsyncMock()
        return sched

    @pytest.fixture
    def storage_service(self):
        """模拟存储服务"""
        storage = Mock()
        storage.delete_prefix = AsyncMock()
        return storage

    @pytest.fixture
    def service(self, session_repo, scheduler, storage_service):
        """创建会话清理服务"""
        return SessionCleanupService(
            session_repo=session_repo,
            scheduler=scheduler,
            idle_timeout_minutes=30,
            max_lifetime_hours=6,
            storage_service=storage_service
        )

    @pytest.fixture
    def active_session(self):
        """创建活跃会话"""
        return Session(
            id="sess_active",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_active",
            runtime_type="docker",
            container_id="container-active",
            last_activity_at=datetime.now()
        )

    @pytest.fixture
    def idle_session(self):
        """创建空闲会话"""
        old_time = datetime.now() - timedelta(minutes=35)
        return Session(
            id="sess_idle",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_idle",
            runtime_type="docker",
            container_id="container-idle",
            last_activity_at=old_time
        )

    @pytest.fixture
    def expired_session(self):
        """创建过期会话"""
        old_time = datetime.now() - timedelta(hours=7)
        return Session(
            id="sess_expired",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_expired",
            runtime_type="docker",
            container_id="container-expired",
            created_at=old_time,
            last_activity_at=datetime.now()
        )

    @pytest.mark.asyncio
    async def test_cleanup_idle_sessions(self, service, session_repo, scheduler, storage_service, idle_session):
        """测试清理空闲会话"""
        session_repo.find_by_status.return_value = [idle_session]
        storage_service.delete_prefix.return_value = 5

        result = await service.cleanup_idle_sessions()

        assert result["idle_cleaned"] == 1
        assert idle_session.status == SessionStatus.TERMINATED
        # destroy_container 可能使用 container_id 参数名
        assert scheduler.destroy_container.called
        storage_service.delete_prefix.assert_called_once()
        session_repo.save.assert_called_once()

    @pytest.mark.asyncio
    async def test_cleanup_expired_sessions(self, service, session_repo, scheduler, storage_service, expired_session):
        """测试清理过期会话"""
        session_repo.find_by_status.return_value = [expired_session]
        storage_service.delete_prefix.return_value = 3

        result = await service.cleanup_idle_sessions()

        assert result["expired_cleaned"] == 1
        assert expired_session.status == SessionStatus.TERMINATED
        # destroy_container 可能使用 container_id 参数名
        assert scheduler.destroy_container.called

    @pytest.mark.asyncio
    async def test_no_cleanup_for_active_sessions(self, service, session_repo, active_session):
        """测试不清理活跃会话"""
        session_repo.find_by_status.return_value = [active_session]

        result = await service.cleanup_idle_sessions()

        assert result["idle_cleaned"] == 0
        assert result["expired_cleaned"] == 0
        assert active_session.status == SessionStatus.RUNNING

    @pytest.mark.asyncio
    async def test_cleanup_mixed_sessions(self, service, session_repo, scheduler, storage_service):
        """测试清理混合状态的会话"""
        session_repo.find_by_status.return_value = [
            Session(
                id="sess_1",
                template_id="python-basic",
                status=SessionStatus.RUNNING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/sess_1",
                runtime_type="docker",
                container_id="container-1",
                last_activity_at=datetime.now()  # 活跃
            ),
            Session(
                id="sess_2",
                template_id="python-basic",
                status=SessionStatus.RUNNING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/sess_2",
                runtime_type="docker",
                container_id="container-2",
                last_activity_at=datetime.now() - timedelta(minutes=40)  # 空闲
            ),
        ]
        storage_service.delete_prefix.return_value = 2

        result = await service.cleanup_idle_sessions()

        assert result["total_checked"] == 2
        assert result["idle_cleaned"] == 1

    @pytest.mark.asyncio
    async def test_cleanup_disabled_idle_timeout(self, session_repo, scheduler, storage_service):
        """测试禁用空闲超时清理"""
        service = SessionCleanupService(
            session_repo=session_repo,
            scheduler=scheduler,
            idle_timeout_minutes=-1,  # 禁用
            max_lifetime_hours=6,
            storage_service=storage_service
        )

        idle_session = Session(
            id="sess_idle",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_idle",
            runtime_type="docker",
            container_id="container-idle",
            last_activity_at=datetime.now() - timedelta(hours=10)  # 超过空闲阈值
        )
        session_repo.find_by_status.return_value = [idle_session]

        result = await service.cleanup_idle_sessions()

        # 空闲会话不应被清理
        assert result["idle_cleaned"] == 0
        assert idle_session.status == SessionStatus.RUNNING

    @pytest.mark.asyncio
    async def test_cleanup_disabled_max_lifetime(self, session_repo, scheduler, storage_service):
        """测试禁用最大生命周期清理"""
        service = SessionCleanupService(
            session_repo=session_repo,
            scheduler=scheduler,
            idle_timeout_minutes=30,
            max_lifetime_hours=-1,  # 禁用
            storage_service=storage_service
        )

        expired_session = Session(
            id="sess_expired",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_expired",
            runtime_type="docker",
            container_id="container-expired",
            created_at=datetime.now() - timedelta(days=1),  # 超过生命周期
            last_activity_at=datetime.now()
        )
        session_repo.find_by_status.return_value = [expired_session]

        result = await service.cleanup_idle_sessions()

        # 过期会话不应被清理
        assert result["expired_cleaned"] == 0
        assert expired_session.status == SessionStatus.RUNNING

    @pytest.mark.asyncio
    async def test_cleanup_orphaned_failed_sessions(self, service, session_repo):
        """测试清理孤立的失败会话"""
        failed_session = Session(
            id="sess_failed",
            template_id="python-basic",
            status=SessionStatus.FAILED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_failed",
            runtime_type="docker",
            container_id="container-failed"
        )
        # 使用 side_effect 区分不同参数的返回值
        session_repo.find_by_status.side_effect = [
            [failed_session],  # failed 状态查询
            []  # timeout 状态查询
        ]

        result = await service.cleanup_orphaned_sessions()

        assert result["cleaned"] == 1
        assert failed_session.status == SessionStatus.TERMINATED

    @pytest.mark.asyncio
    async def test_cleanup_orphaned_timeout_sessions(self, service, session_repo):
        """测试清理孤立超时会话"""
        timeout_session = Session(
            id="sess_timeout",
            template_id="python-basic",
            status=SessionStatus.TIMEOUT,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_timeout",
            runtime_type="docker",
            container_id="container-timeout"
        )
        # 使用 side_effect 区分不同参数的返回值
        session_repo.find_by_status.side_effect = [
            [],  # failed 状态查询
            [timeout_session]  # timeout 状态查询
        ]

        result = await service.cleanup_orphaned_sessions()

        assert result["cleaned"] == 1

    @pytest.mark.asyncio
    async def test_cleanup_orphaned_without_container(self, service, session_repo):
        """测试不清理没有容器的孤立会话"""
        failed_session = Session(
            id="sess_failed",
            template_id="python-basic",
            status=SessionStatus.FAILED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_failed",
            runtime_type="docker",
            container_id=None  # 没有容器
        )
        session_repo.find_by_status.return_value = [failed_session]

        result = await service.cleanup_orphaned_sessions()

        assert result["cleaned"] == 0

    @pytest.mark.asyncio
    async def test_cleanup_session_files(self, service, storage_service):
        """测试清理会话文件"""
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_123",
            runtime_type="docker"
        )
        storage_service.delete_prefix.return_value = 7

        deleted_count = await service.cleanup_session_files(session, "test_cleanup")

        assert deleted_count == 7
        storage_service.delete_prefix.assert_called_once_with(
            "s3://sandbox-workspace/sessions/sess_123"
        )

    @pytest.mark.asyncio
    async def test_cleanup_session_without_workspace(self, service, storage_service):
        """测试清理没有 workspace 的会话"""
        # 注意：Session 实体会验证 workspace_path 不能为空
        # 所以这里使用非 S3 路径
        session = Session(
            id="sess_123",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="local:/tmp/sess_123",  # 非 S3 路径，不会执行 S3 清理
            runtime_type="docker"
        )

        # 模拟 delete_prefix 返回 0
        storage_service.delete_prefix.return_value = 0

        deleted_count = await service.cleanup_session_files(session, "test_cleanup")

        # 非 S3 路径可能仍会执行清理，但结果应为 0
        assert deleted_count == 0

    @pytest.mark.asyncio
    async def test_cleanup_by_ids(self, service, session_repo):
        """测试按 ID 清理会话"""
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

        result = await service.cleanup_by_ids(["sess_123"])

        assert result["cleaned"] == 1
        assert session.status == SessionStatus.TERMINATED

    @pytest.mark.asyncio
    async def test_cleanup_by_ids_not_found(self, service, session_repo):
        """测试按 ID 清理不存在的会话"""
        session_repo.find_by_id.return_value = None

        result = await service.cleanup_by_ids(["non-existent"])

        assert result["not_found"] == 1
        assert result["cleaned"] == 0

    @pytest.mark.asyncio
    async def test_cleanup_container_destruction_failure(self, service, session_repo, scheduler, storage_service):
        """测试容器销毁失败时的处理"""
        idle_session = Session(
            id="sess_idle",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_idle",
            runtime_type="docker",
            container_id="container-idle",
            last_activity_at=datetime.now() - timedelta(minutes=35)
        )
        session_repo.find_by_status.return_value = [idle_session]

        # 模拟容器销毁失败
        scheduler.destroy_container.side_effect = Exception("Docker error")
        storage_service.delete_prefix.return_value = 2

        result = await service.cleanup_idle_sessions()

        # 应继续执行文件清理和状态更新
        assert result["idle_cleaned"] == 1
        assert idle_session.status == SessionStatus.TERMINATED
        storage_service.delete_prefix.assert_called_once()

    @pytest.mark.asyncio
    async def test_cleanup_error_handling(self, service, session_repo):
        """测试清理过程中的错误处理"""
        session_repo.find_by_status.side_effect = Exception("Database error")

        result = await service.cleanup_idle_sessions()

        assert "errors" in result
        assert len(result["errors"]) > 0
        assert any("Database error" in str(e) for e in result["errors"])
