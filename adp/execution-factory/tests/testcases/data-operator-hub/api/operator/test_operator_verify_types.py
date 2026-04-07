# -*- coding:UTF-8 -*-

import pytest
import allure
import json
import time

from jsonschema import Draft7Validator

from common.get_content import GetContent
from common.get_case_and_params import GetCaseAndParams
from lib.operator import Operator

@allure.feature("算子注册与管理接口测试：针对返回体类型进行校验")
class TestOperatorIntegration:
    client = Operator()
    
    def _register_operator_with_retry(self, data, headers, max_retries=3, retry_delay=2):
        """
        带重试机制的算子注册方法
        
        参数：
            data: 注册数据
            headers: 请求头
            max_retries: 最大重试次数，默认3次
            retry_delay: 重试间隔（秒），默认2秒
        
        返回：
            (status_code, response_data) 元组
        
        说明：
            对于临时错误（500, 502, 503, 504）会自动重试。
            对于业务错误（400, 403, 409等）不会重试，直接返回。
        """
        result = None
        
        for attempt in range(max_retries):
            result = self.client.RegisterOperator(data, headers)
            
            # 如果成功，直接返回
            if result[0] == 200:
                return result
            
            # 如果是临时错误且还有重试机会，则重试
            if result[0] in [500, 502, 503, 504] and attempt < max_retries - 1:
                wait_time = retry_delay * (attempt + 1)
                print(f"注册算子返回 {result[0]}，{wait_time} 秒后重试（尝试 {attempt + 1}/{max_retries}）...")
                time.sleep(wait_time)
            else:
                # 非临时错误或重试次数用完，直接返回
                return result
        
        return result
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

        # 所有测试用例都使用重试机制
        # 重试机制内部会智能判断：临时错误（500,502,503,504）会重试，业务错误（400,403,409）不会重试
        # 这样即使期望失败的测试用例遇到服务不可用（503），也会自动重试
        result = self._register_operator_with_retry(data, Headers)

        operator_register_success = GetContent("./response/data-operator-hub/agent-operator-integration/operator_register_response_success.json").jsonfile()
        # operator_list_success = json.loads(json.dumps(register_success))

        if "注册失败" in title:
            validator = Draft7Validator(self.failed_resp)
            assert validator.is_valid(result), f"验证失败，状态码: {result[0]}, 响应: {result}"
        else:
            validator = Draft7Validator(operator_register_success)
            assert validator.is_valid(result), f"验证失败，状态码: {result[0]}, 响应: {result}。请检查后端服务是否正常运行。"

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