# coding: utf-8
"""
概念分组型删除测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_group_api import ConceptGroupApi
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

positive_data = data_reader.read_type_cases(
    "ontology/concept_groups/concept_groups_create.yaml", "positive"
)
negative_data = data_reader.read_type_cases(
    "ontology/concept_groups/concept_groups_create.yaml", "negative"
)


@allure.feature("知识网络概念分组管理")
class TestConceptGroupDelete(BaseTest):
    api = ConceptGroupApi()

    @allure.story("删除概念分组")
    @allure.title("删除概念分组-单个删除")
    def test_create_concept_group(self):
        self.api.create_concept_groups(
            kn_id=self.test_kn_id,
            groups_id="delete_groups",
            groups_name="概念分组删除-单个",
        )

        resp = self.api.delete_concept_groups(self.test_kn_id, "delete_groups")

        self.assert_response_code(resp, 204)

        self.assert_db_record_not_exists(
            table=self.concept_group_table,
            condition="f_id = %s and f_kn_id = %s",
            params=("delete_groups", self.test_kn_id),
        )
