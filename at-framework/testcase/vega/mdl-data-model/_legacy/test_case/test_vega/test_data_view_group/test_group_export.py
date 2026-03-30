import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.assertions import asserts

@allure.feature("数据视图管理")
class TestDataViewGroupExport:

    @allure.story("按分组导出数据视图")
    @allure.title("按分组导出数据视图")
    def test_export_views_by_group(self, group_id):
        """测试按分组导出数据视图"""
        with allure.step(f"按分组导出数据视图: {group_id}"):
            resp = DataViewsApi().export_views_by_group(group_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            assert isinstance(response_data, list)
            # 即使分组中没有数据视图，也应该返回空列表

    @allure.story("按分组导出数据视图-异常场景")
    @allure.title("导出不存在分组的数据视图")
    def test_export_views_non_existent_group(self):
        """测试导出不存在分组的数据视图"""
        non_existent_group_id = "non_existent_group"
        
        with allure.step(f"尝试导出不存在分组的数据视图: {non_existent_group_id}"):
            resp = DataViewsApi().export_views_by_group(non_existent_group_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 404)