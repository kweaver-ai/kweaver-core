"""
获取会话查询

定义获取会话的查询对象。
"""
from dataclasses import dataclass


@dataclass
class GetSessionQuery:
    """获取会话查询"""
    session_id: str

    def __post_init__(self):
        """初始化后验证"""
        if not self.session_id:
            raise ValueError("session_id cannot be empty")
