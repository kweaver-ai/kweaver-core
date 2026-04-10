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

    # ===== 证据注入功能配置 =====

    # 证据注入功能总开关
    enable_evidence_injection: bool = False

    # EvidenceStore 配置
    evidence_store_max_size: int = 1000
    evidence_store_ttl_seconds: int = 3600

    # LLM 提取配置
    llm_extraction_timeout: int = 30
    llm_extraction_model: str = ""

    # 规则匹配配置
    enable_alias_match: bool = True
    min_sentence_length: int = 10

    # LLM 标注配置
    llm_annotation_timeout: int = 30
    llm_annotation_model: str = "deepseek-v3.2"
    # 仅标注最终答案（避免流式输出卡顿）
    # True: 只对最终答案进行 LLM 标注，中间输出快速通过
    # False: 对每个 LLM 输出都进行标注（会阻塞流式输出）
    annotate_only_final_answer: bool = True

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
            # 证据注入配置
            enable_evidence_injection=data.get(
                "enable_evidence_injection", False
            ),
            evidence_store_max_size=int(data.get(
                "evidence_store_max_size", 1000
            )),
            evidence_store_ttl_seconds=int(data.get(
                "evidence_store_ttl_seconds", 3600
            )),
            llm_extraction_timeout=int(data.get(
                "llm_extraction_timeout", 30
            )),
            llm_extraction_model=data.get(
                "llm_extraction_model", ""
            ),
            enable_alias_match=data.get(
                "enable_alias_match", True
            ),
            min_sentence_length=int(data.get(
                "min_sentence_length", 10
            )),
            llm_annotation_timeout=int(data.get(
                "llm_annotation_timeout", 30
            )),
            llm_annotation_model=data.get(
                "llm_annotation_model", "deepseek-v3.2"
            ),
            annotate_only_final_answer=data.get(
                "annotate_only_final_answer", True
            ),
        )
