from enum import Enum


class UserAccountType(str, Enum):
    """用户账号类型枚举"""

    APP = "app"  # 应用账号
    USER = "user"  # 普通用户
    ANONYMOUS = "anonymous"  # 匿名用户（暂不支持）


def check_user_account_type(user_account_type: str) -> bool:
    return user_account_type in [UserAccountType.APP, UserAccountType.USER]
