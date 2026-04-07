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

@allure.feature("MCP服务管理接口测试：调用指定MCP服务下的工具")
class TestCallMCPTool:
    
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

    @allure.title("调用MCP服务下的工具，参数正确，调试成功")
    def test_call_mcp_tool_01(self, Headers):
        global mcp_id
        data = {
            "tool_name": "map_geocode",
            "parameters": {
                "address": "上海东方明珠"
            }
        }
        result = self.client.CallMCPtool(mcp_id, data, Headers)
        assert result[0] == 200
        assert "content" in result[1]
        assert isinstance(result[1]["content"], list)
        assert result[1]["is_error"] == False

    @allure.title("调用MCP服务下的工具，parameters格式不正确，调试失败")
    def test_call_mcp_tool_02(self, Headers):
        global mcp_id
        data = {
            "tool_name": "map_geocode",
            "parameters": "invalid_parameters"
        }
        result = self.client.CallMCPtool(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("调用MCP服务下的工具，mcp_id不存在，调试失败")
    def test_call_mcp_tool_03(self, Headers):
        mcp_id = str(uuid.uuid4())
        data = {
            "tool_name": "map_geocode",
            "parameters": {
                "address": "上海东方明珠"
            }
        }
        result = self.client.CallMCPtool(mcp_id, data, Headers)
        assert result[0] == 404

    @allure.title("调用MCP服务下的工具，name不存在，调试失败")
    def test_call_mcp_tool_04(self, Headers):
        global mcp_id
        data = {
            "tool_name": "not_exist_tool",
            "parameters": {
                "address": "上海东方明珠"
            }
        }
        result = self.client.CallMCPtool(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["is_error"] == True

    @allure.title("调用工具转换成的MCP服务下的工具，参数正确，调试成功")
    def test_call_mcp_tool_05(self, Headers):
        global mcp_id1
        data = {
            "tool_name": "获取工具箱列表",
            "parameters": {
                "header": Headers
            } 
        }
        result = self.client.CallMCPtool(mcp_id1, data, Headers)
        assert result[0] == 200
        assert result[1]["is_error"] == False