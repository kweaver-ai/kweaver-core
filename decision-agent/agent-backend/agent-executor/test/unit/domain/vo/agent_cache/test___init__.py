"""单元测试 - domain/vo/agent_cache/__init__ 模块"""


class TestAgentCacheInit:
    """测试 domain/vo/agent_cache/__init__ 导出"""

    def test_module_exports_agent_cache_id_vo(self):
        """测试模块导出AgentCacheIdVO"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        assert AgentCacheIdVO is not None

    def test_module_exports_cache_data_vo(self):
        """测试模块导出CacheDataVo"""
        from app.domain.vo.agent_cache import CacheDataVo

        assert CacheDataVo is not None

    def test_module_all_list(self):
        """测试__all__列表"""
        from app.domain.vo.agent_cache import __all__

        assert "AgentCacheIdVO" in __all__
        assert "CacheDataVo" in __all__
