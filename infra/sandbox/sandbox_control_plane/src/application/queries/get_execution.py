"""
获取执行查询

定义获取执行详情的查询 DTO。
"""
from dataclasses import dataclass


@dataclass
class GetExecutionQuery:
    """获取执行查询"""
    execution_id: str
