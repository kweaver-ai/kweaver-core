# -*- coding:UTF-8 -*-

import pytest
import allure
import uuid
import string
import random

from lib.operator import Operator
from common.get_content import GetContent

operators = []

@allure.feature("算子注册与管理接口测试：获取算子历史版本详情")
class TestGetOperatorHistoryDetail:
    '''历史版本只能是已发布或者已下架的状态'''
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operators

        # 注册算子并发布
        filepath = "./resource/openapi/compliant/test0.json"
        api_data = GetContent(filepath).jsonfile()
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(data, Headers)
        assert re[0] == 200
        publish_data_list = []
        for operator in re[1]:
            if operator["status"] == "success":
                op = {
                    "operator_id": operator["operator_id"],
                    "version": operator["version"]
                }
                operators.append(op)

                publish_data = {
                    "operator_id": operator["operator_id"],
                    "status": "published"
                }
                publish_data_list.append(publish_data)

        result = self.client.UpdateOperatorStatus(publish_data_list, Headers)
        assert result[0] == 200
                

    @allure.title("获取算子历史版本详情，算子只有一个已发布版本，获取成功")
    def test_get_operator_history_detail_01(self, Headers):
        global operators
        result = self.client.GetOperatorHistoryDetail(operators[0]["operator_id"], operators[0]["version"], Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operators[0]["operator_id"]
        assert result[1]["version"] == operators[0]["version"]
        assert result[1]["status"] == "published"

    @allure.title("获取算子历史版本详情，算子存在不同状态的版本，获取成功")
    def test_get_operator_history_detail_02(self, Headers):
        global operators
        operator_id = operators[1]["operator_id"]
        version = operators[1]["version"]
        # 变更算子状态 
        data = [{
            "operator_id": operator_id,
            "status": "offline"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200
        # 编辑算子
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
        assert re[1]["status"] == "unpublish"
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["version"] != version
        new_version = result[1]["version"]

        result = self.client.GetOperatorHistoryDetail(operator_id, version, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == version
        assert result[1]["status"] == "offline"

        # 变更算子状态 
        data = [{
            "operator_id": operator_id,
            "status": "published"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200
        # 编辑算子
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operator_id,
            "description": "test edit 1234567 dgfh",
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200

        result = self.client.GetOperatorHistoryDetail(operator_id, version, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == version
        assert result[1]["status"] == "offline"

        result = self.client.GetOperatorHistoryDetail(operator_id, new_version, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == new_version
        assert result[1]["status"] == "published"

    @allure.title("获取算子历史版本详情，算子不存在，获取失败")
    def test_get_operator_history_detail_03(self, Headers):
        operator_id = str(uuid.uuid4())
        version = str(uuid.uuid4())
        result = self.client.GetOperatorHistoryDetail(operator_id, version, Headers)
        assert result[0] == 404 