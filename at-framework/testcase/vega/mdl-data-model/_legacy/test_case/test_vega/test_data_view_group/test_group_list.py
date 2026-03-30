import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.assertions import asserts

@allure.feature("数据视图管理")
class TestDataViewGroupList:

    @allure.story("获取数据视图分组列表")
    @allure.title("获取数据视图分组列表-默认参数")
    def test_list_groups_default(self):
        """测试默认参数获取数据视图分组列表"""
        with allure.step("获取数据视图分组列表"):
            resp = DataViewsApi().list_groups()

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)

    @allure.story("获取数据视图分组列表")
    @allure.title("获取数据视图分组列表-带参数")
    def test_list_groups_with_params(self):
        """测试带参数获取数据视图分组列表"""
        with allure.step("获取数据视图分组列表-带参数"):
            resp = DataViewsApi().list_groups(
                sort="name",
                direction="desc",
                offset=0,
                limit=20
            )

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)

    @allure.story("获取数据视图分组列表")
    @allure.title("获取内置数据视图分组")
    def test_list_builtin_groups(self):
        """测试获取内置数据视图分组"""
        with allure.step("获取内置数据视图分组"):
            resp = DataViewsApi().list_groups(builtin=True)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])
            assert isinstance(response_data['entries'], list)
            assert isinstance(response_data['total_count'], int)
            
            # 验证返回的都是内置分组
            for entry in response_data['entries']:
                assert entry.get('builtin') == True