# -*- coding:UTF-8 -*-

import pytest
import allure
import string
import random
import uuid

from lib.operator import Operator
from common.get_content import GetContent

operator_id = ""

@allure.feature("算子注册与管理接口测试：获取算子历史版本列表")
class TestGetOperatorHistoryList:
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operator_id

        # 注册算子并发布
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        re = self.client.RegisterOperator(data, Headers)
        assert re[0] == 200
        operator_id = re[1][0]["operator_id"]

        # 编辑并发布生成新版本
        re = self.client.GetOperatorInfo(operator_id, Headers)
        assert re[0] == 200
        operator_info = re[1]
        metadata_info = operator_info["metadata"]
        metadata_info["summary"] = "test_edit_1234567"
        metadata_info["description"] = "test edit 1234567"
        metadata_info["path"] = "/edit/getid"
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "metadata": metadata_info,
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"
        data = [{
            "operator_id": operator_id,
            "status": "published"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200

        #下架后编辑，生成新版本
        data = [{
            "operator_id": operator_id,
            "status": "offline"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200
        metadata_info["summary"] = "test_edit_1234567_dfhgkjh"
        metadata_info["description"] = "test edit 1234567 xfhg"
        metadata_info["path"] = "/edit/getid}"
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "metadata": metadata_info,
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"

        # 发布并编辑生成新版本
        data = [{
            "operator_id": operator_id,
            "status": "published"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200
        metadata_info["summary"] = "test_edit_hgfjhgj7"
        metadata_info["description"] = "test edit ghkjkj"
        metadata_info["path"] = "/edit/getlist}"
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "metadata": metadata_info,
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"
        
    @allure.title("获取算子历史版本列表，算子存在，获取成功")
    def test_get_operator_history_list_01(self, Headers):
        global operator_id
        # 查询历史版本列表
        result = self.client.GetOperatorHistoryList(operator_id, Headers)
        assert result[0] == 200
        assert isinstance(result[1], list)
        assert len(result[1]) == 3
        assert any(item["operator_id"] == operator_id for item in result[1])

    @allure.title("获取算子历史版本列表，算子不存在，获取失败")
    def test_get_operator_history_list_02(self, Headers):
        operator_id = str(uuid.uuid4())
        result = self.client.GetOperatorHistoryList(operator_id, Headers)
        assert result[0] == 404 