"""
功能特性相关配置
"""

from dataclasses import dataclass


@dataclass
class FeaturesConfig:
    """特性开关配置"""

    # 是否使用explore_block v2版本
    use_explore_block_v2: bool = True

    # 是否禁用dolphin sdk缓存
    disable_dolphin_sdk_llm_cache: bool = False

    # 是否启用dolphin agent输出变量控制
    enable_dolphin_agent_output_variables_ctrl: bool = True

    # 技能agent调用是否需要返回progress
    is_skill_agent_need_progress: bool = False

    # 是否在 API tool proxy 请求中透传 TraceAI evidence 开关
    enable_traceai_evidence: bool = False

    # 是否开启 Skill 功能
    skill_enabled: bool = True

    @classmethod
    def from_dict(cls, data: dict) -> "FeaturesConfig":
        """从字典创建配置对象"""
        return cls(
            use_explore_block_v2=data.get("use_explore_block_v2", True),
            disable_dolphin_sdk_llm_cache=data.get(
                "disable_dolphin_sdk_llm_cache", False
            ),
            enable_dolphin_agent_output_variables_ctrl=data.get(
                "enable_dolphin_agent_output_variables_ctrl", True
            ),
            is_skill_agent_need_progress=data.get(
                "is_skill_agent_need_progress", False
            ),
            enable_traceai_evidence=data.get("enable_traceai_evidence", False),
            skill_enabled=data.get(
                "skill_enabled", True
            ),
        )
