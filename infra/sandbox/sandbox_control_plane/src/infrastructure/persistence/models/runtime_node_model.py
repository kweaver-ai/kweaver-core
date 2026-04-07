"""
运行时节点 ORM 模型

SQLAlchemy 模型定义，用于数据库持久化。
按照数据表命名规范: t_{module}_{entity}, f_{field_name}
"""
from datetime import datetime
from decimal import Decimal

from sqlalchemy import Column, String, Integer, Text, Numeric, BigInteger, Index
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy import func

from src.infrastructure.persistence.database import Base


class RuntimeNodeModel(Base):
    """
    运行时节点 ORM 模型 - t_sandbox_runtime_node

    这是基础设施层的实现细节，映射到数据库表。
    """
    __tablename__ = "t_sandbox_runtime_node"

    # Primary fields
    f_node_id: Mapped[str] = mapped_column(String(40), primary_key=True)
    f_hostname: Mapped[str] = mapped_column(String(128), nullable=False, unique=True)

    # Runtime type
    f_runtime_type: Mapped[str] = mapped_column(String(20), nullable=False)

    # Network
    f_ip_address: Mapped[str] = mapped_column(String(45), nullable=False)
    f_api_endpoint: Mapped[str] = mapped_column(String(512), nullable=False, default="")

    # Status
    f_status: Mapped[str] = mapped_column(String(20), nullable=False, default="online")

    # Resources
    f_total_cpu_cores = Column(Numeric(5, 1), nullable=False)
    f_total_memory_mb = Column(Integer, nullable=False)
    f_allocated_cpu_cores = Column(Numeric(5, 1), nullable=False, default=0)
    f_allocated_memory_mb = Column(Integer, nullable=False, default=0)

    # Container capacity
    f_running_containers = Column(Integer, nullable=False, default=0)
    f_max_containers = Column(Integer, nullable=False)

    # Cache
    f_cached_images = Column(Text, nullable=False, default="")
    f_labels = Column(Text, nullable=False, default="")

    # Timestamps (BIGINT - millisecond timestamps)
    f_last_heartbeat_at = Column(BigInteger, nullable=False, default=0)
    f_created_at = Column(BigInteger, nullable=False, default=0)
    f_created_by = Column(String(40), nullable=False, default="")
    f_updated_at = Column(BigInteger, nullable=False, default=0)
    f_updated_by = Column(String(40), nullable=False, default="")
    f_deleted_at = Column(BigInteger, nullable=False, default=0)
    f_deleted_by = Column(String(36), nullable=False, default="")

    # Indexes
    __table_args__ = (
        Index("t_sandbox_runtime_node_uk_hostname_deleted_at", "f_hostname", "f_deleted_at", unique=True),
        Index("t_sandbox_runtime_node_idx_status", "f_status"),
        Index("t_sandbox_runtime_node_idx_runtime_type", "f_runtime_type"),
        Index("t_sandbox_runtime_node_idx_created_at", "f_created_at"),
        Index("t_sandbox_runtime_node_idx_deleted_at", "f_deleted_at"),
    )

    def to_runtime_node(self):
        """转换为领域 RuntimeNode 值对象"""
        from src.domain.services.scheduler import RuntimeNode

        # 计算资源使用率
        cpu_usage = (
            float(self.f_allocated_cpu_cores) / float(self.f_total_cpu_cores)
            if self.f_total_cpu_cores > 0
            else 0.0
        )
        mem_usage = (
            self.f_allocated_memory_mb / self.f_total_memory_mb
            if self.f_total_memory_mb > 0
            else 0.0
        )

        # 将状态映射到 RuntimeNode 的状态
        status_mapping = {
            "online": "healthy",
            "offline": "unhealthy",
            "draining": "draining",
            "maintenance": "unhealthy",
        }
        status = status_mapping.get(self.f_status, "unhealthy")

        return RuntimeNode(
            id=self.f_node_id,
            type=self.f_runtime_type,
            url=self.f_api_endpoint or f"http://{self.f_ip_address}:2375",
            status=status,
            cpu_usage=cpu_usage,
            mem_usage=mem_usage,
            session_count=self.f_running_containers,
            max_sessions=self.f_max_containers,
            cached_templates=self._parse_json(self.f_cached_images) or [],
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
