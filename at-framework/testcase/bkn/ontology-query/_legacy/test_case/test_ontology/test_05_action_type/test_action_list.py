# coding: utf-8
"""
行动类型创建测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_action_type_api import ActionTypeApi
from src.common.global_var import global_vars
from test_case.ontology_base_test import BaseTest


@allure.feature("知识网络行动类型管理")
class TestActionTypeList(BaseTest):
    api = ActionTypeApi()
    box_id = global_vars.get_var("box_id")
    tool_id = global_vars.get_var("tool_id")
    parameters = global_vars.get_var("parameters")
    action_source = {"box_id": box_id, "tool_id": tool_id, "type": "tool"}

    @allure.story("获取行动类列表")
    @allure.title("获取行动类列表")
    def test_action_list(self):
        # 查询
        resp = self.api.list_action_types(self.test_kn_id)
        self.assert_response_code(resp, 200)
        # 返回数据entries、total_count
        self.assert_json_structure(json.loads(resp.text), ["entries", "total_count"])

        self.assert_json_structure(json.loads(resp.text)["entries"][0], ["id", "name"])

    @allure.story("获取行动类列表")
    @allure.title("获取行动类列表-筛选")
    def test_action_list(self, get_tool_id_parameters):
        self.api.create_action_type(
            kn_id=self.test_kn_id,
            name="测试筛选",
            id="test_list",
            action_type="add",
            object_type_id="person",
            affect_object_type_id="person",
            condition={
                "value_from": "const",
                "operation": "==",
                "field": "p_personid",
                "object_type_id": "person",
                "value": 94,
            },
            action_source=self.action_source,
            parameters=self.parameters,
        )
        # 查询
        resp = self.api.list_action_types(self.test_kn_id, "测试筛选")
        self.assert_response_code(resp, 200)
        self.assert_has_length(json.loads(resp.text)["entries"], 1)
        self.assert_equal("测试筛选", json.loads(resp.text)["entries"][0]["name"])
        # 删除
        self.api.delete_action_type(self.test_kn_id, "test_list")
