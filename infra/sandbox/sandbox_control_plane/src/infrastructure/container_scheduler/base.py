"""
容器调度器接口

定义容器操作的抽象接口。
"""
from abc import ABC, abstractmethod
from dataclasses import dataclass
from typing import Optional, Dict, Any


@dataclass
class ContainerConfig:
    """容器配置"""
    image: str
    name: str
    env_vars: Dict[str, str]
    cpu_limit: str  # 如 "1", "2"
    memory_limit: str  # 如 "512Mi", "1Gi"
    disk_limit: str  # 如 "1Gi", "10Gi"
    workspace_path: str  # S3路径，如 "s3://bucket/sessions/{session_id}/"
    labels: Dict[str, str]
    network_name: str = "sandbox_network"  # Docker 网络名称，默认 sandbox_network


@dataclass
class ContainerInfo:
    """容器信息"""
    id: str
    name: str
    image: str
    status: str  # created, running, paused, exited, deleting
    ip_address: Optional[str]
    created_at: str
    started_at: Optional[str]
    exited_at: Optional[str]
    exit_code: Optional[int]


@dataclass
class ContainerResult:
    """容器执行结果"""
    status: str
    stdout: str
    stderr: str
    exit_code: int


class IContainerScheduler(ABC):
    """
    容器调度器接口

    定义容器生命周期管理操作。
    """

    @abstractmethod
    async def create_container(
        self,
        config: ContainerConfig
    ) -> str:
        """
        创建容器

        返回容器ID
        """
        pass

    @abstractmethod
    async def start_container(self, container_id: str) -> None:
        """启动容器"""
        pass

    @abstractmethod
    async def stop_container(
        self,
        container_id: str,
        timeout: int = 10
    ) -> None:
        """停止容器"""
        pass

    @abstractmethod
    async def remove_container(
        self,
        container_id: str,
        force: bool = True
    ) -> None:
        """删除容器"""
        pass

    @abstractmethod
    async def get_container_status(
        self,
        container_id: str
    ) -> ContainerInfo:
        """获取容器状态"""
        pass

    @abstractmethod
    async def is_container_running(
        self,
        container_id: str
    ) -> bool:
        """
        检查容器是否正在运行

        直接通过 Docker API 查询，不依赖数据库。
        此方法供 StateSyncService 使用。

        Args:
            container_id: 容器 ID

        Returns:
            bool: 容器是否运行中
        """
        pass

    @abstractmethod
    async def get_container_logs(
        self,
        container_id: str,
        tail: int = 100,
        since: Optional[str] = None
    ) -> str:
        """获取容器日志"""
        pass

    @abstractmethod
    async def wait_container(
        self,
        container_id: str,
        timeout: Optional[int] = None
    ) -> ContainerResult:
        """等待容器执行完成"""
        pass

    @abstractmethod
    async def ping(self) -> bool:
        """检查调度器连接状态"""
        pass
