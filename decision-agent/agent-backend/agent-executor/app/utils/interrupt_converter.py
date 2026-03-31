# -*- coding:utf-8 -*-
"""中断信息类型转换工具"""

from dolphin.core.coroutine.resume_handle import ResumeHandle
from app.domain.vo.interrupt.interrupt_handle import InterruptHandle


def interrupt_handle_to_resume_handle(
    interrupt_handle: InterruptHandle,
) -> ResumeHandle:
    """将 Pydantic 的 InterruptHandle 转换为 Dolphin SDK 的 ResumeHandle

    用于从 API 层转换到内部逻辑层（请求处理）

    Args:
        interrupt_handle: Pydantic 的 InterruptHandle 对象

    Returns:
        ResumeHandle: Dolphin SDK 的 ResumeHandle 对象
    """
    return ResumeHandle(
        frame_id=interrupt_handle.frame_id,
        snapshot_id=interrupt_handle.snapshot_id,
        resume_token=interrupt_handle.resume_token,
        interrupt_type=interrupt_handle.interrupt_type,
        current_block=interrupt_handle.current_block,
        restart_block=interrupt_handle.restart_block,
    )
