# -*- coding:utf-8 -*-
"""工具中断信息 VO"""

from typing import Optional
from pydantic import BaseModel, Field

from .interrupt_handle import InterruptHandle
from .interrupt_data import InterruptData


class ToolInterruptInfo(BaseModel):
    """工具中断信息"""

    handle: Optional[InterruptHandle] = Field(None, description="恢复句柄")
    data: Optional[InterruptData] = Field(None, description="中断详情")
