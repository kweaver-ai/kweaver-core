# -*- coding:utf-8 -*-
from typing import Any, Dict, Optional
from datetime import datetime
from pydantic import BaseModel, Field


class AgentCacheManageRes(BaseModel):
    """Agent缓存管理响应"""

    cache_id: str = Field(
        ...,
        description="缓存ID",
        json_schema_extra={"example": "agent_id:agent_version:config_version_flag"},
    )

    ttl: int = Field(
        ...,
        description="缓存有效期（秒）",
        json_schema_extra={"example": 600},
    )

    created_at: Optional[datetime] = Field(
        default=None,
        description="缓存创建时间",
        json_schema_extra={"example": "2024-10-21T10:30:00Z"},
    )

    cache_data: Optional[Dict[str, Any]] = Field(
        default=None,
        description="缓存的数据（仅在DEBUG模式下返回）",
    )
