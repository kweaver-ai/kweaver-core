"""单元测试 - domain/vo 模块"""


class TestDomainVoModule:
    """测试 domain/vo 模块"""

    def test_module_exists(self):
        """测试模块可以导入"""
        from app.domain.vo import agent_cache

        assert agent_cache is not None
