# -*- coding:utf-8 -*-
from typing import Any, Dict, List, Optional

from pydantic import BaseModel, Field


class AgentInputVo(BaseModel):
    """Agent输入模型"""

    query: str  # 用户输入
    history: Optional[List[Dict[str, str]]] = None  # 历史记录，包含角色和内容的对话记录
    tool: Optional[Dict[str, Any]] = Field(
        default_factory=dict
    )  # 发生中断时，需要用户进行干预的工具信息
    header: Optional[Dict[str, Any]] = Field(default_factory=dict)  # 请求头信息
    self_config: Optional[Dict[str, Any]] = Field(
        default_factory=dict
    )  # 当前agent的配置信息

    class Config:
        """Pydantic配置"""

        extra = "allow"  # 允许额外的字段
        populate_by_name = True  # 允许使用别名初始化模型
        populate_by_alias = True  # 允许使用别名初始化模型

    def get_value(self, key: str, default: Any = None) -> Any:
        """灵活获取字段值，无论字段是否在模型定义中"""
        # 先检查是否为定义的属性
        if hasattr(self, key):
            return getattr(self, key)

        # 再检查额外字段
        data = self.model_dump()
        return data.get(key, default)

    def set_value(self, key: str, value: Any) -> None:
        """设置字段值，无论字段是否在模型定义中"""
        # 对于Pydantic模型，直接使用setattr设置属性
        # 对于定义的字段，会应用验证逻辑
        # 对于额外字段，会存储在模型的内部数据结构中
        setattr(self, key, value)

    def model_dump(self, *args, **kwargs):
        data = super().model_dump(*args, **kwargs)
        if not self.tool:
            data.pop("tool")
        return data
