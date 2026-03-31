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
box_ids = []

@allure.feature("MCP服务管理接口测试：编辑MCP_Server配置")
class TestEditMCP:
    
    client = MCP()
    tool_client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global mcp_id, mcp_id1
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

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "add mcp server config",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": []
        }
        result = self.client.RegisterMCP(data, Headers)
        assert result[0] == 200
        mcp_id1 = result[1]["mcp_id"]

        # 创建工具箱
        for i in range(5):
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
            box_ids.append(box_id) 
        result = self.tool_client.UpdateToolboxStatus(box_ids[4], {"status": "published"}, Headers)
        assert result[0] == 200       

    @allure.title("编辑MCP_Server配置，mcp存在，参数正确，编辑成功")
    def test_update_mcp_01(self, Headers):
        global mcp_id
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "updated test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "headers": {
                "Content-Type": "application/json"
            },
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 200

        # 验证更新结果
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["name"] == name
        assert result[1]["base_info"]["description"] == "updated test mcp server"

    @allure.title("编辑MCP_Server配置，缺少必需参数，编辑失败")
    def test_update_mcp_02(self, Headers):
        global mcp_id
        data = {
            "description": "test mcp server",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("编辑MCP_Server配置，mode参数不正确，编辑失败")
    def test_update_mcp_03(self, Headers):
        global mcp_id
        data = {
            "name": "test_mcp",
            "mode": "invalid_mode",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("编辑MCP_Server配置，mcp不存在，编辑失败")
    def test_update_mcp_04(self, Headers):
        mcp_id = str(uuid.uuid4())
        data = {
            "name": "test_mcp",
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 404

    @allure.title("编辑MCP_Server配置，name已存在，编辑失败")
    def test_update_mcp_05(self, Headers):
        global mcp_id
        # 先创建另一个MCP Server并发布
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
        mcp_id1 = result[1]["mcp_id"]
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id1, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        # 尝试将第一个MCP Server的name更新为第二个MCP Server的name
        update_data = {
            "name": name,
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, update_data, Headers)
        assert result[0] == 400 

    @allure.title("编辑MCP_Server配置，名称超过50个字符，编辑失败")
    def test_update_mcp_06(self, Headers):
        global mcp_id
        data = {
            "name": "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa",
            "description": "updated test mcp server",
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("编辑MCP_Server配置，名称包含特殊字符，编辑失败")
    @pytest.mark.parametrize("name", ["invalid name","name~","name@","name`","name#","name$","name%","name^","name^","name&", 
                                      "name*","name()","name-","name+","name=","name[]","name{}","name|","name\\","name:",
                                      "name;","name'","name,","name.","name?","name/","name<","name>","name；","name“","name：",
                                      "name’","name【】","name《","name》","name？","name·","name、","name，","name。"]) 
    def test_update_mcp_07(self, name, Headers):
        global mcp_id
        data = {
            "name": name,
            "description": "updated test mcp server",
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("编辑MCP_Server配置，描述超过255个字符，编辑失败")
    def test_update_mcp_08(self, Headers):
        global mcp_id
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "invalid_desc: more then 255 characters, aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
            "mode": "sse",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("编辑已发布MCP，编辑后mcp状态为editing，再次编辑后仍为editing")
    def test_update_mcp_09(self, Headers):
        global mcp_id

        # 发布mcp
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"

        # 编辑mcp
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "Edit test mcp server",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 200
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "editing"

        # 编辑mcp
        data = {
            "name": name,
            "description": "Edit test mcp server 2",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 200
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "editing"

    @allure.title("编辑已下架MCP，编辑后mcp状态为unpublish，再次编辑后仍为unpublish")
    def test_update_mcp_10(self, Headers):
        global mcp_id
        # 发布mcp
        data = {
            "status": "published"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "published"
        # 下架mcp
        data = {
            "status": "offline"
        }
        result = self.client.MCPReleaseAction(mcp_id, data, Headers)
        assert result[0] == 200
        assert result[1]["status"] == "offline"

        # 编辑mcp
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "Edit the offline mcp server first",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 200
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "unpublish"

        # 编辑mcp
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "Edit the offline mcp server sencend",
            "mode": "sse",
            "url": "https://mcp.map.baidu.com/sse?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 200
        # 验证状态
        result = self.client.GetMCPDetail(mcp_id, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["status"] == "unpublish"

    @allure.title("编辑MCP_Server配置，类型无效，编辑失败")
    def test_update_mcp_11(self, Headers):
        global mcp_id
        data = {
            "name": "test_mcp",
            "mode": "sse",
            "category": "invalid_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 400

    @allure.title("编辑自定义mcp，mode修改为stream，编辑成功")
    def test_update_mcp_12(self, Headers):
        global mcp_id
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "Edit the mcp server to stream",
            "mode": "stream",
            "url": "https://mcp.map.baidu.com/mcp?ak=bW9A9vyhGcYmdKRvWJCkySpekiBUTeUL",
            "category": "data_analysis"
        }
        result = self.client.EditMCP(mcp_id, data, Headers)
        assert result[0] == 200

    @allure.title("编辑从工具箱导入的mcp，添加工具，工具符合要求，编辑成功")
    def test_update_mcp_13(self, Headers):
        global mcp_id1, box_ids
        box_id = box_ids[0]
        result = self.tool_client.GetBoxToolsList(box_id, None, Headers)
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
                "box_name": "box_name",
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
            "description": "Edit the mcp server: add tool",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.EditMCP(mcp_id1, data, Headers)
        assert result[0] == 200

    @allure.title("编辑从工具箱导入的mcp，添加工具，工具存在同名，编辑失败")
    def test_update_mcp_14(self, Headers):
        global mcp_id1, box_ids
        box_id = box_ids[3]
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
                "box_name": "box_name",
                "tool_id": tool["tool_id"],
                "tool_name": "dup_name",
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
            "description": "Edit the mcp server: add tool",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.EditMCP(mcp_id1, data, Headers)
        assert result[0] == 400

    @allure.title("编辑从工具箱导入的mcp，添加工具，添加后工具列表总数超过30个，编辑失败")
    def test_update_mcp_15(self, Headers):
        global mcp_id1
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
            "description": "Edit the mcp server: add tool",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.EditMCP(mcp_id1, data, Headers)
        assert result[0] == 400

    @allure.title("编辑从工具箱导入的mcp，工具名称不合法，编辑失败")
    @pytest.mark.parametrize("invalid_name", ["space name", "invalied-name", "special!@#$%^&*()|【《，}", "invalid_name:_more_then_50_characters_aaaaaaaaaaaaa"])
    def test_update_mcp_16(self, invalid_name, Headers):
        global mcp_id1, box_ids
        box_id = box_ids[4]
        result = self.tool_client.GetBoxToolsList(box_ids[4], None, Headers)
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
                "box_id": box_ids[4],
                "box_name": "box_name",
                "tool_id": tool["tool_id"],
                "tool_name": invalid_name,
                "description": tool["description"],
                "use_rule": "all"
            }
            tool_configs.append(tool_config)       
        result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "Edit the mcp server: add tool",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.EditMCP(mcp_id1, data, Headers)
        assert result[0] == 400

    @allure.title("编辑mcp，工具名称和描述不传，编辑成功")
    def test_update_mcp_17(self, Headers):
        global mcp_id1, box_ids
        box_id = box_ids[4]
        result = self.tool_client.GetBoxToolsList(box_ids[4], None, Headers)
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
                "box_id": box_ids[4],
                "box_name": "box_name",
                "tool_id": tool["tool_id"]
            }
            tool_configs.append(tool_config)       
        result = self.tool_client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "name": name,
            "description": "Edit the mcp server: add tool",
            "category": "data_analysis",
            "creation_type": "tool_imported",
            "tool_configs": tool_configs
        }
        result = self.client.EditMCP(mcp_id1, data, Headers)
        assert result[0] == 200
        result = self.client.GetMCPDetail(mcp_id1, Headers)
        assert result[0] == 200
        assert result[1]["base_info"]["category"] == "data_analysis"
        assert result[1]["base_info"]["creation_type"] == "tool_imported"
        assert len(result[1]["base_info"]["tool_configs"]) > 0