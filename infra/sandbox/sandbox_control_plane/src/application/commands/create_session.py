"""
创建会话命令

定义创建会话的命令对象。
扩展支持依赖安装，按照 sandbox-design-v2.1.md 章节 5 设计。
"""
from dataclasses import dataclass, field
from typing import Dict, List, Optional

from src.domain.value_objects.resource_limit import ResourceLimit
from src.shared.utils.dependencies import DEFAULT_PYTHON_PACKAGE_INDEX_URL


@dataclass
class CreateSessionCommand:
    """
    创建会话命令

    扩展支持 Python 依赖安装功能。
    """
    template_id: str
    timeout: int = 300
    resource_limit: ResourceLimit | None = None
    env_vars: Dict[str, str] | None = None
    id: Optional[str] = None  # 手动指定会话 ID（可选）

    # 依赖安装相关字段（新增）
    dependencies: List[str] = field(default_factory=list)
    install_timeout: int = 300
    fail_on_dependency_error: bool = True
    allow_version_conflicts: bool = False
    python_package_index_url: str = DEFAULT_PYTHON_PACKAGE_INDEX_URL

    def __post_init__(self):
        """初始化后验证"""
        if self.timeout <= 0:
            raise ValueError("timeout must be positive")

        # 设置默认值
        if self.resource_limit is None:
            self.resource_limit = ResourceLimit.default()
        if self.env_vars is None:
            self.env_vars = {}

        # 验证安装超时
        if self.install_timeout < 30 or self.install_timeout > 1800:
            raise ValueError("install_timeout must be between 30 and 1800 seconds")
