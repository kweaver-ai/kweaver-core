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
    install_timeout: int = 300

    def __post_init__(self):
        """初始化后验证。"""
        if self.install_timeout < 30 or self.install_timeout > 1800:
            raise ValueError("install_timeout must be between 30 and 1800 seconds")
