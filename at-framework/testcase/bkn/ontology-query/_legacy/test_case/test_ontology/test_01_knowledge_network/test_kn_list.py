import allure
import json
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.data_gen import random_str
from test_case.ontology_base_test import BaseTest


@allure.feature("业务知识网络")
class TestKnowledgeNetworkList(BaseTest):
    api = KnowledgeNetworkApi()

    @allure.story("获取业务知识网络列表")
    @allure.title("获取业务知识网络")
    def test_list_kn(self):
        # 查询
        resp = self.api.list_kn()
        self.assert_response_code(resp, 200)
        # 返回数据entries、total_count
        self.assert_response_structure(resp, ["entries", "total_count"])
        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])

    @allure.story("获取业务知识网络列表")
    @allure.title("获取业务知识网络-精确筛选")
    def test_list_kn_by_name(self):
        kn_name = "列表查看" + random_str()
        kn_id = "list_test" + random_str()
        self.api.create_kn(kn_name, kn_id)
        resp = self.api.list_kn(kn_name)
        self.assert_response_code(resp, 200)
        self.assert_response_structure(resp, ["entries", "total_count"])
        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])
        self.assert_equal(kn_name, json.loads(resp.text)["entries"][0]["name"])
        self.assert_has_length(json.loads(resp.text)["entries"], 1)

        self.api.delete_kn(kn_id)
