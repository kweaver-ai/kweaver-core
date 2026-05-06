"""
REST API 请求模式

定义 FastAPI 的请求 Pydantic 模型。
"""

import re
from pathlib import PurePosixPath
from pydantic import BaseModel, ConfigDict, Field, field_validator
from typing import Literal, Optional, Dict, List


class DependencySpec(BaseModel):
    """
    依赖包规范

    用于指定会话创建时需要安装的 Python 包。
    按照 sandbox-design-v2.1.md 章节 5.3.1 设计。
    """

    name: str = Field(..., min_length=1, max_length=100, description="包名称")
    version: Optional[str] = Field(None, description="版本约束 (如: ==2.31.0, >=1.0)")

    @field_validator("name")
    @classmethod
    def validate_package_name(cls, v: str) -> str:
        """
        验证包名格式

        禁止：
        - 路径穿越字符 (..)
        - 绝对路径 (/)
        - URL (://)
        - 非法字符（仅允许字母、数字、._-）
        - 版本号混合在包名中 (如 pandas2.3.3 应该是 name="pandas", version="==2.3.3")
        """
        # 禁止路径穿越
        if ".." in v or v.startswith("/"):
            raise ValueError("Package name cannot contain path traversal characters")
        # 禁止 URL
        if "://" in v:
            raise ValueError("Package name cannot contain URL")
        # PyPI 包名规范：仅允许字母、数字、._-
        if not re.match(r"^[a-zA-Z0-9._-]+$", v):
            raise ValueError("Invalid package name format")

        # 检测常见的版本号混合错误（如 pandas2.3.3, numpy1.24.0）
        # 模式：字母开头 + 数字 + 点 + 数字（可能是版本号）
        if re.match(r"^[a-zA-Z]+[0-9]+\.[0-9]", v):
            # 提取包名和版本号用于错误提示
            package_name = re.sub(r"[0-9]+\.[0-9].*", "", v)
            version_num = re.sub(r"^[a-zA-Z]+", "", v)
            raise ValueError(
                f"Invalid package name '{v}'. It looks like a version number is mixed with the package name. "
                f"Use separate 'name' and 'version' fields: "
                f'{{"name": "{package_name}", "version": "=={version_num}"}}'
            )

        return v

    def to_pip_spec(self) -> str:
        """
        转换为 pip 安装规范

        自动为没有操作符的版本号添加 == 前缀。
        例如：version="2.3.3" 会被转换为 "==2.3.3"

        Returns:
            pip 包规范字符串，如 "requests==2.31.0" 或 "pandas"
        """
        if self.version:
            # 检查 version 是否以操作符开头
            # pip 支持的操作符: ==, >=, <=, >, <, ~=, !=, ~, =
            version_operators = ("==", ">=", "<=", ">", "<", "~=", "!=", "~", "=")
            if not self.version.startswith(version_operators):
                # 如果没有操作符，默认添加 ==
                return f"{self.name}=={self.version}"
            return f"{self.name}{self.version}"
        return self.name


class CreateSessionRequest(BaseModel):
    """
    创建会话请求

    按照 sandbox-design-v2.1.md 章节 5.3.1 设计，扩展支持依赖安装。
    """

    id: Optional[str] = Field(
        None, min_length=1, max_length=64, description="会话 ID（可选，手动指定时需确保唯一性）"
    )
    template_id: Optional[str] = Field(
        None, min_length=1, max_length=64, description="模板 ID；未传时使用默认模板配置"
    )
    timeout: int = Field(300, ge=1, le=3600, description="超时时间（秒）")
    cpu: str = Field("1", description="CPU 核心数")
    memory: str = Field("512Mi", description="内存限制")
    disk: str = Field("1Gi", description="磁盘限制")
    env_vars: Dict[str, str] = Field(default_factory=dict, description="环境变量")
    event: Optional[Dict] = Field(None, description="事件数据")

    # 依赖安装相关字段（新增）
    dependencies: List[DependencySpec] = Field(
        default_factory=list, max_length=50, description="会话级依赖包列表"
    )
    install_timeout: int = Field(300, ge=30, le=1800, description="依赖安装超时时间（秒）")
    fail_on_dependency_error: bool = Field(True, description="依赖安装失败时是否终止会话创建")
    allow_version_conflicts: bool = Field(
        False, description="是否允许版本冲突（Template 预装包 vs 用户请求包）"
    )
    python_package_index_url: Optional[str] = Field(
        None, max_length=512, description="Python 软件包仓库地址，默认 https://pypi.org/simple/"
    )

    @field_validator("cpu")
    @classmethod
    def validate_cpu(cls, v: str) -> str:
        try:
            float(v)
        except ValueError:
            raise ValueError("Invalid cpu format")
        return v


