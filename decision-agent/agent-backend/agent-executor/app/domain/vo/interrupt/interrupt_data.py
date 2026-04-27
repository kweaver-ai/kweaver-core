# -*- coding:utf-8 -*-
"""中断详情 VO"""

from typing import List, Any, Optional
from pydantic import BaseModel, Field


class ToolArg(BaseModel):
    """工具参数"""

    key: str = Field(..., description="参数名称")
    value: Any = Field(..., description="参数值")
    type: str = Field(..., description="参数类型")


class InterruptConfig(BaseModel):
    """中断配置"""

    requires_confirmation: bool = Field(..., description="是否需要用户确认")
    confirmation_message: str = Field(..., description="确认提示消息")


class InterruptData(BaseModel):
    """中断详情"""

    tool_name: str = Field(..., description="工具名称")
    tool_description: Optional[str] = Field(None, description="工具描述")
    tool_args: List[ToolArg] = Field(default_factory=list, description="工具参数列表")
    interrupt_config: Optional[InterruptConfig] = Field(None, description="中断配置")
