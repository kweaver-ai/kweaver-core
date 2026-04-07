# -*- coding:utf-8 -*-
from typing import Any, List, Optional
from enum import Enum

from pydantic import BaseModel, Field, validator


class DataSourceTypeEnum(str, Enum):
    """数据源类型枚举"""

    INHERIT_MAIN = "inherit_main"  # 继承主 Agent 数据源
    SELF_CONFIGURED = "self_configured"  # 使用自身配置


class SpecificInheritEnum(str, Enum):
    """继承范围枚举"""

    DOCS_ONLY = "docs_only"  # 仅继承文档数据源
    GRAPH_ONLY = "graph_only"  # 仅继承图谱数据源
    ALL = "all"  # 继承所有类型数据源


class LlmConfigTypeEnum(str, Enum):
    """大模型配置类型枚举"""

    INHERIT_MAIN = "inherit_main"  # 继承主 Agent 大模型
    SELF_CONFIGURED = "self_configured"  # 使用自身配置


class PmsCheckStatusEnum(str, Enum):
    """权限检查状态枚举"""

    EMPTY = ""  # 未检查
    SUCCESS = "success"  # 检查通过
    FAILED = "failed"  # 检查失败


class DataSourceConfigVo(BaseModel):
    """数据源配置"""

    class Config:
        # 序列化时使用枚举的字符串值而不是枚举对象
        use_enum_values = True

    type: DataSourceTypeEnum = Field(..., description="数据源来源")
    specific_inherit: Optional[SpecificInheritEnum] = Field(
        None, description="继承范围"
    )

    @validator("specific_inherit", pre=True)
    def validate_specific_inherit(cls, v):
        """处理空字符串，将其转换为None"""
        if v == "" or v is None:
            return None
        return v


class LlmConfigVo(BaseModel):
    """大模型配置"""

    type: LlmConfigTypeEnum = Field(..., description="大模型来源")


class AgentInputVo(BaseModel):
    """Agent输入参数配置"""

    enable: bool = Field(..., description="是否启用")
    input_name: str = Field(..., description="输入名称")
    input_type: str = Field(..., description="输入类型")
    map_type: str = Field(..., description="值类型")
    map_value: Optional[Any] = Field(None, description="值")
    input_desc: Optional[str] = Field(None, description="输入描述")


class AgentSkillInnerDto(BaseModel):
    """Agent技能内部数据传输对象，用于存储动态添加的属性"""

    class Config:
        # 允许额外字段
        extra = "allow"

    agent_info: Optional[dict] = Field(None, description="Agent信息")
    HOST_AGENT_EXECUTOR: Optional[str] = Field(None, description="Agent执行器主机地址")
    PORT_AGENT_EXECUTOR: Optional[str] = Field(None, description="Agent执行器端口")
    agent_options: Optional[dict] = Field(
        default_factory=dict, description="Agent选项配置"
    )


class AgentSkillVo(BaseModel):
    """Agent类型技能配置"""

    class Config:
        use_enum_values = True

    agent_key: str = Field(..., description="agent key")
    agent_version: Optional[str] = Field("latest", description="agent 版本")
    agent_input: Optional[List[AgentInputVo]] = Field(
        default_factory=list, description="输入参数配置"
    )
    intervention: Optional[bool] = Field(False, description="是否启用干预")
    intervention_confirmation_message: Optional[str] = Field(
        None, description="人工干预确认消息"
    )
    data_source_config: Optional[DataSourceConfigVo] = Field(
        None, description="数据源配置"
    )
    datasource_config: Optional[DataSourceConfigVo] = Field(
        None, description="数据源配置（兼容旧字段）"
    )
    llm_config: Optional[LlmConfigVo] = Field(None, description="大模型配置")
    current_pms_check_status: Optional[PmsCheckStatusEnum] = Field(
        PmsCheckStatusEnum.EMPTY, description="当前此技能的使用权限状态"
    )
    current_is_exists_and_published: Optional[bool] = Field(
        None, description="当前是否存在并已发布"
    )

    agent_timeout: int = Field(1800, description="agent 工具调用超时时间，默认1800s")

    # 内部使用的字段，用于存储动态添加的属性（如 agent_info、HOST_AGENT_EXECUTOR 等）
    inner_dto: AgentSkillInnerDto = Field(
        default_factory=AgentSkillInnerDto, description="内部数据传输对象，存储动态属性"
    )
