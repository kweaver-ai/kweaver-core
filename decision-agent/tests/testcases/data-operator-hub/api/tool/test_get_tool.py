# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = ""
tools_id = []

@allure.feature("工具注册与管理接口测试：获取单个工具")
class TestGetTool:
    
    client = ToolBox()

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

    @allure.title("获取工具，工具存在，获取成功")
    def test_get_tool_01(self, Headers):
        global box_id
        global tools_id

        # 获取工具
        result = self.client.GetTool(box_id, tools_id[0], Headers)
        assert result[0] == 200
        assert result[1]["tool_id"] == tools_id[0]

    @allure.title("获取工具，工具箱不存在，获取失败")
    def test_get_tool_02(self, Headers):
        global tools_id

        box_id = str(uuid.uuid4())
        result = self.client.GetTool(box_id, tools_id[0], Headers)
        assert result[0] == 400

    @allure.title("获取工具，工具不存在，获取失败")
    def test_get_tool_03(self, Headers):
        global box_id
        
        tool_id = str(uuid.uuid4())
        result = self.client.GetTool(box_id, tool_id, Headers)
        assert result[0] == 400
