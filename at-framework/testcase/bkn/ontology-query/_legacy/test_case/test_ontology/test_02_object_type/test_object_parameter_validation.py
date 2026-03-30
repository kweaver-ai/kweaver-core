# coding: utf-8
"""
对象类型参数校验测试用例
"""
import pytest
import allure
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.common.business_util import random_data_properties
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

validation_data = data_reader.read_yaml(
    "ontology/object_type/object_parameter_validation.yaml"
)


@allure.feature("知识网络对象类型管理")
class TestObjectTypeParameterValidation(BaseTest):
    """对象类型参数校验相关测试"""

    api = ObjectTypeApi()

    @allure.story("创建对象类-参数校验")
    @allure.title("{case_name}")
    @pytest.mark.parametrize("test_data", validation_data)
    def test_object_type_parameter_validation(self, test_data, setup_object_kn_id):
        """参数校验测试用例：验证必填参数、参数格式、参数类型等"""
        test_data = test_data[0]
        case_name = test_data.get("case_name")
        object_id = test_data.get("object_id")
        object_name = test_data.get("object_name")
        test_type = test_data.get("test_type")
        exclude_fields = test_data.get("exclude_fields", [])
        expected_status = test_data.get("expected_status", 201)
        should_exist_in_db = test_data.get("should_exist_in_db", True)
        note = test_data.get("note", "")
        data_properties_count = test_data.get("data_properties", 4)
        primary_keys = test_data.get("primary_keys")
        display_key = test_data.get("display_key")
        tags = test_data.get("tags")

        with allure.step(f"执行参数校验测试: {case_name}"):
            if note:
                allure.attach(note, name="测试说明")

            data_properties = random_data_properties(data_properties_count)

            resp = self.api.create_object_types(
                kn_id=setup_object_kn_id,
                object_id=object_id,
                name=object_name,
                data_source="",
                data_properties=data_properties,
                primary_keys=primary_keys,
                display_key=display_key,
                exclude_fields=exclude_fields,
            )

        with allure.step(f"验证HTTP响应状态码: 期望 {expected_status}"):
            self.assert_response_code(resp, expected_status)

        if test_type == "positive":
            with allure.step("验证响应数据结构"):
                self.assert_response_structure_list(resp, ["id"])

            if should_exist_in_db:
                with allure.step("验证数据库记录存在"):
                    self.assert_db_record_exists(
                        table=self.object_table,
                        condition="f_name = %s",
                        params=(object_name,),
                    )
        else:
            with allure.step("验证错误响应结构"):
                self.assert_response_structure(resp, ["error_code", "description"])

            if not should_exist_in_db:
                with allure.step("验证数据库记录不存在"):
                    self.assert_db_record_not_exists(
                        table=self.object_table,
                        condition="f_name = %s",
                        params=(object_name,),
                    )
