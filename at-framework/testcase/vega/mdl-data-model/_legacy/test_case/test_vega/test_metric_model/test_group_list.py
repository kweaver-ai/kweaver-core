import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common.read_data import data_reader
from src.common.assertions import asserts

positive_data = data_reader.read_type_cases("vega/metric_model/group_list.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/group_list.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelGroupList:

    @allure.story("查询指标模型分组列表-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", positive_data)
    def test_list_groups_positive(self, group_data, title):
        with allure.step(f"查询指标模型分组列表: {group_data.get('case_name')}"):
            resp = MetricModelApi().list_metric_model_groups(
                sort=group_data.get("sort"),
                direction=group_data.get("direction"),
                offset=group_data.get("offset"),
                limit=group_data.get("limit")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 200)
            asserts.response_status_code(resp, expected_status)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])

    @allure.story("查询指标模型分组列表-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", negative_data)
    def test_list_groups_negative(self, group_data, title):
        with allure.step(f"尝试查询指标模型分组列表: {group_data.get('case_name')}"):
            resp = MetricModelApi().list_metric_model_groups(
                sort=group_data.get("sort"),
                direction=group_data.get("direction"),
                offset=group_data.get("offset"),
                limit=group_data.get("limit")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)