"""单元测试 - domain/constant 模块"""


class TestDomainConstantModule:
    """测试 domain/constant 模块"""

    def test_module_exists(self):
        """测试模块可以导入"""
        from app.domain.constant import agent_cache_constants
        from app.domain.constant import agent_version

        assert agent_cache_constants is not None
        assert agent_version is not None
