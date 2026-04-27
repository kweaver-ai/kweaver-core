# -*- coding:utf-8 -*-
from dataclasses import dataclass
from datetime import datetime

from app.domain.vo.agent_cache import AgentCacheIdVO
from app.domain.vo.agent_cache.cache_data_vo import CacheDataVo


@dataclass
class AgentCacheEntity:
    """Agent缓存实体

    缓存的value存储的信息：
    - tools_info_dict
    - skill_agent_info_dict
    - llm_config_dict
    """

    cache_id_vo: AgentCacheIdVO
    agent_id: str
    agent_version: str
    cache_data: CacheDataVo  # 缓存数据结构
    cache_data_last_set_timestamp: int
    created_at: datetime
