"""单元测试 - domain/enum/common/user_account_type 模块"""

from app.domain.enum.common.user_account_type import (
    UserAccountType,
    check_user_account_type,
)


class TestUserAccountType:
    """测试 UserAccountType 枚举"""

    def test_enum_values(self):
        """测试枚举值"""
        assert UserAccountType.APP.value == "app"
        assert UserAccountType.USER.value == "user"
        assert UserAccountType.ANONYMOUS.value == "anonymous"

    def test_enum_as_string(self):
        """测试枚举作为字符串使用"""
        # Enum is a str subclass, so it compares equal to its value
        assert UserAccountType.APP == "app"
        assert UserAccountType.USER == "user"
        # The .value attribute gets the actual string value
        assert UserAccountType.APP.value == "app"

    def test_enum_members(self):
        """测试枚举成员"""
        assert hasattr(UserAccountType, "APP")
        assert hasattr(UserAccountType, "USER")
        assert hasattr(UserAccountType, "ANONYMOUS")


class TestCheckUserAccountType:
    """测试 check_user_account_type 函数"""

    def test_valid_app_type(self):
        """测试有效的 app 类型"""
        assert check_user_account_type("app") is True

    def test_valid_user_type(self):
        """测试有效的 user 类型"""
        assert check_user_account_type("user") is True

    def test_valid_enum_app(self):
        """测试使用枚举值 APP"""
        assert check_user_account_type(UserAccountType.APP) is True

    def test_valid_enum_user(self):
        """测试使用枚举值 USER"""
        assert check_user_account_type(UserAccountType.USER) is True

    def test_invalid_anonymous(self):
        """测试无效的 anonymous 类型"""
        assert check_user_account_type("anonymous") is False

    def test_invalid_type(self):
        """测试无效的类型"""
        assert check_user_account_type("invalid") is False
        assert check_user_account_type("") is False
        assert check_user_account_type("APP") is False  # 大小写敏感

    def test_case_sensitive(self):
        """测试大小写敏感"""
        assert check_user_account_type("App") is False
        assert check_user_account_type("USER") is False
