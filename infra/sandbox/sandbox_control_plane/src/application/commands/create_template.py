"""
创建模板命令

定义创建模板的命令 DTO。
"""
from dataclasses import dataclass
from typing import Optional, Dict


@dataclass
class CreateTemplateCommand:
    """创建模板命令"""
    template_id: str
    name: str
    image_url: str
    runtime_type: str
    default_cpu_cores: float
    default_memory_mb: int
    default_disk_mb: int
    default_timeout_sec: int
    default_env_vars: Optional[Dict[str, str]] = None
