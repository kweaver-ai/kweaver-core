import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common.read_data import data_reader
from src.common.assertions import asserts

positive_data = data_reader.read_type_cases("vega/metric_model/model_list.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/model_list.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelList:

    @allure.story("查询指标模型列表-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("model_data,title", positive_data)
    def test_list_models_positive(self, model_data, title):
        with allure.step(f"查询指标模型列表: {model_data.get('case_name')}"):
            resp = MetricModelApi().list_metric_models(
                name_pattern=model_data.get("name_pattern"),
                name=model_data.get("name"),
                metric_type=model_data.get("metric_type"),
                query_type=model_data.get("query_type"),
                sort=model_data.get("sort"),
                direction=model_data.get("direction"),
                offset=model_data.get("offset"),
                limit=model_data.get("limit"),
                tag=model_data.get("tag")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = model_data.get('expected_status', 200)
            asserts.response_status_code(resp, expected_status)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data, ['entries', 'total_count'])

    @allure.story("查询指标模型列表-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("model_data,title", negative_data)
    def test_list_models_negative(self, model_data, title):
        with allure.step(f"尝试查询指标模型列表: {model_data.get('case_name')}"):
            resp = MetricModelApi().list_metric_models(
                name_pattern=model_data.get("name_pattern"),
                name=model_data.get("name"),
                metric_type=model_data.get("metric_type"),
                query_type=model_data.get("query_type"),
                offset=model_data.get("offset"),
                limit=model_data.get("limit")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = model_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)