"""单元测试 - get_account_id 依赖函数"""

import pytest
from app.common.errors import ParamException


@pytest.mark.asyncio
class TestGetAccountId:
    """测试 get_account_id 依赖函数"""

    async def test_valid_new_header(self):
        """测试有效的新版header x-account-id"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        account_id = await get_account_id(x_account_id="test_account_123")
        assert account_id == "test_account_123"

    async def test_valid_old_header(self):
        """测试有效的旧版header x-user"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        account_id = await get_account_id(x_account_id=None, x_user_id="old_user_456")
        assert account_id == "old_user_456"

    async def test_new_header_takes_precedence(self):
        """测试新版header优先级高于旧版"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        account_id = await get_account_id(
            x_account_id="new_account", x_user_id="old_user"
        )
        assert account_id == "new_account"  # Should prefer x_account_id

    async def test_missing_both_headers_raises_error(self):
        """测试缺少两个header时抛出异常"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        with pytest.raises(ParamException) as exc_info:
            await get_account_id(x_account_id=None, x_user_id=None)

        assert "account" in str(exc_info.value).lower()

    async def test_empty_string_header(self):
        """测试空字符串header"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        # Empty string is falsy in Python, so it should raise error
        with pytest.raises(ParamException):
            await get_account_id(x_account_id="", x_user_id=None)

    async def test_whitespace_header(self):
        """测试空白字符header"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        # Whitespace is not empty, so it should be returned
        account_id = await get_account_id(x_account_id="   ")
        assert account_id == "   "

    async def test_unicode_account_id(self):
        """测试Unicode account ID"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        account_id = await get_account_id(x_account_id="测试账号_123")
        assert account_id == "测试账号_123"

    async def test_special_characters_in_account_id(self):
        """测试account ID中的特殊字符"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        account_id = await get_account_id(x_account_id="account@#$%^&*()")
        assert account_id == "account@#$%^&*()"

    async def test_long_account_id(self):
        """测试长account ID"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        long_id = "a" * 1000
        account_id = await get_account_id(x_account_id=long_id)
        assert account_id == long_id

    async def test_account_id_with_newlines(self):
        """测试包含换行符的account ID"""
        from app.router.agent_controller_pkg.dependencies.get_account_id import (
            get_account_id,
        )

        account_id = await get_account_id(x_account_id="line1\nline2")
        assert "\n" in account_id
