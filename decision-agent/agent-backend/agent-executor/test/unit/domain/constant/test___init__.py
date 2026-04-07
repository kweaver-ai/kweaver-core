"""单元测试 - domain/constant/__init__ 模块"""


class TestDomainConstantInit:
    """测试 domain/constant/__init__ 导出"""

    def test_module_exports_agent_cache_ttl(self):
        """测试模块导出AGENT_CACHE_TTL"""
        from app.domain.constant import AGENT_CACHE_TTL

        assert AGENT_CACHE_TTL == 60

    def test_module_exports_agent_cache_data_update_pass_second(self):
        """测试模块导出AGENT_CACHE_DATA_UPDATE_PASS_SECOND"""
        from app.domain.constant import AGENT_CACHE_DATA_UPDATE_PASS_SECOND

        assert AGENT_CACHE_DATA_UPDATE_PASS_SECOND == 10

    def test_module_all_contains_constants(self):
        """测试__all__包含常量"""
        from app.domain.constant import __all__

        assert "AGENT_CACHE_TTL" in __all__
        assert "AGENT_CACHE_DATA_UPDATE_PASS_SECOND" in __all__
