"""
执行代码命令

定义执行代码的命令对象。
"""
from dataclasses import dataclass
from typing import Literal, Optional


@dataclass
class ExecuteCodeCommand:
    """执行代码命令"""
    session_id: str
    code: str
    language: Literal["python", "javascript", "shell"]
    async_mode: bool = False
    stdin: Optional[str] = None
    timeout: int = 30
    event_data: Optional[dict] = None

    def __post_init__(self):
        """初始化后验证"""
        if not self.code:
            raise ValueError("code cannot be empty")
        if self.timeout <= 0:
            raise ValueError("timeout must be positive")
        if self.language not in {"python", "javascript", "shell"}:
            raise ValueError(f"Unsupported language: {self.language}")
