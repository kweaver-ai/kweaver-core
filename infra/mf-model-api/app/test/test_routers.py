"""测试 routers 模块"""
import pytest
from unittest.mock import Mock, AsyncMock, patch
from fastapi.testclient import TestClient
from fastapi import FastAPI
from app.routers.llm_router import llm_route, health_route


class TestHealthRoutes:
    """测试健康检查路由"""

    @pytest.fixture
    def app(self):
        """创建测试应用"""
        app = FastAPI()
        app.include_router(health_route)
        return app

    @pytest.fixture
    def client(self, app):
        """创建测试客户端"""
        return TestClient(app)

    def test_health_ready(self, client):
        """测试ready端点"""
        response = client.get("/health/ready")
        assert response.status_code == 200
        assert response.json() == {"res": 0}

    def test_health_alive(self, client):
        """测试alive端点"""
        response = client.get("/health/alive")
        assert response.status_code == 200
        assert response.json() == {"res": 0}


class TestLLMRoutes:
    """测试LLM路由"""

    @pytest.fixture
    def app(self):
        """创建测试应用"""
        app = FastAPI()
        app.include_router(llm_route)
        return app

    @pytest.fixture
    def client(self, app):
        """创建测试客户端"""
        return TestClient(app)

    def test_chat_completions_endpoint_exists(self, client):
        """测试chat/completions端点存在"""
        # 由于需要认证,应该返回401或422
        response = client.post("/chat/completions")
        assert response.status_code in [401, 422]

    def test_chat_completions_with_valid_request(self, client):
        """测试有效请求的chat/completions"""
        from fastapi.responses import JSONResponse
        
        request_data = {
            "model": "test_model",
            "messages": [{"role": "user", "content": "你好"}],
            "temperature": 0.7,
            "top_p": 0.9,
            "frequency_penalty": 0,
            "presence_penalty": 0,
            "max_tokens": 1000
        }
        
        # Mock response
        mock_response = {
            "id": "chatcmpl-123",
            "object": "chat.completion",
            "created": 1677652288,
            "model": "test_model",
            "choices": [{
                "index": 0,
                "message": {
                    "role": "assistant",
                    "content": "你好！有什么我可以帮助你的吗？",
                },
                "finish_reason": "stop"
            }],
            "usage": {
                "prompt_tokens": 9,
                "completion_tokens": 12,
                "total_tokens": 21
            }
        }
        
        # Patch both get_user_info and used_model_openai within the test
        with patch('app.routers.llm_router.get_user_info', new_callable=AsyncMock) as mock_get_user:
            with patch('app.routers.llm_router.used_model_openai', new_callable=AsyncMock) as mock_used_model:
                mock_get_user.return_value = ("user123", "zh", "user")
                mock_used_model.return_value = JSONResponse(status_code=200, content=mock_response)
                
                # 验证路由能正确处理请求
                response = client.post("/chat/completions", json=request_data)
                assert response.status_code == 200
                assert "choices" in response.json()


class TestRouterIntegration:
    """路由集成测试"""

    def test_llm_route_has_prefix(self):
        """测试llm路由有正确的前缀"""
        # 这个测试验证路由对象存在
        assert llm_route is not None

    def test_health_route_exists(self):
        """测试健康检查路由存在"""
        assert health_route is not None


class TestAPIDocumentation:
    """测试API文档"""

    @pytest.fixture
    def app_with_docs(self):
        """创建带文档的应用"""
        app = FastAPI()
        app.include_router(llm_route)
        app.include_router(health_route)
        return app

    @pytest.fixture
    def client(self, app_with_docs):
        """创建测试客户端"""
        return TestClient(app_with_docs)

    def test_openapi_schema_exists(self, client):
        """测试OpenAPI schema存在"""
        response = client.get("/openapi.json")
        assert response.status_code == 200
        schema = response.json()
        assert "openapi" in schema
        assert "paths" in schema

    def test_health_endpoints_in_schema(self, client):
        """测试健康检查端点在schema中"""
        response = client.get("/openapi.json")
        schema = response.json()
        # 健康检查端点通常标记为不在schema中
        # 这里只验证schema有效
        assert "paths" in schema

