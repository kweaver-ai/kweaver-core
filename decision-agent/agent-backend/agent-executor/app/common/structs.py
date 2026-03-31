# -*- coding:utf-8 -*-
from dataclasses import dataclass, field
from typing import Any, Dict, List, Optional, Union

from pydantic import BaseModel, Field, validator

from app.utils.snow_id import snow_id
from app.domain.vo.agentvo.agent_config_vos import (
    SkillVo,
    AgentSkillVo,
    AgentInputVo,
    DataSourceConfigVo,
    LlmConfigVo,
)
from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
    DataSourceTypeEnum,
    LlmConfigTypeEnum,
)


@dataclass
class LogicBlock:
    id: str = None
    name: str = None
    type: str = None  # retriever_block llm_block
    output: str = None
    llm_config: dict = None


@dataclass
class AugmentBlock:
    input: list = field(default_factory=list)
    augment_data_source: dict = field(default_factory=dict)
    need_augment_content: bool = False
    augment_entities: dict = field(default_factory=dict)


@dataclass
class RetrieverBlock(LogicBlock):
    input: str = None
    headers_info: dict = field(default_factory=dict)  # Young:透传AS的身份信息
    body: dict = field(default_factory=dict)  # Young:透传AS的请求体
    data_source: dict = field(default_factory=dict)
    augment_data_source: dict = field(
        default_factory=dict
    )  # Young:query增强的concept数据

    processed_query: dict = field(default_factory=dict)  # Young:query处理后得到的结果
    retrival_slices: dict = field(default_factory=dict)  # Young:保存召回原始切片
    rank_slices: dict = field(default_factory=dict)  # Young:保存精排之后的排序切片

    rank_rough_slices: dict = field(default_factory=dict)
    rank_rough_slices_num: dict = field(default_factory=dict)

    rank_accurate_slices: dict = field(default_factory=dict)
    rank_accurate_slices_num: dict = field(default_factory=dict)

    snippets_slices: dict = field(default_factory=dict)
    cites_slices: dict = field(default_factory=dict)  # Young:保存cite拼接结果
    format_out: list = field(default_factory=list)

    faq_retrival_qas: list = field(default=list)
    faq_rank_qas: list = field(default=list)
    faq_find_answer: bool = False
    faq_format_out_qas: Union[list, dict] = field(default_factory=list)

    security_token: set = field(
        default_factory=set
    )  # Feature-736016 百胜召回支持外置后过滤功能
    """ 召回后会返回security_token，在后续调用大模型时将security_token作为header传给模型工厂 """


class AgentConfig(BaseModel):
    """Agent配置模型"""

    # agent-factory & 前端页面 传入
    input: Optional[Dict[str, Any]] = None
    llms: Optional[List[Dict[str, Any]]] = None

    skills: Optional[SkillVo] = None

    data_source: Optional[Dict[str, Any]] = {}

    system_prompt: Optional[str] = None
    is_dolphin_mode: bool = False
    dolphin: Optional[str] = None

    pre_dolphin: Optional[List[Dict[str, Any]]] = []
    post_dolphin: Optional[List[Dict[str, Any]]] = []

    output: Optional[Dict[str, Any]] = {}
    memory: Optional[Dict[str, Any]] = {}
    related_question: Optional[Dict[str, Any]] = {}

    # plan_mode 任务规划模式
    plan_mode: Optional[Dict[str, bool]] = None

    # agent-app 传入
    agent_id: Optional[str] = None
    conversation_id: Optional[str] = None
    session_id: Optional[str] = None

    # agent-executor 由_options参数传入，在调用agent工具时使用
    output_vars: Optional[List[str]] = None
    incremental_output: bool = False

    @validator("skills", pre=True, always=True)
    def validate_skills(cls, v):
        """验证skills字段，如果传入null则转换为SkillVo对象"""
        if v is None:
            return SkillVo()

        # 如果已经是SkillVo对象，直接返回
        if isinstance(v, SkillVo):
            return v

        # 如果是字典，转换为SkillVo对象
        if isinstance(v, dict):
            return SkillVo(**v)

        return SkillVo()

    @validator("conversation_id", pre=True, always=True)
    def set_conversation_id(cls, v):
        """如果conversation_id为空，则自动生成"""
        if not v:
            return str(snow_id())

        return v

    @validator("pre_dolphin", "post_dolphin", pre=True, always=True)
    def handle_none(cls, v):
        """处理None, 将 None 转换为 []"""
        return [] if v is None else v

    def is_plan_mode(self) -> bool:
        return self.plan_mode and self.plan_mode.get("is_enabled", False)

    def append_task_plan_agent(self):
        if self.is_plan_mode():
            # 确保skills不为None
            if self.skills is None:
                self.skills = SkillVo()

            # 创建Task_Plan_Agent配置
            task_plan_agent = AgentSkillVo(
                agent_key="Task_Plan_Agent",
                agent_version="latest",
                agent_input=[
                    AgentInputVo(
                        input_name="query",
                        input_type="string",
                        map_type="auto",
                        enable=True,
                        input_desc="用户输入的问题",
                    )
                ],
                intervention=False,
                data_source_config=DataSourceConfigVo(
                    type=DataSourceTypeEnum.SELF_CONFIGURED
                ),
                llm_config=LlmConfigVo(type=LlmConfigTypeEnum.SELF_CONFIGURED),
            )

            self.skills.agents.append(task_plan_agent)


class AgentOptions(BaseModel):
    """Agent运行选项模型(由agent-executor定义)"""

    output_vars: Optional[List[str]] = None
    incremental_output: Optional[bool] = None
    data_source: Optional[Dict[str, Any]] = None
    llm_config: Optional[Dict[str, Any]] = None
    tmp_files: Optional[List] = None


class AgentInput(BaseModel):
    """Agent输入模型"""

    query: str
    history: Optional[List[Dict[str, str]]] = None
    tool: Optional[Dict[str, Any]] = Field(default_factory=dict)
    header: Optional[Dict[str, Any]] = Field(default_factory=dict)
    self_config: Optional[Dict[str, Any]] = Field(default_factory=dict)

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


DEFAULT_INPUTS = ["history", "tool", "header", "self_config"]

if __name__ == "__main__":
    inputs = {"query": "爱数是一家怎样的公司？", "json_input": "{}"}
    agent_input = AgentInput(**inputs)
    dump = agent_input.model_dump()
    print(dump)
