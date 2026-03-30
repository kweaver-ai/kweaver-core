import pytest
import allure
import json
from src.api.mdl.metric_model_api import MetricModelApi
from src.common import data_gen as DG

@pytest.fixture(scope="session")
def metric_model_group_setup():
    """会话级别的指标模型分组设置"""
    api = MetricModelApi()
    
    with allure.step("创建指标模型测试分组"):
        group_resp = api.create_metric_model_group(
            f"指标模型测试分组_{DG.ts_uuid()}",
            "用于指标模型测试的分组"
        )
        group_id = json.loads(group_resp.text).get('id')
    
    yield group_id
    
    with allure.step("清理指标模型测试分组"):
        api.delete_metric_model_group(group_id, force=True)

@pytest.fixture(scope="function")
def create_temp_group():
    """函数级别的临时分组"""
    api = MetricModelApi()
    
    with allure.step("创建临时测试分组"):
        group_resp = api.create_metric_model_group(
            f"临时测试分组_{DG.ts_uuid()}",
            "临时测试分组"
        )
        group_id = json.loads(group_resp.text).get('id')
    
    yield group_id
    
    with allure.step("清理临时测试分组"):
        api.delete_metric_model_group(group_id, force=True)