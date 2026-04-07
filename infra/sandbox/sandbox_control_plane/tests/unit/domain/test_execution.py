"""
执行实体单元测试

测试 Execution 实体的领域行为。
"""
import pytest
from datetime import datetime, timedelta

from src.domain.entities.execution import Execution
from src.domain.value_objects.execution_status import ExecutionStatus, ExecutionState
from src.domain.value_objects.artifact import Artifact, ArtifactType


class TestExecution:
    """执行实体测试"""

    def test_create_execution(self):
        """测试创建执行"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.PENDING)
        )

        assert execution.id == "exec_20240115_abc123"
        assert execution.session_id == "sess_20240115_xyz789"
        assert execution.state.status == ExecutionStatus.PENDING
        assert execution.is_running() is False

    def test_mark_running(self):
        """测试标记为运行中"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.PENDING)
        )

        execution.mark_running()

        assert execution.is_running() is True
        assert execution.last_heartbeat_at is not None

    def test_mark_completed(self):
        """测试标记为已完成"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING)
        )

        artifacts = [
            Artifact.create(
                path="output.txt",
                size=100,
                mime_type="text/plain"
            )
        ]

        execution.mark_completed(
            stdout="hello world\n",
            stderr="",
            exit_code=0,
            execution_time=0.5,
            artifacts=artifacts
        )

        assert execution.state.status == ExecutionStatus.COMPLETED
        assert execution.stdout == "hello world\n"
        assert execution.execution_time == 0.5
        assert len(execution.artifacts) == 1

    def test_mark_failed(self):
        """测试标记为失败"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING)
        )

        execution.mark_failed(
            error_message="SyntaxError: invalid syntax",
            exit_code=1
        )

        assert execution.state.status == ExecutionStatus.FAILED
        assert execution.state.error_message == "SyntaxError: invalid syntax"
        assert execution.state.exit_code == 1

    def test_mark_timeout(self):
        """测试标记为超时"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="while True: pass",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING)
        )

        execution.mark_timeout()

        assert execution.state.status == ExecutionStatus.TIMEOUT

    def test_mark_crashed(self):
        """测试标记为崩溃"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING)
        )

        execution.mark_crashed()

        assert execution.state.status == ExecutionStatus.CRASHED
        assert execution.can_retry() is True

    def test_increment_retry_count(self):
        """测试增加重试计数"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.CRASHED),
            retry_count=0
        )

        execution.increment_retry_count()

        assert execution.retry_count == 1

    def test_can_retry(self):
        """测试是否可以重试"""
        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.CRASHED),
            retry_count=2
        )

        assert execution.can_retry(max_retries=3) is True
        assert execution.can_retry(max_retries=2) is False

    def test_is_heartbeat_timeout(self):
        """测试心跳超时检查"""
        import time

        execution = Execution(
            id="exec_20240115_abc123",
            session_id="sess_20240115_xyz789",
            code="print('hello world')",
            language="python",
            state=ExecutionState(status=ExecutionStatus.RUNNING),
            last_heartbeat_at=datetime.now() - timedelta(seconds=20)
        )

        assert execution.is_heartbeat_timeout(timeout_seconds=15) is True
        assert execution.is_heartbeat_timeout(timeout_seconds=30) is False

    def test_invalid_empty_code(self):
        """测试空代码"""
        with pytest.raises(ValueError, match="code cannot be empty"):
            Execution(
                id="exec_20240115_abc123",
                session_id="sess_20240115_xyz789",
                code="",  # 无效值
                language="python",
                state=ExecutionState(status=ExecutionStatus.PENDING)
            )

    def test_invalid_empty_language(self):
        """测试空语言"""
        with pytest.raises(ValueError, match="language cannot be empty"):
            Execution(
                id="exec_20240115_abc123",
                session_id="sess_20240115_xyz789",
                code="print('hello')",
                language="",  # 无效值
                state=ExecutionState(status=ExecutionStatus.PENDING)
            )
