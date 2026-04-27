"""单元测试 - domain/constant/agent_cache_constants 模块"""

from app.domain.constant.agent_cache_constants import (
    AGENT_CACHE_TTL,
    AGENT_CACHE_DATA_UPDATE_PASS_SECOND,
)


class TestAgentCacheConstants:
    """测试 Agent 缓存常量"""

    def test_agent_cache_ttl(self):
        """测试 AGENT_CACHE_TTL 常量"""
        assert AGENT_CACHE_TTL == 60  # 60秒

    def test_agent_cache_data_update_pass_second(self):
        """测试 AGENT_CACHE_DATA_UPDATE_PASS_SECOND 常量"""
        assert AGENT_CACHE_DATA_UPDATE_PASS_SECOND == 10  # 10秒

    def test_cache_ttl_is_positive(self):
        """测试缓存 TTL 是正数"""
        assert AGENT_CACHE_TTL > 0

    def test_update_threshold_is_less_than_ttl(self):
        """测试更新阈值小于 TTL"""
        assert AGENT_CACHE_DATA_UPDATE_PASS_SECOND < AGENT_CACHE_TTL
