"""
外部LLM配置（非模型工厂）
"""

from dataclasses import dataclass, field
from typing import Dict, Any


@dataclass
class OuterLLMConfig:
    """外部LLM配置（非模型工厂）"""

    api: str = ""
    api_key: str = ""
    model: str = ""
    model_list: Dict[str, Any] = field(default_factory=dict)

    @classmethod
    def from_dict(cls, data: dict) -> "OuterLLMConfig":
        """从字典创建配置对象"""
        model_list_data = data.get("model_list", {})
        if not isinstance(model_list_data, dict):
            model_list_data = {}

        return cls(
            api=data.get("api", ""),
            api_key=data.get("api_key", ""),
            model=data.get("model", ""),
            model_list=model_list_data,
        )
