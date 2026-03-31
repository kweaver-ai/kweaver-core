# -*- coding:UTF-8 -*-

import allure
import pytest
import string
import random

from lib.mcp import MCP
from lib.tool_box import ToolBox
from common.get_content import GetContent

@allure.feature("MCP服务管理接口测试：添加MCP_Server配置")
class TestRegisterMCP:
    
    client = MCP()
    tool_client = ToolBox()

    @allure.title("添加MCP_Server配置，参数正确，添加成功")
    def test_register_mcp_01(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "command": "ls",
            "args": ["a", "b", "c"],
            "headers": {
                "Content-Type": "application/json"
            },
            "env": {},
            "source": "custom",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        assert "mcp_id" in result[1]
        assert result[1]["status"] == "unpublish"

    @allure.title("添加MCP_Server配置，缺少必需参数，添加失败")
    def test_register_mcp_02(self, Headers):
        data = {"description": "test mcp server", "mode": "sse", "category": "data_analysis"}
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，mode参数不正确，添加失败")
    def test_register_mcp_03(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "mode": "invalid_mode",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，name包含特殊字符，添加失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                      "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                      "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                      "name’","name【】","name《","name》","name？","name·","name、","name，","name。"])   
    def test_register_mcp_04(self, name, Headers):
        data = {
            "name": name,
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，name超过50个字符，添加失败")
    def test_register_mcp_05(self, Headers):
        data = {
            "name": "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa",
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，存在同名未发布mcp，添加成功")
    def test_register_mcp_06(self, Headers):
        # 已发布的mcp不允许重名
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

        # 再次添加相同name的MCP Server
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200 
    
    @allure.title("添加MCP_Server配置，存在同名已发布mcp，添加失败")
    def test_register_mcp_07(self, Headers):
        # 已发布的mcp不允许重名
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

        publish_data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, publish_data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        # 再次添加相同name的MCP Server
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400 

    @allure.title("添加MCP_Server配置，描述超过255个字符，添加失败")
    def test_register_mcp_08(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "invalid_desc: more then 255 characters, aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，创建类型非法，添加失败")
    def test_register_mcp_09(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "add mcp server config",
            "mode": "sse",
            "url": "http://localhost:8080/api/v1/tools",
            "category": "data_analysis",
            "creation_type": "invalid_type"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，创建类型为tool_imported，添加成功")
    def test_register_mcp_10(self, Headers):
        # 创建工具箱、启用工具并发布
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
        for tool in tools:
            data = {
                "tool_id": tool["tool_id"],
                "status": "enabled"
            }
            update_data.append(data) 
            tools_id.append(tool["tool_id"])         
        result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        result = self.tool_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
        assert result[0] == 200

        tool_result = self.tool_client.GetTool(box_id, tools_id[0], Headers)
        assert tool_result[0] == 200
        tool_configs = [{
            "box_id": box_id,
            "box_name": name,
            "tool_id": tools_id[0],
            "tool_name": tool_result[1]["name"],
            "description": tool_result[1]["description"],
            "use_rule": "all"
        }]
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

    @allure.title("添加MCP_Server配置，创建类型为tool_imported，存在同名工具，添加失败")
    def test_register_mcp_11(self, Headers):
        tool_configs = []
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        for i in range(2):
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
            for tool in tools:
                data = {
                    "tool_id": tool["tool_id"],
                    "status": "enabled"
                }
                update_data.append(data) 
                tools_id.append(tool["tool_id"])         
            result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
            assert result[0] == 200
            result = self.tool_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
            assert result[0] == 200

            tool_result = self.tool_client.GetTool(box_id, tools_id[0], Headers)
            assert tool_result[0] == 200
            tool_config = {
                "box_id": box_id,
                "box_name": name,
                "tool_id": tools_id[0],
                "tool_name": "mcp_tool_name",
                "description": tool_result[1]["description"],
                "use_rule": "all"
            }
            tool_configs.append(tool_config)
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
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，创建类型为tool_imported，工具名称不合法，添加失败")
    @pytest.mark.parametrize("invalid_name", ["space name", "invalied-name", "special!@#$%^&*()|{【？?<》}", "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"])
    def test_register_mcp_12(self, invalid_name, Headers):
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
        for tool in tools:
            data = {
                "tool_id": tool["tool_id"],
                "status": "enabled"
            }
            update_data.append(data) 
            tools_id.append(tool["tool_id"])         
        result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        result = self.tool_client.UpdateToolboxStatus(box_id, {"status": "published"}, Headers)
        assert result[0] == 200

        tool_result = self.tool_client.GetTool(box_id, tools_id[0], Headers)
        assert tool_result[0] == 200
        tool_configs = [{
            "box_id": box_id,
            "box_name": name,
            "tool_id": tools_id[0],
            "tool_name": invalid_name,
            "description": tool_result[1]["description"],
            "use_rule": "all"
        }]
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
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，创建类型为tool_imported，工具个数超过30个，添加失败")
    def test_register_mcp_13(self, Headers):
        filepath = "./resource/openapi/compliant/capp-model-manage.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.tool_client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]
        result = self.tool_client.GetBoxToolsList(box_id, {"all": True}, Headers)
        tools = result[1]["tools"]
        update_data = []
        tool_configs = []
        for tool in tools:
            data = {
                "tool_id": tool["tool_id"],
                "status": "enabled"
            }
            update_data.append(data) 
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
        assert result[0] == 400

    @allure.title("添加MCP_Server配置，mode为stream，添加成功")
    def test_register_mcp_14(self, Headers):
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "test mcp server",
            "mode": "stream",
            "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "category": "data_analysis"
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200