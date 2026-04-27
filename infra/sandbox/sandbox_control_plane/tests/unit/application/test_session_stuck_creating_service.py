"""
会话创建超时检测服务单元测试

测试 SessionStuckCreatingService 的超时检测逻辑。
"""
import pytest
from unittest.mock import Mock, AsyncMock
from datetime import datetime, timedelta

from src.application.services.session_stuck_creating_service import SessionStuckCreatingService
from src.domain.entities.session import Session
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus
from src.domain.repositories.session_repository import ISessionRepository


class TestSessionStuckCreatingService:
    """会话创建超时检测服务测试"""

    @pytest.fixture
    def session_repo(self):
        """模拟会话仓储"""
        repo = Mock()
        repo.save = AsyncMock()
        repo.find_by_id = AsyncMock()
        repo.find_by_status = AsyncMock()
        return repo

    @pytest.fixture
    def service(self, session_repo):
        """创建会话创建超时检测服务"""
        return SessionStuckCreatingService(
            session_repo=session_repo,
            creating_timeout_seconds=300,  # 5 分钟
        )

    @pytest.fixture
    def creating_session_stuck(self):
        """创建卡在 creating 状态的超时会话"""
        old_time = datetime.now() - timedelta(minutes=6)  # 超过 5 分钟阈值
        return Session(
            id="sess_stuck",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_stuck",
            runtime_type="docker",
            created_at=old_time,
        )

    @pytest.fixture
    def creating_session_recent(self):
        """创建最近创建的 creating 状态会话（未超时）"""
        recent_time = datetime.now() - timedelta(minutes=2)  # 未超过 5 分钟阈值
        return Session(
            id="sess_recent",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_recent",
            runtime_type="docker",
            created_at=recent_time,
        )

    @pytest.mark.asyncio
    async def test_mark_stuck_session_as_failed(self, service, session_repo, creating_session_stuck):
        """测试标记超时的 creating 会话为 failed"""
        session_repo.find_by_status.return_value = [creating_session_stuck]

        result = await service.check_and_mark_stuck_sessions()

        assert result["total_checked"] == 1
        assert result["marked_failed"] == 1
        assert creating_session_stuck.status == SessionStatus.FAILED
        session_repo.save.assert_called_once_with(creating_session_stuck)

    @pytest.mark.asyncio
    async def test_keep_recent_creating_session(self, service, session_repo, creating_session_recent):
        """测试不标记未超时的 creating 会话"""
        session_repo.find_by_status.return_value = [creating_session_recent]

        result = await service.check_and_mark_stuck_sessions()

        assert result["total_checked"] == 1
        assert result["marked_failed"] == 0
        assert creating_session_recent.status == SessionStatus.CREATING
        session_repo.save.assert_not_called()

    @pytest.mark.asyncio
    async def test_check_mixed_creating_sessions(self, service, session_repo):
        """测试检查混合状态的 creating 会话"""
        stuck_time = datetime.now() - timedelta(minutes=6)
        recent_time = datetime.now() - timedelta(minutes=2)

        session_repo.find_by_status.return_value = [
            Session(
                id="sess_1",
                template_id="python-basic",
                status=SessionStatus.CREATING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/sess_1",
                runtime_type="docker",
                created_at=stuck_time,  # 超时
            ),
            Session(
                id="sess_2",
                template_id="python-basic",
                status=SessionStatus.CREATING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/sess_2",
                runtime_type="docker",
                created_at=recent_time,  # 未超时
            ),
        ]

        result = await service.check_and_mark_stuck_sessions()

        assert result["total_checked"] == 2
        assert result["marked_failed"] == 1

    @pytest.mark.asyncio
    async def test_no_creating_sessions(self, service, session_repo):
        """测试没有 creating 会话时的情况"""
        session_repo.find_by_status.return_value = []

        result = await service.check_and_mark_stuck_sessions()

        assert result["total_checked"] == 0
        assert result["marked_failed"] == 0

    @pytest.mark.asyncio
    async def test_custom_timeout_threshold(self, session_repo):
        """测试自定义超时阈值"""
        service = SessionStuckCreatingService(
            session_repo=session_repo,
            creating_timeout_seconds=60,  # 1 分钟
        )

        # 创建 2 分钟前创建的会话（超过自定义阈值）
        old_time = datetime.now() - timedelta(minutes=2)
        stuck_session = Session(
            id="sess_stuck",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_stuck",
            runtime_type="docker",
            created_at=old_time,
        )
        session_repo.find_by_status.return_value = [stuck_session]

        result = await service.check_and_mark_stuck_sessions()

        assert result["marked_failed"] == 1
        assert stuck_session.status == SessionStatus.FAILED

    @pytest.mark.asyncio
    async def test_error_handling(self, service, session_repo):
        """测试错误处理"""
        session_repo.find_by_status.side_effect = Exception("Database error")

        result = await service.check_and_mark_stuck_sessions()

        assert "errors" in result
        assert len(result["errors"]) > 0
        assert any("Database error" in str(e) for e in result["errors"])

    @pytest.mark.asyncio
    async def test_session_without_created_at(self, service, session_repo):
        """测试没有 created_at 的会话（边界情况）"""
        session = Session(
            id="sess_no_created_at",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_no_created_at",
            runtime_type="docker",
            created_at=None,  # 没有 created_at
        )
        session_repo.find_by_status.return_value = [session]

        result = await service.check_and_mark_stuck_sessions()

        # 没有 created_at 的会话不应被标记为失败（无法判断是否超时）
        assert result["marked_failed"] == 0
        assert session.status == SessionStatus.CREATING

    @pytest.mark.asyncio
    async def test_exactly_at_threshold(self, service, session_repo):
        """测试恰好等于阈值的情况"""
        # 创建恰好 5 分钟前的会话
        exact_threshold_time = datetime.now() - timedelta(seconds=300)
        session = Session(
            id="sess_exact",
            template_id="python-basic",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_exact",
            runtime_type="docker",
            created_at=exact_threshold_time,
        )
        session_repo.find_by_status.return_value = [session]

        result = await service.check_and_mark_stuck_sessions()

        # 恰好等于阈值应被标记为失败（因为检查时间略有延迟）
        assert result["marked_failed"] == 1
