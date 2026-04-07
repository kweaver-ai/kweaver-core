"""
会话实体

表示一个沙箱会话，是聚合根。
"""
from dataclasses import dataclass, field
from datetime import datetime, timedelta
from typing import List, Optional

from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus
from src.domain.entities.execution import Execution
from src.shared.utils.dependencies import (
    DEFAULT_PYTHON_PACKAGE_INDEX_URL,
    merge_pip_specs,
    normalize_python_package_index_url,
)


@dataclass
class InstalledDependency:
    """
    已安装的依赖

    用于跟踪会话中实际安装的依赖包信息。
    按照 sandbox-design-v2.1.md 章节 5.6 设计。
    """
    name: str
    version: str
    install_location: str  # 如 "/workspace/.venv/"
    install_time: datetime
    is_from_template: bool  # 是否来自 Template 预装包


@dataclass
class Session:
    """
    会话实体

    聚合根，负责管理会话的生命周期和相关的执行记录。
    扩展支持依赖安装功能，按照 sandbox-design-v2.1.md 章节 5.6 设计。
    """
    id: str
    template_id: str
    status: SessionStatus
    resource_limit: ResourceLimit
    workspace_path: str
    runtime_type: str  # "docker" or "kubernetes"
    runtime_node: str | None = None
    container_id: str | None = None
    pod_name: str | None = None
    env_vars: dict = field(default_factory=dict)
    timeout: int = 300  # 默认 5 分钟
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)
    completed_at: datetime | None = None
    last_activity_at: datetime = field(default_factory=datetime.now)
    _executions: List[Execution] = field(default_factory=list)

    # 依赖安装相关字段（新增）
    requested_dependencies: List[str] = field(default_factory=list)
    installed_dependencies: List[InstalledDependency] = field(default_factory=list)
    dependency_install_status: str = "pending"  # pending/installing/completed/failed
    dependency_install_error: Optional[str] = None
    python_package_index_url: str = DEFAULT_PYTHON_PACKAGE_INDEX_URL
    dependency_install_started_at: datetime | None = None
    dependency_install_completed_at: datetime | None = None

    def __post_init__(self):
        """初始化后验证"""
        if self.timeout <= 0:
            raise ValueError("timeout must be positive")
        if not self.workspace_path:
            raise ValueError("workspace_path cannot be empty")
        self.python_package_index_url = normalize_python_package_index_url(
            self.python_package_index_url
        )

    # ============== 领域行为 ==============

    def mark_as_running(self, runtime_node: str, container_id: str) -> None:
        """标记会话为运行中"""
        if self.status != SessionStatus.CREATING:
            raise ValueError(f"Cannot mark session as running from status: {self.status}")

        self.status = SessionStatus.RUNNING
        self.runtime_node = runtime_node
        self.container_id = container_id
        self.updated_at = datetime.now()

    def mark_as_completed(self) -> None:
        """标记会话为已完成"""
        if self.status != SessionStatus.RUNNING:
            raise ValueError(f"Cannot mark session as completed from status: {self.status}")

        self.status = SessionStatus.COMPLETED
        self.completed_at = datetime.now()
        self.updated_at = datetime.now()

    def mark_as_failed(self) -> None:
        """标记会话为失败"""
        if self.status not in {SessionStatus.CREATING, SessionStatus.RUNNING}:
            raise ValueError(f"Cannot mark session as failed from status: {self.status}")

        self.status = SessionStatus.FAILED
        self.completed_at = datetime.now()
        self.updated_at = datetime.now()

    def mark_as_terminated(self) -> None:
        """终止会话"""
        if self.status == SessionStatus.TERMINATED:
            return  # 已经是终止状态

        self.status = SessionStatus.TERMINATED
        self.completed_at = datetime.now()
        self.updated_at = datetime.now()

    def update_last_activity(self) -> None:
        """更新最后活动时间"""
        self.last_activity_at = datetime.now()
        self.updated_at = datetime.now()

    # ============== 领域查询 ==============

    def is_active(self) -> bool:
        """是否为活跃状态"""
        return self.status in {
            SessionStatus.CREATING,
            SessionStatus.RUNNING
        }

    def is_terminated(self) -> bool:
        """是否已终止"""
        return self.status == SessionStatus.TERMINATED

    def is_idle(self, threshold_minutes: int = 30) -> bool:
        """是否空闲（超过阈值时间未活动）"""
        if not self.is_active():
            return False
        idle_time = datetime.now() - self.last_activity_at
        return idle_time > timedelta(minutes=threshold_minutes)

    def is_expired(self, max_hours: int = 6) -> bool:
        """是否过期（创建超过最大时间）"""
        age = datetime.now() - self.created_at
        return age > timedelta(hours=max_hours)

    def should_cleanup(self, idle_threshold_minutes: int = 30, max_lifetime_hours: int = 6) -> bool:
        """是否应该清理"""
        return self.is_idle(idle_threshold_minutes) or self.is_expired(max_lifetime_hours)

    # ============== 执行管理 ==============

    def add_execution(self, execution: Execution) -> None:
        """添加执行记录"""
        if execution.session_id != self.id:
            raise ValueError("Execution does not belong to this session")
        self._executions.append(execution)
        self.update_last_activity()

    def get_executions(self) -> List[Execution]:
        """获取所有执行记录"""
        return list(self._executions)

    def get_running_executions(self) -> List[Execution]:
        """获取正在运行的执行"""
        return [e for e in self._executions if e.is_running()]

    # ============== 依赖管理 ==============

    def replace_requested_dependencies(
        self, index_url: Optional[str], dependencies: List[str]
    ) -> None:
        """全量替换目标依赖配置。"""
        self.python_package_index_url = normalize_python_package_index_url(index_url)
        self.requested_dependencies = list(dependencies)
        self.updated_at = datetime.now()

    def merge_requested_dependencies(
        self, index_url: Optional[str], dependencies: List[str]
    ) -> None:
        """增量合并目标依赖配置。"""
        self.python_package_index_url = normalize_python_package_index_url(
            index_url or self.python_package_index_url
        )
        self.requested_dependencies = merge_pip_specs(
            self.requested_dependencies,
            dependencies,
        )
        self.updated_at = datetime.now()

    def set_dependencies_installing(self) -> None:
        """兼容旧接口，标记依赖安装中。"""
        self.mark_dependency_installing()

    def mark_dependency_installing(self, started_at: datetime | None = None) -> None:
        """标记依赖安装中。"""
        self.dependency_install_status = "installing"
        self.dependency_install_started_at = started_at or datetime.now()
        self.dependency_install_completed_at = None
        self.dependency_install_error = None
        self.updated_at = datetime.now()

    def set_dependencies_completed(self, installed: List[InstalledDependency]) -> None:
        """兼容旧接口，标记依赖安装完成。"""
        self.mark_dependency_install_completed(installed)

    def mark_dependency_install_completed(
        self,
        installed: List[InstalledDependency],
        completed_at: datetime | None = None,
    ) -> None:
        """标记依赖安装完成。"""
        self.dependency_install_status = "completed"
        self.installed_dependencies = installed
        self.dependency_install_error = None
        self.dependency_install_completed_at = completed_at or datetime.now()
        self.updated_at = datetime.now()

    def set_dependencies_failed(self, error: str) -> None:
        """兼容旧接口，标记依赖安装失败。"""
        self.mark_dependency_install_failed(error)

    def mark_dependency_install_failed(
        self,
        error: str,
        completed_at: datetime | None = None,
    ) -> None:
        """标记依赖安装失败。"""
        self.dependency_install_status = "failed"
        self.dependency_install_error = error
        self.dependency_install_completed_at = completed_at or datetime.now()
        self.updated_at = datetime.now()

    def has_dependencies(self) -> bool:
        """是否有依赖需要安装"""
        return len(self.requested_dependencies) > 0

    def is_dependency_install_pending(self) -> bool:
        """依赖是否正在安装或待安装"""
        return self.dependency_install_status in ("pending", "installing")

    def is_dependency_install_successful(self) -> bool:
        """依赖是否安装成功"""
        return self.dependency_install_status == "completed"
