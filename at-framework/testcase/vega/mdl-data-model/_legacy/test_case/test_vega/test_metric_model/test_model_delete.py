import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common.read_data import data_reader
from src.common.assertions import asserts
from src.common import data_gen as DG

positive_data = data_reader.read_type_cases("vega/metric_model/model_delete.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/model_delete.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelDelete:

    @pytest.fixture(scope="function")
    def create_test_model(self):
        api = MetricModelApi()
        group_resp = api.create_metric_model_group(f"测试删除分组_{DG.ts_uuid()}", "用于删除测试")
        group_id = json.loads(group_resp.text).get('id')
        
        model_resp = api.create_metric_model(
            name=f"测试删除指标_{DG.ts_uuid()}",
            id=f"测试删除指标_{DG.ts_uuid()}",
            metric_type="atomic",
            query_type="promql",
            formula="100",
            data_source_id="__isf_audit_log",
            data_source_type="DSL",
            unit_type="numUnit",
            unit="none",
            group_id=group_id
        )
        model_id = json.loads(model_resp.text)[0].get('id')
        
        yield model_id, group_id
        api.delete_metric_model_group(group_id, force=True)

    @allure.story("删除指标模型-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("model_data,title", positive_data)
    def test_delete_model_positive(self, create_test_model, model_data, title):
        model_id, group_id = create_test_model
        
        if model_data.get('case_name') == '删除单个指标模型':
            with allure.step(f"删除指标模型: {model_data.get('case_name')}"):
                resp = MetricModelApi().delete_metric_model(model_id)

            with allure.step("验证HTTP响应状态码"):
                expected_status = model_data.get('expected_status', 204)
                asserts.response_status_code(resp, expected_status)

        elif model_data.get('case_name') == '批量删除指标模型':
            with allure.step("创建第二个指标用于批量删除测试"):
                api = MetricModelApi()
                model_resp = api.create_metric_model(
                    name=f"测试批量删除指标_{DG.ts_uuid()}",
                    id=f"测试批量删除指标_{DG.ts_uuid()}",
                    metric_type="atomic",
                    query_type="promql",
                    formula="100",
                    data_source_id="__isf_audit_log",
                    data_source_type="DSL",
                    unit_type="numUnit",
                    unit="none",
                    group_id=group_id
                )
                model_id2 = json.loads(model_resp.text)[0].get('id')

            with allure.step(f"批量删除指标模型: {model_data.get('case_name')}"):
                resp = MetricModelApi().delete_metric_models([model_id, model_id2])

            with allure.step("验证HTTP响应状态码"):
                expected_status = model_data.get('expected_status', 204)
                asserts.response_status_code(resp, expected_status)

    @allure.story("删除指标模型-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("model_data,title", negative_data)
    def test_delete_model_negative(self, model_data, title):
        if model_data.get('case_name') == '删除不存在的指标模型':
            with allure.step(f"尝试删除不存在的指标模型: {model_data.get('case_name')}"):
                resp = MetricModelApi().delete_metric_model(model_data.get('model_id'))

            with allure.step("验证HTTP响应状态码"):
                expected_status = model_data.get('expected_status', 404)
                asserts.response_status_code(resp, expected_status)

        elif model_data.get('case_name') == '删除空的ID列表':
            with allure.step(f"尝试删除空的ID列表: {model_data.get('case_name')}"):
                resp = MetricModelApi().delete_metric_models([])

            with allure.step("验证HTTP响应状态码"):
                expected_status = model_data.get('expected_status', 400)
                asserts.response_status_code(resp, expected_status)