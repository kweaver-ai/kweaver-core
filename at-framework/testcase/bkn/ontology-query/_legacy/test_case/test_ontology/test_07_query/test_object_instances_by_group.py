import pytest
import allure
import json
from src.common.assertions import asserts
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from test_case.ontology_base_test import BaseTest


@allure.feature("查询-检索对象类-概念分组")
class TestObjectInstances(BaseTest):
    """对象检索测试"""

    api = ObjectTypeApi()

    @allure.story("查询-检索对象类")
    @allure.title("不指定概念分组,返回所有对象类")
    def test_object_with_no_concept(self):
        with allure.step("检索对象类，不指定概念分组"):
            resp = self.api.search_objects(
                kn_id=self.test_kn_id,
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有对象类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {
                    "person",
                    "comment",
                    "post",
                    "tag",
                    "tagclass",
                    "place",
                    "organisation",
                    "forum",
                },
            )

    @allure.story("查询-检索对象类")
    @allure.title("指定概念分组中存在不存在的概念分组，正常返回")
    def test_object_with_exist_concept(self):
        with allure.step("检索对象类，指定概念分组中存在不存在的概念分组"):
            resp = self.api.search_objects(
                kn_id=self.test_kn_id,
                concept_group=["information", "test"],
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有对象类类"):
            asserts.contains_exact_items(
                "id", objects, {"person", "place", "organisation"}
            )

    @allure.story("查询-检索对象类")
    @allure.title("指定概念分组中存在不存在的概念分组直接报错")
    def test_object_exist_concept(self):
        with allure.step("检索对象类，指定概念分组中存在不存在的概念分组"):
            resp = self.api.search_objects(
                kn_id=self.test_kn_id, concept_group=["test"]
            )

        self.assert_response_code(resp, 500)

    @allure.story("查询-检索对象类")
    @allure.title("指定单个概念分组,仅对象类")
    def test_object_with_single_concept(self):
        with allure.step("检索对象类，指定单个概念分组：information"):
            resp = self.api.search_objects(
                kn_id=self.test_kn_id, concept_group=["information"]
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有对象类类"):
            asserts.contains_exact_items(
                "id", objects, {"person", "place", "organisation"}
            )

    @allure.story("查询-检索对象类")
    @allure.title("指定多个概念分组")
    def test_object_with_concepts(self):
        with allure.step("检索对象类，指定多个概念分组：information、social"):
            resp = self.api.search_objects(
                kn_id=self.test_kn_id,
                concept_group=["information", "social"],
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有对象类类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {"person", "place", "organisation", "comment", "post", "forum"},
            )
