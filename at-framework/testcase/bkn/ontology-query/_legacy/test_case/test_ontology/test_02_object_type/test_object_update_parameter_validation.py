# coding: utf-8
"""
对象类型更新参数校验测试用例
"""
import pytest
import allure
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.common.business_util import random_data_properties
from src.common.read_data import data_reader
from test_case.ontology_base_test import BaseTest

validation_data = data_reader.read_yaml(
    "ontology/object_type/object_update_parameter_validation.yaml"
)


@allure.feature("知识网络对象类型管理")
class TestObjectTypeUpdateParameterValidation(BaseTest):
    """对象类型更新参数校验相关测试"""

    api = ObjectTypeApi()

    @allure.story("更新对象类-参数校验")
    @allure.title("{case_name}")
    @pytest.mark.parametrize("test_data", validation_data)
    def test_object_update_parameter_validation(self, test_data, setup_object_kn_id):
        """参数校验测试用例：验证更新接口的必填参数、参数格式、参数类型等"""
        test_data = test_data[0]
        case_name = test_data.get("case_name")
        object_c_name = test_data.get("object_c_name")
        object_u_name = test_data.get("object_u_name")
        object_id = test_data.get("object_id")
        test_type = test_data.get("test_type")
        exclude_fields = test_data.get("exclude_fields", [])
        expected_status = test_data.get("expected_status", 204)
        note = test_data.get("note", "")
        data_properties_count = test_data.get("data_properties", 4)
        primary_keys = test_data.get("primary_keys")
        display_key = test_data.get("display_key")

        with allure.step("步骤1: 创建对象类"):
            create_resp = self.api.create_object_types(
                kn_id=setup_object_kn_id,
                object_id=object_id,
                name=object_c_name,
                data_source="",
                data_properties=random_data_properties(4),
            )
            self.assert_response_code(create_resp, 201)

        with allure.step(f"步骤2: 执行参数校验测试 - {case_name}"):
            if note:
                allure.attach(note, name="测试说明")

            data_properties = random_data_properties(data_properties_count)

            update_resp = self.api.update_object_types(
                kn_id=setup_object_kn_id,
                object_id=object_id,
                name=object_u_name,
                data_source="",
                data_properties=data_properties,
                primary_keys=primary_keys,
                display_key=display_key,
                exclude_fields=exclude_fields,
            )

        with allure.step(f"步骤3: 验证HTTP响应状态码: 期望 {expected_status}"):
            self.assert_response_code(update_resp, expected_status)

        if test_type == "positive":
            with allure.step("步骤4: 验证数据库记录已更新"):
                self.assert_db_record_exists(
                    table=self.object_table,
                    condition="f_name = %s",
                    params=(object_u_name,),
                )

            with allure.step("步骤5: 清理测试数据"):
                self.api.delete_objects(kn_id=setup_object_kn_id, object_id=object_id)
        else:
            with allure.step("步骤4: 验证数据库记录未被修改"):
                self.assert_db_record_exists(
                    table=self.object_table,
                    condition="f_name = %s",
                    params=(object_c_name,),
                )

            with allure.step("步骤5: 清理测试数据"):
                self.api.delete_objects(kn_id=setup_object_kn_id, object_id=object_id)
