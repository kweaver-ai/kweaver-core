# -*- coding:utf-8 -*-
from enum import Enum
from pydantic import BaseModel, Field


class AgentCacheAction(str, Enum):
    """Agent缓存操作类型"""

    UPSERT = "upsert"
    GET_INFO = "get_info"


class AgentCacheManageReq(BaseModel):
    """Agent缓存管理请求

    agent_config和agent_id 不能同时为空。agent_config优先级高于agent_id。
    """

    action: AgentCacheAction = Field(
        ...,
        description="操作类型: upsert(更新或创建缓存), get_info(获取缓存信息)",
    )

    agent_id: str = Field(
        ...,
        title="agent_id",
        description="agent id",
        json_schema_extra={"example": "xxx"},
    )

    agent_version: str = Field(
        ...,
        title="agent_version",
        description="agent版本号,与agent_id配合使用",
        json_schema_extra={"example": "latest"},
    )
