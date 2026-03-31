# -*- coding:utf-8 -*-
from contextlib import asynccontextmanager

from fastapi import FastAPI, Request, Response
from fastapi.openapi.utils import get_openapi
from limiter import Limiter
from opentelemetry.instrumentation.aiohttp_client import AioHttpClientInstrumentor
from pydantic import BaseModel, Field

from app.common.config import Config, observability_config, server_info
from app.utils.observability.observability import (
    init_observability,
    shutdown_observability,
)

# 导入中间件
from .middleware_pkg import o11y_trace, log_requests


@asynccontextmanager
async def lifespan(app: FastAPI):
    """应用生命周期管理"""
    # 启动时执行
    init_observability(server_info, observability_config)
    AioHttpClientInstrumentor().instrument()
    yield
    # 关闭时执行
    shutdown_observability()


app = FastAPI(lifespan=lifespan)


token_rate = Config.app.rps_limit
token_capacity = 1000
token_consume_limit = 2


# 创建一个令牌桶限流器，设置容量为tokenBucketLimit，每秒产生tokenBucketLimit个令牌
limiter = Limiter(rate=token_rate, capacity=token_capacity, consume=token_consume_limit)


@app.middleware("http")
async def o11y_trace_middleware(request: Request, call_next) -> Response:
    """HTTP请求追踪中间件"""
    return await o11y_trace(request, call_next)


@app.middleware("http")
async def log_requests_middleware(request: Request, call_next) -> Response:
    """请求日志中间件"""
    return await log_requests(request, call_next)


# 健康检查
@app.get(Config.app.host_prefix + "/health/ready", include_in_schema=False)
@app.get(Config.app.host_prefix + "/health/alive", include_in_schema=False)
async def ready():
    return "OK"


# 导入路由
from app.router.agent_controller_pkg.common import router as agent_router_v1  # noqa: E402
from app.router.agent_controller_pkg.common_v2 import router_v2 as agent_router_v2  # noqa: E402
from app.router.agent_controller_pkg import (  # noqa: E402
    run_agent_v2,  # noqa: F401
    run_agent_debug_v2,  # noqa: F401
    agent_cache_manage,  # noqa: F401
)
from app.router.exception_handler import register_exception_handlers  # noqa: E402
from app.router.tool_controller import router as tool_router  # noqa: E402

# 注册异常处理器
register_exception_handlers(app)

# 注册路由
app.include_router(agent_router_v2)
app.include_router(agent_router_v1)
app.include_router(tool_router)


class ErrorResponse(BaseModel):
    Description: str = Field(..., description="错误描述")
    ErrorCode: str = Field(..., description="错误码")
    ErrorDetails: str = Field(..., description="错误详情")
    ErrorLink: str = Field(..., description="错误链接")
    Solution: str = Field(..., description="解决方法")


# openapi
def custom_openapi():
    if app.openapi_schema:
        return app.openapi_schema
    openapi_schema = get_openapi(
        title="agent-executor OpenAPI",
        version="1.0.0",
        description="This is a agent-executor OpenAPI schema",
        routes=app.routes,
    )
    for path in openapi_schema["paths"].values():
        for method in path.values():
            if "responses" in method:
                # 移除422错误响应
                if "422" in method["responses"]:
                    method["responses"].pop("422")
                    # 增加400错误响应
                    method["responses"]["400"] = {
                        "description": "Bad Request",
                        "content": {
                            "application/json": {
                                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
                            }
                        },
                    }
                    method["responses"]["500"] = {
                        "description": "Internal Server Error",
                        "content": {
                            "application/json": {
                                "schema": {"$ref": "#/components/schemas/ErrorResponse"}
                            }
                        },
                    }
    openapi_schema["components"]["schemas"]["ErrorResponse"] = ErrorResponse.schema()
    app.openapi_schema = openapi_schema
    return app.openapi_schema


app.openapi = custom_openapi
