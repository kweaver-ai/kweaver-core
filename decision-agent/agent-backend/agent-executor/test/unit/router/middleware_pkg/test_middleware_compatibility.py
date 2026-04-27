"""
测试中间件在不同 Starlette 版本中的兼容性
"""

import pytest
from fastapi import FastAPI, Request
from fastapi.testclient import TestClient


def test_middleware_body_reading():
    """测试中间件读取请求体后，handler 仍能正常读取"""
    app = FastAPI()

    body_from_middleware = None

    @app.middleware("http")
    async def log_body_middleware(request: Request, call_next):
        nonlocal body_from_middleware
        # 中间件读取请求体（会被缓存）
        body_from_middleware = await request.body()
        response = await call_next(request)
        return response

    @app.post("/test")
    async def test_endpoint(request: Request):
        # Handler 再次读取请求体（应该从缓存获取）
        body = await request.body()
        return {"body": body.decode()}

    client = TestClient(app)
    response = client.post("/test", json={"key": "value"})

    assert response.status_code == 200
    assert body_from_middleware is not None
    assert b"key" in body_from_middleware


def test_middleware_with_streaming_response():
    """测试中间件与流式响应的兼容性"""
    app = FastAPI()

    @app.middleware("http")
    async def log_middleware(request: Request, call_next):
        # 读取请求体
        await request.body()
        response = await call_next(request)
        return response

    @app.get("/stream")
    async def stream_endpoint():
        from fastapi.responses import StreamingResponse

        async def generate():
            for i in range(3):
                yield f"chunk{i}\n".encode()

        return StreamingResponse(generate(), media_type="text/plain")

    client = TestClient(app)
    response = client.get("/stream")

    assert response.status_code == 200
    assert b"chunk0" in response.content


def test_middleware_without_body():
    """测试 GET 请求（无 body）的兼容性"""
    app = FastAPI()

    @app.middleware("http")
    async def log_middleware(request: Request, call_next):
        # 即使是 GET 请求，也尝试读取 body（应该返回空）
        body = await request.body()
        assert body == b""
        response = await call_next(request)
        return response

    @app.get("/test")
    async def test_endpoint():
        return {"status": "ok"}

    client = TestClient(app)
    response = client.get("/test")

    assert response.status_code == 200
    assert response.json() == {"status": "ok"}


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
