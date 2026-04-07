# -*- coding:UTF-8 -*-

import allure
import uuid
import string
import random
import pytest

from lib.operator import Operator
from common.get_content import GetContent

operator_id = ""
version = ""

@allure.feature("算子注册与管理接口测试：获取算子市场指定算子详情")
class TestGetOperatorMarketDetail:
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operator_id, version
        # 注册并发布一个算子
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
        version = re[1][0]["version"]

        # 编辑并发布生成新版本
        re = self.client.GetOperatorInfo(operator_id, Headers)
        assert re[0] == 200
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "description": "test edit 1234567",
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

    @allure.title("获取算子市场详情，参数正确，获取成功，默认获取到最新版本")
    def test_market_detail_01(self, Headers):
        global operator_id, version
        result = self.client.GetOperatorMarketDetail(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] != version

    @allure.title("获取算子市场详情，operator_id不存在，获取失败")
    def test_market_detail_02(self, Headers):
        operator_id = str(uuid.uuid4())
        result = self.client.GetOperatorMarketDetail(operator_id, Headers)
        assert result[0] == 404 

    @allure.title("获取算子市场详情，算子未发布，获取失败")
    def test_market_detail_03(self, Headers):
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(data, Headers)
        assert re[0] == 200
        operator_id = re[1][0]["operator_id"]
        result = self.client.GetOperatorMarketDetail(operator_id, Headers)
        assert result[0] == 404