# -*- coding:utf-8 -*-
from typing import List, Optional
from enum import Enum

from pydantic import BaseModel, Field

from app.infra.common.infra_constant.const import FINAL_ANSWER_DEFAULT_VAR


class DefaultFormatEnum(str, Enum):
    """默认输出格式枚举"""

    JSON = "json"
    MARKDOWN = "markdown"


class OutputVariablesVo(BaseModel):
    """
    输出变量配置

    用于定义Agent的输出变量配置信息:
    - 支持多种输出变量类型: answer_var、doc_retrieval_var、graph_retrieval_var、related_questions_var
    - 支持其他自定义变量: other_vars
    """

    answer_var: Optional[str] = Field(
        None,
        description=(
            "包含最终回答的变量名\n"
            "- 非dolphin模式下，此字段的值固定为'answer'\n"
            "- dolphin模式下，此字段必填"
        ),
    )

    doc_retrieval_var: Optional[str] = Field(
        None,
        description=(
            "包含文档检索结果的变量名\n"
            "- 非dolphin模式下，此字段的值固定为'doc_retrieval_res'"
        ),
    )

    graph_retrieval_var: Optional[str] = Field(
        None,
        description=(
            "包含图谱检索结果的变量名\n"
            "- 非dolphin模式下，此字段的值固定为'graph_retrieval_res'"
        ),
    )

    related_questions_var: Optional[str] = Field(
        None,
        description=(
            "包含相关问题的变量名\n"
            "- 非dolphin模式下，此字段的值固定为'related_questions'"
        ),
    )

    other_vars: Optional[List[str]] = Field(None, description="其他变量数组")


class OutputConfigVo(BaseModel):
    """
    输出变量配置

    用于配置Agent的输出变量和默认输出格式:
    - variables: 输出变量配置（从dolphin中的所有输出变量中进行选择）
    - default_format: 默认输出格式（json或markdown）
    """

    class Config:
        # 序列化时使用枚举的字符串值而不是枚举对象
        use_enum_values = True

    variables: Optional[OutputVariablesVo] = Field(
        None, description=("输出变量（从dolphin中的所有输出变量中进行选择）")
    )

    default_format: DefaultFormatEnum = Field(
        ..., description="默认输出格式", example="markdown"
    )

    def get_all_vars(self) -> List[str]:
        """获取所有输出变量"""
        if not self.variables:
            return []

        vars_list = [
            self.variables.answer_var or FINAL_ANSWER_DEFAULT_VAR,
            self.variables.doc_retrieval_var or "doc_retrieval_res",
            self.variables.graph_retrieval_var or "graph_retrieval_res",
            self.variables.related_questions_var or "related_questions",
        ]

        # 过滤掉空值并添加其他变量
        result = [var for var in vars_list if var]
        if self.variables.other_vars:
            result.extend(self.variables.other_vars)

        return result

    def get_final_answer_var(self) -> str:
        """获取最终回答变量"""
        if self.variables:
            return self.variables.answer_var or FINAL_ANSWER_DEFAULT_VAR
        return FINAL_ANSWER_DEFAULT_VAR
