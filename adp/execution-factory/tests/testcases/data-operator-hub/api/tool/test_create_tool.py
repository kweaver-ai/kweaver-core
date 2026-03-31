# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = ""

@allure.feature("工具注册与管理接口测试：创建工具")
class TestCreateTool:
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

    @allure.title("创建工具，传参正确，创建成功")
    def test_create_tool_01(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/tool.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "metadata_type": "openapi",
            "data": data,
            "use_rule": "quis labore ipsum",
	        "extend_info": {},
            "global_parameters": {
                "in": "cookie",
                "name": "yiyayiya",
                "type": "array",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            },
            "quota_control": {
                "quota_type": "ip",
                "quota_value": 1000,
                "time_window": {
                    "unit": "second",
                    "value": 1
                },
                "burst_capacity": 100,
                "overage_policy": "queue"
            }
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 200
        assert result[1]["success_count"] == 1
        assert result[1]["failure_count"] == 0

    @allure.title("创建工具，工具箱不存在，创建失败")
    def test_create_tool_02(self, Headers):
        filepath = "./resource/openapi/compliant/tool.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "metadata_type": "openapi",
            "data": data
        }
        box_id = str(uuid.uuid4())
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 400

    @allure.title("创建工具，工具名称已存在，创建失败")
    def test_create_tool_03(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/tool.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "data": data,
            "metadata_type": "openapi"
        }

        # 创建同名工具
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 200
        assert result[1]["success_count"] == 0
        assert result[1]["failure_count"] == 1

    @allure.title("创建工具，必填参数metadata_type不传，创建失败")
    def test_create_tool_04(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/mcp.yaml"
        yaml_data = GetContent(filepath).yamlfile()

        # 创建工具，不传metadata_type
        tool_data = {
            "data": yaml_data
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 400

    @allure.title("创建工具，必填参数data不传，创建失败")
    def test_create_tool_05(self, Headers):
        global box_id

        # 创建工具，不传data
        tool_data = {
            "metadata_type": "openapi"
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 400 

    @allure.title("创建工具，参数位置不在支持范围内，创建失败")
    def test_create_tool_06(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/setup.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "metadata_type": "openapi",
            "data": data,
            "global_parameters": {
                "in": "auth",
                "name": "yiyayiya",
                "type": "array",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            }
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 400

    @allure.title("创建工具，参数类型不在支持范围内，创建失败")
    def test_create_tool_07(self, Headers):
        global box_id
        filepath = "./resource/openapi/compliant/test0.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "metadata_type": "openapi",
            "data": data,
            "global_parameters": {
                "in": "body",
                "name": "yiyayiya",
                "type": "number",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            }
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 400