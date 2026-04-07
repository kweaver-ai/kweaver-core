"""
执行请求值对象

表示提交给沙箱执行器的代码执行请求。
"""
from dataclasses import dataclass
from typing import Dict, Any, Optional


@dataclass
class ExecutionRequest:
    """
    执行请求值对象

    包含执行代码所需的所有信息。
    """

    code: str
    language: str
    event: Dict[str, Any]
    timeout: int
    env_vars: Dict[str, str]
    execution_id: Optional[str] = None
    session_id: Optional[str] = None

    def __post_init__(self):
        """验证执行请求"""
        if not self.code:
            raise ValueError("code cannot be empty")

        if not self.language:
            raise ValueError("language cannot be empty")

        if self.timeout < 1 or self.timeout > 3600:
            raise ValueError("timeout must be between 1 and 3600 seconds")

        if self.language not in ("python", "javascript", "shell"):
            raise ValueError(f"unsupported language: {self.language}")
