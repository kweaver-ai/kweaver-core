# -*- coding:utf-8 -*-
"""
Agent缓存管理 - 主入口函数
"""

from fastapi import Request, Depends

from app.driven.dip.agent_factory_service import agent_factory_service
from app.domain.enum.common.user_account_header_key import (
    set_biz_domain_id,
    set_user_account_id,
    set_user_account_type,
)
from app.utils.observability.observability_log import get_logger as o11y_logger

from ..common import router
from ..dependencies import get_account_id, get_account_type, get_biz_domain_id
from ..rdto.v1.req.agent_cache import (
    AgentCacheManageReq,
    AgentCacheAction,
)
from ..rdto.v1.res.agent_cache import AgentCacheManageRes

from .common import prepare_agent_config
from .action_upsert import handle_upsert
from .action_get_info import handle_get_info

from typing import Optional


@router.post(
    "/cache/manage",
    summary="Agent缓存管理",
    response_model=Optional[AgentCacheManageRes],
)
async def manage_agent_cache(
    request: Request,
    req: AgentCacheManageReq,
    account_id: str = Depends(get_account_id),
    account_type: str = Depends(get_account_type),
    biz_domain_id: str = Depends(get_biz_domain_id),
) -> AgentCacheManageRes:
    """
    Agent缓存管理

    功能：
    1. upsert: 更新或创建缓存（存在则更新并恢复TTL，不存在则创建）
    2. get_info: 获取缓存信息（存在返回信息，不存在返回null）

    优势：
    - 降低首次run_agent的响应时间
    - 复用缓存数据
    - 提升用户体验
    """

    try:
        # 0. 设置agent_factory_service的headers
        service_headers = {}

        set_user_account_id(service_headers, account_id)
        set_user_account_type(service_headers, account_type)
        set_biz_domain_id(service_headers, biz_domain_id)
        agent_factory_service.set_headers(service_headers)

        # 1. 获取agent配置
        agent_config = await prepare_agent_config(req, account_id, account_type)

        # 2. 根据action分发处理
        if req.action == AgentCacheAction.UPSERT:
            return await handle_upsert(
                request=request,
                account_id=account_id,
                account_type=account_type,
                agent_id=req.agent_id,
                agent_version=req.agent_version,
                agent_config=agent_config,
            )
        elif req.action == AgentCacheAction.GET_INFO:
            return await handle_get_info(
                request=request,
                account_id=account_id,
                account_type=account_type,
                agent_id=req.agent_id,
                agent_version=req.agent_version,
                agent_config=agent_config,
            )

    except Exception as e:
        o11y_logger().error(f"manage_agent_cache failed: {e}")
        raise
