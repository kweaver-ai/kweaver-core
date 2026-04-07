# -*- coding:utf-8 -*-
"""Agent缓存管理逻辑模块

负责Agent缓存管理服务
缓存的key的构成：agent_id:agent_version:agent_config_version_flag
缓存的value存储的信息：tools_info_dict, skill_agent_info_dict, llm_config_dict

模块结构：
- manager.py: AgentCacheManager主类
- agent_cache_service.py: Redis缓存服务
- create_cache.py: 缓存创建逻辑
- update_cache_data.py: 缓存数据更新逻辑
"""

from .agent_cache_service import AgentCacheService
from .manager import AgentCacheManager

__all__ = ["AgentCacheService", "AgentCacheManager"]
