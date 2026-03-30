import pytest
import allure
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

# 正常场景测试数据
positive_data = data_reader.read_yaml(
    "ontology/knowledge_network/kn_create_postive.yaml"
)
# 异常场景测试数据
negative_data = data_reader.read_yaml(
    "ontology/knowledge_network/kn_create_negative.yaml"
)


@allure.feature("业务知识网络")
class TestKnowledgeNetworkCreate(BaseTest):
    api = KnowledgeNetworkApi()

    @allure.story("创建业务知识网络-正常场景")
    @allure.title("{kn_data[case_name]}")
    @pytest.mark.parametrize("kn_data", positive_data)
    def test_create_kn_positive(self, kn_data):
        """正向测试用例：期望创建成功"""
        if kn_data.get("note"):
            allure.attach(kn_data.get("note"), name="测试说明")
        resp = self.api.create_kn(
            kn_data.get("kn_name"),
            kn_data.get("kn_id"),
            tags=kn_data.get("tags", None),
            exclude_fields=kn_data.get("exclude_fields", None),
        )
        # 验证响应状态码和结构
        self.assert_response_code(resp, 201)
        self.assert_response_structure(resp, ["id"])
        # 验证数据库记录是否存在
        self.assert_db_record_exists(
            table=self.kn_table,
            params=(resp.json().get("id"),),
        )

        # 删除测试创建的知识网络
        self.api.delete_kn(resp.json().get("id"))

    @allure.story("创建业务知识网络-异常场景")
    @allure.title("{kn_data[case_name]}")
    @pytest.mark.parametrize("kn_data", negative_data)
    def test_create_kn_negative(self, kn_data):
        """负向测试用例：期望创建失败"""
        if kn_data.get("note"):
            allure.attach(kn_data.get("note"), name="测试说明")

        if kn_data.get("note") == "重复ID—name":
            self.api.create_kn(kn_data.get("kn_name"), kn_data.get("kn_id"))

        resp = self.api.create_kn(
            kn_data.get("kn_name"),
            kn_data.get("kn_id"),
            tags=kn_data.get("tags", None),
            comment=kn_data.get("comment", None),
            exclude_fields=kn_data.get("exclude_fields", None),
        )

        # 验证响应状态码和结构
        self.assert_response_code(
            resp,
            expected_code=kn_data.get(
                "expected_status", kn_data.get("expected_status")
            ),
        )
        self.assert_response_structure(resp, ["error_code", "description"])
        # 验证数据库记录是否存在
        self.assert_db_record_not_exists(
            table=self.kn_table,
            params=(kn_data.get("kn_id"),),
        )
