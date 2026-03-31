# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random
import uuid

from lib.mcp import MCP

mcp_id = ""
name = ''.join(random.choice(string.ascii_letters) for i in range(8))

@allure.feature("MCP服务管理接口测试：更新MCP服务状态")
class TestUpdateMCPStatus:
    
    client = MCP()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global mcp_id, name
        # 创建MCP Server
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

    @allure.title("MCP服务状态变更，状态正常流转，变更成功")
    def test_update_mcp_status_01(self, Headers):
        '''
        unpublish -> published
        published -> offline
        offline -> published
        offline -> unpublish
        published -> editing
        editing -> published 
        '''
        global mcp_id
        # unpublish -> published
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "published"

        # published -> offline
        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "offline"
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "offline"

        # offline -> published
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "published"

        # offline -> unpublish
        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "offline"

        data = {
            "status": "unpublish"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "unpublish"

        # unpublish -> published
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        # published -> editing
        data = {
            "status": "editing"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "editing"
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "editing"

        # editing -> published
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "published"

    @allure.title("MCP服务状态变更，状态冲突，变更失败")
    def test_update_mcp_status_02(self, Headers):
        global mcp_id
        # published -> unpublish
        data = {
            "status": "unpublish"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 400
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "published"

        # published -> offline
        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "offline"

        # offline -> editing
        data = {
            "status": "editing"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 400
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "offline"

        # offline -> published、published -> editing
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200

        data = {
            "status": "editing"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200

        # editing -> unpublish
        data = {
            "status": "unpublish"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 400
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "editing"

        # editing -> offline
        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 400
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "editing"

        # editing -> published、published -> offline、offline -> unpublish
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200

        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200

        data = {
            "status": "unpublish"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200

        # unpublish -> editing
        data = {
            "status": "editing"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 400
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "unpublish"

    @allure.title("MCP服务发布操作，已存在同名已发布mcp，操作失败")
    def test_update_mcp_status_03(self, Headers):
        global mcp_id, name
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id1 = result[1]["mcp_id"]
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        result = self.client.MCPReleaseAction(mcp_id1, data, Headers)
        assert result[0] == 400

    @allure.title("MCP服务发布操作，status参数不正确，操作失败")
    def test_update_mcp_status_04(self, Headers):
        global mcp_id
        data = {
            "status": "invalid_status"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("MCP服务发布操作，mcp_id不存在，操作失败")
    def test_update_mcp_status_05(self, Headers):
        mcp_id = str(uuid.uuid4())
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 404

    
