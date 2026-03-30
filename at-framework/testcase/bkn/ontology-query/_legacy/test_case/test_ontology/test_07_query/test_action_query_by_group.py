import pytest
import allure
import json
from src.common.assertions import asserts
from src.api.ontology.kn_action_type_api import ActionTypeApi
from test_case.ontology_base_test import BaseTest


@allure.feature("查询-检索行动类-概念分组")
class TestActionInstances(BaseTest):
    """关系检索测试"""

    api = ActionTypeApi()

    @allure.story("查询-检索行动类")
    @allure.title("不指定概念分组,返回所有行动类")
    def test_action_with_no_concept(self):
        with allure.step("检索行动类，不指定概念分组"):
            resp = self.api.search_actions(
                kn_id=self.test_kn_id,
            )
            objects = json.loads(resp.text)["entries"]

        self.assert_response_code(resp, 200)
        self.assert_equal(len(objects), 9)

    @allure.story("查询-检索行动类")
    @allure.title("指定单个概念分组social")
    def test_action_with_single_concept(self):
        with allure.step("检索行动类，指定单个概念分组social"):
            resp = self.api.search_actions(
                kn_id=self.test_kn_id, concept_group=["social"]
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_equal(len(objects), 3)
        with allure.step("验证包含所有关系类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {
                    "action_comment",
                    "action_forum",
                    "action_post",
                },
            )

    @allure.story("查询-检索行动类")
    @allure.title("指定多个概念分组social、information")
    def test_action_with_concepts(self):
        with allure.step("检索行动类，指定多个概念分组social、information"):
            resp = self.api.search_actions(
                kn_id=self.test_kn_id,
                concept_group=["social", "information"],
            )
        objects = json.loads(resp.text)["entries"]

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)

    @allure.story("查询-检索行动类")
    @allure.title("概念分组不存在，直接报错")
    def test_action_exist_concept(self):
        with allure.step("检索对象类，指定概念分组中存在不存在的概念分组"):
            resp = self.api.search_actions(
                kn_id=self.test_kn_id, concept_group=["test"]
            )
        self.assert_response_code(resp, 500)
