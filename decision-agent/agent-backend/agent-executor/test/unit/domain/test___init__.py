"""单元测试 - domain/__init__ 模块"""


class TestDomainInit:
    """测试 domain/__init__ 模块存在性"""

    def test_module_exists(self):
        """测试模块可以导入"""
        from app.domain import constant

        assert constant is not None
