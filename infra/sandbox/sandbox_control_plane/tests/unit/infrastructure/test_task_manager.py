"""
后台任务管理器单元测试

测试 BackgroundTask 和 BackgroundTaskManager 的功能。
"""
import pytest
import asyncio
from unittest.mock import Mock, AsyncMock, patch

from src.infrastructure.background_tasks.task_manager import (
    BackgroundTask,
    BackgroundTaskManager,
)


class TestBackgroundTask:
    """后台任务测试"""

    @pytest.fixture
    def task_func(self):
        """创建模拟任务函数"""
        return AsyncMock()

    @pytest.fixture
    def task(self, task_func):
        """创建后台任务"""
        return BackgroundTask(
            name="test_task",
            func=task_func,
            interval_seconds=1,
            initial_delay_seconds=0,
        )

    def test_init(self, task_func):
        """测试任务初始化"""
        task = BackgroundTask(
            name="test_task",
            func=task_func,
            interval_seconds=5,
            initial_delay_seconds=10,
        )

        assert task.name == "test_task"
        assert task.interval_seconds == 5
        assert task.initial_delay_seconds == 10
        assert task._running is False
        assert task._task is None

    def test_is_running_false_initially(self, task):
        """测试初始状态为未运行"""
        assert task.is_running is False

    @pytest.mark.asyncio
    async def test_start_task(self, task, task_func):
        """测试启动任务"""
        await task.start()

        assert task._running is True
        assert task._task is not None

        # Clean up
        await task.stop()

    @pytest.mark.asyncio
    async def test_start_already_running_task(self, task, task_func):
        """测试启动已在运行的任务"""
        await task.start()

        # Should not create a new task
        original_task = task._task
        await task.start()

        assert task._task is original_task

        # Clean up
        await task.stop()

    @pytest.mark.asyncio
    async def test_stop_task(self, task, task_func):
        """测试停止任务"""
        await task.start()
        await task.stop()

        assert task._running is False

    @pytest.mark.asyncio
    async def test_stop_not_running_task(self, task):
        """测试停止未运行的任务"""
        # Should not raise error
        await task.stop()
        assert task._running is False

    @pytest.mark.asyncio
    async def test_task_executes_func(self, task_func):
        """测试任务执行函数"""
        task = BackgroundTask(
            name="test_task",
            func=task_func,
            interval_seconds=0.1,
            initial_delay_seconds=0,
        )

        await task.start()
        await asyncio.sleep(0.15)  # Wait for at least one execution
        await task.stop()

        task_func.assert_called()

    @pytest.mark.asyncio
    async def test_task_with_initial_delay(self, task_func):
        """测试带初始延迟的任务"""
        task = BackgroundTask(
            name="test_task",
            func=task_func,
            interval_seconds=0.1,
            initial_delay_seconds=0.2,
        )

        await task.start()
        await asyncio.sleep(0.15)  # Before delay finishes

        # Should not have been called yet
        task_func.assert_not_called()

        await asyncio.sleep(0.1)  # After delay
        await task.stop()

        # Now should have been called
        task_func.assert_called()

    @pytest.mark.asyncio
    async def test_task_handles_exception(self, task_func):
        """测试任务处理异常"""
        task_func.side_effect = RuntimeError("Test error")

        task = BackgroundTask(
            name="test_task",
            func=task_func,
            interval_seconds=0.1,
            initial_delay_seconds=0,
        )

        await task.start()
        await asyncio.sleep(0.15)  # Wait for execution
        await task.stop()

        # Task should have been called despite exception
        task_func.assert_called()
        # Task should have stopped gracefully
        assert task._running is False

    @pytest.mark.asyncio
    async def test_is_running_property(self, task, task_func):
        """测试 is_running 属性"""
        assert task.is_running is False

        await task.start()
        assert task.is_running is True

        await task.stop()
        assert task.is_running is False


