import time
import pytest
import allure
import json
from src.common.assertions import asserts
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from test_case.ontology_base_test import BaseTest


@allure.feature("任务管理-创建构建任务")
class TestObjectTypeJob(BaseTest):
    """对象构建测试"""

    @allure.story("任务管理-创建构建任务")
    @allure.title("任务管理-创建构建任务")
    def test_ontology_job_create(self):
        api = KnowledgeNetworkApi()
        with allure.step(f"创建构建任务:{self.test_kn_id}"):
            resp = api.kn_job(
                kn_id=self.test_kn_id,
                kn_name=self.test_kn_id,
            )

        self.assert_response_code(resp, 201)

    @allure.story("任务管理-创建构建任务")
    @allure.title("任务管理-获取构建任务列表,检查任务状态")
    def test_ontology_job_list(self):
        api = KnowledgeNetworkApi()
        with allure.step(f"创建构建任务:{self.test_kn_id}"):
            resp = api.kn_job_list(kn_id=self.test_kn_id)

        self.assert_response_code(resp, 200)

        with allure.step("验证返回数据是否为空"):
            asserts.is_not_empty(json.loads(resp.text)["entries"])

        nums = 3
        result = None
        for i in range(nums):
            time.sleep(30)
            resp = api.kn_job_list(kn_id=self.test_kn_id)
            result = json.loads(resp.text)["entries"][0]["state"]
            if result == "completed":
                break
        with allure.step("验证构建任务是否成功"):
            asserts.equal(result, "completed")
