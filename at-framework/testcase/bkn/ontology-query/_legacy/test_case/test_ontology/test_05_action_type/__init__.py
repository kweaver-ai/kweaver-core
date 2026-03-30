import time

import pytest
import allure
import json
from src.config.setting import config
from src.common.assertions import asserts
from src.api.ontology.kn_network_api import KnowledgeNetworkApi


@allure.feature("任务管理-创建构建任务")
class TestQueryObjectTypeInstances:
    """对象构建测试"""
    @allure.story("任务管理-创建构建任务")
    @allure.title("任务管理-创建构建任务")
    def test_ontology_job_create(self):
        api = KnowledgeNetworkApi()
        with allure.step(f"创建构建任务:{config.get('view_data','kn_id')}"):
            resp=api.kn_job(kn_id=config.get('view_data','kn_id'),
                            kn_name=config.get('view_data','kn_id')
                            )


        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 201)

        time.sleep(60)
    @allure.story("任务管理-创建构建任务")
    @allure.title("任务管理-获取构建任务列表,检查任务状态")
    def test_ontology_job_list(self):
        api = KnowledgeNetworkApi()
        with allure.step(f"创建构建任务:{config.get('view_data','kn_id')}"):
            resp=api.kn_job_list(kn_id=config.get('view_data','kn_id')
                            )


        with allure.step("验证HTTP响应状态码"):
            asserts.response_status_code(resp, 200)


        with allure.step("验证返回数据是否为空"):
            asserts.is_not_empty(json.loads(resp.text)['entries'])

        with allure.step("验证构建任务是否成功"):
            asserts.equal(json.loads(resp.text)['entries'][0]['state'],'completed')