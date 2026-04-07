# -*- coding:UTF-8 -*-

import allure
import pytest
import math
import string
import random

from lib.operator import Operator
from common.assert_tools import AssertTools
from common.get_content import GetContent

operators = []
operator_id = ""

@allure.feature("算子注册与管理接口测试：获取算子市场列表")
class TestGetOperatorMarketList:
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operators, operator_id
        # 注册并发布算子
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

        # 下架算子
        data = [{
            "operator_id": operators[0]["operator_id"],
            "status": "offline"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200

        # 编辑算子
        re = self.client.GetOperatorInfo(operators[1]["operator_id"], Headers)
        assert re[0] == 200
        operator_info = re[1]
        metadata_info = operator_info["metadata"]
        metadata_info["summary"] = "test_edit_1234567"
        metadata_info["description"] = "test edit 1234567"
        metadata_info["path"] = "/edit/getid"
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operators[1]["operator_id"],
            "metadata": metadata_info,
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"

        # 下架后编辑算子
        data = [{
            "operator_id": operators[2]["operator_id"],
            "status": "offline"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200

        re = self.client.GetOperatorInfo(operators[2]["operator_id"], Headers)
        assert re[0] == 200
        operator_info = re[1]
        metadata_info = operator_info["metadata"]
        metadata_info["summary"] = "test_edit_1234567"
        metadata_info["description"] = "test edit 1234567"
        metadata_info["path"] = "/edit/getid"
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operators[2]["operator_id"],
            "metadata": metadata_info,
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "unpublish"

        # 编辑并发布生成新版本
        re = self.client.GetOperatorInfo(operators[3]["operator_id"], Headers)
        assert re[0] == 200
        operator_info = re[1]
        metadata_info = operator_info["metadata"]
        metadata_info["summary"] = "test_edit_1234567"
        metadata_info["description"] = "test edit 1234567"
        metadata_info["path"] = "/edit/getid"
        name = ''.join(random.choice(string.ascii_letters) for i in range(8))
        data = {
            "operator_id": operators[3]["operator_id"],
            "metadata": metadata_info,
            "name": name
        }
        re = self.client.EditOperator(data, Headers)
        assert re[0] == 200
        assert re[1]["status"] == "editing"
        data = [{
            "operator_id": operators[3]["operator_id"],
            "status": "published"
        }]
        result = self.client.UpdateOperatorStatus(data, Headers)
        assert result[0] == 200


    @allure.title("获取算子市场列表，默认参数，获取成功，默认仅能获取到已发布/已下架算子的最新版本，根据更新实现倒序排列")
    def test_market_list_01(self, Headers):
        result = self.client.GetOperatorMarketList(None, Headers)
        assert result[0] == 200
        assert result[1]["page"] == 1
        assert result[1]["page_size"] == 10
        operator_list = result[1]["data"]
        assert len(operator_list) <= 10
        assert result[1]["total_pages"] == math.ceil(result[1]["total"]/result[1]["page_size"])
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True
        operator_ids = []
        update_times = []
        for op in operator_list:
            operator_ids.append(op["operator_id"])
            update_times.append(op["update_time"])
            assert op["status"] == "published" or op["status"] == "offline"
        assert AssertTools.has_duplicates(operator_ids) == False
        assert AssertTools.is_descending_str(update_times) == True

    @allure.title("获取算子市场列表，page参数小于0，获取失败")
    def test_market_list_02(self, Headers):
        params = {"page": -1}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取算子市场列表，page_size为负数，获取成功，默认获取所有")
    @pytest.mark.parametrize("page_size", [-1, -2])
    def test_market_list_03(self, page_size, Headers):
        params = {"page_size": page_size}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == result[1]["total"]

    @allure.title("获取算子市场列表，page_size在[1-100]范围内，获取成功")
    @pytest.mark.parametrize("page_size", [1, 20, 50, 100])
    def test_market_list_04(self, page_size, Headers):
        params = {"page_size": page_size}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        assert result[1]["page_size"] == page_size
        assert len(result[1]["data"]) <= page_size
        if result[1]["total_pages"] > result[1]["page"]:
            assert result[1]["has_next"] == True
        else:
            assert result[1]["has_next"] == False
        if result[1]["page"] == 1:
            assert result[1]["has_prev"] == False
        elif result[1]["total_pages"] > 1:
            assert result[1]["has_prev"] == True
        for op in result[1]["data"]:
            assert op["status"] == "published" or op["status"] == "offline"

    @allure.title("获取算子市场列表，all为True获取所有算子，获取成功")
    def test_market_list_05(self, Headers):
        params = {"all": True}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        assert len(result[1]["data"]) == result[1]["total"]
        for op in result[1]["data"]:
            assert op["status"] == "published" or op["status"] == "offline"

    @allure.title("获取算子市场列表，page_size超出范围，获取失败")
    def test_market_list_06(self, Headers):
        params = {"page_size": 101}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 400

    @allure.title("获取算子市场列表，根据名称获取，获取成功")
    def test_market_list_07(self, Headers):
        params = {"name": "获取"}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        for operator in result[1]["data"]:
            assert "获取" in operator["name"]

    @allure.title("获取算子市场列表，根据类型获取，获取成功")
    def test_market_list_07(self, Headers):
        params = {"category": "data_process"}
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200

    @allure.title("查询数据源算子，获取算子市场列表成功，返回数据源算子")
    def test_market_list_08(self, Headers):
        filepath = "./resource/openapi/compliant/setup.json"
        api_data = GetContent(filepath).jsonfile()

        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi",
            "operator_info": {
                "is_data_source": True
            }
        }

        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_ids = []
        publish_data_list = []
        for op in result[1]:
            if op["status"] == "success":
                operator_ids.append(op["operator_id"])
                publish_data = {
                    "operator_id": op["operator_id"],
                    "status": "published"
                }
                publish_data_list.append(publish_data)

        result = self.client.UpdateOperatorStatus(publish_data_list, Headers)
        assert result[0] == 200

        params = {
            "is_data_source": True
        }
        result = self.client.GetOperatorMarketList(params, Headers)
        assert result[0] == 200
        assert result[1]["total"] == len(operator_ids)
        operators = result[1]["data"]
        for operator in operators:
            assert operator["operator_id"] in operator_ids
            assert operator["operator_info"]["is_data_source"] == True