class TestBackgroundTaskManager:
    """后台任务管理器测试"""

    @pytest.fixture
    def manager(self):
        """创建任务管理器"""
        return BackgroundTaskManager()

    @pytest.fixture
    def task_func(self):
        """创建模拟任务函数"""
        return AsyncMock()

    def test_init(self):
        """测试管理器初始化"""
        manager = BackgroundTaskManager()

        assert manager._tasks == []
        assert manager._running is False

    def test_register_task(self, manager, task_func):
        """测试注册任务"""
        manager.register_task(
            name="test_task",
            func=task_func,
            interval_seconds=5,
            initial_delay_seconds=0,
        )

        assert manager.task_count == 1

    def test_register_multiple_tasks(self, manager, task_func):
        """测试注册多个任务"""
        manager.register_task("task1", task_func, 5)
        manager.register_task("task2", task_func, 10)
        manager.register_task("task3", task_func, 15)

        assert manager.task_count == 3

    def test_task_count(self, manager, task_func):
        """测试任务计数"""
        assert manager.task_count == 0

        manager.register_task("task1", task_func, 5)
        assert manager.task_count == 1

        manager.register_task("task2", task_func, 10)
        assert manager.task_count == 2

    @pytest.mark.asyncio
    async def test_start_all(self, manager, task_func):
        """测试启动所有任务"""
        manager.register_task("task1", task_func, 0.1)
        manager.register_task("task2", task_func, 0.1)

        await manager.start_all()

        assert manager.running is True
        assert manager.task_count == 2

        # Clean up
        await manager.stop_all()

    @pytest.mark.asyncio
    async def test_start_all_already_running(self, manager, task_func):
        """测试在已运行时启动所有任务"""
        manager.register_task("task1", task_func, 0.1)

        await manager.start_all()
        await manager.start_all()  # Second call should be no-op

        assert manager.running is True

        # Clean up
        await manager.stop_all()

    @pytest.mark.asyncio
    async def test_stop_all(self, manager, task_func):
        """测试停止所有任务"""
        manager.register_task("task1", task_func, 0.1)
        manager.register_task("task2", task_func, 0.1)

        await manager.start_all()
        await manager.stop_all()

        assert manager.running is False

    @pytest.mark.asyncio
    async def test_stop_all_not_running(self, manager):
        """测试停止未运行的管理器"""
        # Should not raise error
        await manager.stop_all()
        assert manager.running is False

    @pytest.mark.asyncio
    async def test_lifecycle_context_manager(self, manager, task_func):
        """测试生命周期上下文管理器"""
        manager.register_task("task1", task_func, 0.1)

        async with manager.lifecycle():
            assert manager.running is True

        assert manager.running is False

    @pytest.mark.asyncio
    async def test_lifecycle_context_manager_with_exception(self, manager, task_func):
        """测试生命周期上下文管理器处理异常"""
        manager.register_task("task1", task_func, 0.1)

        with pytest.raises(RuntimeError):
            async with manager.lifecycle():
                assert manager.running is True
                raise RuntimeError("Test error")

        # Should still stop tasks after exception
        assert manager.running is False

    def test_get_task_status(self, manager, task_func):
        """测试获取任务状态"""
        manager.register_task("task1", task_func, 5)
        manager.register_task("task2", task_func, 10)

        status = manager.get_task_status()

        assert "task1" in status
        assert "task2" in status
        assert status["task1"] is False  # Not running initially
        assert status["task2"] is False

    @pytest.mark.asyncio
    async def test_get_task_status_running(self, manager, task_func):
        """测试获取运行中任务状态"""
        manager.register_task("task1", task_func, 0.5)
        manager.register_task("task2", task_func, 0.5)

        await manager.start_all()
        status = manager.get_task_status()

        assert status["task1"] is True
        assert status["task2"] is True

        await manager.stop_all()

    @pytest.mark.asyncio
    async def test_running_property(self, manager, task_func):
        """测试 running 属性"""
        assert manager.running is False

        manager.register_task("task1", task_func, 0.5)
        await manager.start_all()
        assert manager.running is True

        await manager.stop_all()
        assert manager.running is False

    @pytest.mark.asyncio
    async def test_multiple_tasks_execute_independently(self, manager):
        """测试多个任务独立执行"""
        call_counts = {"task1": 0, "task2": 0}

        async def task1_func():
            call_counts["task1"] += 1

        async def task2_func():
            call_counts["task2"] += 1

        manager.register_task("task1", task1_func, 0.05)
        manager.register_task("task2", task2_func, 0.1)

        await manager.start_all()
        await asyncio.sleep(0.2)  # Allow multiple executions
        await manager.stop_all()

        # task1 should have been called more times than task2
        # due to shorter interval
        assert call_counts["task1"] > call_counts["task2"]
