# -*- coding:utf-8 -*-
"""Terminate Agent 请求 DTO"""

from pydantic import BaseModel, Field


class TerminateAgentRequest(BaseModel):
    """终止 Agent 执行请求"""

    agent_run_id: str = Field(..., description="Agent运行ID")
