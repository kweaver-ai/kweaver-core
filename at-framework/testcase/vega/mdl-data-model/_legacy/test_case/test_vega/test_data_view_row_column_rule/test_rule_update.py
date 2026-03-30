import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.read_data import data_reader
from src.common.assertions import asserts

# 读取测试数据
positive_data = data_reader.read_type_cases("vega/data_view_row_column_rule/rule_update.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/data_view_row_column_rule/rule_update.yaml", "negative")

@allure.feature("数据视图行列权限")
class TestDataViewRowColumnRuleUpdate:

    @allure.story("更新数据视图行列权限-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("rule_data,title", positive_data)
    def test_update_rule_positive(self, rule_data, title, rule_id):
        """正向测试用例：期望更新成功"""
        # 替换测试数据中的{rule_id}为实际的rule_id
        actual_rule_id = rule_data.get("rule_id").replace("{rule_id}", rule_id)
        
        with allure.step(f"更新数据视图行列权限: {actual_rule_id}"):
            resp = DataViewsApi().update_row_column_rule(
                actual_rule_id,
                rule_data.get("rule_data")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = rule_data.get('expected_status', 204)
            asserts.response_status_code(resp, expected_status)

    @allure.story("更新数据视图行列权限-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("rule_data,title", negative_data)
    def test_update_rule_negative(self, rule_data, title, rule_id):
        """负向测试用例：期望更新失败"""
        # 替换测试数据中的{rule_id}为实际的rule_id
        actual_rule_id = rule_data.get("rule_id").replace("{rule_id}", rule_id)
        
        with allure.step(f"尝试更新数据视图行列权限: {actual_rule_id}"):
            resp = DataViewsApi().update_row_column_rule(
                actual_rule_id,
                rule_data.get("rule_data")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = rule_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)