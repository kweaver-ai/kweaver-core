# -*- coding:utf-8 -*-
"""Resume Agent 请求 DTO"""

from typing import Optional, List, Any
from pydantic import BaseModel, Field

from app.domain.vo.interrupt.interrupt_handle import InterruptHandle


class ModifiedArg(BaseModel):
    """修改后的参数"""

    key: str = Field(..., description="参数名称")
    value: Any = Field(..., description="参数值")


class ResumeInfo(BaseModel):
    """恢复执行信息"""

    resume_handle: InterruptHandle = Field(..., description="恢复句柄")
    action: str = Field(..., description="操作类型: confirm | skip")
    modified_args: Optional[List[ModifiedArg]] = Field(None, description="修改后的参数")
    data: dict = Field(..., description="中断详情数据（从响应透传，必填）")


class ResumeAgentRequest(BaseModel):
    """恢复 Agent 执行请求"""

    agent_run_id: str = Field(..., description="Agent运行ID")
    resume_info: ResumeInfo = Field(..., description="恢复执行信息")
