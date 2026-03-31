"""
获取业务域ID的依赖注入函数
"""

from typing import Optional
from fastapi import Header

from app.domain.enum.common.user_account_header_key import UserAccountHeaderKey


async def get_biz_domain_id(
    x_business_domain: Optional[str] = Header(
        None, alias=UserAccountHeaderKey.BIZ_DOMAIN_ID.value
    ),
) -> str:
    """
    获取业务域ID
    如果不存在则返回空字符串
    """
    return x_business_domain or ""
