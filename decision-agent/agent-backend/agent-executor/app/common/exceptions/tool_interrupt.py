# -*- coding:utf-8 -*-
"""工具中断异常"""

from dataclasses import dataclass
from typing import Dict, Any, TYPE_CHECKING

# 使用 TYPE_CHECKING 避免运行时导入
if TYPE_CHECKING:
    from dolphin.core.coroutine.resume_handle import ResumeHandle
else:
    # 运行时使用 Any 类型，避免导入错误
    ResumeHandle = Any


def _get_resume_handle_class():
    """延迟获取 ResumeHandle 类

    Returns:
        ResumeHandle 类或 Mock 类
    """
    try:
        # 尝试从 dolphin 导入
        from dolphin.core.coroutine.resume_handle import ResumeHandle

        return ResumeHandle
    except ImportError:
        # 创建 Mock 类
        class MockResumeHandle:
            def __init__(
                self,
                frame_id="",
                snapshot_id="",
                resume_token="",
                interrupt_type="",
                current_block="",
                restart_block="",
            ):
                self.frame_id = frame_id
                self.snapshot_id = snapshot_id
                self.resume_token = resume_token
                self.interrupt_type = interrupt_type
                self.current_block = current_block
                self.restart_block = restart_block

        return MockResumeHandle


# 获取 ResumeHandle 类（延迟加载）
ResumeHandleClass = _get_resume_handle_class()


@dataclass
class ToolInterruptInfo:
    """工具中断信息

    对应接口文档 tool_interrupt_info.yaml

    Attributes:
        handle: 恢复句柄（ResumeHandle 对象或 Mock 对象）
        data: 中断详情（tool_name, tool_description, tool_args, interrupt_config）
    """

    handle: Any  # 使用 Any 类型以兼容 Mock 对象
    data: Dict[str, Any]  # 中断详情


class ToolInterruptException(Exception):
    """自定义工具中断异常

    用于在识别到 Dolphin SDK 的工具中断后，转换为自己的异常类进行处理。
    """

    def __init__(self, interrupt_info: ToolInterruptInfo):
        self.interrupt_info = interrupt_info
        tool_name = (
            interrupt_info.data.get("tool_name", "unknown")
            if interrupt_info.data
            else "unknown"
        )
        super().__init__(f"Tool interrupt: {tool_name}")
