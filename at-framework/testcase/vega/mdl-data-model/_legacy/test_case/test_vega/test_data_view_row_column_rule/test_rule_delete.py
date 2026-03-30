import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.assertions import asserts

@allure.feature("数据视图行列权限")
class TestDataViewRowColumnRuleDelete:

    @allure.story("删除数据视图行列权限-正常场景")
    @allure.title("删除单个数据视图行列权限")
    def test_delete_single_rule(self):
        """测试删除单个数据视图行列权限"""
        # 首先创建一个测试规则
        rule_data = [{
            "name": "待删除的权限规则",
            "view_id": "test_view_id",
            "tags": ["test"],
            "comment": "这是一个待删除的权限规则",
            "fields": ["field1", "field2"]
        }]
        
        create_resp = DataViewsApi().create_row_column_rule(rule_data)
        assert create_resp.status_code == 201
        
        rule_id = json.loads(create_resp.text)[0].get('id')
        assert rule_id is not None
        
        with allure.step(f"删除数据视图行列权限: {rule_id}"):
            resp = DataViewsApi().delete_row_column_rule(rule_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 204)

    @allure.story("删除数据视图行列权限-正常场景")
    @allure.title("删除多个数据视图行列权限")
    def test_delete_multiple_rules(self):
        """测试删除多个数据视图行列权限"""
        # 首先创建两个测试规则
        rule_data = [
            {
                "name": "待删除的权限规则1",
                "view_id": "test_view_id",
                "tags": ["test"],
                "comment": "这是第一个待删除的权限规则",
                "fields": ["field1", "field2"]
            },
            {
                "name": "待删除的权限规则2",
                "view_id": "test_view_id",
                "tags": ["test"],
                "comment": "这是第二个待删除的权限规则",
                "fields": ["field1", "field2"]
            }
        ]
        
        create_resp = DataViewsApi().create_row_column_rule(rule_data)
        assert create_resp.status_code == 201
        
        response_data = json.loads(create_resp.text)
        rule_ids = [item['id'] for item in response_data if 'id' in item]
        assert len(rule_ids) == 2
        
        with allure.step(f"删除多个数据视图行列权限: {','.join(rule_ids)}"):
            resp = DataViewsApi().delete_row_column_rule(','.join(rule_ids))

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 204)

    @allure.story("删除数据视图行列权限-异常场景")
    @allure.title("删除不存在的数据视图行列权限")
    def test_delete_non_existent_rule(self):
        """测试删除不存在的数据视图行列权限"""
        non_existent_rule_id = "non_existent_rule"
        
        with allure.step(f"尝试删除不存在的数据视图行列权限: {non_existent_rule_id}"):
            resp = DataViewsApi().delete_row_column_rule(non_existent_rule_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 404)