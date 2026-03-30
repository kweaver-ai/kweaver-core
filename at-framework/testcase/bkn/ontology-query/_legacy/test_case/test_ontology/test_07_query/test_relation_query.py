import pytest
import allure
import json
from src.common.assertions import asserts
from src.common.read_data import data_reader
from src.api.ontology.ontology_query_api import OntologyQueryApi
from src.common.utils import get_unique_identities, normalize_list_of_dicts
from test_case.ontology_base_test import BaseTest

# 基于起点查询用例
query_datas_by_query_source_target = data_reader.read_yaml(
    "ontology/query/source_target_query.yaml"
)
# 关系查询用例
query_datas_by_query_relation = data_reader.read_yaml(
    "ontology/query/relation_query.yaml"
)


@allure.feature("查询-关系查询")
class TestQueryRelation(BaseTest):
    """关系查询测试"""

    api = OntologyQueryApi()

    @allure.story("查询-基于起点探索")
    @allure.title("{query_data[case_name]}")
    @pytest.mark.parametrize("query_data", query_datas_by_query_source_target)
    def test_query_by_source_target(self, query_data):
        with allure.step(f"执行查询场景: {query_data['case_name']}"):
            resp = self.api.query_objects_by_source_target(
                kn_id=self.test_kn_id,
                source_object_type_id=query_data["source_object_type_id"],
                direction=query_data["direction"],
                limit=query_data["limit"],
                path_length=query_data["path_length"],
                condition=query_data["condition"],
                need_total=query_data["need_total"],
                ignoring_store_cache=query_data["ignoring_store_cache"],
            )

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(json.loads(resp.text)["objects"])

    @allure.story("查询-基于路径查询")
    @allure.title("{query_data[case_name]}")
    @pytest.mark.parametrize("query_data", query_datas_by_query_relation)
    def test_query_by_path(self, query_data):
        with allure.step(f"执行查询场景: {query_data['case_name']}"):
            resp = self.api.query_objects_by_path(
                kn_id=self.test_kn_id,
                limit=query_data["limit"],
                relation_type_paths=query_data["relation_type_paths"],
                ignoring_store_cache=query_data["ignoring_store_cache"],
            )

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(json.loads(resp.text)["entries"][0]["relation_paths"])

        if "assert_data" in query_data:
            with allure.step("验证数据正确性"):
                # 返回数据进行排序
                actual = normalize_list_of_dicts(
                    get_unique_identities(
                        json.loads(resp.text)["entries"][0]["objects"]
                    )
                )
                expected = normalize_list_of_dicts(query_data["assert_data"])
                asserts.equal(actual, expected)
