"""
会话数据传输对象

用于应用层与接口层之间的数据传输。
"""
from dataclasses import dataclass
from datetime import datetime
from typing import Optional

from src.shared.utils.dependencies import parse_pip_spec


@dataclass
class SessionDTO:
    """会话数据传输对象"""
    id: str
    template_id: str
    status: str
    resource_limit: dict
    workspace_path: str
    runtime_type: str
    runtime_node: Optional[str] = None
    container_id: Optional[str] = None
    pod_name: Optional[str] = None
    env_vars: dict = None
    timeout: int = 300
    language_runtime: str = "python3.11"
    python_package_index_url: str = "https://pypi.org/simple/"
    requested_dependencies: list[dict] = None
    installed_dependencies: list[dict] = None
    dependency_install_status: str = "pending"
    dependency_install_error: Optional[str] = None
    dependency_install_started_at: Optional[datetime] = None
    dependency_install_completed_at: Optional[datetime] = None
    created_at: datetime = None
    updated_at: datetime = None
    completed_at: Optional[datetime] = None
    last_activity_at: datetime = None

    def __post_init__(self):
        """初始化默认值"""
        if self.env_vars is None:
            self.env_vars = {}
        if self.created_at is None:
            self.created_at = datetime.now()
        if self.updated_at is None:
            self.updated_at = datetime.now()
        if self.last_activity_at is None:
            self.last_activity_at = datetime.now()
        if self.requested_dependencies is None:
            self.requested_dependencies = []
        if self.installed_dependencies is None:
            self.installed_dependencies = []

    @classmethod
    def from_entity(cls, session) -> "SessionDTO":
        """从领域实体创建 DTO"""
        return cls(
            id=session.id,
            template_id=session.template_id,
            status=session.status.value,
            resource_limit={
                "cpu": session.resource_limit.cpu,
                "memory": session.resource_limit.memory,
                "disk": session.resource_limit.disk,
                "max_processes": session.resource_limit.max_processes,
            },
            workspace_path=session.workspace_path,
            runtime_type=session.runtime_type,
            runtime_node=session.runtime_node,
            container_id=session.container_id,
            pod_name=session.pod_name,
            env_vars=dict(session.env_vars),
            timeout=session.timeout,
            language_runtime=session.runtime_type,
            python_package_index_url=session.python_package_index_url,
            requested_dependencies=[
                parse_pip_spec(dep) for dep in session.requested_dependencies
            ],
            installed_dependencies=[
                {
                    "name": dep.name,
                    "version": dep.version,
                    "install_location": dep.install_location,
                    "install_time": dep.install_time,
                    "is_from_template": dep.is_from_template,
                }
                for dep in session.installed_dependencies
            ],
            dependency_install_status=session.dependency_install_status,
            dependency_install_error=session.dependency_install_error,
            dependency_install_started_at=session.dependency_install_started_at,
            dependency_install_completed_at=session.dependency_install_completed_at,
            created_at=session.created_at,
            updated_at=session.updated_at,
            completed_at=session.completed_at,
            last_activity_at=session.last_activity_at,
        )
