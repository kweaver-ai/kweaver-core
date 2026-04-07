import asyncio
import json
import aiohttp
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse
from starlette.middleware.base import BaseHTTPMiddleware

from app.commons.errors import UnauthorizedError, HydraServiceError
from app.core.config import base_config, server_info, observability_config
from app.logs import log_init, sys_log
from app.mydb.ConnectUtil import get_redis_util
from app.routers import router_init
from app.utils.comment_utils import write_log
from app.utils.observability.observability import init_observability, shutdown_observability


def conf_init(app):
    import os
    environment = os.getenv('ENVIRONMENT', 'development')
    sys_log.info(msg=f'Start app with {environment} environment')
    if environment == 'production':
        app.docs_url = None
        app.redoc_url = None
        app.debug = False


async def start_event():
    await write_log(msg='系统启动')
    # 在应用启动时调用
    try:
        await get_redis_util()
    except Exception as e:
        raise e
    # 初始化可观测模块
    init_observability(server_info, observability_config)


async def shutdown_event():
    await write_log(msg='系统关闭')
    # 关闭可观测模块
    shutdown_observability()


async def auth_middleware(request: Request, call_next):
    path = request.url.path
    if path.startswith("/api/v1/health"):
        pass
    elif path.startswith("/api/private"):
        pass
        # user_id = request.headers.get("x-account-id")
        # if not user_id:
        #     return JSONResponse(
        #         status_code=401,
        #         content=UnauthorizedError
        #     )
    else:
        auth_header = request.headers.get("Authorization")
        if not auth_header or not auth_header.startswith("Bearer "):
            return JSONResponse(
                status_code=401,
                content=UnauthorizedError
            )
        token = auth_header[7:]
        hydra_url = f"http://{base_config.OAUTHADMINHOST}:{base_config.OAUTHADMINPORT}/admin/oauth2/introspect"
        async with aiohttp.ClientSession() as session:
            try:
                payload = {"token": token}
                async with session.post(hydra_url, data=payload) as response:
                    if response.status != 200:
                        error_dict = HydraServiceError.copy()
                        error_dict["detail"] = await response.text()
                        return JSONResponse(
                            status_code=400,
                            content=error_dict
                        )
                    else:
                        res = await response.text()
                        result = json.loads(res)
                        activate = result.get("active", False)
                        user_id = result.get("sub", "")
                        client_id = result.get("client_id", "")
                        role = "user" if client_id != user_id else "app"
                    if activate:
                        request.scope['headers'].append((b"x-account-id", user_id.encode()))
                        request.scope['headers'].append((b"x-account-type", role.encode()))
                    else:
                        return JSONResponse(
                            status_code=401,
                            content=UnauthorizedError
                        )
            except Exception as e:
                return JSONResponse(
                    status_code=400,
                    content=HydraServiceError
                )

    response = await call_next(request)
    return response


class RequestSizeMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        content_length = request.headers.get('content-length')
        if content_length and int(content_length) > 10 * 1024 * 1024:  # 10M限制
            return JSONResponse(
                status_code=413,
                content={"detail": "Payload too large"}
            )
        return await call_next(request)


def create_app():
    app = FastAPI(title="My API",
                  description="",
                  version="1.0.0",
                  on_startup=[start_event],
                  on_shutdown=[shutdown_event])

    # 添加请求体大小检查中间件
    # app.add_middleware(RequestSizeMiddleware)
    # 添加鉴权中间件
    app.add_middleware(BaseHTTPMiddleware, dispatch=auth_middleware)

    # 初始化日志
    log_init()
    # 加载配置
    conf_init(app)
    # 初始化路由配置
    router_init(app)
    return app
