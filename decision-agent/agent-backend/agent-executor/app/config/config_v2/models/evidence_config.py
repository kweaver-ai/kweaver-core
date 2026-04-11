"""
证据注入相关配置
"""

from dataclasses import dataclass


@dataclass
class EvidenceConfig:
    """证据注入配置"""

    # 证据注入功能总开关
    enable: bool = False

    # EvidenceStore 配置
    store_max_size: int = 1000
    store_ttl_seconds: int = 3600

    # 规则匹配配置
    enable_alias_match: bool = True
    min_sentence_length: int = 10

    # LLM 标注配置（文本→证据位置）
    llm_annotation_timeout: int = 30
    llm_annotation_model: str = "deepseek-v3.2"

    # 仅标注最终答案（避免流式输出卡顿）
    # True: 只对最终答案进行 LLM 标注，中间输出快速通过
    # False: 对每个 LLM 输出都进行标注（会阻塞流式输出）
    annotate_only_final_answer: bool = True

    @classmethod
    def from_dict(cls, data: dict) -> "EvidenceConfig":
        """从字典创建配置对象"""
        return cls(
            enable=data.get("enable", False),
            store_max_size=int(data.get("store_max_size", 1000)),
            store_ttl_seconds=int(data.get("store_ttl_seconds", 3600)),
            enable_alias_match=data.get("enable_alias_match", True),
            min_sentence_length=int(data.get("min_sentence_length", 10)),
            llm_annotation_timeout=int(data.get("llm_annotation_timeout", 30)),
            llm_annotation_model=data.get("llm_annotation_model", "deepseek-v3.2"),
            annotate_only_final_answer=data.get("annotate_only_final_answer", True),
        )
