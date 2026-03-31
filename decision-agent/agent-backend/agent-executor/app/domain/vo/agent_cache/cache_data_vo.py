# -*- coding:utf-8 -*-
from dataclasses import dataclass
from typing import Any, Dict


@dataclass
class CacheDataVo:
    """缓存数据VO

    保存创建DolphinAgent所需的静态数据
    """

    # 创建DolphinAgent所需的静态数据
    agent_config: Dict[str, Any]  # Agent配置
    tools_info_dict: Dict[str, Any]  # 工具信息
    skill_agent_info_dict: Dict[str, Any]  # SkillAgent信息
    llm_config_dict: Dict[str, Any]  # LLM配置

    def __init__(self):
        self.agent_config = {}
        self.tools_info_dict = {}
        self.skill_agent_info_dict = {}
        self.llm_config_dict = {}
