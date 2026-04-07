"""
会话实体单元测试

测试 Session 实体的领域行为。
"""
import pytest
from datetime import datetime, timedelta
from zoneinfo import ZoneInfo

from src.domain.entities.session import Session
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus


class TestSession:
    """会话实体测试"""

    def test_create_session(self):
        """测试创建会话"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        assert session.id == "sess_20240115_abc123"
        assert session.template_id == "python-datascience"
        assert session.status == SessionStatus.CREATING
        assert session.is_active() is True

    def test_mark_as_running(self):
        """测试标记会话为运行中"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        session.mark_as_running(
            runtime_node="node-1",
            container_id="container-123"
        )

        assert session.status == SessionStatus.RUNNING
        assert session.runtime_node == "node-1"
        assert session.container_id == "container-123"

    def test_mark_as_running_invalid_transition(self):
        """测试无效状态转换"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        with pytest.raises(ValueError, match="Cannot mark session as running"):
            session.mark_as_running("node-1", "container-123")

    def test_mark_as_terminated(self):
        """测试终止会话"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        session.mark_as_terminated()

        assert session.is_terminated() is True
        assert session.is_active() is False

    def test_is_idle(self):
        """测试空闲检查"""
        # 创建一个超过 30 分钟未活动的会话
        old_time = datetime.now() - timedelta(minutes=35)

        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
            last_activity_at=old_time,
        )

        assert session.is_idle(threshold_minutes=30) is True

    def test_should_cleanup(self):
        """测试是否应该清理"""
        # 空闲会话
        old_time = datetime.now() - timedelta(minutes=35)

        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
            last_activity_at=old_time,
        )

        assert session.should_cleanup() is True

    def test_add_execution(self):
        """测试添加执行记录"""
        from src.domain.entities.execution import Execution
        from src.domain.value_objects.execution_status import ExecutionStatus, ExecutionState

        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        execution = Execution(
            id="exec_20240115_xyz789",
            session_id=session.id,
            code="print('hello')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.PENDING),
        )

        session.add_execution(execution)

        assert len(session.get_executions()) == 1
        assert session.get_executions()[0].id == "exec_20240115_xyz789"

    def test_add_execution_wrong_session(self):
        """测试添加错误会话的执行记录"""
        from src.domain.entities.execution import Execution
        from src.domain.value_objects.execution_status import ExecutionStatus, ExecutionState

        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        execution = Execution(
            id="exec_20240115_xyz789",
            session_id="different_session_id",  # 错误的会话 ID
            code="print('hello')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.PENDING),
        )

        with pytest.raises(ValueError, match="does not belong to this session"):
            session.add_execution(execution)

    def test_invalid_timeout(self):
        """测试无效的超时值"""
        with pytest.raises(ValueError, match="timeout must be positive"):
            Session(
                id="sess_20240115_abc123",
                template_id="python-datascience",
                status=SessionStatus.CREATING,
                resource_limit=ResourceLimit.default(),
                workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
                runtime_type="docker",
                timeout=-1,  # 无效值
            )

    def test_invalid_workspace_path(self):
        """测试无效的 workspace 路径"""
        with pytest.raises(ValueError, match="workspace_path cannot be empty"):
            Session(
                id="sess_20240115_abc123",
                template_id="python-datascience",
                status=SessionStatus.CREATING,
                resource_limit=ResourceLimit.default(),
                workspace_path="",  # 无效值
                runtime_type="docker",
            )

    def test_mark_as_completed(self):
        """测试标记会话为已完成"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        session.mark_as_completed()

        assert session.status == SessionStatus.COMPLETED
        assert session.completed_at is not None

    def test_mark_as_completed_invalid_transition(self):
        """测试标记完成时无效状态转换"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        with pytest.raises(ValueError, match="Cannot mark session as completed"):
            session.mark_as_completed()

    def test_mark_as_failed(self):
        """测试标记会话为失败"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        session.mark_as_failed()

        assert session.status == SessionStatus.FAILED
        assert session.completed_at is not None

    def test_mark_as_failed_from_creating(self):
        """测试从 CREATING 状态标记为失败"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        session.mark_as_failed()

        assert session.status == SessionStatus.FAILED

    def test_mark_as_failed_invalid_transition(self):
        """测试标记失败时无效状态转换"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        with pytest.raises(ValueError, match="Cannot mark session as failed"):
            session.mark_as_failed()

    def test_mark_as_terminated_idempotent(self):
        """测试终止会话幂等性"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        # 再次终止应该不报错
        session.mark_as_terminated()

        assert session.status == SessionStatus.TERMINATED

    def test_is_expired(self):
        """测试过期检查"""
        old_time = datetime.now() - timedelta(hours=7)

        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
            created_at=old_time,
        )

        assert session.is_expired(max_hours=6) is True

    def test_is_not_expired(self):
        """测试未过期"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        assert session.is_expired(max_hours=6) is False

    def test_is_idle_not_active(self):
        """测试非活跃会话不判断空闲"""
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.TERMINATED,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        assert session.is_idle() is False

    def test_update_last_activity(self):
        """测试更新最后活动时间"""
        old_time = datetime.now() - timedelta(minutes=10)
        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
            last_activity_at=old_time,
        )

        session.update_last_activity()

        assert session.last_activity_at > old_time

    def test_get_running_executions(self):
        """测试获取运行中的执行"""
        from src.domain.entities.execution import Execution
        from src.domain.value_objects.execution_status import ExecutionStatus, ExecutionState

        session = Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

        running_execution = Execution(
            id="exec_1",
            session_id=session.id,
            code="print('hello')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING),
        )

        completed_execution = Execution(
            id="exec_2",
            session_id=session.id,
            code="print('world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.COMPLETED),
        )

        session.add_execution(running_execution)
        session.add_execution(completed_execution)

        running = session.get_running_executions()

        assert len(running) == 1
        assert running[0].id == "exec_1"


class TestInstalledDependency:
    """已安装依赖测试"""

    def test_create_installed_dependency(self):
        """测试创建已安装依赖"""
        from src.domain.entities.session import InstalledDependency

        dep = InstalledDependency(
            name="requests",
            version="2.31.0",
            install_location="/workspace/.venv/",
            install_time=datetime.now(),
            is_from_template=False
        )

        assert dep.name == "requests"
        assert dep.version == "2.31.0"
        assert dep.install_location == "/workspace/.venv/"
        assert dep.is_from_template is False


class TestSessionDependencies:
    """会话依赖管理测试"""

    @pytest.fixture
    def session(self):
        """创建测试会话"""
        return Session(
            id="sess_20240115_abc123",
            template_id="python-datascience",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit.default(),
            workspace_path="s3://sandbox-workspace/sessions/sess_20240115_abc123",
            runtime_type="docker",
        )

    def test_set_dependencies_installing(self, session):
        """测试标记依赖安装中"""
        session.set_dependencies_installing()

        assert session.dependency_install_status == "installing"

    def test_set_dependencies_completed(self, session):
        """测试标记依赖安装完成"""
        from src.domain.entities.session import InstalledDependency

        installed = [
            InstalledDependency(
                name="requests",
                version="2.31.0",
                install_location="/opt/sandbox-venv",
                install_time=datetime.now(),
                is_from_template=False
            )
        ]

        session.set_dependencies_completed(installed)

        assert session.dependency_install_status == "completed"
        assert len(session.installed_dependencies) == 1

    def test_set_dependencies_failed(self, session):
        """测试标记依赖安装失败"""
        session.set_dependencies_failed("pip install failed")

        assert session.dependency_install_status == "failed"
        assert session.dependency_install_error == "pip install failed"

    def test_has_dependencies_true(self, session):
        """测试有依赖需要安装"""
        session.requested_dependencies = ["requests", "pandas"]

        assert session.has_dependencies() is True

    def test_has_dependencies_false(self, session):
        """测试没有依赖需要安装"""
        assert session.has_dependencies() is False

    def test_is_dependency_install_pending(self, session):
        """测试依赖是否待安装"""
        assert session.is_dependency_install_pending() is True  # default "pending"

        session.set_dependencies_installing()
        assert session.is_dependency_install_pending() is True

        session.set_dependencies_completed([])
        assert session.is_dependency_install_pending() is False

    def test_is_dependency_install_successful(self, session):
        """测试依赖是否安装成功"""
        session.set_dependencies_completed([])

        assert session.is_dependency_install_successful() is True

    def test_is_dependency_install_successful_false(self, session):
        """测试依赖安装失败"""
        session.set_dependencies_failed("error")

        assert session.is_dependency_install_successful() is False
