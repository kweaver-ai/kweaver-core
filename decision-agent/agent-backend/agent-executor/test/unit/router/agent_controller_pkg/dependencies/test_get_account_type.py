"""单元测试 - get_account_type 依赖函数"""

import pytest
from app.common.errors import ParamException


@pytest.mark.asyncio
class TestGetAccountType:
    """测试 get_account_type 依赖函数"""

    async def test_valid_new_header(self):
        """测试有效的新版header x-account-type"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        account_type = await get_account_type(
            x_account_type="premium", x_user_type=None
        )
        assert account_type == "premium"

    async def test_valid_old_header(self):
        """测试有效的旧版header x-user-type"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        account_type = await get_account_type(
            x_account_type=None, x_user_type="standard"
        )
        assert account_type == "standard"

    async def test_new_header_takes_precedence(self):
        """测试新版header优先级高于旧版"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        account_type = await get_account_type(
            x_account_type="new_type", x_user_type="old_type"
        )
        assert account_type == "new_type"  # Should prefer x_account_type

    async def test_missing_both_headers_raises_error(self):
        """测试缺少两个header时抛出异常"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        with pytest.raises(ParamException) as exc_info:
            await get_account_type(x_account_type=None, x_user_type=None)

        assert "account type" in str(exc_info.value).lower()

    async def test_empty_string_header(self):
        """测试空字符串header"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        # Empty string is falsy, so it should raise error
        with pytest.raises(ParamException):
            await get_account_type(x_account_type="", x_user_type=None)

    async def test_whitespace_header(self):
        """测试空白字符header"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        # Whitespace is not empty, so it should be returned
        account_type = await get_account_type(x_account_type="   ", x_user_type=None)
        assert account_type == "   "

    async def test_common_account_types(self):
        """测试常见的账号类型"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        types = ["free", "premium", "enterprise", "trial", "admin"]
        for account_type in types:
            result = await get_account_type(
                x_account_type=account_type, x_user_type=None
            )
            assert result == account_type

    async def test_numeric_account_type(self):
        """测试数字类型的账号类型"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        account_type = await get_account_type(x_account_type="12345", x_user_type=None)
        assert account_type == "12345"

    async def test_account_type_with_dashes_and_underscores(self):
        """测试包含连字符和下划线的账号类型"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        account_type = await get_account_type(
            x_account_type="premium-user_trial", x_user_type=None
        )
        assert account_type == "premium-user_trial"

    async def test_case_sensitive_account_type(self):
        """测试账号类型区分大小写"""
        from app.router.agent_controller_pkg.dependencies.get_account_type import (
            get_account_type,
        )

        account_type1 = await get_account_type(
            x_account_type="Premium", x_user_type=None
        )
        account_type2 = await get_account_type(
            x_account_type="premium", x_user_type=None
        )

        assert account_type1 == "Premium"
        assert account_type2 == "premium"
        assert account_type1 != account_type2
