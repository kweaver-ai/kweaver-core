"""
健康检查 REST API 路由

定义健康检查和系统监控相关的 HTTP 端点。
"""
from fastapi import APIRouter, Depends
from pydantic import BaseModel
from typing import Optional
import time

from src.interfaces.rest.schemas.response import HealthResponse

router = APIRouter(prefix="/health", tags=["health"])

# 应用启动时间
_start_time = time.time()


class SystemStatus(BaseModel):
    """系统状态"""
    status: str
    version: str
    uptime: float


@router.get("", response_model=HealthResponse)
async def health_check() -> HealthResponse:
    """
    健康检查端点

    返回系统状态和运行时间。
    """
    return HealthResponse(
        status="healthy",
        version="2.1.0",
        uptime=time.time() - _start_time
    )


@router.get("/detailed")
async def detailed_health_check() -> dict:
    """
    详细健康检查

    返回系统状态和依赖项健康状态。
    """
    # TODO: 实现依赖项健康检查
    # - 数据库连接
    # - S3 存储
    # - 运行时节点
    return {
        "status": "healthy",
        "version": "2.1.0",
        "uptime": time.time() - _start_time,
        "dependencies": {
            "database": "healthy",
            "storage": "healthy",
            "runtime_nodes": "healthy"
        }
    }


@router.post("/sync")
async def trigger_state_sync() -> dict:
    """
    手动触发状态同步

    立即执行一次状态同步和健康检查，用于调试和手动恢复。
    """
    from src.infrastructure.dependencies import get_state_sync_service

    state_sync_service = get_state_sync_service()
    result = await state_sync_service.periodic_health_check()

    return {
        "status": "success",
        "message": "State sync completed",
        "result": result
    }
