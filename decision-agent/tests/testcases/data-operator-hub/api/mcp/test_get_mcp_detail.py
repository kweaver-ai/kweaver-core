# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent

@allure.feature("MCP服务管理接口测试：获取MCP_Server详情")
class TestGetMCPDetail:
    
    client = MCP()
    tool_client = ToolBox()
    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.tool_client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        result = self.tool_client.GetBoxToolsList(box_id, None, Headers)
        tools = result[1]["tools"]
        update_data = []
        tools_id = []
        tool_configs = []
        for tool in tools:
            data = {
                "tool_id": tool["tool_id"],
                "status": "enabled"
            }
            update_data.append(data) 
            tools_id.append(tool["tool_id"])  
            tool_config = {
                "box_id": box_id,
                "box_name": name,
                "tool_id": tool["tool_id"],
                "tool_name": tool["name"],
                "description": tool["description"],
                "use_rule": "all"
            }
            tool_configs.append(tool_config)       
        result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        result = self.tool_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
        assert result[0] == 200

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "add mcp server config",
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        TestGetMCPDetail.mcp_id = result[1]["mcp_id"]

    @allure.title("获取MCP_Server详情，mcp存在，url可正常解析，获取成功")
    def test_get_mcp_detail_01(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            },
            "source": "custom",
            "category": "other_category",
            "command": "ls",
            "args": ["a"],
            "env": {
                "name": "test"
            }
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["mcp_id"] == mcp_id
        assert "name" in result[1]["base_info"]
        assert "description" in result[1]["base_info"]
        assert "source" in result[1]["base_info"]
        assert "category" in result[1]["base_info"]
        assert "mode" in result[1]["base_info"]
        assert "command" in result[1]["base_info"]
        assert "url" in result[1]["base_info"]
        assert "tool_configs" not in result[1]["base_info"]
        assert result[1]["connection_info"] == None # 未发布，不返回url
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert "sse_url" in result[1]["connection_info"] # 发布后返回url

    @allure.title("获取MCP_Server详情，mcp存在，url无法解析，获取成功")
    def test_get_mcp_detail_02(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://127.0.0.1/mcp",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["mcp_id"] == mcp_id
        # assert len(result[1]["tools"]) == 0

    @allure.title("获取MCP_Server详情，mcp不存在，获取失败")
    def test_get_mcp_detail_03(self, Headers):
        mcp_id = str(uuid.uuid4())
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("获取MCP_Server详情，mode为stream，获取成功")
    def test_get_mcp_detail_04(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "stream",
            "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]

        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert "base_info" in result[1]
        assert "tool_configs" not in result[1]["base_info"]
        assert result[1]["connection_info"] == None # 未发布，不返回url
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert "stream_url" in result[1]["connection_info"]

    @allure.title("获取从工具箱导入的mcp详情，mcp未发布，详情中返回tool_configs信息，不返回connection_info信息")
    def test_get_mcp_detail_05(self, Headers):
        result = self.client.GetMCPDetail(TestGetMCPDetail.mcp_id, Headers)
        assert result[0] == 200
        assert "base_info" in result[1]
        assert "tool_configs" in result[1]["base_info"]
        assert len(result[1]["base_info"]["tool_configs"]) > 0
        assert result[1]["connection_info"] == None

    @allure.title("获取从工具箱导入的mcp详情，mcp已发布，详情中返回tool_configs和connection_info信息")
    def test_get_mcp_detail_06(self, Headers):
        result = self.client.MCPReleaseAction(TestGetMCPDetail.mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        result = self.client.GetMCPDetail(TestGetMCPDetail.mcp_id, Headers)
        assert result[0] == 200
        assert "base_info" in result[1]
        assert "tool_configs" in result[1]["base_info"]
        assert len(result[1]["base_info"]["tool_configs"]) > 0
        assert "connection_info" in result[1]
        assert "sse_url" in result[1]["connection_info"]
        assert "stream_url" in result[1]["connection_info"]

    @allure.title("获取从工具箱导入的mcp详情，mcp已下架，详情中返回tool_configs信息，不返回connection_info信息")
    def test_get_mcp_detail_07(self, Headers):
        result = self.client.MCPReleaseAction(TestGetMCPDetail.mcp_id, {"status": "offline"}, Headers)
        assert result[0] == 200
        result = self.client.GetMCPDetail(TestGetMCPDetail.mcp_id, Headers)
        assert result[0] == 200
        assert "base_info" in result[1]
        assert "tool_configs" in result[1]["base_info"]
        assert len(result[1]["base_info"]["tool_configs"]) > 0
        assert result[1]["connection_info"] == None

    