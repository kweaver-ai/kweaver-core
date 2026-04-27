"""单元测试 - get_biz_domain_id 依赖函数"""

import pytest


@pytest.mark.asyncio
class TestGetBizDomainId:
    """测试 get_biz_domain_id 依赖函数"""

    async def test_valid_header(self):
        """测试有效的header x-business-domain"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="domain_123")
        assert domain_id == "domain_123"

    async def test_missing_header_returns_empty_string(self):
        """测试缺少header时返回空字符串"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain=None)
        assert domain_id == ""

    async def test_empty_string_header(self):
        """测试空字符串header"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        # Empty string is returned (no exception for biz_domain_id)
        domain_id = await get_biz_domain_id(x_business_domain="")
        assert domain_id == ""

    async def test_whitespace_header(self):
        """测试空白字符header"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="   ")
        assert domain_id == "   "

    async def test_unicode_domain_id(self):
        """测试Unicode域名ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="域测试_123")
        assert domain_id == "域测试_123"

    async def test_numeric_domain_id(self):
        """测试数字域ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="12345")
        assert domain_id == "12345"

    async def test_domain_id_with_dots(self):
        """测试包含点的域ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="domain.example.com")
        assert domain_id == "domain.example.com"

    async def test_domain_id_with_special_characters(self):
        """测试包含特殊字符的域ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="domain@#$%_test-123")
        assert domain_id == "domain@#$%_test-123"

    async def test_long_domain_id(self):
        """测试长域ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        long_domain = "d" * 500
        domain_id = await get_biz_domain_id(x_business_domain=long_domain)
        assert domain_id == long_domain

    async def test_domain_id_with_newlines(self):
        """测试包含换行符的域ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="domain\nline")
        assert "\n" in domain_id

    async def test_domain_id_with_tabs(self):
        """测试包含制表符的域ID"""
        from app.router.agent_controller_pkg.dependencies.get_biz_domain_id import (
            get_biz_domain_id,
        )

        domain_id = await get_biz_domain_id(x_business_domain="domain\ttabbed")
        assert "\t" in domain_id
