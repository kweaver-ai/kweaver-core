"""
增量安装会话依赖命令。
"""
from dataclasses import dataclass
from typing import Optional


@dataclass
class InstallSessionDependenciesCommand:
    """增量安装会话依赖命令。"""

    session_id: str
    dependencies: list[str]
    python_package_index_url: Optional[str] = None
