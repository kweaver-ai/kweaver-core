import pytest
import allure
import json
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

delete_kn_data = data_reader.read_yaml("ontology/knowledge_network/kn_delete.yaml")


@allure.feature("业务知识网络")
class TestKnowledgeNetworkDelete(BaseTest):
    api = KnowledgeNetworkApi()

    @allure.story("删除业务知识网络-正常场景")
    @allure.title("{kn_data[case_name]}")
    @pytest.mark.parametrize("kn_data", delete_kn_data)
    def test_delete_kn_success(self, kn_data):
        # 创建
        resp = self.api.create_kn(kn_data.get("kn_name"), kn_data.get("kn_id"))
        # 删除
        resp = self.api.delete_kn(kn_data.get("kn_id"))
        # 验证响应状态码和结构
        self.assert_response_code(resp, 204)
        # 验证数据库记录是否不存在
        self.assert_db_record_not_exists(self.kn_table, params=kn_data.get("kn_id"))
