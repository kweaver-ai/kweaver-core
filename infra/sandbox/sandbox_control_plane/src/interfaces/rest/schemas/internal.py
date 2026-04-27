"""
内部 API 请求和响应模式

定义 Executor 调用的内部 API 的 Pydantic 模型。
"""
from pydantic import BaseModel, Field
from typing import Optional, Dict, Any, List


class ExecutionMetrics(BaseModel):
    """执行性能指标"""
    duration_ms: float = Field(..., description="墙钟耗时（毫秒）")
    cpu_time_ms: Optional[float] = Field(None, description="CPU 时间（毫秒）")
    peak_memory_mb: Optional[float] = Field(None, description="内存峰值（MB）")
    io_read_bytes: Optional[int] = Field(None, description="读取字节数")
    io_write_bytes: Optional[int] = Field(None, description="写入字节数")


class ArtifactMetadata(BaseModel):
    """文件元数据"""
    path: str = Field(..., description="相对于 workspace 的文件路径")
    size: int = Field(..., description="文件大小（字节）")
    mime_type: str = Field(..., description="MIME 类型")
    type: str = Field(..., description="文件类型: artifact, log, output")
    checksum: Optional[str] = Field(None, description="SHA256 校验和")


class ExecutionResultReport(BaseModel):
    """
    执行结果上报请求

    由 Executor 调用，上报执行结果到控制平面。
    """
    status: str = Field(..., description="执行状态: success, failed, timeout, crashed")
    stdout: str = Field("", description="标准输出")
    stderr: str = Field("", description="标准错误")
    exit_code: int = Field(..., description="进程退出码")
    execution_time: float = Field(..., description="执行耗时（秒）")
    return_value: Optional[Any] = Field(None, description="handler 函数返回值")
    metrics: Optional[ExecutionMetrics] = Field(None, description="性能指标")
    artifacts: List[str] = Field(default_factory=list, description="生成的文件路径列表")


class InternalAPIResponse(BaseModel):
    """内部 API 标准响应"""
    message: str = Field(..., description="响应消息")


class ContainerReadyRequest(BaseModel):
    """容器就绪请求"""
    container_id: str = Field(..., description="容器 ID")
    pod_name: Optional[str] = Field(None, description="Pod 名称（Kubernetes）")
    executor_port: int = Field(8080, description="执行器 HTTP API 端口")
    ready_at: Optional[str] = Field(None, description="就绪时间（ISO 8601）")
