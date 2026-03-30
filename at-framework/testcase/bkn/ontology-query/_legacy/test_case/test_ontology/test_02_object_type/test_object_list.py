import pytest
import allure
import json
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from test_case.ontology_base_test import BaseTest


@allure.feature("知识网络对象类型管理")
class TestObjectList(BaseTest):
    api = ObjectTypeApi()

    @allure.story("获取对象类列表")
    @allure.title("获取对象类列表")
    def test_object_list(self):
        # 查询
        resp = self.api.list_object_types(self.test_kn_id)
        self.assert_response_code(resp, 200)
        # 返回数据entries、total_count
        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])

        # 精确筛选
        resp = self.api.list_object_types(self.test_kn_id, "用户")
        self.assert_response_code(resp, 200)
        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])
        self.assert_equal("用户", json.loads(resp.text)["entries"][0]["name"])
        self.assert_has_length(json.loads(resp.text)["entries"], 1)

    @allure.story("获取对象类列表")
    @allure.title("获取对象类列表-精确筛选")
    def test_object_list_by_name(self):

        # 精确筛选
        resp = self.api.list_object_types(self.test_kn_id, "用户")
        self.assert_response_code(resp, 200)
        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])
        self.assert_equal("用户", json.loads(resp.text)["entries"][0]["name"])
        self.assert_has_length(json.loads(resp.text)["entries"], 1)
