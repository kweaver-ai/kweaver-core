# coding: utf-8
"""
关系类测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_relation_type_api import RelationTypeApi
from src.common.global_var import global_vars
from test_case.ontology_base_test import BaseTest


@allure.feature("知识网络关系类管理")
class TestRelationTypeCreate(BaseTest):
    """关系类型列表查看相关测试"""

    api = RelationTypeApi()

    @allure.story("查看关系类")
    @allure.title("查看关系类-关系类列表")
    def test_relation_type_list(self):
        """测试关系类型"""

        # 列表查看
        resp = self.api.list_relation_types(self.test_kn_id)
        self.assert_response_code(resp, 200)
        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])

    @allure.story("查看关系类")
    @allure.title("查看关系类-关系类列表-精确筛选")
    def test_relation_type_list_by_name(self):
        """测试创建简单的关系类型"""
        # 列表查看
        resp = self.api.list_relation_types(self.test_kn_id, "用户地点")
        self.assert_response_code(resp, 200)
        self.assert_has_length(json.loads(resp.text)["entries"], 1)
        self.assert_equal("用户地点", json.loads(resp.text)["entries"][0]["name"])

    @allure.story("查看关系类")
    @allure.title("查看关系类-关系类列表-精确筛选-双边")
    def test_relation_type_list_by_both_sides(self):
        """测试创建简单的关系类型"""
        # 列表查看
        resp = self.api.list_relation_types(
            self.test_kn_id,
            source_object_type_id="tag",
            target_object_type_id="tagclass",
        )
        self.assert_response_code(resp, 200)
        self.assert_has_length(json.loads(resp.text)["entries"], 1)
        self.assert_equal("标签归属分类", json.loads(resp.text)["entries"][0]["name"])
