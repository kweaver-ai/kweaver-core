# -*- coding:utf-8 -*-
from typing import Any, Dict, List, Optional

from pydantic import BaseModel, validator, Field

from app.utils.snow_id import snow_id
from .agent_config_vos import SkillVo, OutputConfigVo, ConfigMetadataVo


class AgentConfigVo(BaseModel):
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

    output: OutputConfigVo = Field(
        default_factory=lambda: OutputConfigVo(default_format="markdown")
    )
    memory: Optional[Dict[str, Any]] = {}
    related_question: Optional[Dict[str, Any]] = {}

    # plan_mode 任务规划模式
    plan_mode: Optional[Dict[str, bool]] = None
    # config metadata
    metadata: Optional[ConfigMetadataVo] = None

    # agent-app 传入 (也改为由_options参数传入 2025年10月19日)
    agent_id: Optional[str] = None
    conversation_id: Optional[str] = None
    agent_run_id: Optional[str] = None

    # agent-executor 由_options参数传入，在调用agent工具时使用
    output_vars: Optional[List[str]] = None
    incremental_output: bool = False

    # 通过请求参数传入
    agent_version: Optional[str] = None

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

    @validator("output", pre=True, always=True)
    def validate_output(cls, v):
        """验证output字段"""
        if v is None:
            return OutputConfigVo(default_format="markdown")
        # 如果已经是OutputConfigVo对象，直接返回
        if isinstance(v, OutputConfigVo):
            return v
        # 如果是字典，转换为OutputConfigVo对象
        if isinstance(v, dict):
            return OutputConfigVo(**v)
        return OutputConfigVo(default_format="markdown")

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

    @validator("metadata", pre=True, always=True)
    def validate_metadata(cls, v):
        """验证metadata字段，确保符合schema定义"""
        if v is None:
            return ConfigMetadataVo()

        # 如果已经是ConfigMetadataVo对象，直接返回
        if isinstance(v, ConfigMetadataVo):
            return v

        # 如果是字典，转换为ConfigMetadataVo对象
        if isinstance(v, dict):
            return ConfigMetadataVo(**v)

        # 其他情况返回默认对象
        return ConfigMetadataVo()

    def is_plan_mode(self) -> bool:
        return self.plan_mode and self.plan_mode.get("is_enabled", False)

    def append_task_plan_agent(self):
        if self.is_plan_mode():
            from .agent_config_vos import (
                AgentSkillVo,
                AgentInputVo,
                DataSourceConfigVo,
                LlmConfigVo,
            )
            from .agent_config_vos.agent_skill_vo import (
                DataSourceTypeEnum,
                LlmConfigTypeEnum,
            )

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

    def get_config_last_set_timestamp(self) -> int:
        if self.metadata and self.metadata.config_last_set_timestamp:
            return self.metadata.config_last_set_timestamp
        return 0
