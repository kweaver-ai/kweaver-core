"""单元测试 - domain/enum/__init__ 模块"""


class TestDomainEnumInit:
    """测试 domain/enum/__init__ 模块存在性"""

    def test_module_exists(self):
        """测试模块可以导入"""
        from app.domain.enum import common

        assert common is not None
