# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid
import pytest

from common.get_content import GetContent
from lib.tool_box import ToolBox

box_id = ""
name = ''.join(random.choice(string.ascii_letters) for i in range(8))

@allure.feature("工具注册与管理接口测试：更新工具箱状态")
class TestUpdateToolboxStatus:
    '''
    状态流转：
        unpublish -> published
        published -> offline
        offline -> published
        offline -> unpublish
    '''
    
    client = ToolBox()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global box_id
        global name

        # 创建工具箱
        filepath = "./resource/openapi/compliant/test.json"
        json_data = GetContent(filepath).jsonfile()
        
        data = {
            "box_name": name,
            "data": json_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

    @allure.title("更新工具箱状态，状态无冲突更新成功，状态存在冲突，更新失败")
    def test_update_toolbox_status_01(self, Headers):
        global box_id

        # 未发布 -> 下架
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

        # 未发布 -> 已发布
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["status"] == "published"

        # 已发布 -> 已发布
        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

        # 已发布 -> 下架
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200
        assert result[1]["box_id"] == box_id
        assert result[1]["status"] == "offline"

        # 下架 -> 下架
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

        # 下架 -> 未发布
        update_data = {
            "status": "unpublish"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

    @allure.title("更新工具箱状态，工具箱不存在，更新失败")
    def test_update_toolbox_status_02(self, Headers):
        box_id = str(uuid.uuid4())
        update_data = {
            "status": "offline"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("更新工具箱状态，更新为无效状态，更新失败")
    def test_update_toolbox_status_03(self, Headers):
        global box_id

        update_data = {
            "status": "invalid_status"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 400

    @allure.title("发布工具箱，存在同名已发布工具箱，发布失败")
    def test_update_toolbox_status_04(self, Headers):
        # 创建工具箱
        filepath = "./resource/openapi/compliant/template.yaml"
        api_data = GetContent(filepath).yamlfile()
        
        data = {
            "box_name": name,
            "data": api_data,
            "metadata_type": "openapi"
        }
        result = self.client.CreateToolbox(data, Headers)
        box_id = result[1]["box_id"]

        result1 = self.client.CreateToolbox(data, Headers)
        box_id1 = result1[1]["box_id"]

        update_data = {
            "status": "published"
        }
        result = self.client.UpdateToolboxStatus(box_id, update_data, Headers)
        assert result[0] == 200

        result = self.client.UpdateToolboxStatus(box_id1, update_data, Headers)
        assert result[0] == 400