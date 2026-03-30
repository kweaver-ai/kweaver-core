import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi, MetricsModelsApi
from src.common.read_data import data_reader
from src.common.assertions import asserts
from src.common import data_gen as DG

positive_data = data_reader.read_type_cases("vega/metric_model/group_delete.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/group_delete.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelGroupDelete:

    @pytest.fixture(scope="function")
    def create_empty_group(self):
        api = MetricModelApi()
        resp = api.create_metric_model_group(f"空测试分组_{DG.ts_uuid()}", "用于删除测试")
        group_id = json.loads(resp.text).get('id')
        yield group_id

    @pytest.fixture(scope="function")
    def create_non_empty_group(self):
        api = MetricModelApi()
        resp = api.create_metric_model_group(f"非空测试分组_{DG.ts_uuid()}", "用于删除测试")
        group_id = json.loads(resp.text).get('id')
        yield group_id
        api.delete_metric_model_group(group_id, force=True)

    @allure.story("删除指标模型分组-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", positive_data)
    def test_delete_group_positive(self, group_data, title):
        if group_data.get('case_name') == '删除空分组':
            with allure.step("创建空分组用于测试"):
                api = MetricModelApi()
                resp = api.create_metric_model_group(f"空测试分组_{DG.ts_uuid()}", "用于删除测试")
                group_id = json.loads(resp.text).get('id')

            with allure.step(f"删除指标模型分组: {group_data.get('case_name')}"):
                resp = api.delete_metric_model_group(group_id, group_data.get('force', False))

            with allure.step("验证HTTP响应状态码"):
                expected_status = group_data.get('expected_status', 204)
                asserts.response_status_code(resp, expected_status)

        elif group_data.get('case_name') == '强制删除非空分组':
            with allure.step("创建非空分组用于测试"):
                api = MetricModelApi()
                resp = api.create_metric_model_group(f"非空测试分组_{DG.ts_uuid()}", "用于删除测试")
                group_id = json.loads(resp.text).get('id')

            with allure.step(f"强制删除指标模型分组: {group_data.get('case_name')}"):
                resp = api.delete_metric_model_group(group_id, group_data.get('force', True))

            with allure.step("验证HTTP响应状态码"):
                expected_status = group_data.get('expected_status', 204)
                asserts.response_status_code(resp, expected_status)

    @allure.story("删除指标模型分组-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", negative_data)
    def test_delete_group_negative(self, group_data, title):
        if group_data.get('case_name') == '删除非空分组不强制':
            with allure.step("创建非空分组用于测试"):
                metric_api = MetricModelApi()
                resp = metric_api.create_metric_model_group(f"非空测试分组_{DG.ts_uuid()}", "用于删除测试")
                group_id = json.loads(resp.text).get('id')

            with allure.step("在分组内创建指标模型"):
               # 生成一个 UUID 用于指标模型的 name 和 id
               model_uuid = DG.ts_uuid()
               model_resp = metric_api.create_metric_model(
                    name=f"测试指标_{model_uuid}",
                    id=f"测试指标_{model_uuid}",
                    metric_type="atomic",
                    query_type="promql",
                    formula="100",
                    data_source_id="__isf_audit_log",
                    data_source_type="DSL",
                    unit_type="numUnit",
                    unit="none",
                    group_id=group_id
                )
               
               with allure.step("验证指标模型创建成功"):
                   allure.attach(f"指标模型创建响应状态码: {model_resp.status_code}", "响应信息", allure.attachment_type.TEXT)
                   allure.attach(f"指标模型创建响应内容: {model_resp.text}", "响应信息", allure.attachment_type.TEXT)
                   asserts.response_status_code(model_resp, 201)
                   response_data = json.loads(model_resp.text)
                   assert isinstance(response_data, list) and len(response_data) > 0, "指标模型创建响应格式不正确"
                   model_id = response_data[0].get('id')
                   assert model_id is not None, "指标模型ID为空"
                   allure.attach(f"创建的指标模型ID: {model_id}", "响应信息", allure.attachment_type.TEXT)
                   allure.attach(f"关联的分组ID: {group_id}", "响应信息", allure.attachment_type.TEXT)

            with allure.step(f"尝试删除非空分组: {group_data.get('case_name')}"):
                resp = metric_api.delete_metric_model_group(group_id, group_data.get('force', False))

            with allure.step("验证HTTP响应状态码"):
                expected_status = group_data.get('expected_status', 403)
                asserts.response_status_code(resp, expected_status)
            
            with allure.step("记录实际响应信息"):
                allure.attach(f"实际状态码: {resp.status_code}", "响应信息", allure.attachment_type.TEXT)
                allure.attach(f"响应内容: {resp.text}", "响应信息", allure.attachment_type.TEXT)

            with allure.step("清理测试数据"):
                metric_api.delete_metric_model_group(group_id, force=True)

        elif group_data.get('case_name') == '删除不存在的分组':
            with allure.step(f"尝试删除不存在的分组: {group_data.get('case_name')}"):
                resp = MetricModelApi().delete_metric_model_group(
                    group_data.get('group_id'),
                    group_data.get('force', False)
                )

            with allure.step("验证HTTP响应状态码"):
                expected_status = group_data.get('expected_status', 404)
                asserts.response_status_code(resp, expected_status)

        elif group_data.get('case_name') == '无效的force参数':
            with allure.step("创建分组用于测试"):
                api = MetricModelApi()
                resp = api.create_metric_model_group(f"测试分组_{DG.ts_uuid()}", "用于删除测试")
                group_id = json.loads(resp.text).get('id')

            with allure.step(f"尝试使用无效参数删除分组: {group_data.get('case_name')}"):
                resp = api.delete_metric_model_group(group_id, group_data.get('force'))

            with allure.step("验证HTTP响应状态码"):
                expected_status = group_data.get('expected_status', 400)
                asserts.response_status_code(resp, expected_status)

            with allure.step("清理测试数据"):
                api.delete_metric_model_group(group_id, force=True)