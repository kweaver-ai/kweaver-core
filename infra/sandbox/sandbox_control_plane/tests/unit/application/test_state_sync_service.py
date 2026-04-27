"""
状态同步服务单元测试

测试 StateSyncService 的状态同步逻辑。
"""
import pytest
from unittest.mock import Mock, AsyncMock
from datetime import datetime, timedelta

from src.application.services.state_sync_service import StateSyncService
from src.domain.entities.execution import Execution
from src.domain.entities.session import Session
from src.domain.entities.template import Template
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus, ExecutionStatus, ExecutionState
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
        repo.find_sessions = AsyncMock(side_effect=[[], []])
        repo.count_sessions = AsyncMock(return_value=0)
        return repo

    @pytest.fixture
    def container_scheduler(self):
        """模拟容器调度器"""
        scheduler = Mock()
        scheduler.is_container_running = AsyncMock()
        scheduler.create_container = AsyncMock()
        scheduler.start_container = AsyncMock()
        scheduler.remove_container = AsyncMock()
        scheduler.get_container_ownership = AsyncMock()
        return scheduler

    @pytest.fixture
    def template_repo(self):
        """模拟模板仓储"""
        repo = Mock()
        repo.find_by_id = AsyncMock()
        return repo

    @pytest.fixture
    def execution_repo(self):
        """模拟执行仓储"""
        repo = Mock()
        repo.find_by_session_id = AsyncMock(return_value=[])
        repo.save = AsyncMock()
        repo.commit = AsyncMock()
        return repo

    @pytest.fixture
    def service(self, session_repo, container_scheduler, template_repo, execution_repo):
        """创建状态同步服务"""
        return StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
        )

    @pytest.fixture
    def python_template(self):
        """创建测试模板"""
        return Template(
            id="python-basic",
            name="Python Basic",
            image="registry.example.com/custom-python:v2",
            base_image="registry.example.com/custom-python:v2",
            pre_installed_packages=[],
            default_resources=ResourceLimit.default(),
            created_at=datetime.now(),
            updated_at=datetime.now(),
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
    async def test_sync_on_startup_k8s_takeover_recreates_missing_pod(
        self,
        session_repo,
        container_scheduler,
        template_repo,
        execution_repo,
        python_template,
    ):
        """测试 K8s 启动接管时缺失 Pod 会直接重建，并标记中断执行。"""
        service = StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=Mock(
                _cluster_node=Mock(type="kubernetes", id="k8s-cluster"),
                _owner_context=Mock(pod_name="cp-new", pod_uid="uid-new"),
                create_container_for_session=AsyncMock(return_value="new-pod"),
            ),
        )
        session = Session(
            id="sess_missing_pod",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_missing_pod",
            runtime_type="docker",
            container_id="old-pod",
            env_vars={"SESSION_ID": "sess_missing_pod"},
        )
        in_flight_execution = Execution(
            id="exec-missing-pod",
            session_id=session.id,
            code="print('hello')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING),
        )
        session_repo.find_by_status.side_effect = [[session], []]
        container_scheduler.get_container_ownership = AsyncMock(return_value=None)
        container_scheduler.start_container = AsyncMock()
        execution_repo.find_by_session_id.return_value = [in_flight_execution]
        template_repo.find_by_id.return_value = python_template

        result = await service.sync_on_startup()

        assert result["recovered"] == 1
        assert result["takeover_recreated"] == 1
        assert result["interrupted_executions"] == 1
        assert session.container_id == "new-pod"
        assert in_flight_execution.state.status == ExecutionStatus.FAILED
        assert in_flight_execution.state.exit_code == 137

    @pytest.mark.asyncio
    async def test_sync_on_startup_k8s_takeover_keeps_healthy_current_owner(
        self,
        session_repo,
        container_scheduler,
        template_repo,
        execution_repo,
    ):
        """测试当前 owner 且健康的 Pod 会被保留。"""
        service = StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=Mock(
                _cluster_node=Mock(type="kubernetes", id="k8s-cluster"),
                _owner_context=Mock(pod_name="cp-new", pod_uid="uid-new"),
            ),
        )
        session = Session(
            id="sess_keep",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_keep",
            runtime_type="docker",
            container_id="pod-keep",
        )
        session_repo.find_by_status.side_effect = [[session], []]
        container_scheduler.get_container_ownership = AsyncMock(
            return_value=Mock(
                owner_pod_name="cp-new",
                owner_pod_uid="uid-new",
                annotations={"control-plane-pod-name": "cp-new"},
                has_owner_reference=True,
            )
        )
        container_scheduler.is_container_running.return_value = True

        result = await service.sync_on_startup()

        assert result["healthy"] == 1
        assert result["takeover_skipped"] == 1
        container_scheduler.remove_container.assert_not_called()

    @pytest.mark.asyncio
    async def test_sync_on_startup_k8s_takeover_recreates_unhealthy_current_owner_and_marks_execution_failed(
        self,
        session_repo,
        container_scheduler,
        template_repo,
        execution_repo,
        python_template,
    ):
        """测试当前 owner 但不健康的 Pod 会先删除再重建，并标记中断执行。"""
        scheduler = Mock(
            _cluster_node=Mock(type="kubernetes", id="k8s-cluster"),
            _owner_context=Mock(pod_name="cp-new", pod_uid="uid-new"),
            create_container_for_session=AsyncMock(return_value="pod-recreated-current-owner"),
        )
        service = StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=scheduler,
        )
        session = Session(
            id="sess_current_owner_unhealthy",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_current_owner_unhealthy",
            runtime_type="docker",
            container_id="pod-current-owner-unhealthy",
        )
        in_flight_execution = Execution(
            id="exec-current-owner-unhealthy",
            session_id=session.id,
            code="print('hello')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING),
        )
        session_repo.find_by_status.side_effect = [[session], []]
        container_scheduler.get_container_ownership = AsyncMock(
            return_value=Mock(
                owner_pod_name="cp-new",
                owner_pod_uid="uid-new",
                annotations={"control-plane-pod-name": "cp-new"},
                has_owner_reference=True,
            )
        )
        container_scheduler.is_container_running.return_value = False
        container_scheduler.remove_container = AsyncMock()
        container_scheduler.start_container = AsyncMock()
        execution_repo.find_by_session_id.return_value = [in_flight_execution]
        template_repo.find_by_id.return_value = python_template

        result = await service.sync_on_startup()

        assert result["takeover_recreated"] == 1
        assert result["interrupted_executions"] == 1
        assert result["unhealthy"] == 1
        assert session.container_id == "pod-recreated-current-owner"
        assert in_flight_execution.state.status == ExecutionStatus.FAILED
        assert in_flight_execution.state.exit_code == 137
        container_scheduler.remove_container.assert_awaited_once_with("pod-current-owner-unhealthy", force=True)

    @pytest.mark.asyncio
    async def test_sync_on_startup_k8s_takeover_recreates_when_only_annotations_match_current_owner(
        self,
        session_repo,
        container_scheduler,
        template_repo,
        execution_repo,
        python_template,
    ):
        """测试仅 annotation 匹配而无 Pod ownerReference 时仍会接管重建。"""
        scheduler = Mock(
            _cluster_node=Mock(type="kubernetes", id="k8s-cluster"),
            _owner_context=Mock(pod_name="cp-new", pod_uid="uid-new"),
            create_container_for_session=AsyncMock(return_value="pod-recreated-from-annotation-only"),
        )
        service = StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=scheduler,
        )
        session = Session(
            id="sess_annotation_only",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_annotation_only",
            runtime_type="docker",
            container_id="pod-annotation-only",
        )
        session_repo.find_by_status.side_effect = [[session], []]
        container_scheduler.get_container_ownership = AsyncMock(
            return_value=Mock(
                owner_pod_name="cp-new",
                owner_pod_uid="uid-new",
                annotations={"control-plane-pod-uid": "uid-new"},
                has_owner_reference=False,
            )
        )
        container_scheduler.remove_container = AsyncMock()
        container_scheduler.start_container = AsyncMock()
        template_repo.find_by_id.return_value = python_template

        result = await service.sync_on_startup()

        assert result["takeover_recreated"] == 1
        assert result["takeover_skipped"] == 0
        assert session.container_id == "pod-recreated-from-annotation-only"
        container_scheduler.remove_container.assert_awaited_once_with("pod-annotation-only", force=True)

    @pytest.mark.asyncio
    async def test_sync_on_startup_k8s_takeover_recreates_old_owner_and_marks_execution_failed(
        self,
        session_repo,
        container_scheduler,
        template_repo,
        execution_repo,
        python_template,
    ):
        """测试 takeover 删除旧 Pod 并将未完成执行标记为失败。"""
        scheduler = Mock(
            _cluster_node=Mock(type="kubernetes", id="k8s-cluster"),
            _owner_context=Mock(pod_name="cp-new", pod_uid="uid-new"),
            create_container_for_session=AsyncMock(return_value="pod-recreated"),
        )
        service = StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
            scheduler=scheduler,
        )
        session = Session(
            id="sess_takeover",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_takeover",
            runtime_type="docker",
            container_id="old-pod",
            env_vars={"SESSION_ID": "sess_takeover"},
        )
        in_flight_execution = Execution(
            id="exec-1",
            session_id=session.id,
            code="print('hello')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING),
        )
        session_repo.find_by_status.side_effect = [[session], []]
        container_scheduler.get_container_ownership = AsyncMock(
            return_value=Mock(
                owner_pod_name="cp-old",
                owner_pod_uid="uid-old",
                annotations={"control-plane-pod-uid": "uid-old"},
                has_owner_reference=True,
            )
        )
        container_scheduler.remove_container = AsyncMock()
        container_scheduler.start_container = AsyncMock()
        execution_repo.find_by_session_id.return_value = [in_flight_execution]
        template_repo.find_by_id.return_value = python_template

        result = await service.sync_on_startup()

        assert result["takeover_recreated"] == 1
        assert result["interrupted_executions"] == 1
        assert session.status == SessionStatus.RUNNING
        assert session.container_id == "pod-recreated"
        assert in_flight_execution.state.status == ExecutionStatus.FAILED
        assert in_flight_execution.state.exit_code == 137
        container_scheduler.remove_container.assert_awaited_once_with("old-pod", force=True)
        execution_repo.save.assert_awaited_once_with(in_flight_execution)

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
    async def test_recovery_success(
        self,
        service,
        session_repo,
        container_scheduler,
        template_repo,
        python_template,
    ):
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
        template_repo.find_by_id.return_value = python_template

        # 不传入 scheduler 参数，恢复功能会尝试创建新容器
        result = await service._attempt_recovery(session)

        assert result is True
        assert session.container_id == "new-container"
        assert session.status == SessionStatus.RUNNING
        container_scheduler.create_container.assert_awaited_once()
        config = container_scheduler.create_container.await_args.args[0]
        assert config.image == python_template.image

    @pytest.mark.asyncio
    async def test_recovery_success_refreshes_last_activity(
        self,
        service,
        session_repo,
        container_scheduler,
        template_repo,
        python_template,
    ):
        """测试恢复成功后会刷新 last_activity_at，避免被空闲清理立即回收。"""
        stale_last_activity = datetime.now() - timedelta(minutes=45)
        session = Session(
            id="sess_recovery_refresh",
            template_id="python-basic",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_recovery_refresh",
            runtime_type="docker",
            container_id="old-container",
            env_vars={"SESSION_ID": "sess_recovery_refresh"},
            last_activity_at=stale_last_activity,
        )

        container_scheduler.create_container.return_value = "new-container"
        template_repo.find_by_id.return_value = python_template

        result = await service._attempt_recovery(session)

        assert result is True
        assert session.container_id == "new-container"
        assert session.last_activity_at > stale_last_activity

    @pytest.mark.asyncio
    async def test_recovery_failure_marks_failed(
        self,
        service,
        session_repo,
        container_scheduler,
        template_repo,
        python_template,
    ):
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
        template_repo.find_by_id.return_value = python_template

        result = await service._attempt_recovery(session)

        assert result is False
        assert session.status == SessionStatus.FAILED

    @pytest.mark.asyncio
    async def test_recovery_fails_when_template_missing(
        self,
        service,
        container_scheduler,
        template_repo,
    ):
        """测试恢复时模板不存在会标记失败且不创建容器"""
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

        template_repo.find_by_id.return_value = None

        result = await service._attempt_recovery(session)

        assert result is False
        assert session.status == SessionStatus.FAILED
        container_scheduler.create_container.assert_not_awaited()

    @pytest.mark.asyncio
    async def test_sync_on_startup_loads_all_active_sessions_via_pagination(
        self,
        container_scheduler,
        template_repo,
        execution_repo,
    ):
        """测试启动同步会分页加载全部 active sessions，而不是受默认 limit 截断。"""
        class PaginatedSessionRepo:
            def __init__(self, pages):
                self.pages = pages
                self.calls = []

            async def find_sessions(self, status=None, template_id=None, limit=50, offset=0):
                self.calls.append((status, limit, offset))
                return self.pages.get((status, offset), [])

        running_page_1 = [
            Session(
                id=f"running-{index}",
                template_id="python-basic",
                status=SessionStatus.RUNNING,
                resource_limit=ResourceLimit.default(),
                workspace_path=f"s3://sandbox-workspace/sessions/running-{index}",
                runtime_type="docker",
                container_id=f"container-running-{index}",
            )
            for index in range(200)
        ]
        running_page_2 = [
            Session(
                id="running-200",
                template_id="python-basic",
                status=SessionStatus.RUNNING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/running-200",
                runtime_type="docker",
                container_id="container-running-200",
            )
        ]
        creating_page_1 = [
            Session(
                id="creating-0",
                template_id="python-basic",
                status=SessionStatus.CREATING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/creating-0",
                runtime_type="docker",
                container_id="container-creating-0",
            )
        ]
        session_repo = PaginatedSessionRepo(
            {
                (SessionStatus.RUNNING.value, 0): running_page_1,
                (SessionStatus.RUNNING.value, 200): running_page_2,
                (SessionStatus.RUNNING.value, 400): [],
                (SessionStatus.CREATING.value, 0): creating_page_1,
                (SessionStatus.CREATING.value, 200): [],
            }
        )
        service = StateSyncService(
            session_repo=session_repo,
            container_scheduler=container_scheduler,
            execution_repo=execution_repo,
            template_repo=template_repo,
        )
        container_scheduler.is_container_running.return_value = True

        result = await service.sync_on_startup()

        assert result["total"] == 202
        assert result["healthy"] == 202
        assert session_repo.calls == [
            (SessionStatus.RUNNING.value, 200, 0),
            (SessionStatus.RUNNING.value, 200, 200),
            (SessionStatus.CREATING.value, 200, 0),
        ]

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
