"""
资源限制值对象

定义 CPU、内存、磁盘等资源限制。
"""
from dataclasses import dataclass
from typing import Self


@dataclass(frozen=True)
class ResourceLimit:
    """资源限制值对象（不可变）"""
    cpu: str  # 如 "1", "2", "0.5"
    memory: str  # 如 "512Mi", "1Gi", "2Gi"
    disk: str  # 如 "1Gi", "10Gi"
    max_processes: int = 128  # 最大进程数

    def __post_init__(self):
        """验证资源限制值"""
        if self.max_processes <= 0:
            raise ValueError("max_processes must be positive")

        # 验证 CPU 格式
        try:
            cpu_value = float(self.cpu)
        except ValueError:
            raise ValueError(f"Invalid cpu format: {self.cpu}")

        if cpu_value <= 0:
            raise ValueError("cpu must be positive")

        # 验证内存格式
        if not self._validate_size_format(self.memory):
            raise ValueError(f"Invalid memory format: {self.memory}")

        # 验证磁盘格式
        if not self._validate_size_format(self.disk):
            raise ValueError(f"Invalid disk format: {self.disk}")

    @staticmethod
    def _validate_size_format(size: str) -> bool:
        """验证大小格式（如 512Mi, 1Gi）"""
        if not size:
            return False
        if size[-2:] in {"Mi", "Gi"}:
            try:
                int(size[:-2])
                return True
            except ValueError:
                return False
        return False

    def with_cpu(self, cpu: str) -> Self:
        """返回新的 CPU 限制（不修改原对象）"""
        return ResourceLimit(
            cpu=cpu,
            memory=self.memory,
            disk=self.disk,
            max_processes=self.max_processes
        )

    def with_memory(self, memory: str) -> Self:
        """返回新的内存限制（不修改原对象）"""
        return ResourceLimit(
            cpu=self.cpu,
            memory=memory,
            disk=self.disk,
            max_processes=self.max_processes
        )

    @classmethod
    def default(cls) -> Self:
        """返回默认资源限制"""
        return cls(
            cpu="1",
            memory="512Mi",
            disk="1Gi",
            max_processes=128
        )
