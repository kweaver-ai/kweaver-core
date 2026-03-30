import pytest
import allure
import json
from src.common.assertions import asserts
from src.common.read_data import data_reader
from src.api.ontology.ontology_query_api import OntologyQueryApi
from src.common.utils import get_id_data, normalize_list_of_dicts
from test_case.ontology_base_test import BaseTest

# 检索对象用例
query_datas_by_query_object = data_reader.read_yaml("ontology/query/object_query.yaml")


@allure.feature("查询-对象查询")
class TestQueryObject(BaseTest):
    """对象检索测试"""

    api = OntologyQueryApi()

    @allure.story("查询-对象查询")
    @allure.title("{query_data[case_name]}")
    @pytest.mark.parametrize("query_data", query_datas_by_query_object)
    def test_query_object_type(self, query_data):
        with allure.step(f"执行查询场景: {query_data['case_name']}"):
            resp = self.api.query_object_type_instances(
                kn_id=self.test_kn_id,
                ot_id=query_data["ot_id"],
                condition=query_data["condition"],
                limit=query_data["limit"],
                need_total=query_data["need_total"],
                include_type_info=query_data["include_type_info"],
                ignoring_store_cache=query_data["ignoring_store_cache"],
            )

        self.assert_response_code(resp, 200)
        self.assert_is_not_empty(json.loads(resp.text)["datas"])

        if "assert_data" in query_data:
            with allure.step("验证数据正确性"):
                # 返回数据进行排序
                actual = normalize_list_of_dicts(
                    get_id_data(json.loads(resp.text)["datas"])
                )
                expected = normalize_list_of_dicts(query_data["assert_data"])
                asserts.equal(actual, expected)
