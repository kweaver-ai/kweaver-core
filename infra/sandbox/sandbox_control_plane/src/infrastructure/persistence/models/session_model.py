"""
会话 ORM 模型

SQLAlchemy 模型定义，用于数据库持久化。
按照数据表命名规范: t_{module}_{entity}, f_{field_name}
"""
from datetime import datetime

from sqlalchemy import Column, String, Integer, BigInteger, Text, Index
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy import func

from src.infrastructure.persistence.database import Base
from src.shared.utils.dependencies import DEFAULT_PYTHON_PACKAGE_INDEX_URL


class SessionModel(Base):
    """
    会话 ORM 模型 - t_sandbox_session

    这是基础设施层的实现细节，映射到数据库表。
    按照照数据表命名规范实现。
    """
    __tablename__ = "t_sandbox_session"

    # Primary fields
    f_id: Mapped[str] = mapped_column(String(255), primary_key=True)
    f_template_id: Mapped[str] = mapped_column(String(40), nullable=False)

    # Status
    f_status: Mapped[str] = mapped_column(String(20), nullable=False, default="creating")

    # Runtime configuration
    f_runtime_type: Mapped[str] = mapped_column(String(20), nullable=False)

    # Runtime node and container
    f_runtime_node: Mapped[str] = mapped_column(String(128), nullable=False, default="")
    f_container_id: Mapped[str] = mapped_column(String(128), nullable=False, default="")
    f_pod_name: Mapped[str] = mapped_column(String(128), nullable=False, default="")

    # Workspace and resources
    f_workspace_path: Mapped[str] = mapped_column(String(256), nullable=False, default="")
    f_resources_cpu: Mapped[str] = mapped_column(String(16), nullable=False)
    f_resources_memory: Mapped[str] = mapped_column(String(16), nullable=False)
    f_resources_disk: Mapped[str] = mapped_column(String(16), nullable=False)

    # Environment and timeout
    f_env_vars = Column(Text, nullable=False, default="")
    f_timeout = Column(Integer, nullable=False, default=300)
    f_python_package_index_url: Mapped[str] = mapped_column(
        String(512),
        nullable=False,
        default=DEFAULT_PYTHON_PACKAGE_INDEX_URL,
    )

    # Timestamps (BIGINT - millisecond timestamps)
    f_last_activity_at = Column(BigInteger, nullable=False, default=0)
    f_completed_at = Column(BigInteger, nullable=False, default=0)

    # Dependency installation fields
    f_requested_dependencies = Column(Text, nullable=False, default="")
    f_installed_dependencies = Column(Text, nullable=False, default="")
    f_dependency_install_status: Mapped[str] = mapped_column(String(20), nullable=False, default="pending")
    f_dependency_install_error = Column(Text, nullable=False, default="")
    f_dependency_install_started_at = Column(BigInteger, nullable=False, default=0)
    f_dependency_install_completed_at = Column(BigInteger, nullable=False, default=0)

    # Audit fields
    f_created_at = Column(BigInteger, nullable=False, default=0)
    f_created_by = Column(String(40), nullable=False, default="")
    f_updated_at = Column(BigInteger, nullable=False, default=0)
    f_updated_by = Column(String(40), nullable=False, default="")
    f_deleted_at = Column(BigInteger, nullable=False, default=0)
    f_deleted_by = Column(String(36), nullable=False, default="")

    # Indexes
    __table_args__ = (
        Index("t_sandbox_session_idx_template_id", "f_template_id"),
        Index("t_sandbox_session_idx_status", "f_status"),
        Index("t_sandbox_session_idx_runtime_node", "f_runtime_node"),
        Index("t_sandbox_session_idx_last_activity_at", "f_last_activity_at"),
        Index("t_sandbox_session_idx_dependency_install_status", "f_dependency_install_status"),
        Index("t_sandbox_session_idx_created_at", "f_created_at"),
        Index("t_sandbox_session_idx_deleted_at", "f_deleted_at"),
        Index("t_sandbox_session_idx_created_by", "f_created_by"),
    )

    def to_entity(self):
        """转换为领域实体"""
        from src.domain.entities.session import Session, InstalledDependency
        from src.domain.value_objects.resource_limit import ResourceLimit
        from src.domain.value_objects.execution_status import SessionStatus

        session = Session(
            id=self.f_id,
            template_id=self.f_template_id,
            status=SessionStatus(self.f_status),
            resource_limit=ResourceLimit(
                cpu=self.f_resources_cpu,
                memory=self.f_resources_memory,
                disk=self.f_resources_disk,
                max_processes=128  # Default value
            ),
            workspace_path=self.f_workspace_path or "",
            runtime_type=self.f_runtime_type,
            runtime_node=self.f_runtime_node or None,
            container_id=self.f_container_id or None,
            pod_name=self.f_pod_name or None,
            env_vars=self._parse_json(self.f_env_vars) or {},
            timeout=self.f_timeout,
            created_at=self._millis_to_datetime(self.f_created_at) or datetime.now(),
            updated_at=self._millis_to_datetime(self.f_updated_at) or datetime.now(),
            completed_at=self._millis_to_datetime(self.f_completed_at),
            last_activity_at=self._millis_to_datetime(self.f_last_activity_at) or datetime.now(),
            # 依赖安装字段
            python_package_index_url=self.f_python_package_index_url or DEFAULT_PYTHON_PACKAGE_INDEX_URL,
            requested_dependencies=self._parse_json(self.f_requested_dependencies) or [],
            dependency_install_status=self.f_dependency_install_status or "pending",
            dependency_install_error=self.f_dependency_install_error or None,
            dependency_install_started_at=self._millis_to_datetime(
                self.f_dependency_install_started_at
            ),
            dependency_install_completed_at=self._millis_to_datetime(
                self.f_dependency_install_completed_at
            ),
        )

        # 转换 installed_dependencies JSON 为 InstalledDependency 对象列表
        if self.f_installed_dependencies:
            try:
                import json
                deps_list = json.loads(self.f_installed_dependencies)
                session.installed_dependencies = [
                    InstalledDependency(
                        name=dep.get("name"),
                        version=dep.get("version"),
                        install_location=dep.get("install_location", "/workspace/.venv/"),
                        install_time=datetime.fromisoformat(dep["install_time"]) if dep.get("install_time") else datetime.now(),
                        is_from_template=dep.get("is_from_template", False)
                    )
                    for dep in (deps_list if isinstance(deps_list, list) else [])
                ]
            except (json.JSONDecodeError, ValueError, TypeError):
                session.installed_dependencies = []

        return session

    @classmethod
    def from_entity(cls, session):
        """从领域实体创建 ORM 模型"""
        import json

        # 转换 installed_dependencies 对象列表为 JSON
        installed_dependencies_json = ""
        if session.installed_dependencies:
            try:
                deps_list = [
                    {
                        "name": dep.name,
                        "version": dep.version,
                        "install_location": dep.install_location,
                        "install_time": dep.install_time.isoformat(),
                        "is_from_template": dep.is_from_template,
                    }
                    for dep in session.installed_dependencies
                ]
                installed_dependencies_json = json.dumps(deps_list, ensure_ascii=False)
            except (TypeError, ValueError):
                installed_dependencies_json = ""

        now_ms = int(datetime.now().timestamp() * 1000)

        return cls(
            f_id=session.id,
            f_template_id=session.template_id,
            f_status=session.status.value,
            f_runtime_type=session.runtime_type,
            f_runtime_node=session.runtime_node or "",
            f_container_id=session.container_id or "",
            f_pod_name=session.pod_name or "",
            f_workspace_path=session.workspace_path,
            f_resources_cpu=session.resource_limit.cpu,
            f_resources_memory=session.resource_limit.memory,
            f_resources_disk=session.resource_limit.disk,
            f_env_vars=json.dumps(session.env_vars, ensure_ascii=False) if session.env_vars else "",
            f_timeout=session.timeout,
            f_python_package_index_url=session.python_package_index_url,
            f_completed_at=int(session.completed_at.timestamp() * 1000) if session.completed_at else 0,
            f_last_activity_at=int(session.last_activity_at.timestamp() * 1000) if session.last_activity_at else now_ms,
            # 依赖安装字段
            f_requested_dependencies=json.dumps(session.requested_dependencies, ensure_ascii=False) if session.requested_dependencies else "",
            f_installed_dependencies=installed_dependencies_json,
            f_dependency_install_status=session.dependency_install_status,
            f_dependency_install_error=session.dependency_install_error or "",
            f_dependency_install_started_at=int(session.dependency_install_started_at.timestamp() * 1000)
            if session.dependency_install_started_at else 0,
            f_dependency_install_completed_at=int(session.dependency_install_completed_at.timestamp() * 1000)
            if session.dependency_install_completed_at else 0,
            # 审计字段
            f_created_at=int(session.created_at.timestamp() * 1000) if session.created_at else now_ms,
            f_created_by="",
            f_updated_at=int(session.updated_at.timestamp() * 1000) if session.updated_at else now_ms,
            f_updated_by="",
            f_deleted_at=0,
            f_deleted_by="",
        )

    def _parse_json(self, value: str):
        """安全解析 JSON 字符串"""
        if not value or value.strip() == "":
            return None
        try:
            import json
            return json.loads(value)
        except (json.JSONDecodeError, ValueError):
            return None

    def _millis_to_datetime(self, millis: int):
        """将毫秒时间戳转换为 datetime"""
        if not millis or millis == 0:
            return None
        try:
            return datetime.fromtimestamp(millis / 1000)
        except (ValueError, OSError):
            return None
