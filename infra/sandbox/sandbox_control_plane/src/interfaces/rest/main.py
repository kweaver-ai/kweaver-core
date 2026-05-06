"""
FastAPI 主应用

沙箱控制中心的 FastAPI 应用入口。
"""

import time
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from fastapi import FastAPI, Request, status
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from fastapi.middleware.gzip import GZipMiddleware

# Configure logging FIRST before any other imports
from src.infrastructure.config.settings import get_settings
from src.infrastructure.logging import configure_logging, get_logger

# Initialize logging with settings
_settings = get_settings()
configure_logging(
    log_level=_settings.log_level,
    log_format=_settings.log_format,
)

# Now get logger
logger = get_logger(__name__)

# Import routes after logging is configured
from src.interfaces.rest.api.v1 import (
    sessions,
    executions,
    templates,
    health,
    files,
    internal,
)
from src.interfaces.rest.schemas.response import HealthResponse

# 应用启动时间
_start_time = time.time()


@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncGenerator[None, None]:
    """
    应用生命周期管理

    处理应用启动和关闭时的逻辑。
    """
    # 启动时执行
    logger.info("Starting Sandbox Control Plane")

    # 初始化依赖注入
    from src.infrastructure.dependencies import initialize_dependencies, get_storage_service

    initialize_dependencies(app)
    logger.info("Dependencies initialized")

    # 初始化 S3 storage（确保 bucket 存在）
    try:
        storage_service = get_storage_service()
        await storage_service.initialize()
    except Exception as e:
        logger.warning(f"S3 storage initialization failed (continuing): {e}")

    # 初始化数据库并创建表
    from src.infrastructure.persistence.database import db_manager
    from src.infrastructure.config.settings import get_settings

    await db_manager.upgrade_legacy_database_name()
    logger.info("Legacy database upgrade check completed")

    await db_manager.initialize()
    logger.info("Database initialized")
    await db_manager.run_startup_schema_migrations()
    logger.info("Startup schema migrations completed")

    # 根据环境决定是否自动创建表和初始化数据
    from src.infrastructure.config.settings import get_settings

    settings = get_settings()
    if settings.environment in ("development", "staging"):
        from src.infrastructure.persistence.seed.seeder import seed_default_data

        # 创建表
        await db_manager.create_tables()
        logger.info("Database tables created")

        # 初始化默认数据
        seed_stats = await seed_default_data(force=False)
        logger.info(
            "Default data initialized",
            runtime_nodes=seed_stats["runtime_nodes"],
            templates=seed_stats["templates"],
        )

    # ============= 启动时状态同步 =============
    from src.infrastructure.dependencies import get_state_sync_service

    state_sync_service = get_state_sync_service()
    try:
        sync_stats = await state_sync_service.sync_on_startup()
        logger.info(
            "Startup state sync completed",
            total=sync_stats.get("total", 0),
            healthy=sync_stats.get("healthy", 0),
            unhealthy=sync_stats.get("unhealthy", 0),
            recovered=sync_stats.get("recovered", 0),
            failed=sync_stats.get("failed", 0),
        )
        if sync_stats.get("errors"):
            logger.warning("State sync had errors", errors=sync_stats["errors"])
    except Exception as e:
        logger.error("Failed to perform startup state sync", error=str(e), exc_info=True)

    # ============= 启动后台任务管理器 =============
    from src.infrastructure.background_tasks import BackgroundTaskManager
    from src.infrastructure.dependencies import get_state_sync_service

    background_task_manager = BackgroundTaskManager()

    # 注册定时健康检查任务（每 30 秒）
    state_sync_svc = get_state_sync_service()
    background_task_manager.register_task(
        name="health_check",
        func=state_sync_svc.periodic_health_check,
        interval_seconds=30,
        initial_delay_seconds=30,  # 首次执行延迟 30 秒
    )

    # 注册会话清理任务（每 5 分钟）
    from src.application.services.session_cleanup_service import SessionCleanupService
    from src.infrastructure.dependencies import get_docker_scheduler_service, get_storage_service
    from src.infrastructure.persistence.repositories.sql_session_repository import (
        SqlSessionRepository,
    )
    from src.infrastructure.persistence.database import db_manager

    async def session_cleanup_task():
        """会话清理任务（每次执行时创建新的 repository）"""
        async with db_manager.get_session() as session:
            session_repo = SqlSessionRepository(session)
            scheduler = get_docker_scheduler_service(
                runtime_node_repo=None,
                template_repo=None,
            )
            storage_service = get_storage_service()
            cleanup_svc = SessionCleanupService(
                session_repo=session_repo,
                scheduler=scheduler,
                idle_timeout_minutes=settings.idle_threshold_minutes,
                max_lifetime_hours=settings.max_lifetime_hours,
                storage_service=storage_service,
            )
            return await cleanup_svc.cleanup_idle_sessions()

    background_task_manager.register_task(
        name="session_cleanup",
        func=session_cleanup_task,
        interval_seconds=300,  # 5 分钟
        initial_delay_seconds=60,  # 首次执行延迟 1 分钟
    )

    # 注册会话创建超时检测任务（每 5 分钟）
    from src.application.services.session_stuck_creating_service import SessionStuckCreatingService

    async def stuck_creating_check_task():
        """会话创建超时检测任务（每次执行时创建新的 repository）"""
        async with db_manager.get_session() as session:
            session_repo = SqlSessionRepository(session)
            stuck_creating_svc = SessionStuckCreatingService(
                session_repo=session_repo,
                creating_timeout_seconds=settings.creating_timeout_seconds,
            )
            return await stuck_creating_svc.check_and_mark_stuck_sessions()

    background_task_manager.register_task(
        name="stuck_creating_check",
        func=stuck_creating_check_task,
        interval_seconds=300,  # 5 分钟，与清理任务一致
        initial_delay_seconds=60,  # 首次执行延迟 1 分钟
    )

    # 启动所有后台任务
    await background_task_manager.start_all()
    logger.info(f"Background tasks started: {background_task_manager.task_count} tasks")

    # 将后台任务管理器存储到 app.state，以便关闭时使用
    app.state.background_task_manager = background_task_manager

    yield

    # 关闭时执行
    logger.info("Shutting down Sandbox Control Plane")

    # 停止所有后台任务
    if hasattr(app.state, "background_task_manager"):
        await app.state.background_task_manager.stop_all()
        logger.info("Background tasks stopped")

    # 清理依赖项（包括关闭数据库连接）
    from src.infrastructure.dependencies import cleanup_dependencies

    await cleanup_dependencies(app)
    await db_manager.close()


