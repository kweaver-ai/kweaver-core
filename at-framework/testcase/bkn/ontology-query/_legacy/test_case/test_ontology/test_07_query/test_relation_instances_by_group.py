import allure
import json
from src.common.assertions import asserts
from src.api.ontology.kn_relation_type_api import RelationTypeApi
from test_case.ontology_base_test import BaseTest


@allure.feature("查询-检索关系类")
class TestRelationInstances(BaseTest):
    """关系检索测试"""

    api = RelationTypeApi()

    @allure.story("查询-检索关系类")
    @allure.title("不指定概念分组,返回所有关系类")
    def test_relation_with_no_concept(self):
        with allure.step("检索关系类，不指定概念分组"):
            resp = self.api.search_relations(
                kn_id=self.test_kn_id,
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有关系类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {
                    "comment_hascreator_person",
                    "comment_hastag_tag",
                    "comment_islocatedin_place",
                    "comment_replyof_comment",
                    "comment_replyof_post",
                    "forum_containerof_post",
                    "forum_hasmember_person",
                    "forum_hasmoderator_person",
                    "forum_hastag_tag",
                    "organisation_islocatedin_place",
                    "person_hasinterest_tag",
                    "person_islocatedin_place",
                    "person_knows_person",
                    "person_likes_comment",
                    "person_likes_post",
                    "person_studyat_organisation",
                    "person_workat_organisation",
                    "place_ispartof_place",
                    "post_hascreator_person",
                    "post_hastag_tag",
                    "post_islocatedin_place",
                    "tag_hastype_tagclass",
                    "tagclass_issubclassof_tagclass",
                },
            )

    @allure.story("查询-检索关系类")
    @allure.title("指定单个概念分组,仅返回对应关系类")
    def test_relation_with_single_concept(self):
        with allure.step("检索关系类，指定单个概念分组：information"):
            resp = self.api.search_relations(
                kn_id=self.test_kn_id, concept_group=["information"]
            )
        objects = json.loads(resp.text)["entries"]

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有关系类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {
                    "organisation_islocatedin_place",
                    "person_islocatedin_place",
                    "person_knows_person",
                    "person_studyat_organisation",
                    "person_workat_organisation",
                    "place_ispartof_place",
                },
            )

    @allure.story("查询-检索关系类")
    @allure.title("指定多个概念分组")
    def test_relation_with_concepts(self):
        with allure.step("检索关系类，指定多个概念分组：information、social"):
            resp = self.api.search_relations(
                kn_id=self.test_kn_id,
                concept_group=["information", "social"],
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有关系类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {
                    "comment_hascreator_person",
                    "comment_islocatedin_place",
                    "comment_replyof_comment",
                    "comment_replyof_post",
                    "forum_containerof_post",
                    "forum_hasmember_person",
                    "forum_hasmoderator_person",
                    "organisation_islocatedin_place",
                    "person_islocatedin_place",
                    "person_knows_person",
                    "person_likes_comment",
                    "person_likes_post",
                    "person_studyat_organisation",
                    "person_workat_organisation",
                    "place_ispartof_place",
                    "post_hascreator_person",
                    "post_islocatedin_place",
                },
            )

    @allure.story("查询-检索关系类")
    @allure.title("指定单个概念分组,仅返回对应关系类")
    def test_relation_with_exist_concept(self):
        with allure.step("检索关系类，指定单个概念分组：information"):
            resp = self.api.search_relations(
                kn_id=self.test_kn_id,
                concept_group=["information", "test"],
            )
        objects = json.loads(resp.text)["entries"]
        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(objects)
        with allure.step("验证包含所有关系类"):
            asserts.contains_exact_items(
                "id",
                objects,
                {
                    "organisation_islocatedin_place",
                    "person_islocatedin_place",
                    "person_knows_person",
                    "person_studyat_organisation",
                    "person_workat_organisation",
                    "place_ispartof_place",
                },
            )

    @allure.story("查询-检索关系类")
    @allure.title("指定单个概念分组,仅返回对应关系类")
    def test_relation_exist_concept(self):
        with allure.step("检索关系类，指定单个概念分组：information"):
            resp = self.api.search_relations(
                kn_id=self.test_kn_id, concept_group=["test"]
            )

        self.assert_response_code(resp, 500)
