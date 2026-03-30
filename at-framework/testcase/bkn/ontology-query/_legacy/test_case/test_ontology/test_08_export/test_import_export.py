import pytest
import allure
import os
from src.api.ontology.kn_network_api import KnowledgeNetworkApi
from src.common.assertions import asserts
from src.common.utils import check_file_exists_with_os
from test_case.ontology_base_test import BaseTest


@allure.feature("业务知识网络")
class TestExportKnowledgeNetwork(BaseTest):
    api = KnowledgeNetworkApi()

    def teardown_class(self):
        # 清理导出的文件
        if check_file_exists_with_os(f"./export/{self.test_kn_id}.json"):
            os.remove(f"./export/{self.test_kn_id}.json")

    @allure.story("导出业务知识网络")
    @allure.title("导出业务知识网络")
    @pytest.mark.dependency(name="test_export_kn")
    def test_export_kn(self):
        # 查询
        resp = self.api.export_kn(kn_id=self.test_kn_id)
        self.assert_response_code(resp, 200)
        # 返回数据entries、total_count
        with allure.step("验证文件是否导出"):
            asserts.is_true(
                check_file_exists_with_os(f"./export/{self.test_kn_id}.json")
            )

    @allure.story("导入业务知识网络")
    @allure.title("导入业务知识网络")
    @pytest.mark.dependency(depends=["test_export_kn"])
    @pytest.mark.dependency(name="test_import_kn")
    def test_import_kn(self):
        # 查询
        resp = self.api.import_kn(kn_id=self.test_kn_id)
        self.assert_response_code(resp, 201)

    @allure.story("导入业务知识网络")
    @allure.title("导入业务知识网络构建执行")
    def test_import_kn_construct(self):
        # 修改业务知识网络为导入的知识网络
        BaseTest.test_kn_id = self.test_kn_id + "_import"
        from test_case.test_ontology.test_04_ontology_job.test_ontology_job import (
            TestObjectTypeJob,
        )

        TestObjectTypeJob().test_ontology_job_create()
        TestObjectTypeJob().test_ontology_job_list()

    @allure.story("导入业务知识网络")
    @allure.title("导入业务知识网络查询-检索行动类-概念分组")
    def test_import_kn_query_action_group(self):
        from test_case.test_ontology.test_07_query.test_action_query_by_group import (
            TestActionInstances,
        )

        # 查询-检索行动类-概念分组
        TestActionInstances().test_action_with_no_concept()
        TestActionInstances().test_action_with_single_concept()
        TestActionInstances().test_action_with_concepts()
        TestActionInstances().test_action_with_no_concept()

    @allure.story("导入业务知识网络")
    @allure.title("导入业务知识网络查询-检索对象类-概念分组")
    def test_import_kn_query_object_group(self):
        from test_case.test_ontology.test_07_query.test_object_instances_by_group import (
            TestObjectInstances,
        )

        # 查询-检索对象类-概念分组
        TestObjectInstances().test_object_with_no_concept()
        TestObjectInstances().test_object_with_single_concept()
        TestObjectInstances().test_object_with_concepts()
        TestObjectInstances().test_object_with_no_concept()

    @allure.story("导入业务知识网络")
    @allure.title("导入业务知识网络查询-检索关系类-概念分组")
    def test_import_kn_query_relation_group(self):
        from test_case.test_ontology.test_07_query.test_relation_instances_by_group import (
            TestRelationInstances,
        )

        # 查询-检索关系类-概念分组
        TestRelationInstances().test_relation_with_no_concept()
        TestRelationInstances().test_relation_with_single_concept()
        TestRelationInstances().test_relation_with_concepts()
        TestRelationInstances().test_relation_with_no_concept()

    # @allure.story("导入业务知识网络")
    # @allure.title("导入业务知识网络查询-检索对象类")
    # def test_import_kn_query_object_group(self):
    #     from test_case.test_ontology.test_07_query.test_object_query import (
    #         TestQueryObject,
    #     )

    #     # 查询-检索对象查询类
    #     TestQueryObject().test_query_object_type()

    # @allure.story("导入业务知识网络")
    # @allure.title("导入业务知识网络查询-检索关系类")
    # def test_import_kn_query_relation_group(self):
    #     from test_case.test_ontology.test_07_query.test_relation_query import (
    #         TestQueryRelation,
    #     )

    #     # 查询-检索关系查询类
    #     TestQueryRelation().test_query_by_source_target()
    #     TestQueryRelation().test_query_by_path()
