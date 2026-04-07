# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random

from common.get_content import GetContent
from lib.tool_box import ToolBox


@allure.feature("工具注册与管理接口测试：获取工具箱")
class TestGetToolbox:
    
    client = ToolBox()

    @allure.title("获取工具箱，工具箱存在，获取成功")
    def test_get_toolbox_01(self, Headers):
        # 先创建一个工具箱
        filepath = "./resource/openapi/compliant/template.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "box_name": name,
            "data": yaml_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

        # 获取工具箱
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id

    @allure.title("获取工具箱，工具箱不存在，获取失败")
    def test_get_toolbox_02(self, Headers):
        box_id = str(uuid.uuid4())
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 400