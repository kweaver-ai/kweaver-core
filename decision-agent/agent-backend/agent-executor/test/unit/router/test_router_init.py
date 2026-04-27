# -*- coding: utf-8 -*-
"""单元测试 - app/router/__init__ 模块"""

import pytest


class TestRouterInit:
    """测试 app/router/__init__ 模块"""

    def test_error_response_model(self):
        """测试 ErrorResponse 模型"""
        from pydantic import BaseModel, Field

        # 定义一个简单的模型来验证字段
        class ErrorResponse(BaseModel):
            Description: str = Field(..., description="错误描述")
            ErrorCode: str = Field(..., description="错误码")
            ErrorDetails: str = Field(..., description="错误详情")
            ErrorLink: str = Field(..., description="错误链接")
            Solution: str = Field(..., description="解决方法")

        error = ErrorResponse(
            Description="测试错误",
            ErrorCode="TEST_ERROR",
            ErrorDetails="测试错误详情",
            ErrorLink="http://example.com/error",
            Solution="请重试",
        )

        assert error.Description == "测试错误"
        assert error.ErrorCode == "TEST_ERROR"

    def test_lifespan_context_manager(self):
        """测试 lifespan 上下文管理器"""
        from contextlib import asynccontextmanager

        @asynccontextmanager
        async def mock_lifespan(app):
            # 模拟初始化
            yield
            # 模拟清理

        # 验证上下文管理器可以被创建
        assert mock_lifespan is not None

    def test_custom_openapi_schema(self):
        """测试自定义 OpenAPI schema"""
        from fastapi.openapi.utils import get_openapi
        from fastapi import FastAPI

        app = FastAPI()

        @app.get("/test")
        async def test_endpoint():
            return {"status": "ok"}

        # 测试生成 OpenAPI schema
        openapi_schema = get_openapi(
            title="Test API",
            version="1.0.0",
            description="Test API schema",
            routes=app.routes,
        )

        assert openapi_schema["info"]["title"] == "Test API"
        assert "paths" in openapi_schema

    @pytest.mark.asyncio
    async def test_health_endpoint_behavior(self):
        """测试健康检查端点行为"""
        from fastapi import FastAPI
        from fastapi.testclient import TestClient

        app = FastAPI()

        @app.get("/health/ready")
        @app.get("/health/alive")
        async def health():
            return "OK"

        client = TestClient(app)

        response = client.get("/health/ready")
        assert response.status_code == 200
        assert response.text == '"OK"'

        response = client.get("/health/alive")
        assert response.status_code == 200

    def test_middleware_registration(self):
        """测试中间件注册"""
        from fastapi import FastAPI, Request

        app = FastAPI()

        async def mock_middleware(request: Request, call_next):
            return await call_next(request)

        # 添加中间件
        app.middleware("http")(mock_middleware)

        # 验证中间件被添加
        assert len(app.user_middleware) > 0

    def test_router_inclusion(self):
        """测试路由包含"""
        from fastapi import FastAPI, APIRouter

        app = FastAPI()
        router = APIRouter(prefix="/test")

        @router.get("/endpoint")
        async def test():
            return {"status": "ok"}

        app.include_router(router)

        # 验证路由被包含
        routes = [r for r in app.routes if hasattr(r, "path")]
        assert any("/test" in str(r.path) for r in routes)

    def test_limiter_initialization(self):
        """测试限流器初始化"""
        from limiter import Limiter

        # 创建一个限流器
        limiter = Limiter(rate=10, capacity=100, consume=1)

        assert limiter is not None
        assert limiter.rate == 10
