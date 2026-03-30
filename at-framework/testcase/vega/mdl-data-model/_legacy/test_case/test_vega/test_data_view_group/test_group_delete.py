import pytest
import allure
from src.api.vega.data_views_api import DataViewsApi
from src.common.assertions import asserts

@allure.feature("数据视图管理")
class TestDataViewGroupDelete:

    @allure.story("删除数据视图分组-正常场景")
    @allure.title("删除数据视图分组")
    def test_delete_group(self, group_id):
        """测试删除数据视图分组"""
        with allure.step(f"删除数据视图分组: {group_id}"):
            resp = DataViewsApi().delete_group(group_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 204)

    @allure.story("删除数据视图分组-异常场景")
    @allure.title("删除不存在的数据视图分组")
    def test_delete_non_existent_group(self):
        """测试删除不存在的数据视图分组"""
        non_existent_group_id = "non_existent_group"
        
        with allure.step(f"尝试删除不存在的数据视图分组: {non_existent_group_id}"):
            resp = DataViewsApi().delete_group(non_existent_group_id)

        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 404)