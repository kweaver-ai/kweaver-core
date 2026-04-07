"""
文件制品值对象

定义执行生成的文件元数据。
"""
from dataclasses import dataclass
from datetime import datetime
from typing import Literal
from enum import Enum


class ArtifactType(str, Enum):
    """制品类型"""
    ARTIFACT = "artifact"  # 用户生成的文件
    LOG = "log"  # 日志文件
    OUTPUT = "output"  # 标准输出文件


@dataclass(frozen=True)
class Artifact:
    """文件制品值对象（不可变）"""
    path: str  # 相对于 workspace 的路径
    size: int  # 文件大小（字节）
    mime_type: str  # MIME 类型
    type: ArtifactType
    created_at: datetime
    checksum: str | None = None  # SHA256 校验和

    def __post_init__(self):
        """验证制品数据"""
        if self.size < 0:
            raise ValueError("size cannot be negative")
        if not self.path:
            raise ValueError("path cannot be empty")

    def is_log(self) -> bool:
        """是否为日志文件"""
        return self.type == ArtifactType.LOG

    def is_output(self) -> bool:
        """是否为输出文件"""
        return self.type == ArtifactType.OUTPUT

    @classmethod
    def create(
        cls,
        path: str,
        size: int,
        mime_type: str,
        type: Literal["artifact", "log", "output"] = "artifact",
        checksum: str | None = None
    ) -> "Artifact":
        """工厂方法：创建制品"""
        return cls(
            path=path,
            size=size,
            mime_type=mime_type,
            type=ArtifactType(type),
            created_at=datetime.now(),
            checksum=checksum
        )
