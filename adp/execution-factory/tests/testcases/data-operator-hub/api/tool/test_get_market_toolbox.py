# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random

from common.get_content import GetContent
from lib.tool_box import ToolBox


@allure.feature("工具注册与管理接口测试：在市场中获取工具箱")
class TestGetMarketToolbox:
    
    client = ToolBox()

    @allure.title("在市场中获取工具箱，工具箱存在且已发布，获取成功")
    def test_get_market_toolbox_01(self, Headers):
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
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

        # 在市场中获取工具箱
        result = self.client.GetMarketToolbox(box_id, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id

    @allure.title("在市场中获取工具箱，工具箱不存在，获取失败")
    def test_get_market_toolbox_02(self, Headers):
        box_id = str(uuid.uuid4())
        result = self.client.GetMarketToolbox(box_id, Headers)
        assert result[0] == 400

    @allure.title("在市场中获取工具箱，工具箱存在但未发布，获取失败")
    def test_get_market_toolbox_03(self, Headers):
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

        # 在市场中获取工具箱
        result = self.client.GetMarketToolbox(box_id, Headers)
        assert result[0] == 400

    @allure.title("在市场中获取工具箱，工具箱存在但已下架，获取失败")
    def test_get_market_toolbox_04(self, Headers):
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
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

        # 在市场中获取工具箱
        result = self.client.GetMarketToolbox(box_id, Headers)
        assert result[0] == 400