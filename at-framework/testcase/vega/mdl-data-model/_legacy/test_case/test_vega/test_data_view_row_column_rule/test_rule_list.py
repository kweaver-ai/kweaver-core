import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.assertions import asserts

@allure.feature("数据视图行列权限")
class TestDataViewRowColumnRuleList:

    @allure.story("获取数据视图行列权限列表")
    @allure.title("获取数据视图行列权限列表-默认参数")
    def test_list_rules_default(self):
        """测试默认参数获取数据视图行列权限列表"""
        with allure.step("获取数据视图行列权限列表"):
            resp = DataViewsApi().list_row_column_rules()

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)

    @allure.story("获取数据视图行列权限列表")
    @allure.title("获取数据视图行列权限列表-带参数")
    def test_list_rules_with_params(self):
        """测试带参数获取数据视图行列权限列表"""
        with allure.step("获取数据视图行列权限列表-带参数"):
            resp = DataViewsApi().list_row_column_rules(
                name="测试权限规则",
                offset=0,
                limit=20,
                sort="name",
                direction="desc"
            )

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)

    @allure.story("获取数据视图行列权限列表")
    @allure.title("获取数据视图行列权限列表-按视图ID过滤")
    def test_list_rules_by_view_id(self):
        """测试按视图ID获取数据视图行列权限列表"""
        with allure.step("获取数据视图行列权限列表-按视图ID过滤"):
            resp = DataViewsApi().list_row_column_rules(
                view_id="test_view_id"
            )

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)

    @allure.story("获取数据视图行列权限列表")
    @allure.title("获取数据视图行列权限列表-按标签过滤")
    def test_list_rules_by_tag(self):
        """测试按标签获取数据视图行列权限列表"""
        with allure.step("获取数据视图行列权限列表-按标签过滤"):
            resp = DataViewsApi().list_row_column_rules(
                tag="test"
            )

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)