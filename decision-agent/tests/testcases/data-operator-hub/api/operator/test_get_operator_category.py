# -*- coding:UTF-8 -*-

import allure

from jsonschema import Draft7Validator

from common.get_content import GetContent
from lib.operator import Operator


@allure.feature("算子注册与管理接口测试：获取算子分类")
class TestGetOperatorCategory:
    client = Operator()
    operator_category_success = GetContent("./response/data-operator-hub/agent-operator-integration/operator_category_response_success.json").jsonfile()
    
    @allure.title("获取算子分类，获取成功")
    def test_get_operator_category_01(self, Headers):
        result = self.client.GetOperatorCategory(Headers)
        assert result[0] == 200
        validator = Draft7Validator(self.operator_category_success)
        assert validator.is_valid(result)

        sys_categorys = ["other_category", "data_process", "data_transform", "data_store", "data_analysis", "data_query", "data_extract", "data_split", "model_train", "system"]
        categorys = []
        for category in result[1]:
            categorys.append(category["category_type"])

        for category in sys_categorys:
            assert category in categorys