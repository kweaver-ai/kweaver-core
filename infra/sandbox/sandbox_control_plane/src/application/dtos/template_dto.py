"""
模板 DTO

定义模板数据传输对象。
"""
from dataclasses import dataclass
from datetime import datetime
from typing import Optional, Dict

from src.domain.entities.template import Template


@dataclass
class TemplateDTO:
    """模板数据传输对象"""
    id: str
    name: str
    image_url: str
    runtime_type: str
    default_cpu_cores: float
    default_memory_mb: int
    default_disk_mb: int
    default_timeout_sec: int
    default_env_vars: Optional[Dict[str, str]] = None
    is_active: bool = True
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None

    @classmethod
    def from_entity(cls, template: Template) -> "TemplateDTO":
        """从领域实体创建 DTO"""
        import re

        def parse_resource(value: str | None, default: int, resource_type: str) -> int:
            """Parse resource value (handles '512Mi', '1Gi', etc.)"""
            if not value:
                return default

            # Remove any non-numeric characters except decimal point
            numeric_str = re.sub(r'[^0-9.]', '', value)
            if not numeric_str:
                return default

            numeric = float(numeric_str)

            # Convert to MB based on unit
            if 'Gi' in value or 'GB' in value or 'G' in value:
                return int(numeric * 1024)
            elif 'Mi' in value or 'MB' in value or 'M' in value:
                return int(numeric)
            elif 'Ki' in value or 'KB' in value or 'K' in value:
                return int(numeric / 1024)
            else:
                # Assume MB if no unit
                return int(numeric)

        return cls(
            id=template.id,
            name=template.name,
            image_url=template.image,  # Map image to image_url
            runtime_type="python3.11",  # Default, should be from entity if available
            default_cpu_cores=float(template.default_resources.cpu) if template.default_resources.cpu else 0.5,
            default_memory_mb=parse_resource(template.default_resources.memory, 512, "memory"),
            default_disk_mb=parse_resource(template.default_resources.disk, 1024, "disk"),
            default_timeout_sec=template.default_timeout_sec,
            default_env_vars=None,  # Not in entity
            is_active=True,  # Default, not in entity
            created_at=template.created_at,
            updated_at=template.updated_at,
        )

    def to_dict(self) -> dict:
        """转换为字典"""
        return {
            "id": self.id,
            "name": self.name,
            "image_url": self.image_url,
            "runtime_type": self.runtime_type,
            "default_cpu_cores": self.default_cpu_cores,
            "default_memory_mb": self.default_memory_mb,
            "default_disk_mb": self.default_disk_mb,
            "default_timeout_sec": self.default_timeout_sec,
            "default_env_vars": self.default_env_vars,
            "is_active": self.is_active,
            "created_at": self.created_at.isoformat() if self.created_at else None,
            "updated_at": self.updated_at.isoformat() if self.updated_at else None,
        }
