# coding: utf-8
"""
概念分组型创建测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_group_api import ConceptGroupApi
from test_case.ontology_base_test import BaseTest
from src.common.read_data import data_reader
from src.common.assertions import asserts

positive_data = data_reader.read_type_cases(
    "ontology/concept_groups/concept_groups_create.yaml", "positive"
)
negative_data = data_reader.read_type_cases(
    "ontology/concept_groups/concept_groups_create.yaml", "negative"
)


@allure.feature("知识网络概念分组管理")
class TestConceptGroupCreate(BaseTest):
    api = ConceptGroupApi()

    @allure.story("创建概念分组")
    @allure.title("{groups_data[case_name]}")
    @pytest.mark.parametrize("groups_data", positive_data)
    def test_create_concept_group(self, groups_data):

        resp = self.api.create_concept_groups(
            kn_id=self.test_kn_id,
            groups_id=groups_data["groups_id"],
            groups_name=groups_data["groups_name"],
        )

        self.assert_response_code(resp, 201)
        self.assert_response_structure(resp, ["id"])

        self.assert_db_record_exists(
            table=self.concept_group_table,
            condition="f_id = %s and f_kn_id = %s",
            params=(resp.json().get("id"), self.test_kn_id),
        )
        # 删除
        self.api.delete_concept_groups(self.test_kn_id, resp.json().get("id"))

    @allure.story("创建概念分组-异常场景")
    @allure.title("{groups_data[case_name]}")
    @pytest.mark.parametrize("groups_data", negative_data)
    def test_create_concept_group_negative(self, groups_data):
        """负向测试用例：期望创建失败"""
        if groups_data.get("note") == "重复ID—name":
            resp = self.api.create_concept_groups(
                kn_id=self.test_kn_id,
                groups_id=groups_data["groups_id"],
                groups_name=groups_data["groups_name"],
            )
            id = resp.json().get("id")

        with allure.step(f"尝试创建概念分组: {groups_data.get('case_name')}"):
            resp = self.api.create_concept_groups(
                kn_id=self.test_kn_id,
                groups_id=groups_data["groups_id"],
                groups_name=groups_data["groups_name"],
            )

        self.assert_response_code(
            resp,
            expected_code=groups_data.get(
                "expected_status", groups_data.get("expected_status")
            ),
        )

        if not groups_data.get("should_exist_in_db", False):
            # 对于异常场景，检查数据库中不应该存在该记录
            kn_id = self.test_kn_id
            if kn_id:  # 只有当kn_id不为空时才检查
                self.assert_db_record_not_exists(
                    table=self.concept_group_table,
                    condition="f_id = %s and f_kn_id = %s",
                    params=(
                        groups_data["groups_id"],
                        self.test_kn_id,
                    ),
                )
        if groups_data.get("note") == "重复ID—name":
            self.api.delete_concept_groups(
                kn_id=self.test_kn_id,
                groups_id=id,
            )

    @allure.story("创建概念分组-绑定对象类")
    @allure.title("概念分组绑定对象类")
    def test_concept_group(self):
        with allure.step("创建概念分组-社交内容"):
            self.api.create_concept_groups(
                kn_id=self.test_kn_id,
                groups_id="social",
                groups_name="社交内容",
            )
        with allure.step("概念分组-社交内容-绑定对象类"):
            self.api.concept_groups_add_object(
                kn_id=self.test_kn_id,
                groups_id="social",
                objects=["comment", "post", "forum"],
            )
        with allure.step("创建概念分组-个人信息"):
            self.api.create_concept_groups(
                kn_id=self.test_kn_id,
                groups_id="information",
                groups_name="个人信息",
            )
        with allure.step("概念分组-个人信息-绑定对象类"):
            self.api.concept_groups_add_object(
                kn_id=self.test_kn_id,
                groups_id="information",
                objects=["person", "organisation", "place"],
            )

        with allure.step("创建概念分组-分类"):
            self.api.create_concept_groups(
                kn_id=self.test_kn_id,
                groups_id="class",
                groups_name="分类",
            )
        with allure.step("概念分组-分类-绑定对象类"):
            self.api.concept_groups_add_object(
                kn_id=self.test_kn_id,
                groups_id="class",
                objects=["tag", "tagclass"],
            )

    @allure.story("创建概念分组-绑定对象类")
    @allure.title("概念分组详情查看")
    def test_get_concept_group(self):

        with allure.step("创建概念分组-分类"):
            resp = self.api.get_groups_views(kn_id=self.test_kn_id, groups_id="social")

        self.assert_response_code(resp, 200)

        self.assert_response_structure(resp, ["id", "object_types", "relation_types"])

        with allure.step("验证关系类是否存在且仅有三个关系类"):
            relations = json.loads(resp.text)["relation_types"]

            asserts.contains_exact_items(
                "id",
                relations,
                {
                    "comment_replyof_comment",
                    "comment_replyof_post",
                    "forum_containerof_post",
                },
            )

        with allure.step("验证行动类是否存在"):
            actions = json.loads(resp.text)["action_types"]
            asserts.contains_exact_items(
                "id",
                actions,
                {
                    "action_comment",
                    "action_post",
                    "action_forum",
                },
            )
