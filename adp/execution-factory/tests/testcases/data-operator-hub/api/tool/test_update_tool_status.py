# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id1 = ""
tools_id1 = []
box_id2 = ""
tools_id2 = []

@allure.feature("工具注册与管理接口测试：更新工具状态")
class TestUpdateToolStatus:
    
    client = ToolBox()

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

    @allure.title("更新工具状态，工具箱及所有待更新工具均存在，更新成功")
    def test_update_tool_status_01(self, Headers):
        global box_id1
        global tools_id1

        # 更新工具状态
        update_data = [{
            "tool_id": tools_id1[0],
            "status": "enabled"
        },
        {
            "tool_id": tools_id1[1],
            "status": "enabled"
        }]
        result = self.client.UpdateToolStatus(box_id1, update_data, Headers)
        assert result[0] == 200

        for id in tools_id1[2:]:
            result = self.client.GetTool(box_id1, id, Headers)
            assert result[0] == 200
            assert result[1]["status"] == "disabled"

        for id in tools_id1[:2]:
            result = self.client.GetTool(box_id1, id, Headers)
            assert result[0] == 200
            assert result[1]["status"] == "enabled"

    @allure.title("更新工具状态，工具箱不存在，更新失败")
    def test_update_tool_status_02(self, Headers):
        global tools_id1

        box_id = str(uuid.uuid4())
        update_data = [{
            "tool_id": tools_id1[2],
            "status": "enabled"
        }]
        result = self.client.UpdateToolStatus(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具状态，其中部分工具不存在，更新失败")
    def test_update_tool_status_03(self, Headers):
        global box_id2
        global tools_id2

        # 更新工具状态，包含不存在的工具ID
        tool_id = str(uuid.uuid4())
        update_data = [{
            "tool_id": tools_id2[0],
            "status": "enabled"
        },
        {
            "tool_id": tool_id,
            "status": "enabled"
        }]
        result = self.client.UpdateToolStatus(box_id2, update_data, Headers)
        assert result[0] == 400

        result = self.client.GetTool(box_id2, tools_id2[0], Headers)
        assert result[0] == 200
        assert result[1]["status"] == "disabled"

    @allure.title("更新工具状态，其中部分工具不在该工具箱内，更新成功，不在该工具箱内的工具跳过")
    def test_update_tool_status_04(self, Headers):
        global box_id1
        global box_id2
        global tools_id1
        global tools_id2

        # 更新工具状态，包含其他工具箱的工具ID
        update_data = [{
            "tool_id": tools_id1[2],
            "status": "enabled"
        },
        {
            "tool_id": tools_id1[1],
            "status": "enabled"
        }]
        result = self.client.UpdateToolStatus(box_id1, update_data, Headers)
        assert result[0] == 200

        result = self.client.GetTool(box_id1, tools_id1[2], Headers)
        assert result[0] == 200
        assert result[1]["status"] == "enabled"

        result = self.client.GetTool(box_id2, tools_id2[1], Headers)
        assert result[0] == 200
        assert result[1]["status"] == "disabled"

    @allure.title("更新工具状态，状态值不正确，更新失败")
    def test_update_tool_status_05(self, Headers):
        global box_id1
        global tools_id1

        # 更新工具状态为无效值
        update_data = [{
            "tool_id": tools_id1[3],
            "status": "invalid_status"
        }]
        result = self.client.UpdateToolStatus(box_id1, update_data, Headers)
        assert result[0] == 400