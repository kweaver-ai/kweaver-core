"""
FastAPI 依赖注入函数包
"""

from .get_account_id import get_account_id
from .get_account_type import get_account_type
from .get_biz_domain_id import get_biz_domain_id

__all__ = [
    "get_account_id",
    "get_account_type",
    "get_biz_domain_id",
]
