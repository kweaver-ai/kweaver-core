# coding: utf-8
"""
行动类型创建测试用例
"""
import pytest
import allure
import json
from src.api.ontology.kn_action_type_api import ActionTypeApi
from src.common.read_data import data_reader
from src.common.global_var import global_vars
from test_case.ontology_base_test import BaseTest

positive_data = data_reader.read_yaml("ontology/action_type/action_create_postive.yaml")
negative_data = data_reader.read_yaml(
    "ontology/action_type/action_create_negative.yaml"
)


@allure.feature("知识网络行动类型管理")
class TestActionTypeCreate(BaseTest):
    api = ActionTypeApi()
    object_type_id = "person"
    condition = {
        "value_from": "const",
        "operation": "==",
        "field": "p_personid",
        "object_type_id": "person",
        "value": 94,
    }

    @allure.story("创建行动类-正常创建")
    @allure.title("{action_data[case_name]}")
    @pytest.mark.parametrize("action_data", positive_data)
    def test_create_action_type_positive(self, action_data, get_tool_id_parameters):
        box_id = global_vars.get_var("box_id")
        tool_id = global_vars.get_var("tool_id")
        parameters = global_vars.get_var("parameters")
        action_source = {"box_id": box_id, "tool_id": tool_id, "type": "tool"}

        resp = self.api.create_action_type(
            kn_id=self.test_kn_id,
            name=action_data["name"],
            id=action_data["id"],
            action_type=action_data["action_type"],
            object_type_id=action_data.get("object_type_id", self.object_type_id),
            affect_object_type_id=action_data.get("affect_object_type_id", ""),
            condition=action_data.get("condition", self.condition),
            action_source=action_source,
            parameters=parameters,
            tags=action_data.get("tags", []),
            comment=action_data.get("comment", ""),
            concept_groups=action_data.get("concept_groups", []),
            exclude_fields=action_data.get("exclude_fields", None),
        )
        # 验证响应状态码和结构
        self.assert_response_code(resp, 201)
        action_id = json.loads(resp.text)[0]["id"]
        # 验证数据库记录存在
        self.assert_db_record_exists(
            table=self.action__table,
            condition="f_id = %s and f_kn_id = %s",
            params=(action_id, self.test_kn_id),
        )
        # 删除行动类
        if action_data.get("is_delete", True):
            resp = self.api.delete_action_type(self.test_kn_id, action_id)
            self.assert_response_code(resp, 204)

    @allure.story("创建行动类-异常场景")
    @allure.title("{action_data[case_name]}")
    @pytest.mark.parametrize("action_data", negative_data)
    def test_create_action_type_negative(self, action_data, get_tool_id_parameters):
        box_id = global_vars.get_var("box_id")
        tool_id = global_vars.get_var("tool_id")
        parameters = global_vars.get_var("parameters")
        action_source = {"box_id": box_id, "tool_id": tool_id, "type": "tool"}
        # 重复ID、name验证先新建一个正常的
        if action_data["note"] == "重复ID/name":
            resp = self.api.create_action_type(
                kn_id=self.test_kn_id,
                name=action_data["name"],
                id=action_data["id"],
                action_type=action_data["action_type"],
                object_type_id=action_data.get("object_type_id", self.object_type_id),
                affect_object_type_id=action_data.get("affect_object_type_id", ""),
                condition=action_data.get("condition", self.condition),
                action_source=action_source,
                parameters=parameters,
                tags=action_data.get("tags", []),
                comment=action_data.get("comment", ""),
                concept_groups=action_data.get("concept_groups", []),
                exclude_fields=action_data.get("exclude_fields", None),
            )
            self.assert_response_code(resp, 201)
            action_id = json.loads(resp.text)[0]["id"]

        resp = self.api.create_action_type(
            kn_id=self.test_kn_id,
            name=action_data["name"],
            id=action_data["id"],
            action_type=action_data["action_type"],
            object_type_id=action_data.get("object_type_id", self.object_type_id),
            affect_object_type_id=action_data.get("affect_object_type_id", ""),
            condition=action_data.get("condition", self.condition),
            action_source=action_source,
            parameters=parameters,
            tags=action_data.get("tags", []),
            comment=action_data.get("comment", ""),
            concept_groups=action_data.get("concept_groups", []),
            exclude_fields=action_data.get("exclude_fields", None),
        )
        # 验证响应状态码和结构
        self.assert_response_code(resp, action_data.get("expected_status"))
        # 验证数据库记录不存在
        if action_data["note"] != "重复ID/name":
            self.assert_db_record_not_exists(
                self.action__table,
                condition="f_id = %s AND f_kn_id = %s",
                params=(action_data["id"], self.test_kn_id),
            )

        if action_data["note"] == "重复ID/name":
            # 删除行动类
            resp = self.api.delete_action_type(self.test_kn_id, action_id)
            self.assert_response_code(resp, 204)
