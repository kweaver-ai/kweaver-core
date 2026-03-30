import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common.read_data import data_reader
from src.common.assertions import asserts
from src.common import data_gen as DG

positive_data = data_reader.read_type_cases("vega/metric_model/group_update.yaml", "positive")
negative_data = data_reader.read_type_cases("vega/metric_model/group_update.yaml", "negative")

@allure.feature("指标模型")
class TestMetricModelGroupUpdate:

    @pytest.fixture(scope="function")
    def create_test_group(self):
        api = MetricModelApi()
        resp = api.create_metric_model_group(f"测试更新分组_{DG.ts_uuid()}", "原始备注")
        group_id = json.loads(resp.text).get('id')
        yield group_id
        api.delete_metric_model_group(group_id, force=True)

    @allure.story("修改指标模型分组-正常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", positive_data)
    def test_update_group_positive(self, create_test_group, group_data, title):
        with allure.step(f"修改指标模型分组: {group_data.get('case_name')}"):
            resp = MetricModelApi().update_metric_model_group(
                create_test_group,
                group_data.get("new_name"),
                group_data.get("new_comment")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 204)
            asserts.response_status_code(resp, expected_status)

    @allure.story("修改指标模型分组-异常场景")
    @allure.title("{title}")
    @pytest.mark.parametrize("group_data,title", negative_data)
    def test_update_group_negative(self, create_test_group, group_data, title):
        case_name = group_data.get('case_name')
        
        if case_name == "修改为重复名称":
            api = MetricModelApi()
            existing_name = f"已存在分组_{DG.ts_uuid()}"
            api.create_metric_model_group(existing_name, "用于重复名称测试")
            group_data = group_data.copy()
            group_data["new_name"] = existing_name
        
        with allure.step(f"尝试修改指标模型分组: {case_name}"):
            resp = MetricModelApi().update_metric_model_group(
                create_test_group,
                group_data.get("new_name"),
                group_data.get("new_comment")
            )

        with allure.step("验证HTTP响应状态码"):
            expected_status = group_data.get('expected_status', 400)
            asserts.response_status_code(resp, expected_status)