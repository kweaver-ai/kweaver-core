"""
模板 ORM 模型

SQLAlchemy 模型定义，用于数据库持久化。
按照数据表命名规范: t_{module}_{entity}, f_{field_name}
"""
from datetime import datetime
from decimal import Decimal

from sqlalchemy import Column, String, DateTime, Integer, Text, Numeric, BigInteger, Index
from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy import func

from src.infrastructure.persistence.database import Base


class TemplateModel(Base):
    """
    模板 ORM 模型 - t_sandbox_template

    这是基础设施层的实现细节，映射到数据库表。
    """
    __tablename__ = "t_sandbox_template"

    # Primary fields
    f_id: Mapped[str] = mapped_column(String(40), primary_key=True)
    f_name: Mapped[str] = mapped_column(String(128), nullable=False)
    f_description: Mapped[str] = mapped_column(String(500), nullable=False, default="")

    # Image configuration
    f_image_url: Mapped[str] = mapped_column(String(512), nullable=False)
    f_base_image: Mapped[str] = mapped_column(String(256), nullable=False, default="")

    # Runtime type
    f_runtime_type: Mapped[str] = mapped_column(String(30), nullable=False)

    # Default resource limits
    f_default_cpu_cores = Column(Numeric(3, 1), nullable=False, default=0.5)
    f_default_memory_mb = Column(Integer, nullable=False, default=512)
    f_default_disk_mb = Column(Integer, nullable=False, default=1024)
    f_default_timeout_sec = Column(Integer, nullable=False, default=300)

    # Pre-installed packages and environment
    f_pre_installed_packages = Column(Text, nullable=False, default="")
    f_default_env_vars = Column(Text, nullable=False, default="")

    # Security context
    f_security_context = Column(Text, nullable=False, default="")

    # Status
    f_is_active = Column(Integer, nullable=False, default=1)  # TINYINT mapped to Integer

    # Timestamps (BIGINT - millisecond timestamps)
    f_created_at = Column(BigInteger, nullable=False, default=0)
    f_created_by = Column(String(40), nullable=False, default="")
    f_updated_at = Column(BigInteger, nullable=False, default=0)
    f_updated_by = Column(String(40), nullable=False, default="")
    f_deleted_at = Column(BigInteger, nullable=False, default=0)
    f_deleted_by = Column(String(36), nullable=False, default="")

    # Indexes
    __table_args__ = (
        Index("t_sandbox_template_uk_name_deleted_at", "f_name", "f_deleted_at", unique=True),
        Index("t_sandbox_template_idx_runtime_type", "f_runtime_type"),
        Index("t_sandbox_template_idx_is_active", "f_is_active"),
        Index("t_sandbox_template_idx_created_at", "f_created_at"),
        Index("t_sandbox_template_idx_deleted_at", "f_deleted_at"),
    )

    def to_entity(self):
        """转换为领域实体"""
        from src.domain.entities.template import Template
        from src.domain.value_objects.resource_limit import ResourceLimit

        # 将数据库中的数字转换为带单位的格式
        def format_resource(value):
            """将数字转换为带单位的格式"""
            if isinstance(value, (int, float)):
                return f"{int(value)}Mi"
            return str(value)

        return Template(
            id=self.f_id,
            name=self.f_name,
            image=self.f_image_url,
            base_image=self.f_base_image or self.f_image_url,
            pre_installed_packages=self._parse_json(self.f_pre_installed_packages) or [],
            default_resources=ResourceLimit(
                cpu=str(self.f_default_cpu_cores),
                memory=format_resource(self.f_default_memory_mb),
                disk=format_resource(self.f_default_disk_mb),
                max_processes=128,  # Default value
            ),
            default_timeout_sec=self.f_default_timeout_sec,
            security_context=self._parse_json(self.f_security_context) or {},
            created_at=self._millis_to_datetime(self.f_created_at) or datetime.now(),
            updated_at=self._millis_to_datetime(self.f_updated_at) or datetime.now(),
        )

    @classmethod
    def from_entity(cls, template):
        """从领域实体创建 ORM 模型"""
        import re

        def parse_mb_value(value: str) -> int:
            """解析资源值（将 '512Mi', '1Gi' 等转换为 MB）"""
            if not value:
                return 512  # 默认值

            # 提取数字部分
            numeric_str = re.sub(r'[^0-9.]', '', value)
            if not numeric_str:
                return 512

            numeric = float(numeric_str)

            # 根据单位转换
            if 'Gi' in value or 'GB' in value or 'G' in value:
                return int(numeric * 1024)
            elif 'Mi' in value or 'MB' in value or 'M' in value:
                return int(numeric)
            elif 'Ki' in value or 'KB' in value or 'K' in value:
                return int(numeric / 1024)
            else:
                # 如果没有单位，假设是 MB
                return int(numeric)

        import json
        now_ms = int(datetime.now().timestamp() * 1000)

        return cls(
            f_id=template.id,
            f_name=template.name,
            f_description="",
            f_image_url=template.image,
            f_base_image=template.base_image,
            f_pre_installed_packages=json.dumps(template.pre_installed_packages, ensure_ascii=False),
            f_runtime_type="python3.11",  # Default, should be inferred
            f_default_cpu_cores=Decimal(template.default_resources.cpu),
            f_default_memory_mb=parse_mb_value(template.default_resources.memory),
            f_default_disk_mb=parse_mb_value(template.default_resources.disk),
            f_default_timeout_sec=template.default_timeout_sec,
            f_default_env_vars="",
            f_security_context=json.dumps(template.security_context, ensure_ascii=False),
            f_is_active=1,
            f_created_at=int(template.created_at.timestamp() * 1000) if template.created_at else now_ms,
            f_created_by="",
            f_updated_at=int(template.updated_at.timestamp() * 1000) if template.updated_at else now_ms,
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
