# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox
from lib.operator import Operator

box_id1 = ""
tools_id1 = []
box_id2 = ""
tools_id2 = []

@allure.feature("工具注册与管理接口测试：删除工具")
class TestDeleteTool:
    
    client = ToolBox()
    client1 = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id1
        global tools_id1
        global box_id2
        global tools_id2

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
        box_id1 = result[1]["box_id"]

        filepath = "./resource/openapi/compliant/setup.json"
        json_data = GetContent(filepath).jsonfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id2 = result[1]["box_id"]

        # 获取工具箱内工具列表
        result = self.client.GetBoxToolsList(box_id1, None, Headers)
        tools = result[1]["tools"]
        for tool in tools:
            tools_id1.append(tool["tool_id"])

        result = self.client.GetBoxToolsList(box_id2, None, Headers)
        tools = result[1]["tools"]
        for tool in tools:
            tools_id2.append(tool["tool_id"])

    @allure.title("批量删除工具，工具箱及所有待删除工具均存在，删除成功")
    def test_batch_delete_tools_01(self, Headers):
        global box_id1
        global tools_id1

        tool_ids = tools_id1[:2]
        delete_data = {
            "tool_ids": tool_ids
        }
        result = self.client.BatchDeleteTools(box_id1, delete_data, Headers)
        assert result[0] == 200

        # 验证工具已被删除
        for tool_id in tool_ids:
            result = self.client.GetTool(box_id1, tool_id, Headers)
            assert result[0] == 400

    @allure.title("批量删除工具，工具箱不存在，删除失败")
    def test_batch_delete_tools_02(self, Headers):
        global tools_id1

        box_id = str(uuid.uuid4())
        tool_ids = tools_id1[2:]
        delete_data = {
            "tool_ids": tool_ids
        }
        result = self.client.BatchDeleteTools(box_id, delete_data, Headers)
        assert result[0] == 400

    @allure.title("批量删除工具，其中部分工具不存在，删除失败")
    def test_batch_delete_tools_03(self, Headers):
        global box_id1
        global tools_id1

        tool_id = str(uuid.uuid4())
        tool_ids = [tool_id, tools_id1[2]]

        delete_data = {
            "tool_ids": tool_ids
        }
        result = self.client.BatchDeleteTools(box_id1, delete_data, Headers)
        assert result[0] == 400

        # 工具仍存在
        result = self.client.GetTool(box_id1, tools_id1[2], Headers)
        assert result[0] == 200

    @allure.title("批量删除工具，其中部分工具不在该工具箱内，删除失败")
    def test_batch_delete_tools_04(self, Headers):
        global box_id1
        global box_id2
        global tools_id1
        global tools_id2

        tool_ids = [tools_id1[3], tools_id2[0]]

        delete_data = {
            "tool_ids": tool_ids
        }
        result = self.client.BatchDeleteTools(box_id1, delete_data, Headers)
        assert result[0] == 400

        # 工具仍存在
        result = self.client.GetTool(box_id1, tools_id1[3], Headers)
        assert result[0] == 200
        result = self.client.GetTool(box_id2, tools_id2[0], Headers)
        assert result[0] == 200

    @allure.title("批量删除工具，包含通过算子转换成的工具，删除成功")
    def test_batch_delete_tools_05(self, Headers):
        global box_id1
        global tools_id1

        # 注册算子
        filepath = "./resource/openapi/compliant/test3.yaml"
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
            "box_id": box_id1,
            "operator_id": operator_id,
            "operator_version": operator_version
        }
        result = self.client.ConvertOperatorToTool(convert_data, Headers)
        assert result[0] == 200
        tool_id = result[1]["tool_id"]
        tool_ids = [tool_id, tools_id1[2]]

        delete_data = {
            "tool_ids": tool_ids
        }
        result = self.client.BatchDeleteTools(box_id1, delete_data, Headers)
        assert result[0] == 200

        # 工具不存在
        result = self.client.GetTool(box_id1, tools_id1[2], Headers)
        assert result[0] == 400

        result = self.client.GetTool(box_id1, tool_id, Headers)
        assert result[0] == 400

        # 算子仍存在
        result = self.client1.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200