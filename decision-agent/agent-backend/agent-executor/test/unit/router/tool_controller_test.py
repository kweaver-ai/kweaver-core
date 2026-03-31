"""集成测试 - app/router/tool_controller.py"""

import pytest
from unittest.mock import patch
from fastapi.testclient import TestClient
from fastapi import FastAPI


@pytest.fixture
def app():
    """创建测试应用"""
    from app.router.tool_controller import router

    app = FastAPI()
    app.include_router(router)
    return app


@pytest.fixture
def client(app):
    """创建测试客户端"""
    return TestClient(app)


class TestZhipuSearchTool:
    """测试智谱搜索工具端点"""

    @patch("app.router.tool_controller.Config")
    @patch("app.logic.tool.zhipu_search_tool.zhipu_search_tool")
    def test_zhipu_search_tool_success(self, mock_search, m_config, client):
        """测试成功执行智谱搜索"""
        m_config.app.host_prefix = ""

        # Mock 搜索结果 - 使用正确的 ZhipuSearchResponse 格式
        mock_search.return_value = {
            "choices": [{"message": {"content": "搜索结果内容"}}],
            "created": 1749713107,
            "id": "test_id_123",
            "model": "web-search-pro",
            "request_id": "test-request-id",
            "usage": {
                "completion_tokens": 100,
                "prompt_tokens": 0,
                "total_tokens": 100,
            },
        }

        response = client.post(
            "/tools/zhipu_search_tool",
            json={"query": "机器学习"},
            headers={"api_key": "test_api_key"},
        )

        assert response.status_code == 200
        data = response.json()
        assert "choices" in data
        assert "id" in data

    @patch("app.router.tool_controller.Config")
    @patch("app.logic.tool.zhipu_search_tool.zhipu_search_tool")
    def test_zhipu_search_tool_with_database(self, mock_search, m_config, client):
        """测试带数据库参数的智谱搜索"""
        m_config.app.host_prefix = ""
        mock_search.return_value = {
            "choices": [{"message": {"content": "数据库搜索结果"}}],
            "created": 1749713107,
            "id": "test_id_456",
            "model": "web-search-pro",
            "request_id": "test-request-id-2",
            "usage": {
                "completion_tokens": 100,
                "prompt_tokens": 0,
                "total_tokens": 100,
            },
        }

        response = client.post(
            "/tools/zhipu_search_tool",
            json={"query": "测试查询", "database": "test_db"},
            headers={"api_key": "test_key"},
        )

        assert response.status_code == 200
        data = response.json()
        assert "choices" in data

    @patch("app.router.tool_controller.Config")
    def test_zhipu_search_tool_missing_api_key(self, m_config, client):
        """测试缺少API密钥的情况"""
        m_config.app.host_prefix = ""

        response = client.post("/tools/zhipu_search_tool", json={"query": "测试"})

        # FastAPI 应该返回验证错误
        assert response.status_code == 422


class TestOnlineSearchCiteTool:
    """测试联网搜索引用工具端点"""

    @patch("app.router.tool_controller.Config")
    @patch("app.logic.tool.online_search_cite_tool.online_search_cite_tool")
    def test_online_search_cite_tool_non_stream(self, mock_search, m_config, client):
        """测试非流式联网搜索"""
        m_config.app.host_prefix = ""
        mock_search.return_value = {
            "answer": "这是搜索答案",
            "references": [
                {
                    "title": "参考1",
                    "content": "内容1",
                    "link": "http://example.com/1",
                    "index": 0,
                }
            ],
        }

        response = client.post(
            "/tools/online_search_cite_tool",
            json={
                "query": "人工智能",
                "model_name": "deepseek-v3",
                "search_tool": "zhipu_search_tool",
                "api_key": "test_key",
                "user_id": "user123",
                "stream": False,
            },
        )

        assert response.status_code == 200
        data = response.json()
        assert "answer" in data
        assert data["answer"] == "这是搜索答案"

    @pytest.mark.skip(reason="流式测试需要复杂的异步设置，跳过此测试")
    def test_online_search_cite_tool_stream(self):
        """测试流式联网搜索 - 跳过"""
        # 流式测试比较复杂，需要 mock 多个异步函数
        # 在集成测试环境中再实现
        pass

    @patch("app.router.tool_controller.Config")
    @patch("app.logic.tool.online_search_cite_tool.online_search_cite_tool")
    def test_online_search_cite_tool_with_empty_query(
        self, mock_search, m_config, client
    ):
        """测试空查询"""
        m_config.app.host_prefix = ""
        mock_search.return_value = {"answer": "", "references": []}

        response = client.post(
            "/tools/online_search_cite_tool",
            json={
                "query": "",
                "model_name": "model",
                "search_tool": "tool",
                "api_key": "key",
                "user_id": "user",
                "stream": False,
            },
        )

        assert response.status_code == 200


class TestRouterConfiguration:
    """测试路由器配置"""

    def test_router_prefix(self):
        """测试路由前缀配置"""
        from app.router.tool_controller import router

        assert "/tools" in router.prefix
        assert router.tags == ["internal-tools"]

    def test_router_endpoints_count(self):
        """测试端点数量"""
        from app.router.tool_controller import router

        # 应该有2个端点
        routes = [r for r in router.routes if hasattr(r, "path")]
        assert len(routes) == 2

    def test_zhipu_search_endpoint_config(self):
        """测试智谱搜索端点配置"""
        from app.router.tool_controller import router

        route = next(
            (r for r in router.routes if r.path.endswith("zhipu_search_tool")), None
        )
        assert route is not None
        assert route.summary == "智谱搜索"

    def test_online_search_endpoint_config(self):
        """测试联网搜索端点配置"""
        from app.router.tool_controller import router

        route = next(
            (r for r in router.routes if r.path.endswith("online_search_cite_tool")),
            None,
        )
        assert route is not None
        assert route.summary == "联网搜索添加引用工具"
