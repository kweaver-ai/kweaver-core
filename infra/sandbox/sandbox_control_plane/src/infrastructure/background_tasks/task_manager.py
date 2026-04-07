"""
后台任务管理器

管理周期性后台任务的启动和停止，支持优雅关闭。
"""
import asyncio
import logging
from contextlib import asynccontextmanager
from typing import Callable, List, Optional

logger = logging.getLogger(__name__)


class BackgroundTask:
    """
    后台任务

    表示一个周期性运行的后台任务。
    """

    def __init__(
        self,
        name: str,
        func: Callable,
        interval_seconds: int,
        initial_delay_seconds: int = 0,
    ):
        """
        初始化后台任务

        Args:
            name: 任务名称
            func: 异步函数，任务的实际执行逻辑
            interval_seconds: 执行间隔（秒）
            initial_delay_seconds: 首次执行前的延迟（秒）
        """
        self.name = name
        self.func = func
        self.interval_seconds = interval_seconds
        self.initial_delay_seconds = initial_delay_seconds
        self._task: Optional[asyncio.Task] = None
        self._stop_event = asyncio.Event()
        self._running = False

    async def start(self) -> None:
        """
        启动后台任务

        如果任务已在运行，则不执行任何操作。
        """
        if self._running:
            logger.warning(f"Task {self.name} is already running")
            return

        self._stop_event.clear()
        self._running = True
        self._task = asyncio.create_task(self._run())
        logger.info(f"Started background task: {self.name}")

    async def stop(self) -> None:
        """
        停止后台任务

        等待任务完成当前执行后停止，最多等待 30 秒。
        """
        if not self._running:
            return

        self._stop_event.set()
        self._running = False

        try:
            await asyncio.wait_for(self._task, timeout=30)
            logger.info(f"Stopped background task: {self.name}")
        except asyncio.TimeoutError:
            logger.warning(f"Task {self.name} did not stop gracefully, cancelling")
            self._task.cancel()

    async def _run(self) -> None:
        """
        任务运行循环

        执行流程：
        1. 等待初始延迟（如果配置）
        2. 循环执行：
           - 执行任务函数
           - 等待间隔或停止事件
        """
        try:
            # 初始延迟
            if self.initial_delay_seconds > 0:
                await asyncio.sleep(self.initial_delay_seconds)

            # 任务循环
            while not self._stop_event.is_set():
                try:
                    # 执行任务函数
                    await self.func()
                except Exception as e:
                    logger.error(
                        f"Error in background task {self.name}: {e}",
                        exc_info=True,
                    )

                # 等待间隔或停止事件
                try:
                    await asyncio.wait_for(
                        self._stop_event.wait(),
                        timeout=self.interval_seconds,
                    )
                    # 如果 wait_for 完成，说明停止事件被设置
                    break
                except asyncio.TimeoutError:
                    # 超时，继续循环
                    continue

        except asyncio.CancelledError:
            logger.info(f"Background task {self.name} was cancelled")
        except Exception as e:
            logger.error(
                f"Unexpected error in background task {self.name}: {e}",
                exc_info=True,
            )

    @property
    def is_running(self) -> bool:
        """检查任务是否正在运行"""
        return self._running and self._task is not None and not self._task.done()


class BackgroundTaskManager:
    """
    后台任务管理器

    管理多个后台任务的启动和停止。
    """

    def __init__(self):
        self._tasks: List[BackgroundTask] = []
        self._running = False

    def register_task(
        self,
        name: str,
        func: Callable,
        interval_seconds: int,
        initial_delay_seconds: int = 0,
    ) -> None:
        """
        注册一个新的后台任务

        Args:
            name: 任务名称
            func: 异步函数，任务的实际执行逻辑
            interval_seconds: 执行间隔（秒）
            initial_delay_seconds: 首次执行前的延迟（秒）
        """
        task = BackgroundTask(
            name=name,
            func=func,
            interval_seconds=interval_seconds,
            initial_delay_seconds=initial_delay_seconds,
        )
        self._tasks.append(task)
        logger.info(
            f"Registered background task: {name} "
            f"(interval: {interval_seconds}s, delay: {initial_delay_seconds}s)"
        )

    async def start_all(self) -> None:
        """
        启动所有已注册的后台任务

        如果任务已在运行，则不执行任何操作。
        """
        if self._running:
            logger.warning("Background tasks already running")
            return

        self._running = True

        for task in self._tasks:
            await task.start()

        logger.info(f"Started {len(self._tasks)} background tasks")

    async def stop_all(self) -> None:
        """
        停止所有正在运行的后台任务

        等待所有任务完成当前执行后停止，最多等待 30 秒。
        """
        if not self._running:
            return

        self._running = False

        # 并行停止所有任务
        tasks_to_stop = [task.stop() for task in self._tasks]
        await asyncio.gather(*tasks_to_stop, return_exceptions=True)

        logger.info("Stopped all background tasks")

    @asynccontextmanager
    async def lifecycle(self):
        """
        任务生命周期上下文管理器

        用法：
            async with task_manager.lifecycle():
                # 任务正在运行
                pass
            # 任务已停止
        """
        await self.start_all()
        try:
            yield
        finally:
            await self.stop_all()

    @property
    def running(self) -> bool:
        """检查管理器是否正在运行"""
        return self._running

    @property
    def task_count(self) -> int:
        """获取已注册的任务数量"""
        return len(self._tasks)

    def get_task_status(self) -> dict:
        """
        获取所有任务的状态

        Returns:
            dict: 任务名称到运行状态的映射
        """
        return {task.name: task.is_running for task in self._tasks}
