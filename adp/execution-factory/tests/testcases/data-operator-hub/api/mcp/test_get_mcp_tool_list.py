# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent

@allure.feature("MCP服务管理接口测试：获取指定MCP服务的工具列表")
class TestGetMCPTools:
    
    client = MCP()
    tool_client = ToolBox()

    @allure.title("获取指定MCP服务的工具列表，mcp存在，获取成功")
    def test_get_mcp_tools_01(self, Headers):
        # 创建MCP Server
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
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
        result = self.client.GetMCPToolList(mcp_id, Headers)
        assert result[0] == 200
        assert len(result[1]["tools"]) > 0

    @allure.title("获取指定MCP服务的工具列表，mcp不存在，获取失败")
    def test_get_mcp_tools_02(self, Headers):
        mcp_id = str(uuid.uuid4())
        result = self.client.GetMCPToolList(mcp_id, Headers)
        assert result[0] == 404

    @allure.title("获取指定MCP服务的工具列表，url无法解析，获取失败")
    def test_get_mcp_tools_03(self, Headers):
        # 创建MCP Server
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "http://1270.0.01/sse/tools",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id = result[1]["mcp_id"]
        result = self.client.GetMCPToolList(mcp_id, Headers)
        assert result[0] == 504

    @allure.title("获取指定MCP服务的工具列表，mcp为从工具导入的mcp，获取成功")
    def test_get_mcp_tools_04(self, Headers):
        # 创建MCP Server
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
        result = self.client.GetMCPToolList(mcp_id, Headers)
        assert result[0] == 200
        assert len(result[1]["tools"]) > 0