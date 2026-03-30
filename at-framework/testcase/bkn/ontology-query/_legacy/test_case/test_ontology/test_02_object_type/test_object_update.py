import pytest
import allure
import json
from src.api.ontology.kn_object_type_api import ObjectTypeApi
from src.common.read_data import data_reader
from src.common.business_util import random_data_properties
from test_case.ontology_base_test import BaseTest

# 读取测试数据
positive_data = data_reader.read_yaml("ontology/object_type/object_update_postive.yaml")
negative_data = data_reader.read_yaml(
    "ontology/object_type/object_update_negative.yaml"
)
negative_by_properties_data = data_reader.read_yaml(
    "ontology/object_type/object_update_negative_by_properties.yaml"
)


@allure.feature("知识网络对象类型管理")
class TestObjectTypeUpdate(BaseTest):
    api = ObjectTypeApi()

    @allure.story("更新业务知识网络对象类-正常场景")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", positive_data)
    def test_update_object_success(self, object_data, test_kn_id_use):
        # 创建
        resp = self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_c_name"],
            data_source="",
            data_properties=random_data_properties(
                object_data.get("data_properties", 4)
            ),
        )
        self.assert_response_code(resp, 201)
        # 修改
        resp = self.api.update_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_u_name"],
            data_source="",
            tags=object_data.get("tags", None),
            comment=object_data.get("comment", None),
            data_properties=random_data_properties(
                object_data["data_properties"], static=object_data.get("static", False)
            ),
            primary_keys=object_data.get("primary_keys", None),
            incremental_key=object_data.get("incremental_key", None),
            display_key=object_data.get("display_key", None),
            exclude_fields=object_data.get("exclude_fields", None),
        )
        # 验证响应
        self.assert_response_code(resp, 204)
        # 验证数据库记录更新
        self.assert_db_record_exists(
            table=self.object_table,
            condition="f_name = %s",
            params=object_data["object_u_name"],
        )

    @allure.story("更新业务知识网络对象类-参数校验异常场景")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", negative_data)
    def test_update_object_type_negative(self, object_data, test_kn_id_use):
        # 创建
        resp = self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_c_name"],
            data_source="",
            data_properties=random_data_properties(4),
        )
        self.assert_response_code(resp, 201)

        # 修改
        resp = self.api.update_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_u_name"],
            data_source="",
            tags=object_data.get("tags", None),
            comment=object_data.get("comment", None),
            data_properties=random_data_properties(
                object_data.get("data_properties", 4),
                static=object_data.get("static", False),
            ),
            primary_keys=object_data.get("primary_keys", None),
            incremental_key=object_data.get("incremental_key", None),
            display_key=object_data.get("display_key", None),
            exclude_fields=object_data.get("exclude_fields", None),
        )

        # 验证响应
        self.assert_response_code(resp, object_data["expected_status"])

        # 如果是负面测试，验证数据库记录未被修改
        if object_data["expected_status"] != 204:
            self.assert_db_record_exists(
                table=self.object_table,
                condition="f_name = %s",
                params=object_data["object_c_name"],
            )
        else:
            self.assert_db_record_exists(
                table=self.object_table,
                condition="f_name = %s",
                params=object_data["object_u_name"],
            )

        # 清理测试数据
        self.api.delete_objects(
            kn_id=test_kn_id_use, object_id=object_data["object_id"]
        )

    @allure.story("更新业务知识网络对象类-数据属性校验异常场景")
    @allure.title("{object_data[case_name]}")
    @pytest.mark.parametrize("object_data", negative_by_properties_data)
    def test_update_object_type_negative_by_properties(
        self, object_data, test_kn_id_use
    ):
        # 创建
        resp = self.api.create_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_c_name"],
            data_source="",
            data_properties=random_data_properties(4),
        )
        self.assert_response_code(resp, 201)

        # 执行更新
        resp = self.api.update_object_types(
            kn_id=test_kn_id_use,
            object_id=object_data["object_id"],
            name=object_data["object_u_name"],
            data_source="",
            data_properties=[object_data["data_properties"]],
            primary_keys=["test"],
            display_key="test",
        )

        # 验证响应
        self.assert_response_code(resp, object_data["expected_status"])

        # 验证数据库记录未被修改
        self.assert_db_record_exists(
            table=self.object_table,
            condition="f_name = %s",
            params=object_data["object_c_name"],
        )

        # 清理测试数据
        self.api.delete_objects(
            kn_id=test_kn_id_use, object_id=object_data["object_id"]
        )
