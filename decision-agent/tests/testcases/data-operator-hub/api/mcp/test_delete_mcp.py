# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from lib.mcp import MCP

mcp_id = ""

@allure.feature("MCP服务管理接口测试：删除MCP_Server配置")
class TestDeleteMCP:
    
    client = MCP()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global mcp_id
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

    @allure.title("删除MCP_Server配置，mcp存在且未发布，删除成功")
    def test_delete_mcp_01(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        result = self.client.DeleteMCP(mcp_id, Headers)
        assert result[0] == 200

        # 验证删除结果
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("删除MCP_Server配置，mcp不存在，删除失败")
    def test_delete_mcp_02(self, Headers):
        mcp_id = str(uuid.uuid4())
        result = self.client.DeleteMCP(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("删除MCP_Server配置，mcp存在且已发布，删除失败")
    def test_delete_mcp_03(self, Headers):
        global mcp_id
        
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        result = self.client.DeleteMCP(mcp_id, Headers)
        assert result[0] == 400

        # 验证删除结果
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200

    @allure.title("删除MCP_Server配置，mcp存在且状态为已发布编辑中，删除失败")
    def test_delete_mcp_04(self, Headers):
        global mcp_id
        
        data = {
            "status": "editing"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "editing"

        result = self.client.DeleteMCP(mcp_id, Headers)
        assert result[0] == 400

        # 验证删除结果
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200

    @allure.title("删除MCP_Server配置，mcp已下架且未被引用，删除成功")
    def test_delete_mcp_05(self, Headers):
        global mcp_id
        
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "offline"

        result = self.client.DeleteMCP(mcp_id, Headers)
        assert result[0] == 200

        # 验证删除结果
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 404