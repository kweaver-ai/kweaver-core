"""
获取账号ID的依赖注入函数
"""

from typing import Optional
from fastapi import Header

from app.domain.enum.common.user_account_header_key import UserAccountHeaderKey


async def get_account_id(
    x_account_id: Optional[str] = Header(
        None, alias=UserAccountHeaderKey.ACCOUNT_ID.value
    ),
    x_user_id: Optional[str] = Header(
        None, alias=UserAccountHeaderKey.ACCOUNT_ID_OLD.value
    ),
) -> str:
    """
    获取账号ID，支持新旧两种 header 键
    优先使用新的 x-account-id，如果不存在则使用旧的 x-user
    """
    account_id = x_account_id or x_user_id or ""
    return account_id
