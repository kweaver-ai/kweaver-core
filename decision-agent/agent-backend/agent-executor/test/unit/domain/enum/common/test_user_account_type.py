"""单元测试 - domain/enum/common/user_account_type 模块"""


class TestUserAccountType:
    """测试 UserAccountType 枚举"""

    def test_enum_values(self):
        """测试枚举值"""
        from app.domain.enum.common.user_account_type import UserAccountType

        assert UserAccountType.APP.value == "app"
        assert UserAccountType.USER.value == "user"
        assert UserAccountType.ANONYMOUS.value == "anonymous"

    def test_enum_is_string_enum(self):
        """测试枚举是字符串枚举"""
        from app.domain.enum.common.user_account_type import UserAccountType

        assert isinstance(UserAccountType.APP.value, str)
        assert isinstance(UserAccountType.APP, str)

    def test_enum_members(self):
        """测试枚举成员"""
        from app.domain.enum.common.user_account_type import UserAccountType

        assert UserAccountType.APP in UserAccountType
        assert UserAccountType.USER in UserAccountType
        assert UserAccountType.ANONYMOUS in UserAccountType


class TestCheckUserAccountType:
    """测试 check_user_account_type 函数"""

    def test_check_app_type(self):
        """测试检查app类型"""
        from app.domain.enum.common.user_account_type import check_user_account_type

        result = check_user_account_type("app")

        assert result is True

    def test_check_user_type(self):
        """测试检查user类型"""
        from app.domain.enum.common.user_account_type import check_user_account_type

        result = check_user_account_type("user")

        assert result is True

    def test_check_anonymous_type(self):
        """测试检查anonymous类型（暂不支持）"""
        from app.domain.enum.common.user_account_type import check_user_account_type

        result = check_user_account_type("anonymous")

        assert result is False

    def test_check_invalid_type(self):
        """测试检查无效类型"""
        from app.domain.enum.common.user_account_type import check_user_account_type

        result = check_user_account_type("invalid")

        assert result is False

    def test_check_empty_string(self):
        """测试检查空字符串"""
        from app.domain.enum.common.user_account_type import check_user_account_type

        result = check_user_account_type("")

        assert result is False

    def test_check_case_sensitivity(self):
        """测试检查大小写敏感"""
        from app.domain.enum.common.user_account_type import check_user_account_type

        # Uppercase should not match
        result = check_user_account_type("APP")

        assert result is False

        result = check_user_account_type("App")

        assert result is False
