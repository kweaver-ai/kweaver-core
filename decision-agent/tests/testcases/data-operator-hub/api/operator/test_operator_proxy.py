# -*- coding:UTF-8 -*-

import allure
import uuid
import pytest
import os

from common.get_content import GetContent
from lib.operator import Operator
from lib.operator_internal import InternalOperator
from lib.impex import Impex

operator_list = []

@allure.feature("算子注册与管理接口测试：代理执行算子")
class TestOperatorProxy:

    client = Operator()
    internal_client = InternalOperator()
    impex_client = Impex()

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
            update_data = [
                {
                    "operator_id": operator["operator_id"],
                    "status": "published"
                }
            ]
            result = self.client.UpdateOperatorStatus(update_data, Headers)    # 发布算子
            assert result[0] == 200
            result = self.client.GetOperatorList({"all": "true"}, Headers)
            operator_list = result[1]["data"]

    @allure.title("代理执行算子，get接口，path传参正确，执行成功")
    def test_operator_proxy_01(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "获取算子信息":
                operator_id = operator["operator_id"]
        proxy_data = {
            "header": Headers,
            "path": {
                "operator_id": operator_id
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200

    @allure.title("代理执行算子，get接口，query传参正确，执行成功")
    def test_operator_proxy_02(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "获取算子列表":
                operator_id = operator["operator_id"]
        proxy_data = {
            "header": Headers,
            "query": {
                "page_size": "5",
                "sort_by": "name",
                "sort_order": "asc"
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200

    @allure.title("代理执行算子，get接口，header传参正确，执行成功")
    def test_operator_proxy_03(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "获取算子分类":
                operator_id = operator["operator_id"]
        proxy_data = {
            "header": Headers
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200

    @allure.title("代理执行算子，post接口，传参正确，执行成功")
    def test_operator_proxy_04(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "更新算子信息":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]
        filepath = "./resource/openapi/compliant/test3.yaml"
        api_data = GetContent(filepath).yamlfile()
        proxy_data = {
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "data": str(api_data),
                "operator_metadata_type": "openapi"
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200

    @allure.title("代理执行算子，delete接口，传参正确，执行成功")
    def test_operator_proxy_05(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "删除算子":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]
        proxy_data = {
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200

    @allure.title("代理执行算子，算子不存在或未发布，执行失败")
    def test_operator_proxy_06(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "注册算子":
                operator_id = operator["operator_id"]
                update_data = [
                    {
                        "operator_id": operator_id,
                        "status": "offline"
                    }
                ]
                result = self.client.UpdateOperatorStatus(update_data, Headers)    # 下架算子
                assert result[0] == 200
        operator_ids = [operator_id, str(uuid.uuid4())] 
        for op_id in operator_ids:      
            proxy_data = {
                "header": Headers
            }
            result = self.internal_client.ProxyOperator(op_id, proxy_data, UserHeaders)
            assert result[0] == 404

    @allure.title("代理执行算子，缺少heander信息：x-account-id，执行失败")
    def test_operator_proxy_07(self, Headers):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "更新算子状态":
                operator_id = operator["operator_id"]
                operator_version = operator["version"]            
        proxy_data = {
            "version": operator_version,
            "header": Headers,
            "body": {
                "operator_id": operator_id,
                "version": operator_version,
                "status": "published"
            }
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, None)
        assert result[0] == 400

    @allure.title("代理执行算子，指定超时时间，算子执行超时")
    def test_operator_proxy_08(self, Headers, UserHeaders):
        # 导入算子
        file = "./resource/openapi/compliant/zuhesuanzi.json"
        files = {"data": (os.path.basename(file), open(file, "rb"))}
        result = self.impex_client.importation("operator", files, {"mode":"upsert"}, Headers)
        operator_id = "54005a4b-49df-41ec-a7b9-5e73124d7db3"
        proxy_data = {
            "header": Headers,
            "body": {"a": "aaa"},
            "timeout": 1
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 200
        assert result[1]["status_code"] == 500

    @allure.title("代理执行算子，算子运行模式为异步算子，执行失败")
    def test_operator_proxy_09(self, Headers, UserHeaders):
        global operator_list
        for operator in operator_list:
            if operator["name"] == "获取算子分类":
                operator_id = operator["operator_id"]
                # 修改为异步算子
                data = {
                    "operator_id": operator_id,
                    "operator_info": {
                        "execution_mode": "async",
                        "category": "data_processing"
                    }
                }
                re = self.client.EditOperator(data, Headers)
                assert re[0] == 200
                update_data = [{
                        "operator_id": operator_id,
                        "status": "published"
                    }]
                result = self.client.UpdateOperatorStatus(update_data, Headers)    # 发布算子
                assert result[0] == 200
        proxy_data = {
            "header": Headers
        }
        result = self.internal_client.ProxyOperator(operator_id, proxy_data, UserHeaders)
        assert result[0] == 400
