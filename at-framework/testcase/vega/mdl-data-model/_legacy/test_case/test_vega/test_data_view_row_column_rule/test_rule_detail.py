import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.assertions import asserts

@allure.feature("数据视图行列权限")
class TestDataViewRowColumnRuleDetail:

    @allure.story("获取数据视图行列权限详情")
    @allure.title("获取单个数据视图行列权限详情")
    def test_get_single_rule_detail(self):
        """测试获取单个数据视图行列权限详情"""
        # 首先创建一个测试规则
        rule_data = [{
            "name": "测试权限规则详情",
            "view_id": "test_view_id",
            "tags": ["test"],
            "comment": "这是一个测试权限规则",
            "fields": ["field1", "field2"]
        }]
        
        create_resp = DataViewsApi().create_row_column_rule(rule_data)
        assert create_resp.status_code == 201
        
        rule_id = json.loads(create_resp.text)[0].get('id')
        assert rule_id is not None
        
        try:
            with allure.step(f"获取数据视图行列权限详情: {rule_id}"):
                resp = DataViewsApi().get_row_column_rule(rule_id)

            with allure.step("验证HTTP响应状态码"):
                asserts.response_status_code(resp, 200)

            with allure.step("验证响应数据结构"):
                response_data = json.loads(resp.text)
                assert isinstance(response_data, list)
                assert len(response_data) > 0
                
                rule_detail = response_data[0]
                expected_keys = ['id', 'name', 'view_id', 'view_name', 'tags', 'comment', 
                               'create_time', 'update_time', 'creator', 'updater', 'fields']
                for key in expected_keys:
                    assert key in rule_detail, f"Key '{key}' is missing in rule detail"
        finally:
            # 清理：删除测试规则
            try:
                DataViewsApi().delete_row_column_rule(rule_id)
            except Exception as e:
                print(f"清理规则时出错: {e}")

    @allure.story("获取数据视图行列权限详情")
    @allure.title("获取多个数据视图行列权限详情")
    def test_get_multiple_rule_details(self):
        """测试获取多个数据视图行列权限详情"""
        # 首先创建两个测试规则
        rule_data = [
            {
                "name": "测试权限规则详情1",
                "view_id": "test_view_id",
                "tags": ["test"],
                "comment": "这是第一个测试权限规则",
                "fields": ["field1", "field2"]
            },
            {
                "name": "测试权限规则详情2",
                "view_id": "test_view_id",
                "tags": ["test"],
                "comment": "这是第二个测试权限规则",
                "fields": ["field1", "field2"]
            }
        ]
        
        create_resp = DataViewsApi().create_row_column_rule(rule_data)
        assert create_resp.status_code == 201
        
        response_data = json.loads(create_resp.text)
        rule_ids = [item['id'] for item in response_data if 'id' in item]
        assert len(rule_ids) == 2
        
        try:
            with allure.step(f"获取多个数据视图行列权限详情: {','.join(rule_ids)}"):
                resp = DataViewsApi().get_row_column_rule(','.join(rule_ids))

            with allure.step("验证HTTP响应状态码"):
                asserts.response_status_code(resp, 200)

            with allure.step("验证响应数据结构"):
                response_data = json.loads(resp.text)
                assert isinstance(response_data, list)
                assert len(response_data) == 2
                
                for rule_detail in response_data:
                    expected_keys = ['id', 'name', 'view_id', 'view_name', 'tags', 'comment', 
                                   'create_time', 'update_time', 'creator', 'updater', 'fields']
                    for key in expected_keys:
                        assert key in rule_detail, f"Key '{key}' is missing in rule detail"
        finally:
            # 清理：删除测试规则
            try:
                DataViewsApi().delete_row_column_rule(','.join(rule_ids))
            except Exception as e:
                print(f"清理规则时出错: {e}")

    @allure.story("获取数据视图行列权限详情-异常场景")
    @allure.title("获取不存在的数据视图行列权限详情")
    def test_get_non_existent_rule_detail(self):
        """测试获取不存在的数据视图行列权限详情"""
        non_existent_rule_id = "non_existent_rule"
        
        with allure.step(f"尝试获取不存在的数据视图行列权限详情: {non_existent_rule_id}"):
            resp = DataViewsApi().get_row_column_rule(non_existent_rule_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 404)