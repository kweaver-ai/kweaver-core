from datetime import datetime
from typing import List
from pydantic import BaseModel, Field


class Memory(BaseModel):
    """记忆实体"""

    id: str = Field(..., description="记忆ID")
    content: str = Field(..., description="记忆内容")
    embedding: List[float] = Field(..., description="记忆内容的向量表示")
    metadata: dict = Field(default_factory=dict, description="元数据")
    created_at: datetime = Field(default_factory=datetime.now, description="创建时间")
    updated_at: datetime = Field(default_factory=datetime.now, description="更新时间")
    importance: float = Field(default=1.0, description="重要性权重")
    source: str = Field(..., description="记忆来源")
    tags: List[str] = Field(default_factory=list, description="标签")


class MemoryChunk(BaseModel):
    """记忆分块"""

    id: str = Field(..., description="分块ID")
    memory_id: str = Field(..., description="所属记忆ID")
    content: str = Field(..., description="分块内容")
    embedding: List[float] = Field(..., description="分块内容的向量表示")
    chunk_index: int = Field(..., description="分块索引")
    metadata: dict = Field(default_factory=dict, description="元数据")
