# -*- coding:UTF-8 -*-

import allure
import pytest

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：删除算子")
class TestDeleteOperator:
    client = Operator()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        global operator_list

        filepath = "./resource/openapi/compliant/relations.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200

    @allure.title("单个算子删除，算子存在且未发布，删除算子，删除成功")
    def test_delete_operator_01(self, Headers):
        filepath = "./resource/openapi/compliant/del-test1.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        assert len(result[1]) == 1
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(data, Headers)
        assert result[0] == 200

        data = {
            "page_size": -1,
            "status": "unpublish"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        ids = []
        for operator in result[1]["data"]:
            ids.append(operator["operator_id"])
        assert operator_id not in ids

    @allure.title("单个算子删除，算子不存在，删除算子，删除失败")
    def test_delete_operator_02(self, Headers):
        operator_id = "test"
        version = "1.0.0"
        data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(data, Headers)
        assert result[0] == 404

    @allure.title("批量算子删除，算子均存在且未发布，删除算子，删除成功")
    def test_delete_operator_03(self, Headers):
        filepath = "./resource/openapi/compliant/del-test2.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_ids = []
        data = []
        for operator in result[1]:
            operator_ids.append(operator["operator_id"])
            operator_data = {
                "operator_id": operator["operator_id"],
                "version": operator["version"]
            }
            data.append(operator_data)

        result = self.client.DeleteOperator(data, Headers)
        assert result[0] == 200

        data = {
            "page_size": -1,
            "status": "unpublish"
        }
        result = self.client.GetOperatorList(data, Headers)
        assert result[0] == 200
        ids = []
        for operator in result[1]["data"]:
            ids.append(operator["operator_id"])
        for id in operator_ids:
            assert id not in ids

    @allure.title("批量算子删除，部分算子不存在且未发布，删除算子，删除失败")
    def test_delete_operator_04(self, Headers):
        filepath = "./resource/openapi/compliant/del-test2.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi"
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_ids = []
        for operator in result[1]:
            operator_ids.append(operator["operator_id"])

        data = []
        for operator_id in operator_ids:
            result = self.client.GetOperatorInfo(operator_id, Headers)
            assert result[0] == 200
            version = result[1]["version"]
            operator_data = {
                "operator_id": operator_id,
                "version": version
            }
            data.append(operator_data)

        not_exits_operator = {
            "operator_id": "test",
            "version": "V1"
        }
        data.append(not_exits_operator)

        result = self.client.DeleteOperator(data, Headers)
        assert result[0] == 404

    @allure.title("算子为已发布状态，删除算子，删除失败")
    def test_delete_operator_04(self, Headers):
        filepath = "./resource/openapi/compliant/del-test3.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        del_data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(del_data, Headers)
        assert result[0] == 400  # 已发布算子不支持删除

    @allure.title("算子已下架但无引用关系，删除算子，删除成功")
    def test_delete_operator_05(self, Headers):
        filepath = "./resource/openapi/compliant/del-test1.yaml"
        yaml_data = GetContent(filepath).yamlfile()
        data = {
            "data": str(yaml_data),
            "operator_metadata_type": "openapi",
            "direct_publish": True
        }
        result = self.client.RegisterOperator(data, Headers)
        assert result[0] == 200
        operator_id = result[1][0]["operator_id"]
        version = result[1][0]["version"]
        
        data = [
            {
                "operator_id": operator_id,
                "status": "offline"
            }
        ]
        result = self.client.UpdateOperatorStatus(data, Headers)  # 下架算子
        assert result[0] == 200  

        del_data = [
            {
                "operator_id": operator_id,
                "version": version
            }
        ]
        result = self.client.DeleteOperator(del_data, Headers)
        assert result[0] == 200