# coding: utf-8
"""
概念分组型创建测试用例
"""
import pytest
import allure
from src.api.ontology.kn_group_api import ConceptGroupApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

add_objects_datas = data_reader.read_type_cases(
    "ontology/concept_groups/concept_groups_add_object.yaml", "add"
)
remove_objects_datas = data_reader.read_type_cases(
    "ontology/concept_groups/concept_groups_add_object.yaml", "remove"
)


@allure.feature("知识网络概念分组管理")
class TestConceptGroupAddAndRemoveObject(BaseTest):
    api = ConceptGroupApi()

    @allure.story("创建概念分组-绑定对象类")
    @allure.title("{add_objects_data[case_name]}")
    @pytest.mark.parametrize("add_objects_data", add_objects_datas)
    def test_concept_add_objects(
        self, add_objects_data, create_test_concept_group_add_objects
    ):
        with allure.step("概念分组-测试添加对象类"):
            resp = self.api.concept_groups_add_object(
                kn_id=self.test_kn_id,
                groups_id="test_add",
                objects=add_objects_data["objects"],
            )
        self.assert_response_code(resp, add_objects_data["expected_status"])

    @allure.story("创建概念分组-绑定对象类")
    @allure.title("{remove_objects_data[case_name]}")
    @pytest.mark.parametrize("remove_objects_data", remove_objects_datas)
    def test_concept_remove_objects(
        self, remove_objects_data, create_test_concept_group_remove_objects
    ):

        with allure.step("概念分组-测试移除对象类-移除单个对象类"):
            resp = self.api.concept_groups_remove_object(
                kn_id=self.test_kn_id,
                groups_id="test_remove",
                objects=remove_objects_data["objects"],
            )
        self.assert_response_code(resp, remove_objects_data["expected_status"])
