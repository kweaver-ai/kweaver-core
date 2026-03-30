# coding: utf-8
"""
对象类型创建测试用例
"""
import pytest
import allure
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.api.vega.data_views_api import DataViewsApi
from src.api.ontology.kn_relation_type_api import RelationTypeApi
from src.common.utils import convert_db_type_to_support_type
from src.common.read_data import data_reader
from src.common.global_var import global_vars
from test_case.ontology_base_test import BaseTest
from src.config.setting import config


ldbc_object_data = data_reader.read_yaml("ontology/ldbc/ldbc_object.yaml")
ldbc_relation_data = data_reader.read_yaml("ontology/ldbc/ldbc_relation.yaml")


@allure.feature("ldbc社交网络网络创建")
class TestLDBCCreate(BaseTest):
    """对象类型创建相关测试"""

    object_api = ObjectTypeApi()
    view_api = DataViewsApi()
    relation_api = RelationTypeApi()

    @allure.story("创建对象类-ldbc")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", ldbc_object_data)
    @pytest.mark.dependency(name="test_query_kn")
    def test_create_object_type_with_views(
        self, object_data, update_kn_id, get_views_id
    ):
        """测试创建简单的对象类型"""
        BaseTest.test_kn_id = config.get("view_data", "kn_id")
        ldbc_display = {
            "person": "p_firstname",
            "comment": "c_content",
            "forum": "f_title",
            "organisation": "o_name",
            "place": "p_name",
            "post": "ps_content",
            "tag": "t_name",
            "tagclass": "tc_name",
        }
        ldbc_index = {
            "person": "p_personid",
            "comment": "c_commentid",
            "forum": "f_forumid",
            "organisation": "o_organisationid",
            "place": "p_placeid",
            "post": "ps_postid",
            "tag": "t_tagid",
            "tagclass": "tc_tagclassid",
        }
        index_config = {
            "keyword_config": {"enabled": True, "ignore_above_len": 1024},
            "fulltext_config": {"enabled": True, "analyzer": "standard"},
            "vector_config": {"enabled": False, "model_id": ""},
        }
        views_id = global_vars.get_var("objects_list")
        object_view_id = [
            item
            for item in views_id
            if item["technical_name"] == object_data["object_id"]
        ]
        with allure.step("构建对象属性"):
            fields = self.view_api.get_views_fields(object_view_id[0]["id"])
            data_properties = []
            for j in fields:
                properties = {
                    "name": j["name"],
                    "type": j["type"],
                    "display_name": j["display_name"],
                    "comment": j["comment"],
                    "index": True,
                    "incremental_key": False,
                    "mapped_field": {
                        "name": j["name"],
                    },
                }
                if j["type"] == "string" or "text":
                    properties["index_config"] = index_config
                    if j["name"] == "tc_name":
                        properties["index_config"]["vector_config"] = {
                            "enabled": True,
                            "model_id": self.embedding_model_id,
                        }
                if j["name"] == ldbc_index[object_data["object_id"]]:
                    properties["incremental_key"] = True

                data_properties.append(properties)
        with allure.step(f"创建对象类: {object_data['object_name']}"):
            resp = self.object_api.create_object_types(
                kn_id=self.test_kn_id,
                object_id=object_data["object_id"],
                name=object_data["object_name"],
                data_source=object_view_id[0]["id"],
                data_properties=data_properties,
                display_key=ldbc_display[object_data["object_id"]],
                primary_keys=ldbc_index[object_data["object_id"]],
                incremental_key=ldbc_index[object_data["object_id"]],
            )

        self.assert_response_code(resp, 201)
        self.assert_db_record_exists(
            self.object_table,
            condition="f_id = %s and f_kn_id = %s",
            params=(object_data["object_id"], self.test_kn_id),
        )

    @allure.story("创建关系类-绑定数据视图")
    @allure.title("{relations_data[case_name]}")
    @pytest.mark.parametrize("relations_data", ldbc_relation_data)
    def test_create_relation_type_with_views(self, relations_data, get_views_id):
        """测试创建简单的关系类型"""
        test_views_id = global_vars.get_var("relations_list")
        relation_view_id = [
            item
            for item in test_views_id
            if item["technical_name"] == relations_data["relation_id"]
        ]
        relations_data["mapping_rules"]["backing_data_source"]["id"] = relation_view_id[
            0
        ]["id"]
        resp = self.relation_api.create_relation_type(
            kn_id=self.test_kn_id,
            name=relations_data["relation_name"],
            id=relations_data["relation_id"],
            type="data_view",
            source_object_type_id=relations_data["source_object_type_id"],
            target_object_type_id=relations_data["target_object_type_id"],
            mapping_rules=relations_data["mapping_rules"],
        )

        self.assert_response_code(resp, 201)

        self.assert_db_record_exists(
            table=self.relation_table,
            condition="f_id = %s and f_kn_id = %s",
            params=(
                relations_data["relation_id"],
                self.test_kn_id,
            ),
        )