def create_app() -> FastAPI:
    """
    创建 FastAPI 应用

    使用工厂模式创建应用，便于测试和配置。
    """
    app = FastAPI(
        title="Sandbox Control Plane",
        description="代码沙箱管理平台 API",
        version="2.1.0",
        docs_url="/docs",
        redoc_url="/redoc",
        openapi_url="/openapi.json",
        lifespan=lifespan,
    )

    # 配置 CORS
    app.add_middleware(
        CORSMiddleware,
        allow_origins=["*"],  # 生产环境应配置具体域名
        allow_credentials=True,
        allow_methods=["*"],
        allow_headers=["*"],
    )

    # 添加 Gzip 压缩
    app.add_middleware(GZipMiddleware, minimum_size=1000)

    # 注册异常处理器
    _register_exception_handlers(app)

    # 注册中间件
    _register_middleware(app)

    # 注册路由
    _register_routes(app)

    return app


def _register_exception_handlers(app: FastAPI) -> None:
    """注册异常处理器"""
    from src.shared.errors.domain import NotFoundError, ValidationError

    @app.exception_handler(NotFoundError)
    async def not_found_exception_handler(request: Request, exc: NotFoundError) -> JSONResponse:
        """404 异常处理"""
        logger.warning(
            "Resource not found",
            path=request.url.path,
            method=request.method,
            error=str(exc),
        )
        return JSONResponse(
            status_code=status.HTTP_404_NOT_FOUND,
            content={
                "error": "Not Found",
                "message": exc.message,
                "detail": str(exc),
            },
        )

    @app.exception_handler(ValidationError)
    async def validation_exception_handler(request: Request, exc: ValidationError) -> JSONResponse:
        """409 Conflict 异常处理"""
        logger.warning(
            "Validation error (conflict)",
            path=request.url.path,
            method=request.method,
            error=str(exc),
        )
        return JSONResponse(
            status_code=status.HTTP_409_CONFLICT,
            content={
                "error": "Conflict",
                "message": exc.message,
                "detail": str(exc),
            },
        )

    @app.exception_handler(Exception)
    async def global_exception_handler(request: Request, exc: Exception) -> JSONResponse:
        """全局异常处理"""
        logger.error(
            "Unhandled exception",
            path=request.url.path,
            method=request.method,
            error=str(exc),
            exc_info=exc,
        )
        return JSONResponse(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            content={
                "error": "Internal Server Error",
                "message": "An unexpected error occurred",
                "detail": str(exc) if app.debug else None,
            },
        )


def _register_routes(app: FastAPI) -> None:
    """注册路由"""

    # 注册所有 API 路由
    app.include_router(health.router, prefix="/api/v1")
    app.include_router(sessions.router, prefix="/api/v1")
    app.include_router(executions.router, prefix="/api/v1")
    app.include_router(templates.router, prefix="/api/v1")
    app.include_router(files.router, prefix="/api/v1")
    app.include_router(internal.router, prefix="/api/v1")  # 内部 API

    # 根端点
    @app.get("/", tags=["root"])
    async def root() -> dict:
        """根端点"""
        return {
            "name": "Sandbox Control Plane",
            "version": "2.1.0",
            "status": "operational",
            "features": [
                "session_management",
                "code_execution",
                "template_management",
                "file_operations",
                "state_sync",
            ],
            "documentation": {
                "swagger": "/docs",
                "redoc": "/redoc",
                "openapi": "/openapi.json",
            },
        }


def _register_middleware(app: FastAPI) -> None:
    """注册中间件"""

    # Add request logging middleware first (wraps all other middleware)
    from src.interfaces.rest.middleware import RequestLoggingMiddleware

    app.add_middleware(RequestLoggingMiddleware)


# 创建应用实例
app = create_app()


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(
        "sandbox_control_plane.src.interfaces.rest.main:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        log_level="info",
    )
