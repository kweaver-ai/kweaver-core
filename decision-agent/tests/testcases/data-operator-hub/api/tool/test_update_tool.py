# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox
from lib.operator import Operator

box_id = ""
tools_id = []

@allure.feature("工具注册与管理接口测试：更新工具")
class TestUpdateTool:
    
    client = ToolBox()
    client1 = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global tools_id

        # 创建工具箱
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

        # 获取工具箱内工具列表
        result = self.client.GetBoxToolsList(box_id, None, Headers)
        tools = result[1]["tools"]
        for tool in tools:
            tools_id.append(tool["tool_id"])

    @allure.title("更新工具，传参正确，更新成功")
    def test_update_tool_01(self, Headers):
        global box_id
        global tools_id

        # 更新工具
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description",
            "use_rule": "quis labore ipsum",
	        "extend_info": {},
            "global_parameters": {
                "in": "query",
                "name": "www",
                "type": "string",
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
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 200

    @allure.title("更新工具，工具箱不存在，更新失败")
    def test_update_tool_02(self, Headers):
        global tools_id

        box_id = str(uuid.uuid4())
        update_data = {
            "name": "test_update_tool",
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具，工具不存在，更新失败")
    def test_update_tool_03(self, Headers):
        global box_id

        tool_id = str(uuid.uuid4())
        update_data = {
            "name": "test_tool_update",
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tool_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具，工具名称已存在，更新失败")
    def test_update_tool_04(self, Headers):
        global box_id
        global tools_id

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 200

        result = self.client.UpdateTool(box_id, tools_id[1], update_data, Headers)
        assert result[0] == 409 

    @allure.title("更新工具，必填参数name不传，更新失败")
    def test_update_tool_05(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400        

    @allure.title("更新工具，必填参数description不传，更新失败")
    def test_update_tool_06(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400  

    @allure.title("更新工具，参数位置不在支持范围内，更新失败")
    def test_update_tool_07(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name",
            "description": "test tool update description",
            "global_parameters": {
                "in": "session",
                "name": "www",
                "type": "string",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            }
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400  

    @allure.title("更新工具，参数类型不在支持范围内，更新失败")
    def test_update_tool_08(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name",
            "description": "test tool update description",
            "global_parameters": {
                "in": "header",
                "name": "www",
                "type": "json",
                "value": "pariatur est eu ex sed",
                "required": True,
                "description": "test desctiption"
            }
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400

    @allure.title("更新通过算子转换成的工具，更新成功")
    def test_update_tool_09(self, Headers):
        global box_id
        global tools_id

        # 注册算子
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }

        result = self.client1.RegisterOperator(data, Headers)
        assert result[0] == 200
        operators = result[1]
        operator_id = operators[0]["operator_id"]
        operator_version = operators[0]["version"]

        # 转换算子为工具
        convert_data = {
            "box_id": box_id,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 200
        tool_id = result[1]["tool_id"]
        tools_id.append(tool_id)

        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        update_data = {
            "name": name,
            "description": "test tool update description"
        }
        result = self.client.UpdateTool(box_id, tool_id, update_data, Headers)
        assert result[0] == 200

        # 验证更新结果
        result = self.client.GetTool(box_id, tool_id, Headers)
        assert result[0] == 200
        assert result[1]["tool_id"] == tool_id
        assert result[1]["name"] == name
        assert result[1]["description"] == "test tool update description"

        # 算子不变
        result = self.client1.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["name"] != name

    @allure.title("更新工具元数据，metadata_type为其他类型，更新失败")
    def test_update_tool_10(self, Headers):
        global box_id
        global tools_id

        update_data = {
            "name": "test_tool_update_name_10",
            "description": "test tool update description",
            "metadata_type": "tool"
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 400

    @allure.title("更新算子转换成的工具的元数据，更新失败")
    def test_update_tool_11(self, Headers):
        global box_id
        global tools_id

        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        update_data = {
            "name": "test_tool_update_name_11",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": api_data
        }
        result = self.client.UpdateTool(box_id, tools_id[-1], update_data, Headers)
        assert result[0] == 405

    @allure.title("更新工具元数据，未匹配到当前工具，编辑失败")
    def test_update_tool_12(self, Headers):
        global box_id
        global tools_id

        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        update_data = {
            "name": "test_tool_update_name_12",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": api_data
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 404

    @allure.title("更新工具元数据，openapi中包含多个工具，可匹配到当前工具，更新成功")
    def test_update_tool_13(self, Headers):
        global box_id
        global tools_id

        filepath = "./resource/openapi/compliant/test.json"
        api_data = GetContent(filepath).jsonfile()
        update_data = {
            "name": "test_tool_update_name_13",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": api_data
        }
        result = self.client.UpdateTool(box_id, tools_id[0], update_data, Headers)
        assert result[0] == 200

    @allure.title("更新工具元数据，openapi中仅包含一个工具，可匹配到当前工具，更新成功")
    def test_update_tool_14(self, Headers):
        global box_id

        filepath = "./resource/openapi/compliant/tool.json"
        data = GetContent(filepath).jsonfile()
        tool_data = {
            "metadata_type": "openapi",
            "data": data
        }
        result = self.client.CreateTool(box_id, tool_data, Headers)
        assert result[0] == 200
        tool_id = result[1]["success_ids"][0]
        update_data = {
            "name": "test_tool_update_name_14",
            "description": "test tool update description",
            "metadata_type": "openapi",
            "data": data
        }
        result = self.client.UpdateTool(box_id, tool_id, update_data, Headers)
        assert result[0] == 200