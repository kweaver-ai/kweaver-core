# -*- coding:UTF-8 -*-

import allure
import string
import random
import uuid

from common.get_content import GetContent
from lib.tool_box import ToolBox


@allure.feature("工具注册与管理接口测试：删除工具箱")
class TestDeleteToolbox:
    
    client = ToolBox()

    @allure.title("删除工具箱，工具箱存在，删除成功")
    def test_delete_toolbox_01(self, Headers):
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

        # 获取工具箱中的工具列表
        params = {
            "page_size": 20
        }
        result = self.client.GetBoxToolsList(box_id, params, Headers)
        assert result[0] == 200
        tool_list = result[1]["tools"]

        # 删除工具箱
        result = self.client.DeleteToolbox(box_id, Headers)
        assert result[0] == 200

        # 验证工具箱已被删除
        result = self.client.GetToolbox(box_id, Headers)
        assert result[0] == 400

    @allure.title("删除工具箱，工具箱不存在，删除失败")
    def test_delete_toolbox_02(self, Headers): 
        box_id = str(uuid.uuid4())
        result = self.client.DeleteToolbox(box_id, Headers)
        assert result[0] == 400