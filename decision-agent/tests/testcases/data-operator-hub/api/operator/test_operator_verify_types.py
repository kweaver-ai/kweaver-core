# -*- coding:UTF-8 -*-

import pytest
import allure
import json

from jsonschema import Draft7Validator

from common.get_content import GetContent
from common.get_case_and_params import GetCaseAndParams
from lib.operator import Operator

@allure.feature("算子注册与管理接口测试：针对返回体类型进行校验")
class TestOperatorIntegration:
    client = Operator()
    register_operator = GetCaseAndParams("./data/data-operator-hub/agent-operator-integration/register_operator_data.json")
    register_titles, register_params = register_operator.get_case_and_params()

    delete_operator = GetCaseAndParams("./data/data-operator-hub/agent-operator-integration/delete_operator_data.json")
    delete_operator_titles, delete_operator_params = delete_operator.get_case_and_params()

    update_operator_status = GetCaseAndParams("./data/data-operator-hub/agent-operator-integration/update_operator_status_data.json")
    update_operator_status_titles, update_operator_status_params = update_operator_status.get_case_and_params()

    edit_operator = GetCaseAndParams("./data/data-operator-hub/agent-operator-integration/edit_operator_data.json")
    edit_titles, edit_params = edit_operator.get_case_and_params()

    failed_resp = GetContent("./response/data-operator-hub/agent-operator-integration/response_failed.json").jsonfile()
    # failed_resp = json.loads(json.dumps(failed_resp))

    @pytest.mark.parametrize('title, data', zip(register_titles, register_params), ids=register_titles)
    def test_register_operator(self, title, data, Headers):
        allure.title(title)
        if data.get("data") is not None:
            filename = data["data"]
            if (type(filename) == str) and (".yaml" in filename or ".json" in filename):
                openapi_data = GetContent(filename).yamlfile()
                data["data"] = str(openapi_data)

        if data.get("user_token") is not None:
            data["user_token"] = Headers["Authorization"][7:]

        result = self.client.RegisterOperator(data, Headers)

        operator_register_success = GetContent("./response/data-operator-hub/agent-operator-integration/operator_register_response_success.json").jsonfile()
        # operator_list_success = json.loads(json.dumps(register_success))

        if "注册失败" in title:
            validator = Draft7Validator(self.failed_resp)
            assert validator.is_valid(result)
        else:
            validator = Draft7Validator(operator_register_success)
            assert validator.is_valid(result)

    @pytest.mark.parametrize('title, data', zip(delete_operator_titles, delete_operator_params), ids=delete_operator_titles)
    def test_delete_operator(self, title, data, Headers):
        allure.title(title)

        result = self.client.DeleteOperator(data, Headers)

        if "失败" in title:
            validator = Draft7Validator(self.failed_resp)
            assert validator.is_valid(result)
        else:
            assert result[0] == 200

    @pytest.mark.parametrize('title, data', zip(update_operator_status_titles, update_operator_status_params), ids=update_operator_status_titles)
    def test_update_operator_status(self, title, data, Headers):
        allure.title(title)

        result = self.client.UpdateOperatorStatus(data, Headers)

        if "失败" in title:
            validator = Draft7Validator(self.failed_resp)
            assert validator.is_valid(result)
        else:
            assert result[0] == 200

    @pytest.mark.parametrize('title, data', zip(edit_titles, edit_params), ids=edit_titles)
    def test_edit_operator(self, title, data, Headers):
        allure.title(title)
        result = self.client.EditOperator(data, Headers)

        if "编辑失败" in title:
            validator = Draft7Validator(self.failed_resp)
            assert validator.is_valid(result)
        else:
            return "success"