from typing import Any, Dict, List

from pydantic import BaseModel, Field


class ZhipuSearchResponse(BaseModel):
    """智谱搜索响应"""

    """样例
    {'choices': [{...}], 'created': 1749713107, 'id': '20250612152507b7e80ee738244e30', 'model': 'web-search-pro', 'request_id': 'a983a178-7f9a-4ef2-a4d0-6cb5e17570bf', 'usage': {'completion_tokens': 3000, 'prompt_tokens': 0, 'total_tokens': 3000}}
    """

    choices: List[Dict[str, Any]] = Field(..., description="搜索结果列表")
    created: int = Field(..., description="创建时间")
    id: str = Field(..., description="搜索ID")
    model: str = Field(..., description="模型")
    request_id: str = Field(..., description="请求ID")
    usage: Dict[str, Any] = Field(..., description="使用情况统计")


class ReferenceResult(BaseModel):
    """引用内容"""

    title: str = Field(..., description="引用标题")
    content: str = Field(..., description="引用内容")
    index: int = Field(..., description="引用索引")
    link: str = Field(..., description="引用链接")


class OnlineSearchCiteResponse(BaseModel):
    """联网搜索添加引用响应"""

    references: List[ReferenceResult] = Field(..., description="引用内容列表")
    answer: str = Field(..., description="添加引用标记的回答")


class NL2NGQLResponse(BaseModel):
    """NL2NGQL响应"""

    outputs: List[Dict[str, Any]] = Field(..., description="输出结果列表")


class SchemaInfo(BaseModel):
    """图数据库模式信息"""

    schema_data: Dict[str, Any] = Field(..., description="数据库模式", alias="schema")
