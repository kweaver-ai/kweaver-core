# -*- coding:utf-8 -*-
from typing import Any, List, Optional

from pydantic import BaseModel, Field


# class InputTypeEnum(str, Enum):
#     """输入类型枚举"""
#     STRING = "string"
#     FILE = "file"
#     OBJECT = "object"


# class MapTypeEnum(str, Enum):
#     """值类型枚举"""
#     FIXED_VALUE = "fixedValue"  # 固定值
#     VAR = "var"  # 引用变量
#     MODEL = "model"  # 选择模型
#     AUTO = "auto"  # 模型生成


class SkillInputVo(BaseModel):
    """
    技能输入参数配置

    用于定义技能输入参数的配置信息,支持嵌套参数结构:
    - 支持多种映射类型: fixedValue(固定值)、var(变量)、model(选择模型)、auto(模型自动生成)
    - 支持嵌套参数: 通过 children 字段定义子参数
    """

    enable: Optional[bool] = Field(default=None, description="是否启用此参数")

    input_name: str = Field(..., description="输入参数名称,必传字段")
    input_type: str = Field(..., description="输入参数类型,必传字段")
    input_desc: Optional[str] = Field(None, description="输入参数描述")

    map_type: Optional[str] = Field(
        default=None, description="映射类型。可选值:fixedValue、var、model、auto"
    )
    map_value: Optional[Any] = Field(None, description="映射值")

    children: Optional[List["SkillInputVo"]] = Field(
        default=None, description="子参数列表,非必传字段,用于嵌套参数场景"
    )

    # def is_enabled(self) -> bool:
    #     """检查参数是否启用"""
    #     return self.enable == True or self.enable is None

    # def get_map_value(self) -> Union[str, Dict[str, Any]]:
    #     """获取映射值"""
    #     return self.map_value if self.map_value is not None else ""

    # def get_map_type(self) -> str:
    #     """获取映射类型"""
    #     return self.map_type if self.map_type is not None else "auto"

    # class Config:
    #     """Pydantic配置"""

    #     # 允许从字典创建模型
    #     from_attributes = True

    #     # 允许额外字段
    #     extra = "allow"

    #     # 验证时使用赋值
    #     validate_assignment = True
