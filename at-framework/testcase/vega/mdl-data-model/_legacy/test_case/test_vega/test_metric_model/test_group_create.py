import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common.read_data import data_reader
from src.common.assertions import asserts, get_db_asserts

positive_data = data_reader.read_type_cases("vega/metric_model/group_create.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/group_create.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelGroupCreate:

    @allure.story("创建指标模型分组-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", positive_data)
    def test_create_group_positive(self, group_data, title):
        with allure.step(f"创建指标模型分组: {group_data.get('case_name')}"):
            resp = MetricModelApi().create_metric_model_group(
                group_data.get("name"),
                group_data.get("comment")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 201)
            asserts.response_status_code(resp, expected_status)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['id'])

    @allure.story("创建指标模型分组-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", negative_data)
    def test_create_group_negative(self, group_data, title):
        if group_data.get('case_name') == '重复名称创建分组':
            first_resp = MetricModelApi().create_metric_model_group(group_data.get("name"), group_data.get("comment"))
            if first_resp.status_code == 201:
                with allure.step(f"尝试创建重复名称的指标模型分组: {group_data.get('case_name')}"):
                    resp = MetricModelApi().create_metric_model_group(
                        group_data.get("name"),
                        group_data.get("comment")
                    )
            else:
                resp = first_resp
        else:
            with allure.step(f"尝试创建指标模型分组: {group_data.get('case_name')}"):
                resp = MetricModelApi().create_metric_model_group(
                    group_data.get("name"),
                    group_data.get("comment")
                )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)