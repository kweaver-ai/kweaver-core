# -*- coding:utf-8 -*-
from pydantic import BaseModel, Field


class McpSkillVo(BaseModel):
    """MCP类型技能配置"""

    mcp_server_id: str = Field(..., description="mcp server id")
