import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common.read_data import data_reader
from src.common.assertions import asserts
from src.common import data_gen as DG

positive_data = data_reader.read_type_cases("vega/metric_model/model_create.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/model_create.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelCreate:

    @pytest.fixture(scope="function")
    def create_test_group(self):
        api = MetricModelApi()
        resp = api.create_metric_model_group(f"测试指标分组_{DG.ts_uuid()}", "用于指标创建测试")
        group_id = json.loads(resp.text).get('id')
        yield group_id
        api.delete_metric_model_group(group_id, force=True)

    @allure.story("创建指标模型-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("model_data,title", positive_data)
    def test_create_model_positive(self, create_test_group, model_data, title):
        with allure.step(f"创建指标模型: {model_data.get('case_name')}"):
            resp = MetricModelApi().create_metric_model(
                name=model_data.get("name"),
                id=model_data.get("name"),
                metric_type=model_data.get("metric_type"),
                query_type=model_data.get("query_type"),
                formula=model_data.get("formula"),
                formula_config=model_data.get("formula_config"),
                data_source_id=model_data.get("data_source_id"),
                data_source_type=model_data.get("data_source_type"),
                date_field=model_data.get("date_field"),
                date_format=model_data.get("date_format"),
                measure_field=model_data.get("measure_field"),
                unit_type=model_data.get("unit_type"),
                unit=model_data.get("unit"),
                group_id=create_test_group,
                tags=model_data.get("tags"),
                task=model_data.get("task")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = model_data.get('expected_status', 201)
            asserts.response_status_code(resp, expected_status)

        with allure.step("验证响应数据结构"):
            response_data = json.loads(resp.text)
            asserts.dict_contains_keys(response_data[0], ['id'])

    @allure.story("创建指标模型-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("model_data,title", negative_data)
    def test_create_model_negative(self, create_test_group, model_data, title):
        if model_data.get('case_name') == '重复名称创建指标':
            with allure.step("先创建一个指标"):
                MetricModelApi().create_metric_model(
                    name=model_data.get("name"),
                    id=model_data.get("name"),
                    metric_type=model_data.get("metric_type"),
                    query_type=model_data.get("query_type"),
                    formula=model_data.get("formula"),
                    data_source_id=model_data.get("data_source_id"),
                    data_source_type=model_data.get("data_source_type"),
                    unit_type=model_data.get("unit_type"),
                    unit=model_data.get("unit"),
                    group_id=create_test_group
                )

        with allure.step(f"尝试创建指标模型: {model_data.get('case_name')}"):
            resp = MetricModelApi().create_metric_model(
                name=model_data.get("name"),
                id=model_data.get("name"),
                metric_type=model_data.get("metric_type"),
                query_type=model_data.get("query_type"),
                formula=model_data.get("formula"),
                data_source_id=model_data.get("data_source_id"),
                data_source_type=model_data.get("data_source_type"),
                unit_type=model_data.get("unit_type"),
                unit=model_data.get("unit"),
                group_id=create_test_group
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = model_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)