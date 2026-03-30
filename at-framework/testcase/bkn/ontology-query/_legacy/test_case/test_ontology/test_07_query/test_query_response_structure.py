import pytest
import allure
import json
from src.common.assertions import asserts
from src.common.read_data import data_reader
from src.api.ontology.ontology_query_api import OntologyQueryApi

from src.common.utils import get_id_data, normalize_list_of_dicts
from test_case.ontology_base_test import BaseTest


@allure.feature("查询-响应结构体校验")
class TestQueryResponseStructure(BaseTest):
    """查询接口响应参数结构校验测试"""

    api = OntologyQueryApi()

    @allure.story("查询-对象检索查询结构体校验")
    def test_object_data_response_structure(self):
        """测试对象查询接口响应结构"""
        object_query = data_reader.read_type_cases(
            "ontology/query/query_structure.yaml", "object"
        )[0]

        # 解析响应
        resp = self.api.query_object_type_instances(
            kn_id=self.test_kn_id,
            ot_id=object_query["ot_id"],
            condition=object_query["condition"],
            limit=object_query["limit"],
            need_total=object_query["need_total"],
            include_type_info=object_query["include_type_info"],
            ignoring_store_cache=object_query["ignoring_store_cache"],
        )

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(json.loads(resp.text)["datas"])

        # 校验数据结构
        self.assert_json_structure(
            json.loads(resp.text)["datas"][0], object_query["response_structure"]
        )

    @allure.story("查询-起点查询结构体校验")
    def test_object_by_source_target_response_structure(self):
        """测试对象查询接口响应结构"""
        target_query = data_reader.read_type_cases(
            "ontology/query/query_structure.yaml", "source_target"
        )[0]

        # 解析响应
        resp = self.api.query_objects_by_source_target(
            kn_id=self.test_kn_id,
            source_object_type_id=target_query["ot_id"],
            condition=target_query["condition"],
            limit=target_query["limit"],
            need_total=target_query["need_total"],
            ignoring_store_cache=target_query["ignoring_store_cache"],
        )

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(json.loads(resp.text)["objects"])

        # 校验数据结构
        self.assert_json_structure(
            json.loads(resp.text)["objects"]["tag-3"],
            target_query["response_structure"],
        )

    @allure.story("查询-关系查询结构体校验")
    def test_relation_response_structure(self):
        """测试关系查询接口响应结构"""
        relation_query = data_reader.read_type_cases(
            "ontology/query/query_structure.yaml", "relation"
        )[0]

        resp = self.api.query_objects_by_path(
            kn_id=self.test_kn_id,
            limit=relation_query["limit"],
            relation_type_paths=relation_query["relation_type_paths"],
            ignoring_store_cache=relation_query["ignoring_store_cache"],
        )

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(json.loads(resp.text)["entries"][0]["relation_paths"])

        # 校验数据结构
        self.assert_json_structure(
            json.loads(resp.text)["entries"][0]["objects"]["tagclass-148"],
            relation_query["response_structure"],
        )
