# -*- coding:UTF-8 -*-

import allure
import uuid
import pytest

from common.get_content import GetContent
from lib.operator import Operator

operator_list = []

@allure.feature("算子注册与管理接口测试：算子调试")
class TestOperatorDebug:

    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operator_list

        file = GetContent("./config/env.ini")
        config = file.config()
        host = config["server"]["host"]
        filepath = "./resource/openapi/compliant/operator.json"
        api_data = GetContent(filepath).jsonfile()
        server = api_data["servers"][0]
        server["url"] = f"https://{host}"
        
        data = {
            "data": str(api_data),
            "operator_metadata_type": "openapi"
        }

        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operators = result[1]
        for operator in operators:
            assert operator["status"] == "success"

            data = {
                "page_size": -1,
                "status": "unpublish"
            }
            result = self.client.GetOperatorList(data, Headers)
            operator_list = result[1]["data"]

    @allure.title("算子调试，get接口，path传参正确，调试成功")
    def test_operator_debug_01(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "获取算子信息":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "path": {
                "operator_id": operator_id
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200

    @allure.title("算子调试，get接口，query传参正确，调试成功")
    def test_operator_debug_02(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "获取算子列表":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "query": {
                "page_size": "5",
                "sort_by": "name",
                "sort_order": "asc"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200

    @allure.title("算子调试，get接口，header传参正确，调试成功")
    def test_operator_debug_03(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "获取算子分类":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200

    @allure.title("算子调试，post接口，传参正确，调试成功")
    def test_operator_debug_04(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "更新算子信息":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "data": str(api_data),
                "operator_metadata_type": "openapi"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200

    @allure.title("算子调试，delete接口，传参正确，调试成功")
    def test_operator_debug_05(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "删除算子":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]

        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 200

    @allure.title("算子调试，算子不存在，调试失败")
    def test_operator_debug_06(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "更新算子状态":
                operator_id = operator["operator_id"]
        
        operator_version = str(uuid.uuid4())
                
        debug_data = {
            "operator_id": operator_id,
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 404

    @allure.title("算子调试，缺少必填参数operator_id，调试失败")
    def test_operator_debug_07(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "更新算子状态":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]
                
        debug_data = {
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 400

    @allure.title("算子调试，缺少必填参数version，调试失败")
    def test_operator_debug_08(self, Headers):
        global operator_list

        for operator in operator_list:
            if operator["name"] == "更新算子状态":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]
                
        debug_data = {
            "operator_id": operator_id,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        result = self.client.OperatorDebug(debug_data, Headers)
        assert result[0] == 400