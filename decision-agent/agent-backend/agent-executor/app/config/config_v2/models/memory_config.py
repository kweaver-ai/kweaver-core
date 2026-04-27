"""
记忆相关配置
"""

from dataclasses import dataclass


@dataclass
class MemoryConfig:
    """记忆相关配置"""

    # 召回记忆条数
    limit: int = 50

    # 召回记忆向量评分阈值
    threshold: float = 0.5

    # 召回记忆重排序评分阈值
    rerank_threshold: float = 0.1

    @classmethod
    def from_dict(cls, data: dict) -> "MemoryConfig":
        """从字典创建配置对象"""
        return cls(
            limit=int(data.get("limit", 50)),
            threshold=float(data.get("threshold", 0.5)),
            rerank_threshold=float(data.get("rerank_threshold", 0.1)),
        )
