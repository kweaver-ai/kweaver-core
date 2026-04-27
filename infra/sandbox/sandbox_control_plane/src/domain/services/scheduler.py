"""
调度器领域服务接口

定义调度器的抽象接口，负责选择最优运行时节点。
"""
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import List, Optional, TYPE_CHECKING

from src.domain.value_objects.resource_limit import ResourceLimit

if TYPE_CHECKING:
    from src.domain.value_objects.execution_request import ExecutionRequest


@dataclass
class RuntimeNode:
    """运行时节点值对象"""
    id: str
    type: str  # "docker" or "kubernetes"
    url: str  # 节点 API 地址
    status: str  # "healthy", "unhealthy", "draining"
    cpu_usage: float  # 0.0 - 1.0
    mem_usage: float  # 0.0 - 1.0
    session_count: int
    max_sessions: int
    cached_templates: List[str]

    def is_healthy(self) -> bool:
        """是否健康"""
        return self.status == "healthy"

    def get_load_ratio(self) -> float:
        """获取负载比率 (会话数/最大会话数)"""
        return self.session_count / self.max_sessions if self.max_sessions > 0 else 1.0

    def has_template(self, template_id: str) -> bool:
        """是否已缓存指定模板"""
        return template_id in self.cached_templates


@dataclass
class ScheduleRequest:
    """调度请求"""
    template_id: str
    resource_limit: ResourceLimit
    session_id: str | None = None


class IScheduler(ABC):
    """
    调度器接口

    这是领域层定义的 Port，由基础设施层实现 Adapter。
    """

    @abstractmethod
    async def schedule(self, request: ScheduleRequest) -> RuntimeNode:
        """
        调度会话到最优节点

        调度策略：
        1. 优先考虑模板亲和性（镜像已缓存）
        2. 使用负载均衡选择健康节点
        """
        pass

    @abstractmethod
    async def get_node(self, node_id: str) -> Optional[RuntimeNode]:
        """获取指定节点"""
        pass

    @abstractmethod
    async def get_healthy_nodes(self) -> List[RuntimeNode]:
        """获取所有健康节点"""
        pass

    @abstractmethod
    async def mark_node_unhealthy(self, node_id: str) -> None:
        """标记节点为不健康"""
        pass

    @abstractmethod
    async def execute(
        self,
        session_id: str,
        container_id: str,
        execution_request: "ExecutionRequest",
    ) -> str:
        """
        提交执行请求到容器内的执行器

        Args:
            session_id: 会话 ID
            container_id: 容器 ID
            execution_request: 执行请求

        Returns:
            execution_id: 执行任务 ID

        Raises:
            ConnectionError: 无法连接到执行器
            TimeoutError: 执行器响应超时
        """
        pass
