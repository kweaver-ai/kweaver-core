"""
REST API 响应模式

定义 FastAPI 的响应 Pydantic 模型。
"""

from pydantic import BaseModel, Field
from typing import Optional, List, Dict, Any
from datetime import datetime


class ResourceLimitResponse(BaseModel):
    """资源限制响应"""

    cpu: str
    memory: str
    disk: str
    max_processes: Optional[int] = 128


class DependencyResponse(BaseModel):
    """依赖响应。"""

    name: str
    version: Optional[str] = None


class InstalledDependencyResponse(BaseModel):
    """已安装依赖响应。"""

    name: str
    version: str
    install_location: str
    install_time: datetime
    is_from_template: bool = False


class SessionResponse(BaseModel):
    """会话响应"""

    id: str
    template_id: str
    status: str
    resource_limit: Optional[ResourceLimitResponse] = None
    workspace_path: Optional[str] = None
    language_runtime: str
    runtime_type: str
    runtime_node: Optional[str] = None
    container_id: Optional[str] = None
    pod_name: Optional[str] = None
    env_vars: Dict[str, str] = {}
    timeout: int
    python_package_index_url: str = "https://pypi.org/simple/"
    requested_dependencies: List[DependencyResponse] = []
    installed_dependencies: List[InstalledDependencyResponse] = []
    dependency_install_status: str = "pending"
    dependency_install_error: Optional[str] = None
    dependency_install_started_at: Optional[datetime] = None
    dependency_install_completed_at: Optional[datetime] = None
    created_at: datetime
    updated_at: datetime
    completed_at: Optional[datetime] = None
    last_activity_at: Optional[datetime] = None


class SessionListResponse(BaseModel):
    """会话列表响应"""

    items: List[SessionResponse]
    total: int
    limit: int
    offset: int
    has_more: bool


class ArtifactResponse(BaseModel):
    """文件制品响应"""

    path: str
    size: int
    mime_type: str
    type: str
    created_at: datetime
    checksum: Optional[str] = None


class ExecutionResponse(BaseModel):
    """执行响应"""

    id: str
    session_id: str
    status: str
    code: Optional[str] = None
    language: Optional[str] = None
    timeout: Optional[int] = None
    exit_code: Optional[int] = None
    error_message: Optional[str] = None
    execution_time: Optional[float] = None
    stdout: Optional[str] = None
    stderr: Optional[str] = None
    artifacts: List[ArtifactResponse] = []
    retry_count: int = 0
    created_at: Optional[datetime] = None
    started_at: Optional[datetime] = None
    completed_at: Optional[datetime] = None
    return_value: Optional[Any] = None
    metrics: Optional[Dict[str, Any]] = None


class ExecuteCodeResponse(BaseModel):
    """执行代码响应"""

    execution_id: str
    session_id: str
    status: str
    created_at: Optional[datetime] = None


class TemplateResponse(BaseModel):
    """模板响应"""

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


class ContainerResponse(BaseModel):
    """容器响应"""

    id: str
    session_id: str
    runtime_type: str
    node_id: str
    container_name: str
    image_url: str
    status: str
    ip_address: Optional[str] = None
    cpu_cores: float
    memory_mb: int
    disk_mb: int
    created_at: Optional[datetime] = None
    started_at: Optional[datetime] = None
    exited_at: Optional[datetime] = None


class ErrorResponse(BaseModel):
    """错误响应"""

    error: str
    message: str
    detail: Optional[str] = None


class HealthResponse(BaseModel):
    """健康检查响应"""

    status: str = "healthy"
    version: str = "2.1.0"
    uptime: float
