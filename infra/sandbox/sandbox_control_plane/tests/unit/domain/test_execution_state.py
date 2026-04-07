"""
执行状态值对象单元测试

测试 ExecutionState 值对象的行为。
"""
import pytest

from src.domain.value_objects.execution_status import (
    SessionStatus,
    ExecutionStatus,
    ExecutionState
)


class TestSessionStatus:
    """会话状态枚举测试"""

    def test_session_status_values(self):
        """测试会话状态枚举值"""
        assert SessionStatus.CREATING == "creating"
        assert SessionStatus.RUNNING == "running"
        assert SessionStatus.COMPLETED == "completed"
        assert SessionStatus.FAILED == "failed"
        assert SessionStatus.TIMEOUT == "timeout"
        assert SessionStatus.TERMINATED == "terminated"

    def test_session_status_is_string(self):
        """测试会话状态是字符串类型"""
        assert isinstance(SessionStatus.RUNNING, str)


class TestExecutionStatus:
    """执行状态枚举测试"""

    def test_execution_status_values(self):
        """测试执行状态枚举值"""
        assert ExecutionStatus.PENDING == "pending"
        assert ExecutionStatus.RUNNING == "running"
        assert ExecutionStatus.COMPLETED == "completed"
        assert ExecutionStatus.FAILED == "failed"
        assert ExecutionStatus.TIMEOUT == "timeout"
        assert ExecutionStatus.CRASHED == "crashed"

    def test_execution_status_is_string(self):
        """测试执行状态是字符串类型"""
        assert isinstance(ExecutionStatus.RUNNING, str)


class TestExecutionState:
    """执行状态值对象测试"""

    def test_create_pending_state(self):
        """测试创建等待状态"""
        state = ExecutionState(status=ExecutionStatus.PENDING)

        assert state.status == ExecutionStatus.PENDING
        assert state.exit_code is None
        assert state.error_message is None

    def test_create_running_state(self):
        """测试创建运行中状态"""
        state = ExecutionState(status=ExecutionStatus.RUNNING)

        assert state.status == ExecutionStatus.RUNNING
        assert state.is_terminal() is False

    def test_create_completed_state(self):
        """测试创建完成状态"""
        state = ExecutionState(
            status=ExecutionStatus.COMPLETED,
            exit_code=0
        )

        assert state.status == ExecutionStatus.COMPLETED
        assert state.exit_code == 0
        assert state.is_terminal() is True
        assert state.can_retry() is False

    def test_create_failed_state_with_error(self):
        """测试创建失败状态（带错误信息）"""
        state = ExecutionState(
            status=ExecutionStatus.FAILED,
            exit_code=1,
            error_message="Syntax error"
        )

        assert state.status == ExecutionStatus.FAILED
        assert state.exit_code == 1
        assert state.error_message == "Syntax error"
        assert state.is_terminal() is True
        assert state.can_retry() is False

    def test_create_failed_state_without_error(self):
        """测试创建失败状态（不带错误信息）"""
        state = ExecutionState(
            status=ExecutionStatus.FAILED,
            exit_code=1
        )

        assert state.status == ExecutionStatus.FAILED
        assert state.exit_code == 1
        assert state.error_message is None
        assert state.is_terminal() is True

    def test_create_timeout_state(self):
        """测试创建超时状态"""
        state = ExecutionState(
            status=ExecutionStatus.TIMEOUT,
            exit_code=124
        )

        assert state.status == ExecutionStatus.TIMEOUT
        assert state.is_terminal() is True
        assert state.can_retry() is False

    def test_create_crashed_state(self):
        """测试创建崩溃状态"""
        state = ExecutionState(
            status=ExecutionStatus.CRASHED
        )

        assert state.status == ExecutionStatus.CRASHED
        assert state.is_terminal() is False
        assert state.can_retry() is True

    def test_is_terminal_for_pending(self):
        """测试等待状态不是终态"""
        state = ExecutionState(status=ExecutionStatus.PENDING)
        assert state.is_terminal() is False

    def test_is_terminal_for_running(self):
        """测试运行中状态不是终态"""
        state = ExecutionState(status=ExecutionStatus.RUNNING)
        assert state.is_terminal() is False

    def test_is_terminal_for_completed(self):
        """测试完成状态是终态"""
        state = ExecutionState(status=ExecutionStatus.COMPLETED)
        assert state.is_terminal() is True

    def test_is_terminal_for_failed(self):
        """测试失败状态是终态"""
        state = ExecutionState(status=ExecutionStatus.FAILED)
        assert state.is_terminal() is True

    def test_is_terminal_for_timeout(self):
        """测试超时状态是终态"""
        state = ExecutionState(status=ExecutionStatus.TIMEOUT)
        assert state.is_terminal() is True

    def test_can_retry_for_crashed(self):
        """测试崩溃状态可以重试"""
        state = ExecutionState(status=ExecutionStatus.CRASHED)
        assert state.can_retry() is True

    def test_can_retry_for_completed(self):
        """测试完成状态不能重试"""
        state = ExecutionState(status=ExecutionStatus.COMPLETED)
        assert state.can_retry() is False

    def test_can_retry_for_failed(self):
        """测试失败状态不能重试"""
        state = ExecutionState(status=ExecutionStatus.FAILED)
        assert state.can_retry() is False

    def test_can_retry_for_timeout(self):
        """测试超时状态不能重试"""
        state = ExecutionState(status=ExecutionStatus.TIMEOUT)
        assert state.can_retry() is False

    def test_immutability(self):
        """测试不可变性"""
        state = ExecutionState(
            status=ExecutionStatus.PENDING,
            exit_code=0,
            error_message="test"
        )

        # frozen=True 意味着不可修改
        with pytest.raises(Exception):  # FrozenInstanceError
            state.exit_code = 1

    def test_state_equality(self):
        """测试状态相等性"""
        state1 = ExecutionState(
            status=ExecutionStatus.COMPLETED,
            exit_code=0
        )
        state2 = ExecutionState(
            status=ExecutionStatus.COMPLETED,
            exit_code=0
        )
        assert state1 == state2

        state3 = ExecutionState(
            status=ExecutionStatus.COMPLETED,
            exit_code=1
        )
        assert state1 != state3

    def test_all_terminal_statuses(self):
        """测试所有终态状态"""
        terminal_states = [
            ExecutionStatus.COMPLETED,
            ExecutionStatus.FAILED,
            ExecutionStatus.TIMEOUT
        ]

        for status in terminal_states:
            state = ExecutionState(status=status)
            assert state.is_terminal() is True

    def test_non_terminal_statuses(self):
        """测试所有非终态状态"""
        non_terminal_states = [
            ExecutionStatus.PENDING,
            ExecutionStatus.RUNNING,
            ExecutionStatus.CRASHED
        ]

        for status in non_terminal_states:
            state = ExecutionState(status=status)
            assert state.is_terminal() is False
