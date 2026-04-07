from typing import List, Dict, Any, Optional
from pydantic import BaseModel, Field, model_validator
from enum import Enum


class VisitorType(str, Enum):
    """访客类型"""

    REALNAME = "realname"
    ANONYMOUS = "anonymous"
    BUSINESS = "business"


class RequestContext(BaseModel):
    """请求上下文"""

    user_id: Optional[str] = Field(None, description="用户ID/应用账号ID")
    visitor_type: Optional[VisitorType] = Field(None, description="账号类型")


class Message(BaseModel):
    """消息模型"""

    role: str = Field(..., description="消息角色")
    content: str = Field(..., description="消息内容")


class BuildMemoryRequest(BaseModel):
    """构建记忆请求"""

    messages: List[Message] = Field(..., description="消息列表")
    user_id: Optional[str] = Field(None, description="用户ID")
    agent_id: Optional[str] = Field(None, description="代理ID")
    run_id: Optional[str] = Field(None, description="运行ID")
    metadata: Optional[Dict[str, Any]] = Field(None, description="元数据")
    infer: bool = Field(True, description="是否推理")
    memory_type: Optional[str] = Field(None, description="记忆类型")
    prompt: Optional[str] = Field(None, description="提示词")
    context: Optional[RequestContext] = Field(None, description="请求上下文")


class RetrievalMemoryRequest(BaseModel):
    """召回记忆请求"""

    query: str = Field(..., description="查询文本")
    user_id: Optional[str] = Field(None, description="用户ID")
    agent_id: Optional[str] = Field(None, description="代理ID")
    run_id: Optional[str] = Field(None, description="运行ID")
    limit: int = Field(5, description="返回结果数量限制")
    filters: Optional[Dict[str, Any]] = Field(None, description="过滤条件")
    threshold: Optional[float] = Field(None, description="相似度阈值")
    rerank_threshold: Optional[float] = Field(None, description="Rerank 阈值")
    context: Optional[RequestContext] = Field(None, description="请求上下文")


class UpdateMemoryRequest(BaseModel):
    """更新记忆请求"""

    data: str = Field(..., description="更新数据")


class GetAllMemoriesRequest(BaseModel):
    """获取所有记忆请求"""

    user_id: Optional[str] = Field(None, description="用户ID")
    agent_id: Optional[str] = Field(None, description="Agent ID")
    run_id: Optional[str] = Field(None, description="Run ID")
    limit: int = Field(100, description="返回结果数量限制")

    @model_validator(mode="after")
    def check_at_least_one_identifier(self):
        if not any([self.user_id, self.agent_id, self.run_id]):
            raise ValueError(
                "At least one of 'user_id', 'agent_id' or 'run_id' must be provided."
            )
        return self


class SearchMemoryItemResponse(BaseModel):
    """搜索记忆响应项"""

    id: str = Field(..., description="记忆ID")
    memory: str = Field(..., description="记忆内容")
    hash: Optional[str] = Field(None, description="记忆哈希值")
    metadata: Optional[Dict[str, Any]] = Field(None, description="元数据")
    score: Optional[float] = Field(None, description="相似度分数")
    rerank_score: Optional[float] = Field(None, description="Rerank评分")
    created_at: str = Field(..., description="创建时间")
    updated_at: Optional[str] = Field(None, description="更新时间")
    user_id: Optional[str] = Field(None, description="用户ID")
    agent_id: Optional[str] = Field(None, description="代理ID")
    run_id: Optional[str] = Field(None, description="运行ID")


class SearchMemoryResponse(BaseModel):
    """搜索响应"""

    result: List[SearchMemoryItemResponse] = []


class MemoryResponse(BaseModel):
    """记忆响应"""

    id: str = Field(..., description="记忆ID")
    memory: str = Field(..., description="记忆内容")
    hash: Optional[str] = Field(None, description="记忆哈希值")
    metadata: Optional[Dict[str, Any]] = Field(None, description="元数据")
    # score: Optional[float] = Field(None, description="相似度分数")
    created_at: str = Field(..., description="创建时间")
    updated_at: Optional[str] = Field(None, description="更新时间")
    user_id: Optional[str] = Field(None, description="用户ID")
    agent_id: Optional[str] = Field(None, description="代理ID")
    run_id: Optional[str] = Field(None, description="运行ID")


class MemoryListResponse(BaseModel):
    """记忆列表响应"""

    result: List[MemoryResponse] = Field(..., description="记忆列表")
    total: int = Field(..., description="总数")


class MemoryHistoryResponse(BaseModel):
    """记忆历史响应"""

    history: List[Dict[str, Any]] = Field(..., description="历史记录")


class ErrorResponse(BaseModel):
    """错误响应模型"""

    error_code: str
    description: str  # 支持多语言的错误描述
    solution: str  # 支持多语言的解决方案
    error_link: str  # 支持多语言的错误链接
    error_details: Optional[Dict[str, Any]] = None  # 可选的错误详情
