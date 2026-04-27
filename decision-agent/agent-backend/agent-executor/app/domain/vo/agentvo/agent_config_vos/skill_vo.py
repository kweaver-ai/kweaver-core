# -*- coding:utf-8 -*-
from typing import List, Optional

from pydantic import BaseModel, Field, validator

from .tool_skill_vo import ToolSkillVo
from .agent_skill_vo import AgentSkillVo
from .mcp_skill_vo import McpSkillVo


class SkillVo(BaseModel):
    """技能配置"""

    tools: Optional[List[ToolSkillVo]] = Field(
        default_factory=list, description="工具列表"
    )
    agents: Optional[List[AgentSkillVo]] = Field(
        default_factory=list, description="Agent列表"
    )
    mcps: Optional[List[McpSkillVo]] = Field(
        default_factory=list, description="MCP列表"
    )

    @validator("tools", "agents", "mcps", pre=True, always=True)
    def ensure_list(cls, v):
        """确保字段为列表类型"""
        if v is None:
            return []
        return v
