# -*- coding:utf-8 -*-
"""
Agent缓存管理 - 公共模块

包含:
- 全局AgentCacheManager实例
- 准备Agent配置
- 构造AgentCacheIdVO
- 创建Cache并返回响应
- 获取缓存数据
"""

import json
from dataclasses import asdict
from typing import Dict, Any
from fastapi import Request

from app.common.errors import (
    AgentPermissionException,
    CodeException,
    APIError,
)
from app.common.config import Config
from app.driven.dip.agent_factory_service import agent_factory_service
from app.domain.vo.agentvo import AgentConfigVo
from app.domain.vo.agent_cache import AgentCacheIdVO
from app.logic.agent_core_logic_v2.agent_cache_manage_logic import AgentCacheManager
from app.utils.observability.observability_log import get_logger as o11y_logger

from ..rdto.v1.req.agent_cache import AgentCacheManageReq
from ..rdto.v1.res.agent_cache import AgentCacheManageRes

# 全局AgentCacheManager实例
cache_manager = AgentCacheManager()


async def prepare_agent_config(
    req: AgentCacheManageReq,
    account_id: str,
    account_type: str,
) -> AgentConfigVo:
    """准备Agent配置

    通过agent_id获取配置

    Args:
        req: 请求对象
        account_id: 账户ID
        account_type: 账户类型

    Returns:
        AgentConfigVo: Agent配置
    """

    # 1. 通过agent_id获取配置
    config_data = await agent_factory_service.get_agent_config(req.agent_id)

    config_str = config_data["config"]
    if isinstance(config_str, str):
        config_dict = json.loads(config_str)
    else:
        config_dict = config_str

    agent_config = AgentConfigVo(**config_dict)

    # 2. 设置agent_id（如果未设置）
    if agent_config.agent_id is None:
        agent_config.agent_id = req.agent_id

    # 3. 检查权限
    if not await agent_factory_service.check_agent_permission(
        agent_config.agent_id, account_id, account_type
    ):
        o11y_logger().error(
            f"check_agent_permission failed: agent_id = {agent_config.agent_id}, account_id = {account_id}, account_type = {account_type}"
        )
        raise AgentPermissionException(agent_config.agent_id, account_id)

    return agent_config


def build_cache_id_vo(
    account_id: str,
    account_type: str,
    agent_id: str,
    agent_version: str,
    agent_config: AgentConfigVo,
) -> AgentCacheIdVO:
    """构造AgentCacheIdVO

    Args:
        account_id: 账户ID
        account_type: 账户类型
        agent_id: Agent ID
        agent_version: Agent版本
        agent_config: Agent配置

    Returns:
        AgentCacheIdVO: Agent缓存ID值对象
    """
    return AgentCacheIdVO(
        account_id=account_id,
        account_type=account_type,
        agent_id=agent_id,
        agent_version=agent_version,
        agent_config_version_flag=str(agent_config.get_config_last_set_timestamp()),
    )


def get_cache_data_for_debug_mode(cache_entity) -> Dict[str, Any]:
    """获取缓存数据（仅debug模式）

    Args:
        cache_entity: Agent缓存实体对象

    Returns:
        Dict[str, Any]: 缓存数据，非debug模式返回空字典
    """
    if Config.is_debug_mode() and cache_entity and cache_entity.cache_data:
        return asdict(cache_entity.cache_data)
    return {}


async def create_cache_and_build_response(
    request: Request,
    account_id: str,
    account_type: str,
    agent_id: str,
    agent_version: str,
    agent_config: AgentConfigVo,
) -> AgentCacheManageRes:
    """创建Agent缓存并构建响应

    Args:
        request: FastAPI请求对象
        account_id: 账户ID
        account_type: 账户类型
        agent_id: Agent ID
        agent_version: Agent版本
        agent_config: Agent配置

    Returns:
        AgentCacheManageRes: 响应对象
    """
    # 1. 构造headers
    headers = dict(request.headers)

    # 2. 创建Cache
    cache_entity = await cache_manager.create_cache(
        account_id=account_id,
        account_type=account_type,
        agent_id=agent_id,
        agent_version=agent_version,
        agent_config=agent_config,
        headers=headers,
    )

    # 3. 获取实时TTL
    ttl = await cache_manager.cache_service.get_ttl(cache_entity.cache_id_vo)
    if ttl <= 0:
        raise CodeException(
            error=APIError(
                error_code="AgentExecutor.InternalServerError.CacheTTLInvalid",
                description="Agent cache TTL is invalid after creation",
                solution="Please check Redis connection and try again",
            ),
            error_details=f"cache_id={cache_entity.cache_id_vo}, ttl={ttl}",
        )

    # 4. 获取缓存数据（debug模式）
    cache_data = get_cache_data_for_debug_mode(cache_entity)

    # 5. 返回响应
    return AgentCacheManageRes(
        cache_id=cache_entity.cache_id_vo.get_cache_id(),
        ttl=ttl,
        created_at=cache_entity.created_at,
        cache_data=cache_data,
    )
