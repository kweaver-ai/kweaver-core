"""
更新模板命令

定义更新模板的命令 DTO。
"""
from dataclasses import dataclass
from typing import Optional, Dict


@dataclass
class UpdateTemplateCommand:
    """更新模板命令"""
    template_id: str
    name: Optional[str] = None
    image_url: Optional[str] = None
    default_cpu_cores: Optional[float] = None
    default_memory_mb: Optional[int] = None
    default_disk_mb: Optional[int] = None
    default_timeout_sec: Optional[int] = None
    default_env_vars: Optional[Dict[str, str]] = None
