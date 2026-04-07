# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent

mcp_id = ""
mcp_id1 = ""

@allure.feature("MCP服务管理接口测试：MCP工具调试")
class TestMCPToolDebug:
    
    client = MCP()
    tool_client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global mcp_id, mcp_id1
        # 创建自定义MCP Server
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
        # 创建从工具导入的mcp
        filepath = "./resource/openapi/compliant/toolbox.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.tool_client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        file = GetContent("./config/env.ini")
        config = file.config()
        host = config["server"]["host"]
        update_data = {
            "box_name": name,
            "box_desc": "test toolbox update description",
            "box_svc_url": f"https://{host}/api/agent-operator-integration",
            "box_icon": "icon-color-tool-FADB14",
            "box_category": "data_process"
        }
        result = self.tool_client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200
        result = self.tool_client.GetBoxToolsList(box_id, {"all": True}, Headers)
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
        mcp_id1 = result[1]["mcp_id"]

    @allure.title("MCP工具调试，参数正确，调试成功")
    def test_debug_tool_01(self, Headers):
        global mcp_id
        data = {"address": "上海东方明珠"}
        result = self.client.MCPToolDebug(mcp_id, "map_geocode", data, Headers)
        assert result[0] == 200
        assert "content" in result[1]
        assert isinstance(result[1]["content"], list)
        assert result[1]["is_error"] == False

    @allure.title("MCP工具调试，参数错误，调试失败")
    def test_debug_tool_02(self, Headers):
        global mcp_id
        data = {
            "parameters": "invalid_parameters"
        }
        result = self.client.MCPToolDebug(mcp_id, "map_geocode", data, Headers)
        assert result[0] == 200
        assert result[1]["is_error"] == True

    @allure.title("MCP工具调试，mcp_id不存在，调试失败")
    def test_debug_tool_03(self, Headers):
        mcp_id = str(uuid.uuid4())
        data = {"address": "上海东方明珠"}
        result = self.client.MCPToolDebug(mcp_id, "map_geocode", data, Headers)
        assert result[0] == 404

    @allure.title("MCP工具调试，name不存在，调试失败")
    def test_debug_tool_04(self, Headers):
        global mcp_id
        data = {"address": "上海东方明珠"}
        result = self.client.MCPToolDebug(mcp_id, "not_exist_tool", data, Headers)
        assert result[0] == 200
        assert result[1]["is_error"] == True

    @allure.title("调试工具转换成的MCP下的工具，传参正确，调试成功")
    def test_debug_tool_05(self, Headers):
        global mcp_id1
        data = {"header": Headers}
        result = self.client.MCPToolDebug(mcp_id1, "获取工具箱列表", data, Headers)
        assert result[0] == 200
        assert result[1]["is_error"] == False