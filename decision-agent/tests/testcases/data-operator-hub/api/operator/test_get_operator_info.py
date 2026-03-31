# -*- coding:UTF-8 -*-

import pytest
import allure

from jsonschema import Draft7Validator

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：算子信息获取")
class TestGetOperatorInfo:
    client = Operator()

    failed_resp = GetContent("./response/data-operator-hub/agent-operator-integration/response_failed.json").jsonfile()

    @pytest.fixture(scope="class", autouse=True)
    def setup(self, Headers):
        filepath = "./resource/openapi/compliant/test2.yaml"
        operator_data = GetContent(filepath).yamlfile()
        req_data = {
            "data": str(operator_data),
            "operator_metadata_type": "openapi"
        }
        re = self.client.RegisterOperator(req_data, Headers)
        assert re[0] == 200
        count = int(len(re[1])/2)
        operators = re[1]
        for i in range(count):
            operator_id = operators[i]["operator_id"]
            version = operators[i]["version"]            
            update_data = [{
                "operator_id": operator_id,
                "version": version,
                "status": "published"
            }]
            re = self.client.UpdateOperatorStatus(update_data, Headers)
            assert re[0] == 200

    @allure.title("获取算子信息，算子不存在，获取失败")
    def test_get_operator_info_01(self, Headers):
        operator_id = "b6ea4229-c2a1-4007-b2f0-ab8301ccd31c"
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 404
        validator = Draft7Validator(self.failed_resp)
        assert validator.is_valid(result)

    @allure.title("获取算子信息，算子存在并已发布，获取成功")
    def test_get_operator_info_02(self, Headers):
        data = {"status": "published"}
        result = self.client.GetOperatorList(data, Headers)
        operators = result[1]["data"]
        operator_id = operators[0]["operator_id"]
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id

    @allure.title("获取算子信息，算子存在但未发布，获取成功")
    def test_get_operator_info_03(self, Headers):
        data = {"status": "unpublish"}
        result = self.client.GetOperatorList(data, Headers)
        operators = result[1]["data"]
        operator_id = operators[0]["operator_id"]
        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id

    @allure.title("获取算子信息，算子存在多个版本，默认获取到最新版本")
    def test_get_operator_info_04(self, Headers):
        data = {"status": "published"}
        result = self.client.GetOperatorList(data, Headers)
        operators = result[1]["data"]
        operator_id = operators[0]["operator_id"]
        version = operators[0]["version"]

        re = self.client.GetOperatorInfo(operator_id, Headers)
        assert re[0] == 200

        data = {
            "operator_id": operator_id,
            "description": "test edit 1234567"
        }
        edit_re = self.client.EditOperator(data, Headers)    # 编辑算子
        assert edit_re[0] == 200
        assert edit_re[1]["version"] != version
        assert edit_re[1]["status"] == "editing"

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == edit_re[1]["version"]
        assert result[1]["status"] == "editing"

        data = [{
            "operator_id": edit_re[1]["operator_id"],
            "status": "published"
        }]
        re = self.client.UpdateOperatorStatus(data, Headers)    # 发布算子
        assert re[0] == 200

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == edit_re[1]["version"]
        assert result[1]["status"] == "published"

        data = [{
            "operator_id": edit_re[1]["operator_id"],
            "status": "offline"
        }]
        re = self.client.UpdateOperatorStatus(data, Headers)    # 下架算子
        assert re[0] == 200

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == edit_re[1]["version"]
        assert result[1]["status"] == "offline"

        data = {
            "operator_id": operator_id,
            "description": "test edit 123fhg4567"
        }
        edit_re1 = self.client.EditOperator(data, Headers)    # 编辑算子
        assert edit_re1[0] == 200
        assert edit_re1[1]["version"] != edit_re[1]["version"]
        assert edit_re1[1]["status"] == "unpublish"

        result = self.client.GetOperatorInfo(operator_id, Headers)
        assert result[0] == 200
        assert result[1]["operator_id"] == operator_id
        assert result[1]["version"] == edit_re1[1]["version"]
        assert result[1]["status"] == "unpublish"
        