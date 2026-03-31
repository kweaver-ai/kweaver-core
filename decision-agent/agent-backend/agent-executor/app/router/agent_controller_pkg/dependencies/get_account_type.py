"""
获取账号类型的依赖注入函数
"""

from typing import Optional
from fastapi import Header

from app.domain.enum.common.user_account_header_key import UserAccountHeaderKey


async def get_account_type(
    x_account_type: Optional[str] = Header(
        None, alias=UserAccountHeaderKey.ACCOUNT_TYPE.value
    ),
    x_user_type: Optional[str] = Header(
        None, alias=UserAccountHeaderKey.ACCOUNT_TYPE_OLD.value
    ),
) -> str:
    """
    获取账号类型，支持新旧两种 header 键
    优先使用新的 x-account-type，如果不存在则使用旧的 x-user-type
    """
    account_type = x_account_type or x_user_type or ""
    return account_type
