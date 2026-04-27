"""
后台任务管理

提供后台任务的启动、停止和生命周期管理。
"""
from src.infrastructure.background_tasks.task_manager import (
    BackgroundTask,
    BackgroundTaskManager,
)

__all__ = [
    "BackgroundTask",
    "BackgroundTaskManager",
]
