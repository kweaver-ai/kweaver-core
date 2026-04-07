"""
获取模板查询

定义获取模板详情的查询 DTO。
"""
from dataclasses import dataclass


@dataclass
class GetTemplateQuery:
    """获取模板查询"""
    template_id: str