class ExecuteCodeRequest(BaseModel):
    """执行代码请求"""

    code: str = Field(
        ...,
        min_length=1,
        max_length=102400,
        description="要执行的代码。language=python 时需符合 AWS Lambda handler 格式；language=shell 时表示 shell 脚本内容。",
    )
    language: Literal["python", "javascript", "shell"] = Field(..., description="编程语言")
    timeout: int = Field(30, ge=1, le=3600, description="执行超时（秒）")
    event: Optional[Dict] = Field(None, description="事件数据")
    working_directory: Optional[str] = Field(
        None, description="可选执行目录，相对于 workspace 根目录；未传时默认使用 workspace 根目录。"
    )

    @field_validator("working_directory")
    @classmethod
    def validate_working_directory(cls, value: Optional[str]) -> Optional[str]:
        if value is None:
            return value

        stripped = value.strip()
        if not stripped or stripped.startswith("/") or "\\" in stripped:
            raise ValueError("working_directory must be a relative workspace path")
        if re.match(r"^[A-Za-z]:", stripped):
            raise ValueError("working_directory must be a relative workspace path")

        normalized = PurePosixPath(stripped).as_posix()
        parts = PurePosixPath(normalized).parts
        if any(part == ".." for part in parts):
            raise ValueError("working_directory must be a relative workspace path")

        if normalized.startswith("./"):
            normalized = normalized[2:]
        if not normalized:
            raise ValueError("working_directory must be a relative workspace path")
        return normalized

    model_config = ConfigDict(
        json_schema_extra={
            "examples": [
                {
                    "code": 'def handler(event):\n    name = event.get("name", "World")\n    return {"message": f"Hello, {name}!"}',
                    "language": "python",
                    "timeout": 10,
                    "event": {"name": "World"},
                },
                {
                    "code": 'def handler(event):\n    name = event.get("name", "World")\n    age = event.get("age", 0)\n    return {"message": f"Hello, {name}!", "age_doubled": age * 2}',
                    "language": "python",
                    "timeout": 30,
                    "event": {"name": "Alice", "age": 25},
                },
                {"code": "pwd && ls -la", "language": "shell", "timeout": 30},
                {
                    "code": "bash run.sh && python tools/build.py",
                    "language": "shell",
                    "timeout": 30,
                    "working_directory": "skill/mini-wiki",
                },
            ]
        }
    )


class TerminateSessionRequest(BaseModel):
    """终止会话请求"""

    reason: Optional[str] = Field(None, description="终止原因")


class InstallSessionDependenciesRequest(BaseModel):
    """增量安装 Python 依赖请求。"""

    python_package_index_url: Optional[str] = Field(
        None,
        max_length=512,
        description="Python 软件包仓库地址；未传则沿用当前 session 配置",
    )
    dependencies: List[DependencySpec] = Field(
        ...,
        min_length=1,
        max_length=50,
        description="本次增量安装的依赖列表",
    )
    install_timeout: int = Field(
        300,
        ge=30,
        le=1800,
        description="本次依赖安装超时时间（秒）",
    )


class CreateTemplateRequest(BaseModel):
    """创建模板请求"""

    id: str = Field(..., min_length=1, max_length=64, description="模板 ID")
    name: str = Field(..., min_length=1, max_length=255, description="模板名称")
    image_url: str = Field(..., min_length=1, max_length=512, description="镜像 URL")
    runtime_type: Literal["python3.11", "nodejs20", "java17", "go1.21"] = Field(
        ..., description="运行时类型"
    )
    default_cpu_cores: float = Field(0.5, ge=0.1, le=4.0, description="默认 CPU 核心数")
    default_memory_mb: int = Field(512, ge=128, le=8192, description="默认内存（MB）")
    default_disk_mb: int = Field(1024, ge=256, le=51200, description="默认磁盘（MB）")
    default_timeout: int = Field(300, ge=60, le=3600, description="默认超时（秒）")
    default_env_vars: Optional[Dict[str, str]] = Field(None, description="默认环境变量")


class UpdateTemplateRequest(BaseModel):
    """更新模板请求"""

    name: Optional[str] = Field(None, min_length=1, max_length=255, description="模板名称")
    image_url: Optional[str] = Field(None, min_length=1, max_length=512, description="镜像 URL")
    default_cpu_cores: Optional[float] = Field(None, ge=0.1, le=4.0, description="默认 CPU 核心数")
    default_memory_mb: Optional[int] = Field(None, ge=128, le=8192, description="默认内存（MB）")
    default_disk_mb: Optional[int] = Field(None, ge=256, le=51200, description="默认磁盘（MB）")
    default_timeout: Optional[int] = Field(None, ge=60, le=3600, description="默认超时（秒）")
    default_env_vars: Optional[Dict[str, str]] = Field(None, description="默认环境变量")
