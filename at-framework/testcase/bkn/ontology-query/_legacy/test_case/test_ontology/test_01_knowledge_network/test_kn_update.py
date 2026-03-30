import pytest
import allure
import json
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest


# 正向测试数据
positive_data = data_reader.read_yaml(
    "ontology/knowledge_network/kn_update_postive.yaml"
)
# 负向测试数据
negative_data = data_reader.read_yaml(
    "ontology/knowledge_network/kn_update_negative.yaml"
)


@allure.feature("业务知识网络")
class TestKnowledgeNetworkUpdate(BaseTest):
    api = KnowledgeNetworkApi()

    @allure.story("更新业务知识网络-正常场景")
    @allure.title("{kn_data[case_name]}")
    @pytest.mark.parametrize("kn_data", positive_data)
    def test_update_kn_positive(self, kn_data):
        """正向测试用例：期望更新成功"""
        # 添加测试说明
        if kn_data.get("note"):
            allure.attach(kn_data.get("note"), name="测试说明")

        # 创建业务知识网络
        resp = self.api.create_kn(
            kn_data.get("kn_c_name"),
            kn_data.get("kn_id"),
            tags=kn_data.get("kn_c_tags", None),
        )
        self.assert_response_code(resp, 201)

        # 更新业务知识网络
        resp = self.api.update_kn(
            kn_data.get("kn_u_name"),
            kn_data.get("kn_id"),
            tags=kn_data.get("kn_u_tags", None),
            comment=kn_data.get("kn_u_comment", None),
        )
        self.assert_response_code(resp, 204)

        # 验证数据库记录更新成功
        self.assert_db_record_exists(
            table=self.kn_table,
            condition="f_name = %s",
            params=(kn_data.get("kn_u_name"),),
        )

        # 删除测试数据
        self.api.delete_kn(kn_data.get("kn_id"))

    @allure.story("更新业务知识网络-异常场景")
    @allure.title("{kn_data[case_name]}")
    @pytest.mark.parametrize("kn_data", negative_data)
    def test_update_kn_negative(self, kn_data):
        """负向测试用例：期望更新失败"""
        # 添加测试说明
        if kn_data.get("note"):
            allure.attach(kn_data.get("note"), name="测试说明")

        # 创建业务知识网络
        resp = self.api.create_kn(
            kn_data.get("kn_c_name"),
            kn_data.get("kn_id"),
            tags=kn_data.get("kn_c_tags", None),
        )
        self.assert_response_code(resp, 201)

        # 处理特殊测试场景
        update_id = kn_data.get("kn_id")
        if kn_data.get("non_existent_id"):
            update_id = kn_data.get("non_existent_id")
        elif kn_data.get("invalid_id"):
            update_id = kn_data.get("invalid_id")

        # 更新业务知识网络
        resp = self.api.update_kn(
            kn_data.get("kn_u_name"),
            update_id,
            tags=kn_data.get("kn_u_tags", None),
            comment=kn_data.get("kn_u_comment", None),
        )

        # 验证响应状态码
        expected_status = kn_data.get("expected_status", 400)
        self.assert_response_code(resp, expected_status)

        # 如果是参数错误，验证响应结构
        if expected_status >= 400 and expected_status < 500:
            self.assert_response_structure(resp, ["error_code", "description"])

        # 删除测试数据（如果需要）
        if kn_data.get("kn_id"):
            self.api.delete_kn(kn_data.get("kn_id"))
