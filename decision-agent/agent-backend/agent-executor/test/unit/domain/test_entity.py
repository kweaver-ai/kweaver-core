"""单元测试 - domain/entity 模块"""


class TestAgentCacheEntity:
    """测试 domain/entity 模块导出"""

    def test_module_exists(self):
        """测试模块可以导入"""
        from app.domain.entity import agent_cache

        assert agent_cache is not None
