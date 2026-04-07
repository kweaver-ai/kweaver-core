# -*- coding:utf-8 -*-
from typing import List, Optional

from pydantic import BaseModel, Field

from .skill_input_vo import SkillInputVo


class ResultProcessCategoryVo(BaseModel):
    """结果处理策略分类"""

    id: Optional[str] = Field(None, description="结果处理策略分类ID")
    name: Optional[str] = Field(None, description="结果处理策略分类名称")
    description: Optional[str] = Field(None, description="结果处理策略分类描述")


class ResultProcessStrategyDetailVo(BaseModel):
    """结果处理策略详情"""

    id: Optional[str] = Field(None, description="结果处理策略ID")
    name: Optional[str] = Field(None, description="结果处理策略名称")
    description: Optional[str] = Field(None, description="结果处理策略描述")


class ResultProcessStrategyVo(BaseModel):
    """结果处理策略"""

    category: Optional[ResultProcessCategoryVo] = Field(
        None, description="结果处理策略分类"
    )
    strategy: Optional[ResultProcessStrategyDetailVo] = Field(
        None, description="结果处理策略"
    )


class ToolSkillVo(BaseModel):
    """工具（技能）配置"""

    tool_id: str = Field(..., description="工具id")
    tool_box_id: str = Field(..., description="工具箱id")
    tool_timeout: int = Field(300, description="工具调用超时时间")
    tool_input: Optional[List[SkillInputVo]] = Field(
        default_factory=list, description="输入参数配置"
    )
    intervention: Optional[bool] = Field(False, description="是否启用干预")
    intervention_confirmation_message: Optional[str] = Field(
        None, description="人工干预确认消息"
    )
    result_process_strategies: Optional[List[ResultProcessStrategyVo]] = Field(
        default_factory=list, description="结果处理策略"
    )
