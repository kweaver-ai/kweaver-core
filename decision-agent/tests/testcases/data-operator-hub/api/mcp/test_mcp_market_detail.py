 # -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent
from common.assert_tools import AssertTools

mcp_id = ""
mcp_id1 = ""

@allure.feature("MCP服务市场接口测试：获取MCP服务市场详情")
class TestMarketDetail:
    client = MCP()
    tool_client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global mcp_id, mcp_id1
        # 创建MCP Server并发布
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_process",
            "source": "custom",
            "command": "ls",
            "args": ["a"],
            "env": {
                "name": "test"
            }
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        # 发布MCP服务
        release_data = {
                "status": "published"
            }
        result = self.client.MCPReleaseAction(mcp_id, release_data, Headers)
        assert result[0] == 200

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server 1",
            "mode": "stream",
            "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id1 = result[1]["mcp_id"]
        # 发布MCP服务
        release_data = {
                "status": "published"
            }
        result = self.client.MCPReleaseAction(mcp_id1, release_data, Headers)
        assert result[0] == 200

    @allure.title("获取MCP服务市场详情，参数正确，获取成功")
    def test_market_detail_01(self, Headers):
        global mcp_id
        result = self.client.GetMCPMarketDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["mcp_id"] == mcp_id

        # 验证base_info字段
        base_info = result[1]["base_info"]
        assert "name" in base_info
        assert "description" in base_info
        assert "mode" in base_info
        assert "source" in base_info
        assert "category" in base_info
        assert "create_user" in base_info
        assert "create_time" in base_info
        assert "update_time" in base_info
        assert "release_time" in base_info
        assert "release_user" in base_info
        assert "command" in base_info
        assert "headers" in base_info
        assert "env" in base_info
        assert "args" in base_info
        # 验证connection_info
        assert "connection_info" in result[1]
        assert "sse_url" in result[1]["connection_info"]
        # 验证tool_configs
        assert "tool_configs" not in result[1]["base_info"]

    @allure.title("获取MCP服务市场详情，mcp_id不存在，获取失败")
    def test_market_detail_02(self, Headers):
        mcp_id = str(uuid.uuid4())
        result = self.client.GetMCPMarketDetail(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("获取MCP服务市场详情，服务未发布，获取失败")
    def test_market_detail_03(self, Headers):
        # 创建新的MCP Server但不发布
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

        # 尝试获取未发布服务的市场详情
        result = self.client.GetMCPMarketDetail(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("获取MCP服务市场详情，服务url解析失败，获取成功")
    def test_market_detail_04(self, Headers):
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

        # 发布MCP服务
        release_data = {
                "status": "published"
            }
        result = self.client.MCPReleaseAction(mcp_id, release_data, Headers)
        assert result[0] == 200

        result = self.client.GetMCPMarketDetail(mcp_id, Headers)
        assert result[0] == 200

    @allure.title("获取MCP服务市场详情，服务已下架，获取失败")
    def test_market_detail_05(self, Headers):
        global mcp_id

        release_data = {
                "status": "offline"
            }
        result = self.client.MCPReleaseAction(mcp_id, release_data, Headers)
        assert result[0] == 200

        # 尝试获取已下架服务的市场详情
        result = self.client.GetMCPMarketDetail(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("获取MCP服务市场详情，mcp的mode为stream，获取成功，返回stream_url")
    def test_market_detail_06(self, Headers):
        global mcp_id1
        result = self.client.GetMCPMarketDetail(mcp_id1, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["mcp_id"] == mcp_id1
        # 验证connection_info
        assert "connection_info" in result[1]
        assert "stream_url" in result[1]["connection_info"]
        # 验证tool_configs
        assert "tool_configs" not in result[1]["base_info"]

    @allure.title("获取MCP服务市场详情，mcp为从工具箱导入，获取成功，返回connection_info")
    def test_market_detail_07(self, Headers):
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
        mcp_id = result[1]["mcp_id"]
        result = self.client.MCPReleaseAction(mcp_id, {"status": "published"}, Headers)
        assert result[0] == 200
        result = self.client.GetMCPMarketDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["mcp_id"] == mcp_id
        # 验证connection_info
        assert "connection_info" in result[1]
        assert "stream_url" in result[1]["connection_info"]
        assert "sse_url" in result[1]["connection_info"]