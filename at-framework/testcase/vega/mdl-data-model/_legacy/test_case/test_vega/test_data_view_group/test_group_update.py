import pytest
import allure
import json
from src.api.vega.data_views_api import DataViewsApi
from src.common.read_data import data_reader
from src.common.assertions import asserts
from src.common import data_gen as DG

# 读取测试数据
positive_data = data_reader.read_type_cases("vega/data_view_group/group_update.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/data_view_group/group_update.yaml", "negative")

@allure.feature("数据视图管理")
class TestDataViewGroupUpdate:

    @allure.story("更新数据视图分组-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", positive_data)
    def test_update_group_positive(self, group_data, title, group_id):
        """正向测试用例：期望更新成功"""
        # 替换测试数据中的{group_id}为实际的group_id
        actual_group_id = group_data.get("group_id").replace("{group_id}", group_id)
        
        # 处理new_name中的变量替换
        new_name = group_data.get("new_name")
        if isinstance(new_name, str) and "${" in new_name:
            # 简单的变量替换，替换${ts_uuid}等
            if "${ts_uuid}" in new_name:
                new_name = new_name.replace("${ts_uuid}", DG.ts_uuid())
        
        with allure.step(f"更新数据视图分组: {actual_group_id}"):
            resp = DataViewsApi().update_group(
                actual_group_id,
                new_name
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 200)
            asserts.response_status_code(resp, expected_status)

    @allure.story("更新数据视图分组-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", negative_data)
    def test_update_group_negative(self, group_data, title, group_id):
        """负向测试用例：期望更新失败"""
        # 替换测试数据中的{group_id}为实际的group_id
        actual_group_id = group_data.get("group_id").replace("{group_id}", group_id)
        
        # 处理new_name中的变量替换
        new_name = group_data.get("new_name")
        if isinstance(new_name, str) and "${" in new_name:
            # 简单的变量替换，替换${ts_uuid}等
            if "${ts_uuid}" in new_name:
                new_name = new_name.replace("${ts_uuid}", DG.ts_uuid())
        
        with allure.step(f"尝试更新数据视图分组: {actual_group_id}"):
            resp = DataViewsApi().update_group(
                actual_group_id,
                new_name
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)