# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = ""
tools = []


@allure.feature("工具注册与管理接口测试：工具调试")
class TestDebugTool:
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global tools

        # 创建工具箱
        filepath = "./resource/openapi/compliant/toolbox.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

        
        # 更新工具箱
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
        result = self.client.UpdateToolbox(box_id, update_data, Headers)
        assert result[0] == 200

        # 获取工具箱内工具列表
        params = {
            "page_size": 20
        }
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        tools = result[1]["tools"]


       
    @allure.title("工具调试，Get接口，Header传参，传参正确，调试成功")
    def test_debug_tool_01(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "获取工具箱列表":
                tool_id = tool["tool_id"]

        debug_data = {
            "header": Headers
        }

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 200
        

    @allure.title("工具调试，Get接口，query传参，传参正确，调试成功")
    def test_debug_tool_02(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "查询工具列表":
                tool_id = tool["tool_id"]

        debug_data = {
            "header": Headers,
            "query": {
                "page_size": "5"
            }
        }

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 200

    @allure.title("工具调试，Get接口，path传参，传参正确，调试成功")
    def test_debug_tool_03(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "获取工具":
                tool_id = tool["tool_id"]

        debug_data = {
            "header": Headers,
            "path": {
                "box_id": box_id,
                "tool_id": tool_id
            }
        }

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 200

    @allure.title("工具调试，Get接口，传参错误，调试成功，调试接口报错")
    def test_debug_tool_04(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "获取工具":
                tool_id = tool["tool_id"]

        # 缺少必需的path参数
        debug_data = {
            "header": Headers
        }

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 200
        assert result[1]["status_code"] == 400

    @allure.title("工具调试，POST接口，调试成功")
    def test_debug_tool_05(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "创建工具箱":
                tool_id = tool["tool_id"]

        filepath = "./resource/openapi/compliant/template.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        debug_data = {
            "header": Headers,
            "body": {
                "metadata_type": "openapi",
                "data": yaml_data
            }
        }

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 200

    @allure.title("工具调试，工具箱不存在，调试失败")
    def test_debug_tool_06(self, Headers):
        global tools
        for tool in tools:
            if tool["name"] == "获取工具箱列表":
                tool_id = tool["tool_id"]

        debug_data = {
            "header": Headers
        }
        box_id = str(uuid.uuid4())

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 404

    @allure.title("工具调试，工具不存在，调试失败")
    def test_debug_tool_07(self, Headers):
        global box_id

        tool_id = str(uuid.uuid4())
        debug_data = {}

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 400

    @allure.title("工具调试，工具状态为启用，调试成功")
    def test_debug_tool_08(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "获取工具箱列表":
                tool_id = tool["tool_id"]

        # 更新工具状态为禁用
        update_data = [{
            "tool_id": tool_id,
            "status": "enabled"
        }]
        result = self.client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 200

        proxy_data = {
            "header": Headers
        }

        result = self.client.DebugTool(box_id, tool_id, proxy_data, Headers)
        assert result[0] == 200

    @allure.title("工具调试，Delete接口，调试成功")
    def test_debug_tool_09(self, Headers):
        global box_id
        global tools
        for tool in tools:
            if tool["name"] == "删除工具箱ID":
                tool_id = tool["tool_id"]

        debug_data = {
            "header": Headers,
            "path": {
                "box_id": box_id
            }
        }

        result = self.client.DebugTool(box_id, tool_id, debug_data, Headers)
        assert result[0] == 200
        