# -*- coding:utf-8 -*-
"""
Agent缓存相关常量定义
"""

# Agent缓存 TTL (Time To Live) 配置，单位：秒
AGENT_CACHE_TTL = 60  # 60秒

# 缓存数据更新触发阈值，单位：秒
# 当发现缓存 TTL 减少的时间超过此值时，触发缓存数据更新
AGENT_CACHE_DATA_UPDATE_PASS_SECOND = 10  # 10秒
